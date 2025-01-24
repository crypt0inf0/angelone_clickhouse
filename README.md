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
git clone https://github.com/crypt0inf0/broker_clickhouse.git
cd broker_clickhouse
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
SELECT * FROM market_ticks WHERE symbol = '2885' LIMIT 5;
```

## Project Structure

```
broker_clickhouse/
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

```go
const (
    batchSize     = 1000    // Number of ticks per batch
    flushInterval = 5       // Seconds between forced flushes
)
```

Adjust these values based on your requirements:
- Higher batchSize = Better throughput
- Lower flushInterval = Lower latency

## Additional Production Requirements

1. Monitoring and Alerting
- Add Prometheus metrics export
- Setup alerting thresholds
- Integration with monitoring systems

2. Security
- Add API rate limiting
- Implement request authentication
- Add input validation
- Secure credential management

3. Deployment
- Add Kubernetes manifests
- Setup CI/CD pipeline
- Add health checks
- Resource limits configuration

4. Documentation
- Add API documentation
- Add runbook for operations
- Document failover procedures
- Add performance tuning guide

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
