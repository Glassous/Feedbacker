# 部署与运维

## 部署拓扑

```text
接入方前端 ──► https://feedback.glassous.top ──► 反向代理 ──► feedbacker 容器 ──► MongoDB

管理员浏览器 ──► https://feedbackeradmin.vodyanitsa.top（1Panel 静态站，root 指向 dist）
                        └── /api 请求 ──► https://feedback.glassous.top/api（跨域，CORS 白名单）
```

应用容器只提供 API，默认监听 `9090`；管理后台构建为静态站点，独立部署，不参与镜像构建。MongoDB 只暴露在 Docker 内部网络中，无需对外映射端口。

## CI/CD（GitHub Actions）

推送到 `main` 后由 GitHub 云端完成构建，服务器只负责拉取镜像与覆盖静态文件，**不在服务器上编译**。

| 工作流 | 触发路径 | 产物 | 落地位置 |
| --- | --- | --- | --- |
| `.github/workflows/backend.yml` | `server/**`、`Dockerfile` | 镜像 `ghcr.io/glassous/feedbacker:sha-xxxxxxx` | `/opt/feedbacker/docker-compose.yml` 对应的 compose 项目 |
| `.github/workflows/admin.yml` | `admin/**` | `admin/dist` | `/opt/1panel/www/sites/feedbackeradmin.vodyanitsa.top/index/dist` |

### 仓库 Secrets

| 名称 | 值 |
| --- | --- |
| `SERVER_HOST` | 服务器 IP 或可 SSH 的域名 |
| `SERVER_SSH_KEY` | 部署用私钥全文（包含 `-----BEGIN ...` 与 `-----END ...` 行） |
| `SERVER_USER` | `ubuntu` |

镜像推送使用 Actions 自动提供的 `GITHUB_TOKEN`（工作流已声明 `packages: write`），无需额外密钥；首次推送会在仓库的 Packages 下生成同名包，默认私有，工作流部署时会用临时令牌在服务器 `docker login` 后拉取，结束时 `docker logout`。

### 服务器一次性准备

```bash
# 部署用户可使用 Docker（执行后需重新登录 SSH 会话才生效）
sudo usermod -aG docker ubuntu

# 后端目录：.env 必须放在这里
sudo mkdir -p /opt/feedbacker
sudo chown -R ubuntu:ubuntu /opt/feedbacker

# 管理后台站点目录
sudo mkdir -p /opt/1panel/www/sites/feedbackeradmin.vodyanitsa.top/index/dist
sudo chown -R ubuntu:ubuntu /opt/1panel/www/sites/feedbackeradmin.vodyanitsa.top/index
```

- 部署脚本发现目录不可写时会自动退回 `sudo -n`（要求 `ubuntu` 具备免密 sudo），两者都不满足会直接失败并说明原因。
- 若 1Panel 会重置站点目录属主，重新执行上面的 `chown`，或给 `ubuntu` 配置免密 sudo。
- 首次切换到本工作流前，先停掉旧的后端容器，避免 9090 端口冲突导致新容器启动失败。数据卷名仍为 `feedbacker_mongo-data`，不会因为换部署目录而丢失数据。

### 部署流程

后端：

1. 云端 `docker buildx` 构建 `linux/amd64` 镜像（服务器为 ARM 时把工作流中的 `PLATFORMS` 改为 `linux/arm64`），推送 `sha-xxxxxxx` 与 `latest` 两个标签。
2. SSH 上传 `deploy/docker-compose.prod.yml`、`deploy/deploy-backend.sh` 到 `/tmp`。
3. 脚本写入 `/opt/feedbacker/docker-compose.yml`，执行 `docker compose --env-file .env pull app` 与 `up -d`，随后轮询 `http://127.0.0.1:<PORT>/healthz`；失败会输出最近 50 行日志并让工作流失败。

管理后台：

1. `npm ci && npm run build`，构建时注入 `VITE_API_BASE_URL`（见 `.github/workflows/admin.yml` 顶部 `env`）。
2. 打包 `admin/dist` 经 SSH 传至服务器，解压到临时目录后整体替换站点 `dist`，不残留上一版文件。

### 回滚

```bash
# 后端：回到任意历史标签（每次运行的摘要里能看到标签）
cd /opt/feedbacker
APP_IMAGE=ghcr.io/glassous/feedbacker:sha-1a2b3c4 docker compose --env-file .env up -d

# 管理后台：对目标提交重新运行 Admin 工作流（workflow_dispatch）
```

历史镜像不会自动回收，可定期清理：

```bash
docker image prune -a --filter "until=720h"
```

### 生产 `.env` 关键项

```bash
CORS_ORIGINS=https://feedbackeradmin.vodyanitsa.top
TRUST_PROXY=true
```

其余项见下方「环境变量」。

## Docker Compose 部署（手动）

生产环境推荐使用上面的工作流；以下命令用于本地或应急，会在当前机器上构建镜像。

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
| `CORS_ORIGINS` | `*` | 允许跨域调用接口的来源（管理后台域名、接入方站点），逗号分隔，不带结尾斜杠 |
| `TRUST_PROXY` | `false` | 反向代理之后设为 `true`，启用 `X-Forwarded-For` |
| `RATE_LIMIT_PER_MINUTE` | `60` | 每个 IP 每分钟的提交上限 |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `TZ` | `Asia/Shanghai` | 容器时区 |

## 管理后台部署

管理后台只产出静态文件，独立于后端容器部署。接口地址在构建时注入，未设置时使用同源 `/api`（本地开发走 Vite 代理）。

```bash
cd admin
npm ci

# 指向独立部署的 API 域名（跨域），生产构建使用这一条
VITE_API_BASE_URL=https://feedback.glassous.top/api npm run build

# 本地预览构建结果（代理 /api 到 127.0.0.1:9090）
npm run preview
```

线上发布由 `Admin` 工作流完成，落盘位置：

```text
/opt/1panel/www/sites/feedbackeradmin.vodyanitsa.top/index/dist
```

### Nginx（1Panel）

1Panel 静态站的默认站点目录是 `/opt/1panel/www/sites/<域名>/index`，而 dist 发布在其下的 `dist` 子目录，因此需要把站点根目录指向 dist：

- 面板方式：网站 → feedbackeradmin.vodyanitsa.top → 配置 → 网站目录，改为 `/opt/1panel/www/sites/feedbackeradmin.vodyanitsa.top/index/dist`。
- 文件方式：编辑 `/opt/1panel/apps/openresty/openresty/conf/conf.d/feedbackeradmin.vodyanitsa.top.conf`（Nginx 则在 `/opt/1panel/apps/nginx/nginx/conf/conf.d/`），把 `root` 改成上面的 dist 路径，保存后由面板校验并重载（未生效时在「应用」里重启 OpenResty）。

必须确认：

- `root` 指向 dist 目录。Vite 产物用 `/assets/...` 绝对路径引用资源，root 指错会整站白屏。
- 存在 SPA 回退 `try_files $uri $uri/ /index.html`，否则刷新 `/settings` 这类前端路由会 404。
- 开启 SSL 后 1Panel 生成的 443 server 块沿用同一个 root，无需重复配置。
- 静态站点不需要反向代理；只有不想跨域时，才把 `/api` 反代到 `https://feedback.glassous.top`。
- `admin/dist` 中带 hash 的 `assets/*` 可长期缓存，`index.html` 应禁止缓存。

完整示例见 [`admin/deploy/nginx.conf`](../admin/deploy/nginx.conf)。

## 跨域（CORS）

后台页面在 `https://feedbackeradmin.vodyanitsa.top`，接口在 `https://feedback.glassous.top`，属于跨域请求：

- 后端 `.env` 设置 `CORS_ORIGINS=https://feedbackeradmin.vodyanitsa.top`（多个来源用英文逗号分隔，不要带结尾斜杠）。
- 后端 `.env` 设置 `TRUST_PROXY=true`，否则限流与日志记录的 IP 会是代理地址。
- 登录态走 `Authorization: Bearer`，不依赖 Cookie，因此后端不返回 `Access-Control-Allow-Credentials`，前端也不要开启 `withCredentials`。
- 预检 `OPTIONS` 由 `server/internal/middleware/cors.go` 统一返回 204，并带上允许的 method 与 header（`Authorization`、`Content-Type`、`X-API-Key`、`X-Request-ID`），`Access-Control-Max-Age` 为 86400。

自检：

```bash
curl -i -X OPTIONS https://feedback.glassous.top/api/admin/login \
  -H "Origin: https://feedbackeradmin.vodyanitsa.top" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: authorization,content-type"
```

响应中应出现：

```text
access-control-allow-origin: https://feedbackeradmin.vodyanitsa.top
access-control-allow-headers: Content-Type, Authorization, X-API-Key, X-Request-ID
```

改为同源部署时，把 `CORS_ORIGINS` 中的后台域名去掉、构建变量改回 `/api`，并在站点上反代 `/api`。

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

以下命令均在服务器 `/opt/feedbacker` 下执行（工作流部署的 compose 文件即该目录下的 `docker-compose.yml`）。

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

推送到 `main` 后由工作流自动完成。手动触发或回滚见上文「回滚」。

```bash
# 服务器手工同步到某个镜像标签
cd /opt/feedbacker
APP_IMAGE=ghcr.io/glassous/feedbacker:sha-1a2b3c4 docker compose --env-file .env up -d
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
