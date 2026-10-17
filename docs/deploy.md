# 部署与运维

## 部署拓扑

```text
接入方前端 ──► API 反向代理 ──► feedbacker 容器 ──► MongoDB

管理员浏览器 ──► 静态管理后台（Nginx）
                      └── /api 请求 ──► feedbacker 容器
```

应用容器只提供 API，默认监听 `9090`；管理后台构建为静态站点，独立部署，不参与镜像构建。MongoDB 只暴露在 Docker 内部网络中，无需对外映射端口。

## Docker Compose 部署

```bash
cp .env.example .env
docker compose up -d --build
docker compose ps
```

首次启动日志中会输出一次性管理员密码：

```bash
docker compose logs -f app
```

后端启动后仅提供 API 服务，管理后台按下一节单独部署。

### 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `9090` | 宿主机映射端口（compose 使用），容器内始终监听 9090 |
| `SITE_NAME` | `Feedbacker` | 管理后台与页面标题展示的名称 |
| `ADMIN_USERNAME` | `admin` | 首次启动创建的管理员用户名；账号已存在时不生效 |
| `MONGO_URI` | `mongodb://mongo:27017` | MongoDB 连接串，连接外部数据库时覆盖 |
| `MONGO_DB` | `feedbacker` | 数据库名 |
| `JWT_SECRET` | 空 | 令牌签名密钥；留空时自动生成并持久化到 `settings` 集合 |
| `JWT_TTL_HOURS` | `12` | 登录令牌有效期（小时） |
| `FEEDBACK_API_KEY` | 空 | 设置后公开提交接口必须携带 `X-API-Key` |
| `CORS_ORIGINS` | `*` | 允许跨域调用公开提交接口的来源，逗号分隔 |
| `TRUST_PROXY` | `false` | 反向代理之后设为 `true`，启用 `X-Forwarded-For` |
| `RATE_LIMIT_PER_MINUTE` | `60` | 每个 IP 每分钟的提交上限 |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `TZ` | `Asia/Shanghai` | 容器时区 |

## 管理后台部署

管理后台只产出静态文件，可部署到 Nginx、Caddy、对象存储或任意静态托管服务。

```bash
cd admin
npm ci
npm run build

# 本地预览构建结果（代理 /api 到 127.0.0.1:9090）
npm run preview
```

将 `admin/dist` 上传到静态服务器，并由该服务器把 `/api` 代理到后端。后台与 API 同源访问，不需要配置 API 地址或 CORS。示例见 `admin/deploy/nginx.conf`。

- 静态服务器需要配置 SPA 回退：`try_files $uri $uri/ /index.html`。
- `admin/dist` 中带 hash 的 `assets/*` 可长期缓存，`index.html` 应禁止缓存。

## 反向代理

应用位于代理之后时，设置 `TRUST_PROXY=true`，并确保代理传递真实客户端 IP，否则限流会把所有请求算作同一个 IP。

### Nginx

```nginx
server {
    listen 80;
    server_name feedback.example.com;

    client_max_body_size 64k;

    location / {
        proxy_pass http://127.0.0.1:9090;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Caddy

```caddyfile
feedback.example.com {
    reverse_proxy 127.0.0.1:9090
}
```

Caddy 会自动处理 HTTPS 与转发头，此时仍然建议设置 `TRUST_PROXY=true`。

## 运维命令

### 重置管理员密码

```bash
# 随机生成一次性密码
docker compose exec app ./feedbacker admin reset-password

# 指定新密码（不会强制再次修改）
docker compose exec app ./feedbacker admin reset-password --password 'new-password-2026'
```

重置后旧的登录令牌会立即失效。

### 查看版本

```bash
docker compose exec app ./feedbacker version
```

### 健康检查

```bash
curl http://localhost:9090/healthz
```

容器镜像内置 `HEALTHCHECK`，`docker compose ps` 中会显示健康状态。

## 数据备份与恢复

```bash
# 备份
docker compose exec -T mongo mongodump --archive=/tmp/feedbacker.archive --gzip --db feedbacker
docker compose cp mongo:/tmp/feedbacker.archive ./feedbacker-$(date +%Y%m%d).archive.gz

# 恢复
docker compose cp ./feedbacker-20260101.archive.gz mongo:/tmp/feedbacker.archive
docker compose exec mongo mongorestore --archive=/tmp/feedbacker.archive --gzip --drop
```

如需定时备份，可将上述命令放入宿主机 cron 任务。

## 升级

```bash
git pull
docker compose up -d --build
```

索引会在启动时自动创建。升级前建议先执行一次备份；数据卷 `mongo-data` 不会被重建。

## 使用外部 MongoDB

在 `.env` 中覆盖连接串：

```bash
MONGO_URI=mongodb://user:password@mongo.example.com:27017/?authSource=admin
```

同时从 `docker-compose.yml` 中删除 `mongo` 服务与 `depends_on`，或改用独立的 compose 文件。生产环境建议为数据库开启认证，并限制来源 IP。

## 安全清单

- 使用 HTTPS，不要直接暴露 `9090` 到公网。
- 公网环境下设置 `FEEDBACK_API_KEY`，并让接入方通过服务端代理提交，避免密钥出现在浏览器代码中。
- 将 `CORS_ORIGINS` 限制为实际接入方的域名。
- 管理后台是纯静态站点，不要在其中写入 `FEEDBACK_API_KEY`。
- 定期备份 MongoDB 数据卷。
- 不要将 MongoDB 端口映射到宿主机公网地址。

> 提示：如果 `FEEDBACK_API_KEY` 直接写在前端代码里，任何访问者都能拿到。公开页面建议通过自己的后端转发提交，或接受在没有 API Key 的情况下依赖限流防护。

## 故障排查

| 现象 | 排查方向 |
| --- | --- |
| 启动即退出，日志提示连接 MongoDB 失败 | 检查 `MONGO_URI`、MongoDB 是否健康、网络是否互通 |
| 找不到初始密码 | 执行 `docker compose exec app ./feedbacker admin reset-password` |
| 登录后所有接口返回 403 | 仍在使用一次性密码，请按提示修改密码 |
| 提交返回 429 | 触发限流，调整 `RATE_LIMIT_PER_MINUTE` 或检查 `TRUST_PROXY` |
| 提交返回 401 | 服务端配置了 `FEEDBACK_API_KEY`，请求需携带 `X-API-Key` |
| 后台页面刷新后 404 | 静态服务器未配置 SPA 回退，需添加 `try_files $uri $uri/ /index.html` |
| 后台接口请求 404 | 静态服务器未把 `/api` 代理到后端 |
| 限流把所有用户识别为同一 IP | 反向代理未传 `X-Forwarded-For`，或未设置 `TRUST_PROXY=true` |
