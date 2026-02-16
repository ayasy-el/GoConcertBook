# Project Overview

Concert Booking MVP adalah backend booking tiket konser untuk skenario high-concurrency (war tiket).

## Value Proposition

- Mencegah overselling dengan atomic stock reservation di Redis.
- Menjaga fairness dengan reservation TTL dan rollback otomatis.
- Menjaga stabilitas di burst traffic dengan queue gate + worker pool.
- Memisahkan write path lewat Kafka + worker supaya API tetap responsif.
- Menyediakan observability siap pakai via Prometheus + Grafana.

## Core Feature Set

- Admin event management
- Ticket category management
- Realtime availability
- Reserve ticket (TTL)
- Confirm booking (payment simulation)
- Expiry release and stock restoration
- JWT role auth + rate limit
- Metrics endpoint + dashboard

## Non-Functional Targets

- No double booking
- p95 reserve latency < 200ms (k6 benchmark)
- Resilient startup against dependency readiness
- End-to-end deployment via Docker Compose
