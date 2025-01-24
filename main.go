package main

import (
	"broker_clickhouse/angel"
	"broker_clickhouse/db"
	"broker_clickhouse/models"
	"broker_clickhouse/ws"
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Convert port string to int
	port, err := strconv.Atoi(os.Getenv("CLICKHOUSE_PORT"))
	if err != nil {
		log.Fatalf("Invalid CLICKHOUSE_PORT: %v", err)
	}

	// Initialize ClickHouse connection with converted port
	clickhouse, err := db.NewClickHouseDB(
		os.Getenv("CLICKHOUSE_HOST"),
		port,
		os.Getenv("CLICKHOUSE_USER"),
		os.Getenv("CLICKHOUSE_PASSWORD"),
		"",
	)
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}

	// Initialize WebSocket client with AngelOne headers
	headers := map[string]string{
		"Authorization": "Bearer " + os.Getenv("ANGEL_AUTH_TOKEN"),
		"x-api-key":     os.Getenv("ANGEL_API_KEY"),
		"x-client-code": os.Getenv("ANGEL_CLIENT_ID"),
		"x-feed-token":  os.Getenv("ANGEL_FEED_TOKEN"),
	}

	wsClient := ws.NewWebSocketClient("wss://smartapisocket.angelone.in/smart-stream", headers)

	// Buffer for batch inserts
	tickBuffer := make([]models.MarketTick, 0, 1000)
	lastFlush := time.Now()

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
				lastFlush = time.Now()
			}
		}
	}()

	// Handle incoming messages
	wsClient.OnMessage = func(message []byte) {
		var tick models.MarketTick
		if err := json.Unmarshal(message, &tick); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		tickBuffer = append(tickBuffer, tick)

		// Flush buffer if it's full or if enough time has passed
		if len(tickBuffer) >= 1000 || time.Since(lastFlush) > 5*time.Second {
			if err := clickhouse.InsertTicks(context.Background(), tickBuffer); err != nil {
				log.Printf("Error inserting ticks: %v", err)
			}
			tickBuffer = tickBuffer[:0]
			lastFlush = time.Now()
		}
	}

	// Connect to WebSocket
	if err := wsClient.Connect(); err != nil {
		log.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer wsClient.Close()

	// Subscribe to market data
	subscribeReq := angel.SubscribeRequest{
		CorrelationID: "ws_test",
		Action:        1,
		Params: angel.SubscriptionParams{
			Mode: 2,
			TokenList: []angel.TokenSubscription{
				{
					ExchangeType: 1,
					Tokens:       []string{"2885"}, // RELIANCE token
				},
			},
		},
	}

	// Send subscription request
	if err := wsClient.SendJSON(subscribeReq); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// Start listening for messages
	wsClient.Listen()
}
