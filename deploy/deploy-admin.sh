#!/usr/bin/env bash
# 在服务器上执行：把云端构建好的管理后台 dist 覆盖到站点目录。
# 由 .github/workflows/admin.yml 上传到 /tmp 后调用，压缩包从 stdin 传入。
#
# 用法: tar -czf - -C admin/dist . | deploy-admin.sh <dist_dir>
set -euo pipefail

DIST_DIR="${1:?用法: deploy-admin.sh <dist_dir>}"
PARENT_DIR="$(dirname "$DIST_DIR")"

if [ -d "$PARENT_DIR" ] && [ -w "$PARENT_DIR" ]; then
  SUDO=()
elif sudo -n true 2>/dev/null; then
  SUDO=(sudo -n)
else
  echo "用户 $(id -un) 对 ${PARENT_DIR} 无写权限，且未配置免密 sudo" >&2
  exit 1
fi

"${SUDO[@]}" mkdir -p "$PARENT_DIR"

TMP_DIR="${PARENT_DIR}/.dist-incoming.$$"
"${SUDO[@]}" rm -rf "$TMP_DIR"
"${SUDO[@]}" mkdir -p "$TMP_DIR"

echo "==> 解压到 ${TMP_DIR}"
"${SUDO[@]}" tar -xzf - -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/index.html" ]; then
  echo "构建产物缺少 index.html，终止发布" >&2
  "${SUDO[@]}" rm -rf "$TMP_DIR"
  exit 1
fi

"${SUDO[@]}" chmod -R a+rX "$TMP_DIR"

OLD_DIR="${DIST_DIR}.old"
echo "==> 替换 ${DIST_DIR}"
"${SUDO[@]}" rm -rf "$OLD_DIR"
if [ -d "$DIST_DIR" ]; then
  "${SUDO[@]}" mv "$DIST_DIR" "$OLD_DIR"
fi
"${SUDO[@]}" mv "$TMP_DIR" "$DIST_DIR"
"${SUDO[@]}" rm -rf "$OLD_DIR"

echo "==> 发布完成"
"${SUDO[@]}" ls -la "$DIST_DIR"
