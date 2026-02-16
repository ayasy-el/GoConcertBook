# Architecture

## High-Level

```text
Client
  -> API (Go)
     -> Redis (stock + TTL reservation)
     -> Kafka (ticket events)
        -> Worker (Go consumer)
           -> PostgreSQL (durable data)

Observability:
API + Worker -> Prometheus -> Grafana
```

## Components

- API: validasi request, auth, reserve/confirm workflow.
- Redis: source of truth stok realtime, key TTL reservation.
- Kafka: event stream (`ticket.reserved`, `ticket.confirmed`, `ticket.expired`).
- Worker: consume event dan persist reservation.
- PostgreSQL: events, categories, reservations, bookings.

## Consistency Strategy

- Atomic stock decrement dengan Redis Lua.
- One winner semantics pada race reserve.
- Idempotent confirm booking (`CreateIfNotExists`).
- Expiry reaper untuk stock release.

## Scalability Notes

- API stateless.
- Worker horizontal-scale by consumer group.
- Redis/Kafka/Postgres tuning untuk production scale-out.
