#!/bin/bash
# ============================================================
# Offer 捕手（Go 版）— 服务器部署脚本
# 在 Linux 服务器上以 root 执行: bash deploy.sh
# ============================================================
set -euo pipefail

APP_NAME="offer-hunter"
APP_DIR="/opt/offer-hunter"
SERVICE_NAME="offer-hunter"
SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
SERVER_PORT="${SERVER_PORT:-8080}"

echo "=== Offer 捕手（Go 版）部署开始 ==="

# ---- Step 1: 环境检查 ----
echo "[1/6] 检查环境..."
if ! command -v go &>/dev/null; then
    echo "  ✗ 未检测到 Go。请先安装 Go 1.26+，或在本机交叉编译后上传二进制到 ${APP_DIR}/offer-hunter"
    echo "    交叉编译命令：CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o offer-hunter ./cmd/server"
    exit 1
fi
echo "  Go: $(go version)"

# ---- Step 2: 创建目录 ----
echo "[2/6] 创建目录 ${APP_DIR}..."
mkdir -p "${APP_DIR}"

# ---- Step 3: 编译 Linux 二进制 ----
echo "[3/6] 编译 Linux 二进制..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${APP_DIR}/${APP_NAME}" ./cmd/server
echo "  二进制已生成: ${APP_DIR}/${APP_NAME}"

# ---- Step 4: 复制前端静态文件 ----
echo "[4/6] 部署前端静态文件..."
rm -rf "${APP_DIR}/static"
cp -r static "${APP_DIR}/static"
echo "  已复制 static/ 到 ${APP_DIR}/static"

# ---- Step 5: 处理 .env ----
echo "[5/6] 检查 .env 配置..."
if [ ! -f "${APP_DIR}/.env" ]; then
    echo "  ✗ 未找到 ${APP_DIR}/.env"
    echo "    请手动创建该文件（参考仓库根目录 .env.example），填入真实 API Key："
    echo "      cp .env.example ${APP_DIR}/.env  # 然后编辑填入真实密钥"
    echo "    注意：不要把真实密钥提交到 git。"
    exit 1
fi
echo "  .env 已存在（不会打印其内容）"

# ---- Step 6: 写 systemd 服务并启动 ----
echo "[6/6] 配置 systemd 服务..."
cat > "/etc/systemd/system/${SERVICE_NAME}.service" << SVC
[Unit]
Description=Offer Hunter (Go) AI job assistant
After=network.target

[Service]
Type=simple
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/${APP_NAME}
Restart=always
RestartSec=5
Environment=SERVER_HOST=${SERVER_HOST}
Environment=SERVER_PORT=${SERVER_PORT}

[Install]
WantedBy=multi-user.target
SVC

systemctl daemon-reload
systemctl enable "${SERVICE_NAME}"
systemctl restart "${SERVICE_NAME}"
echo "  服务已启动"

# ---- 健康检查 ----
sleep 2
echo "  健康检查:"
curl -s "http://127.0.0.1:${SERVER_PORT}/health" && echo ""

echo ""
echo "=== 部署完成 ==="
echo ""
echo "后续操作:"
echo "1. 查看日志:   journalctl -u ${SERVICE_NAME} -f"
echo "2. 重启服务:   systemctl restart ${SERVICE_NAME}"
echo "3. 前端访问:   http://<服务器IP>:${SERVER_PORT}"
echo ""
echo "如需域名 + SSL，可参考下方 Nginx 反代配置（将 80 端口代理到 ${SERVER_PORT}）："
echo "  location / {"
echo "      proxy_pass http://127.0.0.1:${SERVER_PORT};"
echo "      proxy_set_header Host \$host;"
echo "      proxy_set_header X-Real-IP \$remote_addr;"
echo "      proxy_read_timeout 120s;"
echo "      proxy_buffering off;"
echo "  }"
