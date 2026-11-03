# Money Manager

Ứng dụng quản lý chi tiêu cá nhân (tiếng Việt, responsive mobile/desktop, light/dark).

## Stack

- **Frontend:** Next.js, TypeScript, Tailwind, shadcn/ui, TanStack Query, Zustand, Recharts
- **Backend:** Go, Gin, GORM
- **Data:** PostgreSQL, Redis, Asynq worker
- **Auth:** JWT trong HttpOnly cookie, RBAC cơ bản

## Deploy production (Hypercore / VPS, subdomain money)

Giữ nguyên manage-tool trên `ankhang.name.vn` + `api.ankhang.name.vn`. Money Manager: **`money.ankhang.name.vn`** — Git + PM2 + Nginx.

Chi tiết: **[deploy/DEPLOY-HYPERCORE.md](deploy/DEPLOY-HYPERCORE.md)** · `ecosystem.config.cjs` · `scripts/deploy-production.sh`

## Chạy local (Docker)

```bash
cp .env.example .env
cd deploy
docker compose up --build
```

- UI qua Nginx: http://localhost
- API trực tiếp: http://localhost:8080/health

## Chạy dev (tách service)

1. Postgres + Redis: `cd deploy && docker compose up postgres redis -d`
2. API: `cd apps/api && go run ./cmd/server`
3. Worker: `cd apps/api && go run ./cmd/worker`
4. Web: `cd apps/web && npm run dev`

Web proxy `/api/v1` → API (xem `apps/web/next.config.ts`) để cookie auth hoạt động trên cùng origin.

## Tests

```bash
cd apps/api && go test ./...
cd apps/web && npm test
cd apps/web && npx playwright test
```

## Ảnh chi tiêu (OCR)

- Ảnh lưu tại **`LOCAL_STORAGE_PATH`** (mặc định `./uploads` hoặc volume Docker `/app/uploads`) hoặc **MinIO** khi `STORAGE_DRIVER=s3` (`docker compose --profile s3 up`).
- Màn **Chụp chi tiêu** (`/expenses/daily`) và **Lịch chi tiêu** (`/expenses/calendar`).
- OCR: đặt `OPENAI_API_KEY` trong `.env`; không có key → worker trả dòng mẫu để test UI.

## UI

Frontend dùng **MUI** (Material UI) + `@mui/x-date-pickers`, theme tím Money Manager.

## Phase sau

OAuth 2.0, pgvector search, Sentry, Prometheus/Grafana.
