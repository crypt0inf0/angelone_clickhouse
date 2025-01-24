package db

import (
    "context"
    "fmt"
    "github.com/ClickHouse/clickhouse-go/v2"
    "github.com/ClickHouse/clickhouse-go/v2/lib/driver"
    "broker_clickhouse/models"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS market_ticks (
    timestamp DateTime,
    symbol String,
    last_price Float64,
    volume Int64,
    bid_price Float64,
    ask_price Float64,
    open_price Float64,
    high_price Float64,
    low_price Float64,
    close_price Float64
) ENGINE = MergeTree()
ORDER BY (timestamp, symbol)
`

type ClickHouseDB struct {
    conn driver.Conn
}

func NewClickHouseDB(host string, port int, database string, username string, password string) (*ClickHouseDB, error) {
    conn, err := clickhouse.Open(&clickhouse.Options{
        Addr: []string{fmt.Sprintf("%s:%d", host, port)},
        Auth: clickhouse.Auth{
            Database: username,  // Use username as database name
            Username: username,
            Password: password,
        },
        Protocol: clickhouse.Native,  // Explicitly set to use native protocol
        Debug:    true,              // Enable debug logging
        Settings: clickhouse.Settings{
            "max_execution_time": 60,
        },
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to connect to ClickHouse: %v", err)
    }

    db := &ClickHouseDB{conn: conn}
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
    if (err != nil) {
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
