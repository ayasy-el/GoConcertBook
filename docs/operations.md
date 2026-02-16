# Operations Runbook

## Start

```bash
docker compose up -d --build
docker compose ps
```

## Logs

```bash
docker compose logs --no-color --tail=100 api worker kafka postgres redis prometheus grafana
```

## Restart critical services

```bash
docker compose restart api worker
```

## Full reset

```bash
docker compose down
# or with volume cleanup
docker compose down -v
```

## Health URLs

- API health: `http://localhost:8080/health`
- API metrics: `http://localhost:8080/metrics`
- Worker metrics: `http://localhost:9091/`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

## Smoke flow

1. Generate admin token
2. Create event
3. Create category
4. Generate user token
5. Reserve
6. Confirm
