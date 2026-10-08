# Inventaris Kantor — Sistem Barcode & Stok Real-Time

Aplikasi inventaris kantor dengan barcode scanning dan pemantauan stok real-time.

## Stack

| Layer | Teknologi |
|---|---|
| Backend | Go + Gin + pgx + JWT + Casbin RBAC |
| Frontend | Vue 3 + TypeScript + Vite + Pinia + Vue Router + PrimeVue 4 |
| Database | PostgreSQL 16 |
| Storage | MinIO (foto) |
| Proxy | Caddy (reverse proxy + auto TLS) |

## Struktur

```
├── backend/            # Go API server
│   ├── cmd/server/     # entry point
│   └── internal/       # auth, db, handler, middleware, service, telegram
├── frontend/           # Vue 3 SPA
│   └── src/            # views, components, stores, api
├── docker-compose.yml  # backend + frontend + postgres + minio + caddy
└── Caddyfile           # /api/* → backend:8080, lain → frontend:3000
```

## Jalankan

```bash
cp .env.example .env
make up           # docker compose up -d --build
make logs
```

Dev lokal:

```bash
make dev-backend    # go run ./cmd/server (port 8080)
make dev-frontend   # npm run dev (port 3000)
```

Buat user admin:

```bash
make create-user    # admin@kantor.id
```

## Port

- Caddy: 80 / 443
- backend: 8080 (internal)
- frontend: 3000 (internal, nginx)
- minio console: 9001

## Test

```bash
make test           # go test ./...
```
