# Concert Booking MVP

High-concurrency ticket booking backend for concert war scenarios.

## Quick Links

- Full docs index: [`docs/README.md`](docs/README.md)
- Swagger UI: `http://localhost:8080/swagger/index.html`
- Docker stack: `docker compose up -d --build`

## Fast Start

```bash
docker compose up -d --build
go run ./cmd/token --role=admin --sub=admin-1 --secret=dev-secret
go run ./cmd/token --role=user --sub=user-1 --secret=dev-secret
```

## Key Features

- Atomic anti-oversell reserve
- Realtime availability via Redis
- Kafka async persistence worker
- Idempotent confirm + rollback logic
- Prometheus/Grafana observability
- k6 performance test script
