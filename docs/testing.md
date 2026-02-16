# Testing Guide

## Unit Test

```bash
go test ./...
```

## Manual Functional Test (curl)

1. Create event and category (admin)
2. Reserve ticket (user)
3. Confirm booking (user)
4. Check availability reduced

## Race/Oversell Test

- Setup stock = 1
- Fire 2 reserve requests in parallel
- Expect one `201`, one `409`

## Idempotency Test

- Confirm same reservation twice
- Expect same booking ID

## Payment Rollback Test

- Confirm with `payment_ok=false` -> `402`
- Reserve again should succeed

## k6 Load Test

```bash
USER_TOKEN=<USER_TOKEN> \
EVENT_ID=<EVENT_ID> \
BASE_URL=http://localhost:8080 \
RATE=150 \
DURATION=20s \
PRE_VUS=150 \
MAX_VUS=1000 \
k6 run k6/reserve.js
```

Default SLO threshold:
- `p(95) < 200ms`
