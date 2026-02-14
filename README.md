# URL Shortener

URL shortening service with analytics, QR code generation, and Redis caching. Go + PostgreSQL + Redis.

## Running

```bash
cp .env.example .env
docker-compose up -d
```

Or locally:

```bash
go mod download
go run cmd/api/main.go
```

Requires PostgreSQL 16+ and Redis 7+.

## Endpoints

| Method | Path | What it does |
|--------|------|-------------|
| POST | `/api/urls` | Shorten a URL |
| GET | `/{code}` | Redirect |
| GET | `/api/urls/{code}/stats` | Click analytics |
| GET | `/api/urls/{code}/qr` | QR code (PNG) |
| DELETE | `/api/urls/{id}` | Remove |
| GET | `/api/urls` | List URLs |

See `.env.example` for config (port, DB, Redis, rate limit, cache TTL).

## Testing

```bash
go test ./...
```

## License

MIT with attribution — see [LICENSE](LICENSE).
