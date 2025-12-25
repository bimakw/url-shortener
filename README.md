# URL Shortener

[![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-DC382D?style=flat&logo=redis&logoColor=white)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A high-performance URL shortening service built with Go, featuring analytics, QR code generation, and Redis caching.

## Features

- **URL Shortening**
  - Generate short codes using cryptographically secure nanoid
  - Custom alias support (e.g., `yourdomain.com/my-link`)
  - URL expiration (optional)

- **Analytics & Tracking**
  - Click counting (total & unique)
  - Referrer tracking
  - Device/Browser/OS detection
  - Geographic data (country, city)
  - Time-series click data

- **QR Code Generation**
  - Generate QR codes for any short URL
  - Configurable size (64-1024px)
  - PNG format with caching

- **Performance & Security**
  - Redis caching for fast redirects
  - Rate limiting (configurable)
  - Input validation
  - SQL injection protection

## Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.24 |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Architecture | Clean Architecture |
| Container | Docker |

## Project Structure

```
url-shortener/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── adapter/
│   │   ├── inbound/
│   │   │   └── http/            # HTTP handlers & middleware
│   │   └── outbound/
│   │       ├── postgres/        # PostgreSQL repositories
│   │       └── redis/           # Redis cache
│   ├── application/
│   │   └── usecase/             # Business logic
│   ├── domain/
│   │   ├── entity/              # Domain entities
│   │   └── repository/          # Repository interfaces
│   └── infrastructure/          # Config
├── pkg/
│   └── nanoid/                  # Short code generator
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## API Endpoints

### URL Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/urls` | Create short URL |
| GET | `/api/urls/{code}` | Get URL info |
| GET | `/api/urls/{code}/stats` | Get click statistics |
| GET | `/api/urls/{code}/qr` | Generate QR code |
| DELETE | `/api/urls/{id}` | Delete URL |
| GET | `/api/urls` | List user's URLs |

### Redirect
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/{code}` | Redirect to original URL |

## Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 16+
- Redis 7+ (optional, for caching)
- Docker (optional)

### Run with Docker (Recommended)

```bash
# Clone repository
git clone https://github.com/bimakw/url-shortener.git
cd url-shortener

# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f api
```

### Run Locally

```bash
# Clone repository
git clone https://github.com/bimakw/url-shortener.git
cd url-shortener

# Setup environment
cp .env.example .env
# Edit .env with your database credentials

# Install dependencies
go mod download

# Run application
go run cmd/api/main.go
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Server port |
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | postgres | Database password |
| `DB_NAME` | url_shortener | Database name |
| `REDIS_HOST` | localhost | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `BASE_URL` | http://localhost:8080 | Base URL for short links |
| `SHORT_CODE_LENGTH` | 8 | Length of generated short codes |
| `RATE_LIMIT` | 100 | Requests per second per IP |
| `CACHE_TTL` | 1h | Redis cache TTL |

## Usage Examples

### Create Short URL

```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{
    "original_url": "https://example.com/very/long/url/here",
    "custom_alias": "my-link",
    "expires_in": 24
  }'
```

Response:
```json
{
  "success": true,
  "message": "Short URL created",
  "data": {
    "short_code": "my-link",
    "short_url": "http://localhost:8080/my-link",
    "original_url": "https://example.com/very/long/url/here",
    "expires_at": "2024-01-02T12:00:00Z",
    "created_at": "2024-01-01T12:00:00Z",
    "click_count": 0,
    "qr_code_url": "http://localhost:8080/api/urls/my-link/qr"
  }
}
```

### Get URL Statistics

```bash
curl http://localhost:8080/api/urls/my-link/stats?from=2024-01-01&to=2024-01-31
```

Response:
```json
{
  "success": true,
  "data": {
    "total_clicks": 1250,
    "unique_clicks": 890,
    "clicks_by_date": {
      "2024-01-01": 45,
      "2024-01-02": 67
    },
    "top_referrers": [
      {"referrer": "twitter.com", "count": 320},
      {"referrer": "Direct", "count": 280}
    ],
    "top_countries": [
      {"country": "Indonesia", "count": 650},
      {"country": "USA", "count": 200}
    ],
    "top_browsers": [
      {"browser": "Chrome", "count": 800},
      {"browser": "Safari", "count": 250}
    ]
  }
}
```

### Generate QR Code

```bash
# Default size (256px)
curl http://localhost:8080/api/urls/my-link/qr -o qr.png

# Custom size
curl "http://localhost:8080/api/urls/my-link/qr?size=512" -o qr-large.png
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Client                              │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Handler Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │ URLHandler  │  │ QRHandler   │  │ Middleware          │  │
│  │             │  │             │  │ (CORS, RateLimit,   │  │
│  │             │  │             │  │  Logging, Recovery) │  │
│  └──────┬──────┘  └──────┬──────┘  └─────────────────────┘  │
└─────────┼────────────────┼──────────────────────────────────┘
          │                │
          ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Use Case Layer                           │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   URLUseCase                         │    │
│  │  • CreateShortURL  • GetOriginalURL  • RecordClick  │    │
│  │  • GetStats        • DeleteURL       • GetUserURLs  │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
          │                │
          ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                    Repository Layer                         │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────┐   │
│  │ URLRepository  │  │ ClickRepository│  │ URLCache     │   │
│  │ (PostgreSQL)   │  │ (PostgreSQL)   │  │ (Redis)      │   │
│  └────────────────┘  └────────────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Performance

- **Redirect latency**: ~1-5ms (with Redis cache hit)
- **URL creation**: ~10-20ms
- **Throughput**: 10,000+ requests/second (depends on hardware)

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v
```

| Package | Tests | Coverage |
|---------|-------|----------|
| nanoid | 10 | ID generation, uniqueness, alphabet validation |
| utm | 20 | Build, HasUTM, Strip, edge cases |
| entity | 22 | URL struct, expiry, redirect logic |

## License

This project is licensed under the MIT License with Attribution Requirement - see the [LICENSE](LICENSE) file for details.

## Author

**Bima Kharisma Wicaksana**
- GitHub: [@bimakw](https://github.com/bimakw)
- LinkedIn: [Bima Kharisma Wicaksana](https://www.linkedin.com/in/bima-kharisma-wicaksana-aa3981153/)
