# Task Management — Middle Fullstack Go Assessment

Greenfield implementation. Stack: **Go (Gin) + GORM + MySQL 8 + Redis (go-redis)** backend, **React Native + TypeScript (Expo)** frontend.

> Note: earlier revision used PostgreSQL per request; this final build is **MySQL** as the original spec requires. Keyword search uses `LIKE` (MySQL's default `utf8mb4_0900_ai_ci` collation is case-insensitive, matching Postgres `ILIKE` semantics). Duplicate-title detection uses MySQL error `1062` plus a pre-check, returning `409` either way.

## Repo layout

```
.
├── backend/
│   ├── cmd/api/main.go                 # wiring: pg + redis + gin
│   ├── internal/model/task.go          # Task + TaskFilter
│   ├── internal/repository/            # GORM persistence, filtering/sort/pagination
│   ├── internal/service/               # validation, 409/404 rules
│   ├── internal/handler/               # Gin handlers, Redis cache, error envelope
│   ├── pkg/db/mysql.go                # connect + AutoMigrate + unique index
│   ├── pkg/cache/redis.go              # 60s cache, canonical key, invalidation
│   ├── migrations/001_create_tasks.{up,down}.sql
│   ├── Dockerfile
│   └── .env.example
├── frontend/  (Expo, TypeScript)
│   ├── App.tsx                         # TaskList screen + QueryClientProvider
│   ├── src/api/client.ts               # axios client + types
│   ├── src/hooks/useTasks.ts           # TanStack Query: list / update / delete
│   ├── src/components/                 # SearchInput, StatusFilter, Pagination, EditModal, TaskCard, LoadingState
│   └── __tests__/                      # SearchInput + TaskList tests
├── docker-compose.yml                  # mysql:8 + redis:7 + api
└── README.md
```

## Quickstart

### 1. Infra (MySQL + Redis)

```bash
docker compose up -d mysql redis
# MySQL localhost:3306 (taskuser/taskpass/taskdb), Redis localhost:6379
```

### 2. DB migration + seed

```bash
docker exec -i taskmanager-mysql mysql -h 127.0.0.1 -P 3306 -utaskuser -ptaskpass taskdb \
  < backend/migrations/001_create_tasks.up.sql
# Dev alternative: backend AutoMigrate runs on boot (same schema + unique title).

# Seed 10 demo tasks (idempotent — skips titles that already exist):
cd backend
DATABASE_URL="taskuser:taskpass@tcp(localhost:3306)/taskdb?parseTime=true&charset=utf8mb4&loc=Local" go run ./cmd/seed
```

### 3. Backend

```bash
cd backend
cp .env.example .env
go run ./cmd/api
# :8080, GET /health -> {"status":"ok"}
```

Or full stack:

```bash
docker compose up --build
```

### 4. Frontend (Expo)

```bash
cd frontend
npm install
cp .env.example .env   # EXPO_PUBLIC_API_URL=http://localhost:8080
npx expo start         # then w (web) / a (android) / i (ios)
npm test               # jest
./node_modules/.bin/tsc --noEmit
```

## API

Base `http://localhost:8080`. All errors share the envelope:

```json
{ "error": { "code": "CONFLICT", "message": "task title already exists" } }
```

| Method | Route | Success | Notes |
|---|---|---|---|
| `GET` | `/health` | `200 {"status":"ok"}` | |
| `GET` | `/api/tasks?status=&keyword=&assignee=&page=&limit=&sort=` | `200 {data[], meta{page,limit,total,total_pages}}` + `X-Cache: HIT/MISS` | `status∈{todo,in_progress,done}`, `keyword` LIKE title+description (case-insensitive), `assignee` LIKE, `sort∈{created_at,-created_at,updated_at,-updated_at,title,-title}` default `-created_at`, `page≥1` default 1, `limit 1..100` default 10 |
| `GET` | `/api/tasks/:id` | `200 {data}` / `404` | soft-deleted → `404` |
| `POST` | `/api/tasks {title*,description,status,assignee}` | `201 {data}` / `400` / `409` | create; duplicate `title` → `409 CONFLICT` (was `500` bug); invalidates list cache |
| `PUT` | `/api/tasks/:id {title?,description?,status?,assignee?}` | `200 {data}` / `400` / `404` / `409` | **Task 1** new endpoint; invalidates list cache |
| `DELETE` | `/api/tasks/:id` | `200 {data:{message}}` / `404` | soft delete (`deleted_at=NOW()`); hidden from list/get |

Example:

```bash
curl -s "http://localhost:8080/api/tasks?status=todo&keyword=fix&page=1&limit=10&sort=-created_at" | jq
curl -s -X POST localhost:8080/api/tasks -H 'Content-Type: application/json' \
  -d '{"title":"Fix login","status":"todo","assignee":"budi"}'
```

## Redis caching (Task 2)

- `GET /api/tasks` cached **60s** (`TaskListTTL`).
- Key = `tasks:list:<canonical-sorted-querystring>` via `BuildTasksKey` — **includes query params**, order-independent (`pkg/cache/redis.go`).
- `POST/PUT/DELETE` call `InvalidateTaskLists` → `SCAN tasks:list:*` + `DEL`.
- `X-Cache: HIT/MISS` header for observability; cache failure never fails the request (falls through to DB).
- Unit-tested with `miniredis`: key determinism, TTL=60s, invalidation (`pkg/cache/redis_test.go`).

## Frontend (Task 3 + Task 4 fixes)

- `SearchInput` — debounced 300ms, drives `keyword`.
- `StatusFilter` — `all/todo/in_progress/done` chips, drives `status`.
- `Pagination` — `Prev/Next`, `Page x / y`, drives `page` (limit 10).
- `EditModal` — edits title/desc/assignee/status via `PUT`; shows save error (e.g. duplicate → 409); `invalidateQueries` **refreshes list after update** (bug fix).
- `LoadingState` — `ActivityIndicator` + text; plus empty/error/retry states and pull-to-refresh.
- Soft-deleted tasks never render (backend filters `deleted_at IS NULL`).

## Testing (Task 5)

```bash
cd backend && go test ./... -count=1
cd frontend && npm test
```

- Backend: `service` (update, duplicate-409, search/filter/pagination, soft-delete-hides), `cache` (key includes params, 60s TTL, invalidation), `handler` (MISS→HIT, write-invalidates, 409 shape, PUT/DELETE flow, error envelope). SQLite + miniredis so no live infra needed.
- Frontend: `SearchInput` debounce test (required: search or list) + `TaskList` card/pagination test (RNTL + jest-expo).

## Evaluation mapping

Go 35% (Gin+GORM layered code), SQL 10% (migration + partial unique index + trigram), Redis 15% (TTL/key/invalidate), Frontend 20% (5 required UI pieces), Testing 10%, Code Quality 5% (`go vet` clean, consistent errors), Docs 5% (this README).
