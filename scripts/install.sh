#!/usr/bin/env bash
# Modernchat SSO 最小化生产安装脚本：解包发行物 → 初始化数据库 → 安装 systemd 单元。
# 用法：sudo ./install.sh modernchat-sso-<version>-linux-amd64.tar.gz <port>
# 不会写入任何凭据：JWT_SECRET / DB_DSN / COOKIE_SECURE 由运维在 .env 中填写。
set -euo pipefail

ARCHIVE="${1:?用法: install.sh <release.tar.gz> <port>}"
PORT="${2:?用法: install.sh <release.tar.gz> <port>}"
ROOT="${SSO_ROOT:-/opt/modernchat-sso}"
MYSQL_BIN="${MYSQL_BIN:-mysql}"

if [[ ! -f "$ARCHIVE" ]]; then
  echo "发行包不存在: $ARCHIVE" >&2
  exit 1
fi
if [[ ! "$PORT" =~ ^[0-9]+$ ]]; then
  echo "端口必须是数字: $PORT" >&2
  exit 1
fi

command -v "$MYSQL_BIN" >/dev/null || { echo "未找到 mysql 客户端: $MYSQL_BIN" >&2; exit 1; }

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
tar -xzf "$ARCHIVE" -C "$TMP"

mkdir -p "$ROOT/bin" "$ROOT/server" "$ROOT/web" "$ROOT/deploy"
cp "$TMP/bin/modernchat-sso-server" "$ROOT/bin/"
chmod +x "$ROOT/bin/modernchat-sso-server"
rm -rf "$ROOT/web/dist"
cp -r "$TMP/web/dist" "$ROOT/web/dist"
rm -rf "$ROOT/server/migrations"
cp -r "$TMP/server/migrations" "$ROOT/server/migrations"
cp -r "$TMP/deploy/." "$ROOT/deploy/"

if [[ ! -f "$ROOT/server/.env" ]]; then
  cp "$TMP/server/.env.example" "$ROOT/server/.env"
  echo "已生成 $ROOT/server/.env，请先填写 JWT_SECRET / DB_DSN / PUBLIC_URL / COOKIE_SECURE 后再启动服务。"
fi

if [[ -n "${MYSQL_DATABASE:-}" && -n "${MYSQL_USER:-}" ]]; then
  echo "应用 server/migrations/0001_init.up.sql ..."
  MYSQL_PWD="${MYSQL_PASSWORD:-}" "$MYSQL_BIN" \
    -h "${MYSQL_HOST:-127.0.0.1}" -P "${MYSQL_PORT:-3306}" -u "$MYSQL_USER" "$MYSQL_DATABASE" \
    < "$ROOT/server/migrations/0001_init.up.sql"
else
  echo "未设置 MYSQL_DATABASE/MYSQL_USER，跳过自动迁移；请手工执行 server/migrations/0001_init.up.sql。"
fi

UNIT="modernchat-sso@${PORT}.service"
install -m 0644 "$ROOT/deploy/modernchat-sso@.service" "/etc/systemd/system/$UNIT"
systemctl daemon-reload
systemctl enable --now "$UNIT"
echo "SSO 已安装并启动: systemctl status $UNIT"
