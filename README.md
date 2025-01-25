# AngelOne Market Data Processor

Real-time market data processing system that captures data from AngelOne's WebSocket feed and stores it in ClickHouse database with concurrent processing and monitoring capabilities.

## Features

### Core Features
- Real-time market data capture via WebSocket
- Multi-exchange support (NSE, BSE, MCX)
- Configurable token subscriptions via JSON
- High-performance ClickHouse storage
- Circuit breaker pattern for failure handling
- Binary data parsing for market ticks

### Performance Features
- Concurrent processing with configurable worker pools
- Batch processing with configurable sizes
- Buffer management for high throughput
- Rate limiting and backoff strategies

### Monitoring & Reliability
- Prometheus metrics integration
- Health check endpoints
- Structured logging with rotation
- Automatic reconnection handling
- Error recovery middleware

## Prerequisites

- Go 1.21 or higher
- Docker
- AngelOne Trading Account

## Installation & Setup

1. Clone and setup:
```bash
git clone https://github.com/crypt0inf0/angelone_clickhouse.git
cd angelone_clickhouse
go mod tidy
```

2. Configure ClickHouse:
```bash
docker run -d \
    --name clickhouse \
    -p 9000:9000 \
    -v clickhouse_data:/var/lib/clickhouse \
    clickhouse/clickhouse-server
```

3. Set up ClickHouse using Docker:
```bash
docker run -d \
    --name clickhouse \
    --network=host \
    -v /tmp/clickhouse:/var/lib/clickhouse \
    clickhouse/clickhouse-server
```

4. Configure environment variables:
# Edit .env with your AngelOne credentials and ClickHouse configuration
```bash
cp .env.example .env
```

## Configuration

### Token Configuration
Create `config/tokens.json`:
```json
[
    {
        "symbol": "RELIANCE",
        "token": "2885",
        "exchange": "NSE_CM"
    },
    {
        "symbol": "NIFTY25JAN23200PE",
        "token": "43607",
        "exchange": "NSE_FO"
    }
]
```

### Exchange Types
```go
NSE_CM = 1  // NSE Cash Market
NSE_FO = 2  // NSE F&O
BSE_CM = 3  // BSE Cash Market
BSE_FO = 4  // BSE F&O
MCX_FO = 5  // MCX F&O
NCX_FO = 7  // NCX F&O
CDE_FO = 13 // Currency Derivatives
```

### Environment Variables
```properties
# AngelOne credentials
ANGEL_CLIENT_ID=YOUR_CLIENT_ID
ANGEL_CLIENT_PIN=YOUR_PIN
ANGEL_TOTP_CODE=YOUR_TOTP_CODE
ANGEL_API_KEY=YOUR_API_KEY
ANGEL_CLIENT_LOCAL_IP=YOUR_LOCAL_IP
ANGEL_CLIENT_PUBLIC_IP=YOUR_PUBLIC_IP
ANGEL_MAC_ADDRESS=YOUR_MAC_ADDRESS
ANGEL_STATE_VARIABLE=YOUR_STATE_VARIABLE

# ClickHouse configuration
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=

# Application Settings
BATCH_SIZE=1000           # Number of ticks per batch
FLUSH_INTERVAL=5          # Seconds between forced flushes
MAX_QUEUE_SIZE=10000      # Maximum number of pending ticks
NUM_WORKERS=5             # Number of concurrent workers
```

## Usage

1. Start the application:
```bash
go run main.go
```

2. Verify data storage:
```bash
docker exec -it clickhouse clickhouse-client
```

```sql
SELECT * FROM angelone_market_data WHERE token = '2885' LIMIT 5;
```

## Project Structure

```
angelone_clickhouse/
├── angel/         # AngelOne specific types and utils
├── db/           # ClickHouse database operations
├── models/       # Data models
├── ws/           # WebSocket client implementation
├── main.go       # Application entry point
└── .env          # Configuration file
```

## Troubleshooting

### ClickHouse Connection

1. Verify ClickHouse is running:
```bash
docker ps | grep clickhouse
```

2. Check ClickHouse logs:
```bash
docker logs clickhouse
```

3. Test port accessibility:
```bash
nc -zv localhost 9000
```

### WebSocket Connection

1. Verify your AngelOne credentials
2. Check WebSocket connection logs
3. Ensure your API key is active

## Performance Optimization

The system uses batch processing with configurable parameters:

```properties
BATCH_SIZE=1000           # Number of ticks per batch
FLUSH_INTERVAL=5          # Seconds between forced flushes
MAX_QUEUE_SIZE=10000      # Maximum number of pending ticks
NUM_WORKERS=5             # Number of concurrent workers
```

Adjust these values based on your requirements:
- Higher batchSize = Better throughput
- Lower flushInterval = Lower latency

## Performance Tuning

### ClickHouse Settings
Optimize your ClickHouse configuration for high-frequency trading data:

```sql
max_memory_usage = 20000000000
max_memory_usage_for_user = 20000000000
max_bytes_before_external_group_by = 20000000000
max_threads = 8
max_insert_threads = 8
```

### Batch Processing Configuration

Configure batch sizes and intervals in your `.env`:

```properties
BATCH_SIZE=1000           # Number of ticks per batch
FLUSH_INTERVAL=5          # Seconds between forced flushes
MAX_QUEUE_SIZE=10000      # Maximum number of pending ticks
NUM_WORKERS=5             # Number of concurrent workers
```

## Data Queries

### Basic Queries

#### Get latest prices:
```sql
SELECT
    token,
    last_traded_price,
    timestamp
FROM angelone_market_data
WHERE token IN ('2885', '1594')
ORDER BY timestamp DESC
LIMIT 10;
```

#### Daily OHLCV:
```sql
SELECT
    token,
    toDate(timestamp) as date,
    min(low_price) as day_low,
    max(high_price) as day_high,
    first_value(open_price) as open,
    last_value(close_price) as close,
    sum(volume) as volume
FROM angelone_market_data
WHERE timestamp >= today() - INTERVAL 7 DAY
GROUP BY token, date
ORDER BY date DESC;
```

#### Volume Profile:
```sql
SELECT
    token,
    round(last_traded_price, 2) as price_level,
    count(*) as tick_count,
    sum(volume) as total_volume
FROM angelone_market_data
WHERE timestamp >= now() - INTERVAL 1 DAY
GROUP BY token, price_level
ORDER BY price_level DESC;
```

## Monitoring

### Available Metrics
- `market_data_processed_total`: Total processed ticks
- `market_data_errors_total`: Total error count
- `market_data_last_processed_timestamp`: Last tick timestamp
- `market_data_uptime_seconds`: Application uptime

### Health Check
```bash
curl http://localhost:8080/health
```

### Metrics Endpoint
```bash
curl http://localhost:8080/metrics
```

## Error Handling

### Common Issues
1. WebSocket Disconnection
   - Automatic reconnection with exponential backoff
   - State recovery and data gap detection

2. Database Connection Issues
   - Connection pooling with automatic recovery
   - Query timeout handling
   - Batch insert retry logic

3. Rate Limiting
   - Smart backoff strategy
   - Queue management
   - Throughput monitoring

## Production Checklist

### Security
- [ ] SSL/TLS configuration
- [ ] API authentication
- [ ] Network security groups
- [ ] Credential rotation

### Monitoring
- [ ] Prometheus metrics
- [ ] Grafana dashboards
- [ ] Alert rules
- [ ] Log aggregation

### High Availability
- [ ] Database replication
- [ ] Service redundancy
- [ ] Load balancing
- [ ] Failover procedures

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- AngelOne SmartAPI Documentation
- ClickHouse Documentation
- Market data processing best practices from [marketcalls.in](https://www.marketcalls.in/python/storing-websocket-stock-market-data-in-clickhouse-using-python-a-comprehensive-guide.html)
