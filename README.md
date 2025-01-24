# AngelOne Market Data Processor

High-performance market data processing system that captures real-time stock market data from AngelOne WebSocket feed and stores it in ClickHouse database with concurrent processing and monitoring capabilities.

## Features

### Core Functionality
- Real-time market data capture via WebSocket
- Concurrent processing with worker pools
- High-performance ClickHouse storage
- Automatic reconnection and error recovery
- Binary data parsing for market ticks

### Performance Features
- Configurable worker pool size
- Batch processing with configurable sizes
- Buffer management for high throughput
- Rate limiting and backoff strategies

### Monitoring & Metrics
- HTTP endpoint for health checks (/health)
- Real-time metrics endpoint (/metrics)
- Daily statistics for each symbol
- Structured logging with rotation
- Performance monitoring dashboard

## Prerequisites

- Go 1.21 or higher
- Docker
- AngelOne Trading Account

## Installation

1. Clone the repository:
```bash
git clone https://github.com/crypt0inf0/angelone_clickhouse.git
cd angelone_clickhouse
```

2. Install dependencies:
```bash
go mod tidy
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
```bash
cp .env.example .env
# Edit .env with your AngelOne credentials and ClickHouse configuration
```

## Configuration

Update the `.env` file with your credentials:

```properties
# AngelOne credentials
ANGEL_CLIENT_ID=your_client_id
ANGEL_CLIENT_PIN=your_pin
ANGEL_TOTP_CODE=your_totp_code
ANGEL_API_KEY=your_api_key
ANGEL_CLIENT_LOCAL_IP=localhost
ANGEL_CLIENT_PUBLIC_IP=your_public_ip
ANGEL_MAC_ADDRESS=your_mac_address
ANGEL_STATE_VARIABLE=your_state_variable

# ClickHouse configuration
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
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

## Monitoring Metrics

Available Prometheus metrics:

- `angelone_market_data_processed_ticks_total`
- `angelone_market_data_error_count_total`
- `angelone_market_data_tick_processing_seconds`
- `angelone_market_data_batch_size_current`

## Health Checks

### Database Connectivity:
```bash
curl http://localhost:8080/health
```
```bash
curl http://localhost:8080/metrics
```

### Latest Data Verification:
```sql
SELECT
    token,
    max(timestamp) as last_update,
    now() - max(timestamp) as delay
FROM angelone_market_data
GROUP BY token
HAVING delay > INTERVAL 5 MINUTE;
```

## Error Handling

### Common Error Scenarios and Solutions

#### WebSocket Disconnection
- Automatic reconnection with exponential backoff
- State recovery and data gap detection

#### Database Connection Issues
- Connection pooling with automatic recovery
- Query timeout handling
- Batch insert retry logic

#### Rate Limiting
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
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

### Credit: https://www.marketcalls.in/python/storing-websocket-stock-market-data-in-clickhouse-using-python-a-comprehensive-guide.html
