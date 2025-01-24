package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"angelone_clickhouse/config"
	"angelone_clickhouse/models"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS angelone_market_data (
    token String,
    timestamp DateTime64(3),
    last_traded_price Float64,
    open_price Float64,
    high_price Float64,
    low_price Float64,
    close_price Float64,
    volume Float64
) ENGINE = MergeTree()
ORDER BY timestamp
`

type ClickHouseDB struct {
	conn   driver.Conn
	config *config.Config
}

func NewClickHouseDB(cfg *config.Config) (*ClickHouseDB, error) {
	opts := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouse.Host, cfg.ClickHouse.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouse.Database,
			Username: cfg.ClickHouse.User,
			Password: cfg.ClickHouse.Password,
		},
		Protocol: clickhouse.Native,
		Debug:    cfg.ClickHouse.Debug,
		Settings: clickhouse.Settings{
			"max_execution_time": cfg.ClickHouse.QueryTimeout.Seconds(),
		},
		DialTimeout:          10 * time.Second,
		MaxOpenConns:         cfg.ClickHouse.MaxOpenConns,
		MaxIdleConns:         cfg.ClickHouse.MaxIdleConns,
		ConnMaxLifetime:      cfg.ClickHouse.ConnMaxLifetime,
		Compression:          &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10 << 20, // 10MB
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %v", err)
	}

	db := &ClickHouseDB{
		conn:   conn,
		config: cfg,
	}

	if err := db.createTable(); err != nil {
		return nil, err
	}

	return db, nil
}

func (db *ClickHouseDB) createTable() error {
	return db.conn.Exec(context.Background(), createTableSQL)
}

func (db *ClickHouseDB) InsertTicks(ctx context.Context, ticks []models.MarketTick) error {
	batch, err := db.conn.PrepareBatch(ctx, "INSERT INTO market_ticks")
	if err != nil {
		return err
	}

	for _, tick := range ticks {
		err := batch.AppendStruct(&tick)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

// Add method for single tick insertion
func (db *ClickHouseDB) InsertTick(ctx context.Context, tick models.MarketTick) error {
	ctx, cancel := context.WithTimeout(ctx, db.config.ClickHouse.QueryTimeout)
	defer cancel()

	query := `
        INSERT INTO angelone_market_data (
            token, timestamp, last_traded_price, 
            open_price, high_price, low_price, 
            close_price, volume
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `

	return db.conn.Exec(ctx, query,
		tick.Symbol,
		tick.Timestamp,
		tick.LastPrice,
		tick.OpenPrice,
		tick.HighPrice,
		tick.LowPrice,
		tick.ClosePrice,
		tick.Volume,
	)
}

// Add method to verify data storage
func (db *ClickHouseDB) VerifyLastInserted(ctx context.Context, symbol string) (*models.MarketTick, error) {
	query := `
        SELECT 
            token, timestamp, last_traded_price, 
            open_price, high_price, low_price, 
            close_price, volume
        FROM angelone_market_data 
        WHERE token = ?
        ORDER BY timestamp DESC 
        LIMIT 1
    `

	var tick models.MarketTick
	row := db.conn.QueryRow(ctx, query, symbol)
	err := row.Scan(
		&tick.Symbol,
		&tick.Timestamp,
		&tick.LastPrice,
		&tick.OpenPrice,
		&tick.HighPrice,
		&tick.LowPrice,
		&tick.ClosePrice,
		&tick.Volume,
	)

	if err != nil {
		return nil, fmt.Errorf("error verifying data: %v", err)
	}

	return &tick, nil
}

// Add method to get statistics
func (db *ClickHouseDB) GetDailyStats(ctx context.Context, symbol string) error {
	query := `
        SELECT 
            token,
            toDate(timestamp) as date,
            min(low_price) as day_low,
            max(high_price) as day_high,
            sum(volume) as total_volume,
            count(*) as tick_count
        FROM angelone_market_data 
        WHERE token = ? 
        GROUP BY token, date
        ORDER BY date DESC
        LIMIT 1
    `

	var (
		token           string
		date            time.Time
		dayLow, dayHigh float64
		volume, count   int64
	)

	row := db.conn.QueryRow(ctx, query, symbol)
	if err := row.Scan(&token, &date, &dayLow, &dayHigh, &volume, &count); err != nil {
		return fmt.Errorf("error getting stats: %v", err)
	}

	log.Printf("Daily Stats [%s] %s: Low: %.2f | High: %.2f | Volume: %d | Ticks: %d",
		date.Format("2006-01-02"), token, dayLow, dayHigh, volume, count)

	return nil
}

// Add method for bulk verification
func (db *ClickHouseDB) VerifyMultipleTokens(ctx context.Context, tokens []string) error {
	query := `
        SELECT 
            token,
            max(timestamp) as last_update,
            count(*) as tick_count
        FROM angelone_market_data 
        WHERE token IN (?)
        GROUP BY token
    `

	rows, err := db.conn.Query(ctx, query, tokens)
	if err != nil {
		return fmt.Errorf("error querying multiple tokens: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			token      string
			lastUpdate time.Time
			tickCount  int64
		)
		if err := rows.Scan(&token, &lastUpdate, &tickCount); err != nil {
			return err
		}
		log.Printf("Token %s: Last update @ %s, Total ticks: %d",
			token, lastUpdate.Format("15:04:05"), tickCount)
	}

	return rows.Err()
}
