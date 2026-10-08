# Rewrite Plan — Go + Gin + Vue 3 SPA

## Stack
- Backend: Go 1.25 + Gin + SQLC + Goose + Casbin + Slog + Testify + Excelize + MinIO
- Frontend: Vue 3 + TypeScript + Vite + Pinia + Vue Router + Tailwind CSS 4 + DaisyUI + Axios + Chart.js
- DB: PostgreSQL 18
- Deploy: Docker + Caddy (reverse proxy + auto TLS)

## Struktur Project

```
inventariskantor/
├── backend/
│   ├── cmd/server/main.go          # Gin server
│   ├── internal/
│   │   ├── config/                 # env loading
│   │   ├── db/                     # pgx pool + goose + sqlc generated
│   │   ├── auth/                   # JWT + Casbin RBAC
│   │   ├── handler/                # Gin handlers (REST API JSON)
│   │   ├── service/                # business logic
│   │   ├── middleware/             # auth, RBAC, logging, CORS
│   │   ├── minio/                  # MinIO client (foto storage)
│   │   └── logger/                 # slog setup
│   ├── migrations/                 # goose .sql files
│   ├── sqlc.yaml                   # sqlc config
│   ├── queries/                    # sqlc .sql queries
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── main.ts
│   │   ├── App.vue
│   │   ├── router/                 # Vue Router
│   │   ├── stores/                 # Pinia (auth, items, etc)
│   │   ├── api/                    # Axios instances
│   │   ├── views/                  # Dashboard, Items, Movement, etc
│   │   ├── components/             # reusable Vue components
│   │   ├── composables/            # Vue composables
│   │   └── assets/                 # CSS
│   ├── index.html
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml              # backend + frontend + postgres + minio + caddy
├── Caddyfile                       # reverse proxy + auto HTTPS
└── Makefile
```

## API Endpoints (REST JSON)

### Auth
- POST /api/auth/login → {token, user}
- POST /api/auth/logout
- GET /api/auth/me → {user}

### Items
- GET /api/items?q=&cat=&loc=&page=&per_page= → {data, total, page}
- GET /api/items/:id → {item}
- POST /api/items → {item}
- PUT /api/items/:id → {item}
- DELETE /api/items/:id

### Categories
- GET /api/categories
- POST /api/categories
- PUT /api/categories/:id
- DELETE /api/categories/:id

### Locations
- GET /api/locations
- POST /api/locations
- PUT /api/locations/:id
- DELETE /api/locations/:id

### Movement
- POST /api/movement/in → {result}
- POST /api/movement/out → {result}
- POST /api/movement/adjust → {result}
- POST /api/movement/bulk → {ok, fail, errors}
- GET /api/transactions?sku=&type=&from=&to=&page= → {data, total}

### Barcode
- GET /api/barcode/:id.png?fmt=qr → image/png

### Upload (MinIO)
- POST /api/upload → {url}

### Export
- GET /api/export/items.csv
- GET /api/export/items.xlsx
- GET /api/export/tx.csv
- GET /api/report.pdf

### Dashboard
- GET /api/dashboard → {stats, chart, lowStock, recentTx}

### Users (admin)
- GET /api/users
- POST /api/users
- PUT /api/users/:id/role
- DELETE /api/users/:id

### Audit
- GET /api/audit?page= → {data, total}

### Telegram Bot
- Internal goroutine (not HTTP)

## RBAC (Casbin)
```
r = sub, obj, act
p = admin, /api/users, GET
p = admin, /api/users, POST
p = admin, /api/items/:id, DELETE
p = petugas, /api/items, GET
p = petugas, /api/movement/*, POST
```

## Frontend Pages (Vue Router)
- /login
- / (Dashboard — Chart.js)
- /items (table + pagination + filter)
- /items/new, /items/:id/edit
- /categories
- /locations
- /movement (tabs: in/out/adjust)
- /history (filter + pagination)
- /barcode (QR grid)
- /adjust/bulk (CSV upload)
- /users (admin)
- /audit (admin)
- /password
- /scan/:id (public, no auth)

## docker-compose.yml
```yaml
services:
  postgres:
    image: postgres:18
    volumes: [pgdata:/var/lib/postgresql/data]
  minio:
    image: minio/minio
    command: server /data
    volumes: [miniodata:/data]
  backend:
    build: ./backend
    depends_on: [postgres, minio]
  frontend:
    build: ./frontend
  caddy:
    image: caddy:2
    ports: ["80:80", "443:443"]
    volumes: [./Caddyfile:/etc/caddy/Caddyfile]
```

## Caddyfile
```
{$DOMAIN} {
    handle /api/* {
        reverse_proxy backend:8080
    }
    handle {
        reverse_proxy frontend:3000
    }
}
```

## Migrasi dari app lama
1. Database schema sama (sudah Postgres)
2. Data tidak hilang (volume persistent)
3. API JSON baru menggantikan server-rendered
4. Foto: pindah dari web/uploads → MinIO
5. Telegram bot: reuse logic, pindah ke backend
```
