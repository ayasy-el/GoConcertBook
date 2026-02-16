# Concert Booking API

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Architecture](https://img.shields.io/badge/Architecture-Clean-0A7EA4)](./docs/architecture.md)
[![Docker](https://img.shields.io/badge/Deploy-Docker%20Compose-2496ED?logo=docker&logoColor=white)](./docker-compose.yml)
[![Kafka](https://img.shields.io/badge/Event%20Streaming-Kafka-231F20?logo=apachekafka&logoColor=white)](./docs/architecture.md)
[![Redis](https://img.shields.io/badge/Realtime%20Stock-Redis-DC382D?logo=redis&logoColor=white)](./docs/architecture.md)
[![PostgreSQL](https://img.shields.io/badge/Persistence-PostgreSQL-4169E1?logo=postgresql&logoColor=white)](./migrations/001_init.sql)
[![Observability](https://img.shields.io/badge/Observability-Prometheus%20%2B%20Grafana-E6522C?logo=prometheus&logoColor=white)](./docs/operations.md)
[![Load Test](https://img.shields.io/badge/Load%20Test-k6-7D64FF?logo=k6&logoColor=white)](./k6/reserve.js)
[![Swagger](https://img.shields.io/badge/API-Swagger-85EA2D?logo=swagger&logoColor=black)](http://localhost:8080/swagger/index.html)
[![Status](https://img.shields.io/badge/Status-Production--Ready%20MVP-brightgreen)](./docs/overview.md)

🚀 High-concurrency ticket booking backend designed for **war tiket** scenarios: no overselling, fair reservation, and operational visibility out of the box.

## ✨ Why This Project

Most ticket-war systems fail in one or more of these areas:
- ⚠️ race condition leads to overselling
- ⚠️ synchronous write path melts under spike
- ⚠️ poor incident visibility in real-time
- ⚠️ unfair reserve/confirm flow for users

This MVP addresses those directly:
- ✅ **Atomic reserve in Redis** for real-time stock consistency
- ✅ **Kafka decoupling** for async persistence and better API responsiveness
- ✅ **TTL reservation + rollback** for fairness and stock recovery
- ✅ **Observability-first** with Prometheus and Grafana
- ✅ **Load-testable** using k6 with practical SLO threshold

## 🎯 What You Get

- 🛠️ Admin event and ticket category management
- ⚡ Realtime availability endpoint (backed by Redis stock)
- 🧱 Reserve endpoint with queue/backpressure control
- 💳 Confirm endpoint with idempotency and payment simulation
- ♻️ Expiry reaper for automatic stock release
- 🔐 JWT role auth (`admin` / `user`) and IP throttling
- 📘 Source-generated Swagger/OpenAPI
- 🐳 End-to-end Docker stack (API, Worker, Redis, Kafka, Postgres, Prometheus, Grafana)

## 🧭 System Snapshot

```text
Client
  -> API (Go)
     -> Redis (atomic stock + reservation TTL)
     -> Kafka (reserved/confirmed/expired events)
        -> Worker (Go consumer)
           -> PostgreSQL

Metrics: API + Worker -> Prometheus -> Grafana
```

## ⚡ Quick Start (2 Minutes)

```bash
# 1) 🚀 boot stack
docker compose up -d --build

# 2) 🔑 generate tokens
go run ./cmd/token --role=admin --sub=admin-1 --secret=dev-secret
go run ./cmd/token --role=user --sub=user-1 --secret=dev-secret

# 3) 📊 open docs and dashboards
# Swagger:    http://localhost:8080/swagger/index.html
# Prometheus: http://localhost:9090
# Grafana:    http://localhost:3000  (admin/admin)
```

## 📡 API Surface

Core endpoints:
- `POST /events` (admin)
- `POST /events/{id}/ticket-category` (admin)
- `GET /events/{id}/availability`
- `POST /reserve` (user)
- `POST /confirm` (user)
- `GET /health`
- `GET /metrics`
- `GET /swagger/index.html`

Detailed API examples and schemas:
- [`docs/api.md`](./docs/api.md)

## 🧪 Performance & Validation

k6 scenario included in repository:
- file: [`k6/reserve.js`](./k6/reserve.js)
- target threshold: `p(95) < 200ms`

Run benchmark:

```bash
USER_TOKEN=<USER_TOKEN> \
EVENT_ID=<EVENT_ID> \
BASE_URL=http://localhost:8080 \
RATE=150 DURATION=20s PRE_VUS=150 MAX_VUS=1000 \
k6 run k6/reserve.js
```

Test and validation playbook:
- [`docs/testing.md`](./docs/testing.md)
- [`docs/operations.md`](./docs/operations.md)

## 📚 Documentation Map

- Project overview: [`docs/overview.md`](./docs/overview.md)
- Architecture: [`docs/architecture.md`](./docs/architecture.md)
- API guide: [`docs/api.md`](./docs/api.md)
- Operations runbook: [`docs/operations.md`](./docs/operations.md)
- Testing guide: [`docs/testing.md`](./docs/testing.md)
- Swagger artifacts: [`docs/swagger`](./docs/swagger)

## 👨‍💻 Developer Workflow

```bash
go test ./...
```

CI runs test on push/PR via:
- `.github/workflows/ci.yml`

## 📄 License

MIT License. See `LICENSE`.
