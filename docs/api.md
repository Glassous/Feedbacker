# 接口文档

所有接口均使用 JSON 通信，基础地址为服务部署地址，例如 `http://localhost:9090`。

## 统一约定

- 请求头：`Content-Type: application/json`
- 时间：UTC ISO 8601 字符串，例如 `2026-10-17T06:25:50Z`
- 出错时返回统一结构：

```json
{
  "code": "invalid_argument",
  "message": "请至少填写标题、正文或联系方式中的一项"
}
```

| code | HTTP 状态 | 说明 |
| --- | --- | --- |
| `invalid_argument` | 400 | 参数校验失败 |
| `unauthorized` | 401 | 未登录、令牌失效或 API Key 错误 |
| `password_change_required` | 403 | 仍在使用一次性密码，需要先修改 |
| `not_found` | 404 | 资源不存在 |
| `payload_too_large` | 413 | 请求体超过 32KB |
| `rate_limited` | 429 | 触发限流 |
| `internal` | 500 | 服务端异常 |

---

## 提交反馈（公开）

```http
POST /api/v1/feedback
```

### 请求字段

| 字段 | 类型 | 必填 | 限制 | 说明 |
| --- | --- | --- | --- | --- |
| `title` | string | 否 | ≤ 200 字符 | 反馈标题 |
| `content` | string | 否 | ≤ 5000 字符 | 反馈正文，保留换行 |
| `contact` | string | 否 | ≤ 200 字符 | 联系方式，邮箱/手机号/IM 均可 |
| `project` | string | 否 | ≤ 64 字符 | 来源项目标识，用于后台区分与筛选 |
| `pageUrl` | string | 否 | ≤ 2048 字符 | 反馈发生时所在页面，必须是 http(s) 链接 |

`title`、`content`、`contact` 三者至少填写一项。服务端会自动记录提交者 IP 与 `User-Agent`，无需客户端传入。

### 成功响应

```http
HTTP/1.1 201 Created
```

```json
{
  "id": "6712f3a1c2d3e4f5a6b7c8d9",
  "createdAt": "2026-10-17T06:25:50Z"
}
```

### 示例

```bash
curl -X POST http://localhost:9090/api/v1/feedback \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-shared-key" \
  -d '{
    "title": "无法导出报表",
    "content": "点击导出按钮后页面一直转圈，控制台报 500。",
    "contact": "zhang@example.com",
    "project": "web-console",
    "pageUrl": "https://app.example.com/report?tab=daily"
  }'
```

```javascript
const response = await fetch('http://localhost:9090/api/v1/feedback', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    // 仅在服务端配置了 FEEDBACK_API_KEY 时需要
    // 'X-API-Key': 'your-shared-key',
  },
  body: JSON.stringify({
    content: '在移动端无法上传头像',
    contact: 'user@example.com',
    project: 'mobile-app',
  }),
})

if (!response.ok) {
  const error = await response.json()
  console.error(error.code, error.message)
}
```

```python
import requests

response = requests.post(
    "http://localhost:9090/api/v1/feedback",
    json={
        "title": "结算页金额显示异常",
        "content": "优惠券抵扣后总价没有更新",
        "contact": "user@example.com",
        "project": "checkout-web",
    },
    timeout=10,
)
response.raise_for_status()
print(response.json()["id"])
```

### 限流与防滥用

- 默认每个 IP 每分钟最多提交 60 次，可通过 `RATE_LIMIT_PER_MINUTE` 调整。
- 响应头包含 `X-RateLimit-Limit`、`X-RateLimit-Remaining`，超限时返回 `429` 与 `Retry-After`。
- 请求体上限 32KB；字段长度超限会返回具体的校验提示。
- 若配置了 `FEEDBACK_API_KEY`，请求必须携带 `X-API-Key` 请求头。

---

## 管理员接口

除登录外，所有管理接口都需要请求头：

```
Authorization: Bearer <token>
```

### 登录

```http
POST /api/admin/login
```

```json
{
  "username": "admin",
  "password": "一次性密码或新密码"
}
```

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "username": "admin",
  "mustChangePassword": true
}
```

`mustChangePassword` 为 `true` 时，除 `/api/admin/me`、`/api/admin/password`、`/api/admin/logout` 外的接口都会返回 `403 password_change_required`。

### 当前账号

```http
GET /api/admin/me
```

```json
{
  "username": "admin",
  "mustChangePassword": false
}
```

### 修改密码

```http
POST /api/admin/password
```

```json
{
  "oldPassword": "当前密码",
  "newPassword": "至少 8 个字符"
}
```

成功后返回新的 `token`（旧令牌立即失效），结构同登录响应。

### 反馈列表

```http
GET /api/admin/feedback?page=1&pageSize=20&q=导出&status=open&project=web-console
```

| 参数 | 类型 | 默认 | 说明 |
| --- | --- | --- | --- |
| `page` | int | 1 | 页码 |
| `pageSize` | int | 20 | 每页数量，最大 100 |
| `q` | string | - | 关键词，匹配标题、正文、联系方式（忽略大小写，≤ 64 字符） |
| `status` | string | - | `open` 或 `resolved` |
| `project` | string | - | 来源项目 |

```json
{
  "items": [
    {
      "id": "6712f3a1c2d3e4f5a6b7c8d9",
      "title": "无法导出报表",
      "content": "点击导出按钮后页面一直转圈。",
      "contact": "zhang@example.com",
      "project": "web-console",
      "pageUrl": "https://app.example.com/report",
      "userAgent": "Mozilla/5.0 ...",
      "ip": "203.0.113.10",
      "status": "open",
      "createdAt": "2026-10-17T06:25:50Z",
      "updatedAt": "2026-10-17T06:25:50Z"
    }
  ],
  "total": 1,
  "page": 1,
  "pageSize": 20,
  "counts": {
    "total": 12,
    "open": 5,
    "resolved": 7
  }
}
```

- `total` 是当前筛选条件下的总数，用于分页。
- `counts` 是全部反馈的统计，不受筛选条件影响。

### 反馈详情

```http
GET /api/admin/feedback/:id
```

### 更新状态

```http
PATCH /api/admin/feedback/:id
```

```json
{ "status": "resolved" }
```

状态仅支持 `open`（未处理）与 `resolved`（已处理）。标记为 `resolved` 时自动记录 `resolvedAt`，重新打开会清除该字段。返回更新后的完整反馈。

### 删除反馈

```http
DELETE /api/admin/feedback/:id
```

成功返回 `204 No Content`。

### 来源项目列表

```http
GET /api/admin/projects
```

```json
{ "projects": ["checkout-web", "mobile-app", "web-console"] }
```

### 退出登录

```http
POST /api/admin/logout
```

服务端不维护会话，客户端丢弃本地令牌即可。该接口返回 `204`。

---

## 其他

### 站点信息

```http
GET /api/meta
```

```json
{ "siteName": "Feedbacker", "version": "1.0.0" }
```

### 健康检查

```http
GET /healthz
```

```json
{ "status": "ok", "version": "1.0.0" }
```

MongoDB 不可用时返回 `503`，便于容器编排与负载均衡健康检查。
