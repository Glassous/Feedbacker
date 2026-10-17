# Feedbacker

轻量的自部署意见反馈微服务。任意项目通过一个 HTTP 接口提交反馈，运维者在自带的管理后台中集中查看、筛选与处理。

- 后端：Go 1.26 + Gin + MongoDB
- 管理后台：Vite + Vue 3 + TypeScript + Tailwind CSS v4
- 交付：后端 API 以 Docker 运行，管理后台独立构建为静态站点，可交给任意静态服务器托管

## 特性

- **极简接入**：`POST /api/v1/feedback`，标题、正文、联系方式均可选，至少填写一项即可。
- **来源追踪**：自动记录 `project`、`pageUrl`、`User-Agent`、`IP` 与提交时间，支持按来源筛选。
- **管理后台**：关键词搜索、状态筛选、来源筛选、分页、详情抽屉、未处理/已处理状态流转、删除。
- **无注册体系**：单管理员账号，首次启动生成一次性密码并打印到日志，登录后强制修改。
- **防滥用**：内置按 IP 限流、请求体大小限制、可选 API Key 与 CORS 白名单。
- **前后端分离**：后端 API 容器化运行，管理后台只产出静态文件，不参与容器构建。

## 快速开始

### 1. 启动后端 API

```bash
cp .env.example .env
docker compose up -d --build
```

首次启动时，容器日志会打印管理员账号与一次性密码：

```bash
docker compose logs -f app
```

```text
====================================================================
  已生成初始管理员账号
  用户名:     admin
  密码:       xxxxxxxxxxxxxxxx
  该密码为一次性密码，登录后必须立即修改。
====================================================================
```

后端接口地址为 `http://localhost:9090`。

> 数据库数据保存在 Docker 卷 `feedbacker_mongo-data` 中，删除容器不会丢失反馈。

### 2. 构建并部署管理后台

```bash
cd admin
npm ci
npm run build

# 本地预览（默认把 /api 代理到 127.0.0.1:9090）
npm run preview
```

将 `admin/dist` 上传到 Nginx、Caddy 或任意静态托管服务（示例见 [`admin/deploy/nginx.conf`](admin/deploy/nginx.conf)）。后台与 API 分域名部署时，构建时注入接口地址：

```bash
VITE_API_BASE_URL=https://feedback.glassous.top/api npm run build
```

并在后端 `.env` 中把 `CORS_ORIGINS` 设置为后台域名。若采用同源部署（静态服务器把 `/api` 反代到后端），不设置该变量即可。

前后端的云端构建与服务器发布由 GitHub Actions 完成，见 [部署与运维](docs/deploy.md) 的「CI/CD」一节。

## 接入示例

```bash
curl -X POST http://localhost:9090/api/v1/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "title": "无法导出报表",
    "content": "点击导出按钮后页面一直转圈",
    "contact": "you@example.com",
    "project": "web-console",
    "pageUrl": "https://app.example.com/report"
  }'
```

```javascript
await fetch('http://localhost:9090/api/v1/feedback', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    content: '在移动端无法上传头像',
    contact: 'user@example.com',
    project: 'mobile-app',
  }),
})
```

字段说明、错误码与更多语言示例见 [docs/api.md](docs/api.md)。

## 密码管理

首次启动后密码只打印一次。忘记密码时可在服务器上重置：

```bash
# 生成随机一次性密码（下次登录后仍需修改）
docker compose exec app ./feedbacker admin reset-password

# 或直接指定新密码
docker compose exec app ./feedbacker admin reset-password --password 'new-password-2026'
```

## 本地开发

前置：Go 1.26+、Node.js 20+、MongoDB 7+（或 Docker）。

```bash
# 1. 启动一个本地 MongoDB
docker run -d --name feedbacker-mongo -p 27017:27017 mongo:7

# 2. 启动后端（默认连接 127.0.0.1:27017）
cd server
MONGO_URI=mongodb://127.0.0.1:27017 go run ./cmd/feedbacker

# 3. 启动管理后台（Vite 代理 /api 到 127.0.0.1:9090）
cd admin
npm install
npm run dev
```

访问 `http://localhost:5173`，Vite 会把 `/api` 代理到 `127.0.0.1:9090`。

只构建后端二进制：

```bash
make build
./server/bin/feedbacker
```

## 目录结构

```text
.
├── admin/                  # Vite + Vue + Tailwind v4 管理后台
├── server/                 # Go 服务端
│   ├── cmd/feedbacker/     # 入口与 admin 运维子命令
│   └── internal/
│       ├── config/         # 环境变量配置
│       ├── database/       # MongoDB 连接与索引
│       ├── model/          # 数据模型
│       ├── repository/     # 数据访问
│       ├── service/        # 业务逻辑与认证
│       ├── handler/        # HTTP 接口
│       ├── middleware/     # 日志、CORS、限流、鉴权
│       └── router/         # 路由与中间件编排
├── deploy/                 # 服务器侧部署脚本与生产 compose（CI 使用）
├── docs/                   # 接口与部署文档
├── .github/workflows/      # 云端构建与发布工作流
├── Dockerfile
└── docker-compose.yml
```

## 文档

- [接口文档](docs/api.md)
- [部署与运维](docs/deploy.md)

## 安全建议

- 通过 Nginx/Caddy 反向代理并启用 HTTPS，不要将 MongoDB 端口暴露到公网。
- 公网部署时建议设置 `FEEDBACK_API_KEY`，并在接入方请求中携带 `X-API-Key`。
- 将 `CORS_ORIGINS` 收敛到实际接入的域名，而不是保持 `*`。
- 位于反向代理之后时设置 `TRUST_PROXY=true`，以保证限流基于真实客户端 IP。
