#!/usr/bin/env bash
# 在服务器上执行：更新 compose 文件、拉取镜像、重启后端容器并做健康检查。
# 由 .github/workflows/backend.yml 上传到 /tmp 后调用，服务器上不执行任何构建。
#
# 用法: deploy-backend.sh <image:tag> [app_dir]
set -euo pipefail

IMAGE="${1:?用法: deploy-backend.sh <image:tag> [app_dir]}"
APP_DIR="${2:-/opt/feedbacker}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_SRC="${SCRIPT_DIR}/docker-compose.prod.yml"
COMPOSE_FILE="${APP_DIR}/docker-compose.yml"
ENV_FILE="${APP_DIR}/.env"

command -v docker >/dev/null 2>&1 || { echo "未找到 docker 命令" >&2; exit 1; }
[ -f "$COMPOSE_SRC" ] || { echo "未找到 ${COMPOSE_SRC}" >&2; exit 1; }
[ -f "$ENV_FILE" ] || { echo "未找到 ${ENV_FILE}，请先在服务器上创建该文件" >&2; exit 1; }

# 文件写入权限：优先当前用户，其次免密 sudo
if [ -w "$APP_DIR" ]; then
  SUDO=()
elif sudo -n true 2>/dev/null; then
  SUDO=(sudo -n)
else
  echo "用户 $(id -un) 对 ${APP_DIR} 无写权限，且未配置免密 sudo" >&2
  exit 1
fi

# Docker 访问权限：优先当前用户（docker 组），其次免密 sudo
if docker info >/dev/null 2>&1; then
  DOCKER=()
elif sudo -n docker info >/dev/null 2>&1; then
  DOCKER=(sudo -n)
else
  echo "用户 $(id -un) 无法访问 Docker，请将其加入 docker 组后重新登录" >&2
  exit 1
fi

if grep -qE '^[[:space:]]*CORS_ORIGINS[[:space:]]*=[[:space:]]*\*' "$ENV_FILE"; then
  echo "警告: ${ENV_FILE} 中 CORS_ORIGINS=*，生产环境建议收敛到实际域名" >&2
fi

echo "==> 写入 ${COMPOSE_FILE}"
"${SUDO[@]}" install -m 0644 "$COMPOSE_SRC" "$COMPOSE_FILE"

cd "$APP_DIR"

echo "==> 拉取镜像 ${IMAGE}"
if ! "${DOCKER[@]}" env APP_IMAGE="$IMAGE" docker compose --env-file .env pull app; then
  echo "拉取镜像失败：若 GHCR 包为私有，请确认部署用户在 docker 组内（docker login 凭据按用户家目录保存），或把包设为 public" >&2
  exit 1
fi

echo "==> 启动容器"
"${DOCKER[@]}" env APP_IMAGE="$IMAGE" docker compose --env-file .env up -d --remove-orphans

PORT="$(sed -n 's/^[[:space:]]*PORT[[:space:]]*=[[:space:]]*\([0-9]\{1,\}\).*/\1/p' .env | tail -n1)"
PORT="${PORT:-9090}"

http_ok() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsS --max-time 3 "$1" >/dev/null 2>&1
  else
    wget -qO- --timeout=3 "$1" >/dev/null 2>&1
  fi
}

echo "==> 健康检查 http://127.0.0.1:${PORT}/healthz"
healthy=0
for _ in $(seq 1 30); do
  if http_ok "http://127.0.0.1:${PORT}/healthz"; then
    healthy=1
    break
  fi
  sleep 2
done

if [ "$healthy" -ne 1 ]; then
  echo "健康检查失败，最近日志：" >&2
  "${DOCKER[@]}" docker compose logs --tail=50 app >&2 || true
  exit 1
fi

"${DOCKER[@]}" docker compose ps
echo "==> 后端部署完成：${IMAGE}"
