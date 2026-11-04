# Deploy Money Manager — Hypercore / VPS (Git + PM2 + Nginx)

Chạy **song song** với [manage-tool](https://github.com/Kanghocit/manage-tool) — **không đụng** `ankhang.name.vn` / `api.ankhang.name.vn`.

| Domain | App |
|--------|-----|
| `ankhang.name.vn` | manage-tool FE (Nginx static) |
| `api.ankhang.name.vn` | manage-tool API → `:4000` |
| **`money.ankhang.name.vn`** | Money Manager → `:3001` (web) + `:8081` (API) |

---

## 1. DNS

Thêm bản ghi **A** `money` → cùng IP VPS với `ankhang` / `api.ankhang`.

---

## 2. VPS — phụ thuộc (một lần)

```bash
# Go 1.26+ (hoặc build binary trên máy dev rồi scp lên)
go version

# Redis (manage-tool chưa cần; Money Manager cần cho worker/OCR queue)
sudo apt update
sudo apt install -y redis-server
sudo systemctl enable redis-server

# PostgreSQL: dùng instance sẵn có, tạo DB riêng
sudo -u postgres psql -c "CREATE USER moneymanager WITH PASSWORD 'YOUR_STRONG_PASSWORD';"
sudo -u postgres psql -c "CREATE DATABASE moneymanager OWNER moneymanager;"
```

---

## 3. Clone & env

```bash
sudo mkdir -p /var/www/manage-money
sudo chown $USER:$USER /var/www/manage-money
git clone <repo-url> /var/www/manage-money
cd /var/www/manage-money
```

Tạo **`apps/api/.env`** (không commit):

```env
DATABASE_URL=postgres://moneymanager:YOUR_STRONG_PASSWORD@127.0.0.1:5432/moneymanager?sslmode=disable
REDIS_URL=redis://127.0.0.1:6379/0
API_PORT=8081

JWT_ACCESS_SECRET=<random-32+>
JWT_REFRESH_SECRET=<random-32+>
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h

COOKIE_SECURE=true
CORS_ORIGINS=https://money.ankhang.name.vn

LOCAL_STORAGE_PATH=/var/www/manage-money/uploads
STORAGE_DRIVER=local

# Tuỳ chọn
OPENAI_API_KEY=
SEED_ADMIN_EMAIL=
SEED_ADMIN_PASSWORD=
```

```bash
mkdir -p /var/www/manage-money/uploads
```

---

## 4. Nginx

```bash
sudo cp deploy/nginx/money.ankhang.name.vn.conf.example /etc/nginx/sites-available/money.ankhang.name.vn
sudo ln -sf /etc/nginx/sites-available/money.ankhang.name.vn /etc/nginx/sites-enabled/
sudo certbot --nginx -d money.ankhang.name.vn
sudo nginx -t && sudo systemctl reload nginx
```

**Không sửa** `/etc/nginx/sites-available/ankhang.name.vn` của manage-tool.

---

## 5. Deploy lần đầu & PM2

```bash
cd /var/www/manage-money
chmod +x scripts/deploy-production.sh scripts/pm2/*.sh
export DEPLOY_PATH=/var/www/manage-money
bash scripts/deploy-production.sh

pm2 startup   # làm theo hướng dẫn in ra
pm2 save
```

Kiểm tra:

```bash
pm2 list
curl -sI http://127.0.0.1:8081/health
curl -sI http://127.0.0.1:3001/
curl -sI https://money.ankhang.name.vn/
```

---

## 6. Deploy lại (sau git push)

```bash
ssh user@<vps>
export DEPLOY_PATH=/var/www/manage-money
export GIT_BRANCH=main   # hoặc production
bash /var/www/manage-money/scripts/deploy-production.sh
```

---

## 7. PM2 apps

| Name | Port / ghi chú |
|------|----------------|
| `money-api` | `8081` — đọc `apps/api/.env` qua `scripts/pm2/run-api.sh` |
| `money-worker` | Asynq OCR / jobs |
| `money-web` | `3001` — Next standalone |

manage-tool vẫn chỉ `manage-tool-api` trên `:4000`.

---

## 8. Checklist lỗi thường gặp

- **502** → `pm2 logs money-api` / `money-web`; port 8081/3001 có listen không.
- **Login không giữ session** → `COOKIE_SECURE=true` bắt buộc HTTPS; `NEXT_PUBLIC_API_URL=/api/v1`; `JWT_REFRESH_TTL=720h` (web tự gọi `/auth/refresh` khi access hết hạn).
- **Chi tiêu cũ chỉ có expense-day chưa xác nhận** → mở Daily, nhập lại và **Lưu** (hoặc gọi lại `POST .../confirm`) để đồng bộ sang bảng `transactions`.
- **OCR không chạy** → `money-worker` online + `REDIS_URL` + `OPENAI_API_KEY` (hoặc mock).
- **Ảnh hóa đơn mất sau reboot** → backup thư mục `uploads/`.

---

## Docker (local dev)

Vẫn dùng `deploy/docker-compose.yml` trên máy dev; production Hypercore dùng PM2 theo doc này.
