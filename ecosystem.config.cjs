/** PM2 — production on VPS (Hypercore). Nginx: money.ankhang.name.vn → :3001 + :8081 */
const path = require("path");

const root = __dirname;

module.exports = {
  apps: [
    {
      name: "money-api",
      cwd: path.join(root, "apps/api"),
      script: path.join(root, "scripts/pm2/run-api.sh"),
      interpreter: "bash",
      instances: 1,
      exec_mode: "fork",
      autorestart: true,
      max_memory_restart: "512M",
    },
    {
      name: "money-worker",
      cwd: path.join(root, "apps/api"),
      script: path.join(root, "scripts/pm2/run-worker.sh"),
      interpreter: "bash",
      instances: 1,
      exec_mode: "fork",
      autorestart: true,
      max_memory_restart: "512M",
    },
    {
      name: "money-web",
      cwd: path.join(root, "apps/web/.next/standalone"),
      script: "server.js",
      instances: 1,
      exec_mode: "fork",
      autorestart: true,
      max_memory_restart: "768M",
      env: {
        NODE_ENV: "production",
        PORT: "3001",
        HOSTNAME: "127.0.0.1",
        NEXT_PUBLIC_API_URL: "/api/v1",
        INTERNAL_API_URL: "http://127.0.0.1:8081/api/v1",
      },
    },
  ],
};
