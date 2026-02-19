# Crypto Portfolio Tracker

Full-stack cryptocurrency portfolio tracking system built with **Go**, **ScyllaDB**, **ElasticSearch**, and **Docker**.

## 🚀 Features

- Real-time price tracking from CoinGecko API
- Background worker for automatic price updates
- Historical price data storage
- Fast token search with ElasticSearch
- Analytics and aggregations
- RESTful API

## 🛠 Tech Stack

**Backend:**
- Go 1.24.5
- Fiber (HTTP framework)
- ScyllaDB (time-series data)
- ElasticSearch (search & analytics)
- Docker & Docker Compose

**APIs:**
- CoinGecko API (crypto data)

## 📦 Project Structure


crypto-portfolio-tracker/

├── cmd/api/

│   └── main.go              # Application entry point

├── internal/

│   ├── db/

│   │   ├── scylla.go        # ScyllaDB client

│   │   └── elasticsearch.go # ElasticSearch client

│   ├── handlers/

│   │   └── handlers.go      # HTTP handlers

│   ├── models/

│   │   └── crypto.go        # Data models

│   └── services/

│       ├── coingecko.go     # CoinGecko API client

│       └── worker.go        # Background price worker

├── docker-compose.yml       # Infrastructure setup

├── .env                     # Configuration

└── README.md


## 🏃 Quick Start

### Prerequisites
- Docker Desktop
- Go 1.24+

### Installation

1. **Clone & navigate:**
\`\`\`bash
cd crypto-portfolio-tracker
\`\`\`

2. **Start infrastructure:**
\`\`\`bash
docker-compose up -d
\`\`\`

3. **Install dependencies:**
\`\`\`bash
go mod download
\`\`\`

4. **Run the API:**
\`\`\`bash
go run cmd/api/main.go
\`\`\`

API will be available at **http://localhost:8080**

## 📚 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/v1/health | Health check |
| POST | /api/v1/tokens | Add token manually |
| GET | /api/v1/tokens/:id | Get token by ID |
| GET | /api/v1/search?q=bitcoin | Search tokens |
| POST | /api/v1/sync?limit=10 | Sync from CoinGecko |
| GET | /api/v1/history/:id?limit=100 | Price history |
| GET | /api/v1/analytics | Market analytics |

## 🧪 Examples

**Search for Ethereum:**
\`\`\`bash
curl http://localhost:8080/api/v1/search?q=ethereum
\`\`\`

**Get Bitcoin price history:**
\`\`\`bash
curl http://localhost:8080/api/v1/history/bitcoin?limit=50
\`\`\`

**Get market analytics:**
\`\`\`bash
curl http://localhost:8080/api/v1/analytics
\`\`\`

**Sync top 20 tokens:**
\`\`\`bash
curl -X POST http://localhost:8080/api/v1/sync?limit=20
\`\`\`

## 🔧 Configuration

Edit \`.env\` to customize:
- Database connections
- API endpoints
- Worker sync interval
- Server port

## 🏗 Architecture

**ScyllaDB Schema:**
\`\`\`sql
-- Tokens table
CREATE TABLE tokens (
    id text PRIMARY KEY,
    symbol text,
    name text,
    current_price double,
    market_cap double,
    volume_24h double,
    updated_at timestamp
);

-- Price history (time-series)
CREATE TABLE price_history (
    token_id text,
    timestamp timestamp,
    price double,
    PRIMARY KEY (token_id, timestamp)
) WITH CLUSTERING ORDER BY (timestamp DESC);
\`\`\`

**Background Worker:**
- Runs every 1 minutes (configurable)
- Fetches top tokens from CoinGecko
- Updates both ScyllaDB and ElasticSearch
- Saves price history for charts

## 🐳 Docker Services

\`\`\`yaml
services:
  scylla:      # Port 9042
  elasticsearch: # Port 9200
\`\`\`

## 🚀 Deployment (AWS EC2)

Coming soon...

## 📝 License

MIT

## 👨‍💻 Author

Maksim Jatmanov - Backend Developer specializing in Go, Blockchain & AI
