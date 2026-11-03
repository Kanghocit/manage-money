#!/usr/bin/env bash
# Deploy manage-money on VPS (Git + PM2 + Nginx money.ankhang.name.vn)
# Required: DEPLOY_PATH
# Optional: GIT_BRANCH (default main), GO_BIN (default go)

set -euo pipefail

DEPLOY_PATH="${DEPLOY_PATH:?Set DEPLOY_PATH, e.g. /var/www/manage-money}"
GIT_BRANCH="${GIT_BRANCH:-main}"
GO_BIN="${GO_BIN:-go}"

echo "==> Deploy ${GIT_BRANCH} in ${DEPLOY_PATH}"

cd "${DEPLOY_PATH}"

if [[ ! -d .git ]]; then
  echo "ERROR: ${DEPLOY_PATH} is not a git repository."
  exit 1
fi

git fetch origin "${GIT_BRANCH}"
git reset --hard "origin/${GIT_BRANCH}"

chmod +x scripts/pm2/run-api.sh scripts/pm2/run-worker.sh

echo "==> API (Go)"
cd apps/api
mkdir -p bin
"${GO_BIN}" build -o bin/server ./cmd/server
"${GO_BIN}" build -o bin/worker ./cmd/worker
cd ../..

if [[ ! -f apps/api/.env ]]; then
  echo "WARN: apps/api/.env missing — copy from .env.example and configure on VPS."
fi

echo "==> Web (Next.js standalone)"
cd apps/web
rm -rf node_modules
npm ci
NEXT_PUBLIC_API_URL=/api/v1 INTERNAL_API_URL=http://127.0.0.1:8081/api/v1 npm run build
cp -r public .next/standalone/public
cp -r .next/static .next/standalone/.next/static
cd ../..

echo "==> PM2"
cd "${DEPLOY_PATH}"
if pm2 describe money-api >/dev/null 2>&1; then
  pm2 reload ecosystem.config.cjs --update-env
else
  pm2 start ecosystem.config.cjs
fi
pm2 save || true

echo "==> Deploy finished OK (money-api :8081, money-web :3001)"
