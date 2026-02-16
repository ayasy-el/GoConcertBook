# API Guide

Swagger UI:
- `GET /swagger/index.html`

OpenAPI JSON:
- `GET /swagger/doc.json`

## Auth

Authorization header:
```text
Authorization: Bearer <token>
```

Role claims:
- `admin`
- `user`

Generate token:
```bash
go run ./cmd/token --role=admin --sub=admin-1 --secret=dev-secret
go run ./cmd/token --role=user --sub=user-1 --secret=dev-secret
```

## Endpoints

- `GET /health`
- `GET /metrics`
- `POST /events` (admin)
- `POST /events/{id}/ticket-category` (admin)
- `GET /events/{id}/availability`
- `POST /reserve` (user)
- `POST /confirm` (user)

Lihat detail schema dan response code di Swagger UI.
