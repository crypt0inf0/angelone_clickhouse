package main

import (
	"angelone_clickhouse/angel"
	"angelone_clickhouse/config"
	"angelone_clickhouse/db"
	"angelone_clickhouse/metrics"
	"angelone_clickhouse/models"
	"angelone_clickhouse/parser"
	"angelone_clickhouse/utils"
	"angelone_clickhouse/ws"
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/joho/godotenv"
)

type MarketData struct {
	Token           string  `json:"token"`
	LastTradedPrice float64 `json:"last_traded_price"`
	OpenPrice       float64 `json:"open_price_of_the_day"`
	HighPrice       float64 `json:"high_price_of_the_day"`
	LowPrice        float64 `json:"low_price_of_the_day"`
	ClosedPrice     float64 `json:"closed_price"`
	Volume          float64 `json:"volume_trade_for_the_day"`
}

// Add channel type for data processing
type MarketDataChannel struct {
	data       MarketData
	clickhouse *db.ClickHouseDB
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize metrics
	metricsInstance := metrics.NewMetrics(cfg)

	// Initialize DB
	db, err := db.NewClickHouseDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize worker pool
	jobs := make(chan MarketDataChannel, cfg.App.BufferSize)
	for w := 1; w <= cfg.App.NumWorkers; w++ {
		go processDataWorker(w, jobs, db, metricsInstance)
	}

	// Initialize logger
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Create reconnection context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		operation := func() error {
			return runWebSocket(ctx, cfg, metricsInstance)
		}

		retry := utils.NewExponentialBackoff()
		err := backoff.RetryNotify(operation, retry,
			func(err error, duration time.Duration) {
				log.Printf("Error: %v, retrying in %v...", err, duration)
			})
		if err != nil {
			log.Printf("Max retries reached: %v", err)
		}
	}()

	// Use request logger middleware
	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/health", healthCheck)
	metricsMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metricsHandler(w, r, metricsInstance)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: utils.RequestLogger(metricsMux),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Error(err, "Metrics server error")
		}
	}()

	wg.Wait()
}

func processAndStoreData(data MarketData) {
	// Format output with correct decimal places
	now := time.Now()
	minute := now.Format("15:04")
	log.Printf("[%s] %s - minute: %s | open: %.2f | close: %.2f | high: %.2f | low: %.2f | volume: %.3f",
		now.Format("2006-01-02 15:04:05.000"),
		data.Token,
		minute,
		data.OpenPrice,
		data.LastTradedPrice,
		data.HighPrice,
		data.LowPrice,
		data.Volume)
}

// Update processDataWorker signature and implementation
func processDataWorker(id int, jobs <-chan MarketDataChannel, db *db.ClickHouseDB, metrics *metrics.Metrics) {
	for job := range jobs {
		// Create a MarketTick for ClickHouse storage
		tick := models.MarketTick{
			Timestamp:  time.Now(),
			Symbol:     job.data.Token,
			LastPrice:  job.data.LastTradedPrice,
			Volume:     int64(job.data.Volume),
			OpenPrice:  job.data.OpenPrice,
			HighPrice:  job.data.HighPrice,
			LowPrice:   job.data.LowPrice,
			ClosePrice: job.data.ClosedPrice,
		}

		// Store tick with retry mechanism
		if err := job.clickhouse.InsertTick(context.Background(), tick); err != nil {
			utils.Error(err, "Error storing tick",
				"worker_id", id,
				"token", job.data.Token,
			)
			metrics.IncrementErrors()
			continue
		}

		utils.Logger.Infow("Tick stored",
			"worker_id", id,
			"token", job.data.Token,
			"price", job.data.LastTradedPrice,
		)
		metrics.IncrementProcessed()
	}
}

// Update runWebSocket to use configuration consistently
func runWebSocket(ctx context.Context, cfg *config.Config, metrics *metrics.Metrics) error {
	// Authenticate with AngelOne
	authToken, feedToken, err := angel.Authenticate()
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Set tokens in environment
	os.Setenv("ANGEL_AUTH_TOKEN", authToken)
	os.Setenv("ANGEL_FEED_TOKEN", feedToken)

	// Initialize ClickHouse connection using config
	clickhouse, err := db.NewClickHouseDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}

	// Initialize WebSocket client with AngelOne headers
	headers := map[string]string{
		"Authorization": "Bearer " + authToken,
		"X-Client-Code": os.Getenv("ANGEL_CLIENT_ID"),
		"X-Api-Key":     os.Getenv("ANGEL_API_KEY"),
		"X-Feed-Token":  feedToken,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	wsClient := ws.NewWebSocketClient("wss://smartapisocket.angelone.in/smart-stream", headers)

	// Buffer for batch inserts
	tickBuffer := make([]models.MarketTick, 0, 1000)

	// Add configuration parameters
	const (
		batchSize     = 1000
		flushInterval = 5 * time.Second
	)

	// Create a ticker for regular flushes
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			if len(tickBuffer) > 0 {
				if err := clickhouse.InsertTicks(context.Background(), tickBuffer); err != nil {
					log.Printf("Error inserting ticks: %v", err)
				}
				tickBuffer = tickBuffer[:0]
			}
		}
	}()

	// Add periodic verification
	go func() {
		verifyTicker := time.NewTicker(1 * time.Minute)
		defer verifyTicker.Stop()

		for range verifyTicker.C {
			// Verify last stored data
			tick, err := clickhouse.VerifyLastInserted(context.Background(), "2885")
			if err != nil {
				log.Printf("Verification error: %v", err)
				continue
			}

			// Print verification result
			log.Printf("Last stored data verified: %s @ %s: %.2f",
				tick.Symbol,
				tick.Timestamp.Format("15:04:05"),
				tick.LastPrice)

			// Get daily statistics
			if err := clickhouse.GetDailyStats(context.Background(), "2885"); err != nil {
				log.Printf("Stats error: %v", err)
			}
		}
	}()

	// Create a buffered channel for market data processing
	jobs := make(chan MarketDataChannel, cfg.App.BufferSize)

	// Start worker pool
	for w := 1; w <= cfg.App.NumWorkers; w++ {
		go processDataWorker(w, jobs, clickhouse, metrics)
	}

	// Multiple tokens to subscribe
	tokens := []string{
		"2885",  // RELIANCE
		"1594",  // INFY
		"11536", // TCS
		"3045",  // SBIN
		"3787",  // HDFCBANK
	}

	// Subscribe to market data for multiple tokens
	subscribeReq := angel.SubscribeRequest{
		CorrelationID: "ws_test",
		Action:        1,
		Params: angel.SubscriptionParams{
			Mode: 2,
			TokenList: []angel.TokenSubscription{
				{
					ExchangeType: 1,
					Tokens:       tokens,
				},
			},
		},
	}

	// Update message handling for concurrent processing
	wsClient.OnMessage = func(message []byte) {
		data, err := parser.ParseBinaryData(message)
		if err != nil {
			log.Printf("Error parsing binary data: %v", err)
			return
		}

		adjustedData := MarketData{
			Token:           data.Token,
			LastTradedPrice: data.GetLastTradedPrice(),
			OpenPrice:       data.GetOpenPrice(),
			HighPrice:       data.GetHighPrice(),
			LowPrice:        data.GetLowPrice(),
			ClosedPrice:     data.GetClosedPrice(),
			Volume:          float64(data.VolumeTrade),
		}

		// Send to worker pool through channel
		select {
		case jobs <- MarketDataChannel{data: adjustedData, clickhouse: clickhouse}:
			// Successfully queued
		default:
			log.Printf("Warning: Channel buffer full, dropping tick for %s", data.Token)
		}
	}

	// Connect to WebSocket
	if err := wsClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer wsClient.Close()

	// Send subscription request
	if err := wsClient.SendJSON(subscribeReq); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// Start listening for messages
	wsClient.Listen()

	return nil
}

// Add health check handler
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// Update metricsHandler to use metrics instance
func metricsHandler(w http.ResponseWriter, r *http.Request, metrics *metrics.Metrics) {
	w.Header().Set("Content-Type", "text/plain")
	processed, errors, lastProc, uptime := metrics.GetStats()
	w.Write([]byte(
		"market_data_processed_total " + strconv.FormatUint(processed, 10) + "\n" +
			"market_data_errors_total " + strconv.FormatUint(errors, 10) + "\n" +
			"market_data_last_processed_timestamp " + strconv.FormatInt(lastProc.Unix(), 10) + "\n" +
			"market_data_uptime_seconds " + strconv.FormatFloat(uptime.Seconds(), 'f', 1, 64) + "\n",
	))
}
