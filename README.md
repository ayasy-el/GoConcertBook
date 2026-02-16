# Concert Ticket Booking MVP

Sistem booking tiket konser berbasis Golang dengan fokus high-concurrency, anti-overselling, observability, dan siap dijalankan via Docker Compose.

## Stack
- API & Worker: Go
- Stock & TTL reservation: Redis
- Event stream: Kafka
- Persistence: PostgreSQL
- Monitoring: Prometheus + Grafana
- Load test: k6

## Arsitektur
User -> API -> Redis (atomic stock + TTL) -> Kafka -> Worker -> PostgreSQL

## Endpoint MVP
- `POST /events` (admin)
- `POST /events/{id}/ticket-category` (admin)
- `GET /events/{id}/availability`
- `POST /reserve` (user)
- `POST /confirm` (user)
- `GET /health`
- `GET /metrics`

## Struktur Clean Architecture
- `cmd/` : entrypoints (`api`, `worker`, `token`)
- `internal/domain` : entity + repository/service contracts
- `internal/usecase` : business rules
- `internal/interface/http` : handler/router/middleware
- `internal/infrastructure` : adapter postgres/redis/kafka/memory
- `internal/observability` : metrics

## Jalankan Lokal
1. Start stack:
```bash
docker compose up -d --build
```
2. Generate token admin:
```bash
go run ./cmd/token --role=admin --sub=admin-1 --secret=dev-secret
```
3. Create event:
```bash
curl -X POST http://localhost:8080/events \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Coldplay Live","date":"2026-12-01T19:00:00Z"}'
```
4. Add category:
```bash
curl -X POST http://localhost:8080/events/<EVENT_ID>/ticket-category \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"name":"VIP","total_stock":1000,"price":2500000}'
```
5. Generate user token:
```bash
go run ./cmd/token --role=user --sub=user-1 --secret=dev-secret
```
6. Reserve:
```bash
curl -X POST http://localhost:8080/reserve \
  -H "Authorization: Bearer <USER_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"event_id":"<EVENT_ID>","category":"VIP","qty":1}'
```
7. Confirm (payment simulation):
```bash
curl -X POST http://localhost:8080/confirm \
  -H "Authorization: Bearer <USER_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"reservation_id":"<RESERVATION_ID>","payment_ok":true}'
```

## Monitoring
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` (admin/admin)

Core metrics:
- RPS (`http_requests_total`)
- Reservation success/fail (`reservation_total`, `failed_reservation_total`)
- Kafka lag (`kafka_consumer_lag`)
- Redis memory (`redis_memory_bytes`)
- DB pool (`db_open_connections`)

## Load Test (k6)
```bash
USER_TOKEN=<USER_TOKEN> EVENT_ID=<EVENT_ID> BASE_URL=http://localhost:8080 k6 run k6/reserve.js
```

Target threshold default:
- `p95 < 200ms`

## Commit Workflow
Sebelum commit jalankan:
```bash
go test ./...
```
CI (`.github/workflows/ci.yml`) juga menjalankan test untuk push/PR.
