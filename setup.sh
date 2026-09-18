#!/bin/bash
# ============================================================
# Offer 捕手 LangChain Edition — 服务器一键部署脚本
# 在 Linux 服务器上以 root 执行: bash setup.sh
# ============================================================
set -e

echo "=== Offer 捕手 LangChain Edition 部署开始 ==="

APP_DIR="/opt/offer-hunter"
SERVER_HOST="${SERVER_HOST:-0.0.0.0}"
SERVER_PORT="${SERVER_PORT:-8080}"

# ---- Step 0: 环境检查 ----
echo "[0/7] 检查系统环境..."
OS=$(cat /etc/os-release | grep '^ID=' | cut -d= -f2 | tr -d '"')
echo "  系统: $OS"

if ! command -v python3 &>/dev/null; then
    echo "  安装 Python3..."
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        apt update && apt install -y python3 python3-pip
    elif [ "$OS" = "centos" ] || [ "$OS" = "rhel" ] || [ "$OS" = "tencentos" ]; then
        yum install -y python3 python3-pip
    fi
fi
echo "  Python: $(python3 --version)"

# ---- Step 1: 停止旧服务（如果存在） ----
echo "[1/7] 停止旧服务..."
systemctl stop offer-hunter 2>/dev/null || true
systemctl stop offer-hunter-go 2>/dev/null || true

# ---- Step 2: 创建目录 ----
echo "[2/7] 创建目录..."
mkdir -p "$APP_DIR"

# ---- Step 3: 部署应用文件 ----
echo "[3/7] 部署应用文件..."
# 复制所有 Python 源码
cp -r agent handlers prompts services "$APP_DIR/"
cp main.py models.py config.py requirements.txt "$APP_DIR/"
# 前端
mkdir -p "$APP_DIR/static"
cp static/index.html "$APP_DIR/static/"
# 环境变量（部署用的 .env 单独管理，这里不覆盖已有文件）
if [ ! -f "$APP_DIR/.env" ]; then
    cp .env "$APP_DIR/.env"
    echo "  已创建 .env（请检查 API Key 是否正确）"
else
    echo "  .env 已存在，跳过覆盖（如需更新请手动修改 $APP_DIR/.env）"
fi

# ---- Step 4: 安装依赖 ----
echo "[4/7] 安装 Python 依赖..."
cd "$APP_DIR"
pip3 install -r requirements.txt -q
echo "  依赖安装完成"

# ---- Step 5: 创建 systemd 服务 ----
echo "[5/7] 配置 systemd 服务..."
cat > /etc/systemd/system/offer-hunter.service << SVC
[Unit]
Description=Offer Hunter LangChain Edition
After=network.target

[Service]
Type=simple
WorkingDirectory=$APP_DIR
ExecStart=python3 -m uvicorn main:app --host ${SERVER_HOST} --port ${SERVER_PORT}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SVC

systemctl daemon-reload
systemctl enable offer-hunter
systemctl restart offer-hunter
echo "  后端服务已启动（端口 ${SERVER_PORT}）"

# ---- Step 6: 配置 Nginx ----
echo "[6/7] 配置 Nginx..."
if ! command -v nginx &>/dev/null; then
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        apt update && apt install -y nginx
    elif [ "$OS" = "centos" ] || [ "$OS" = "rhel" ] || [ "$OS" = "tencentos" ]; then
        yum install -y nginx
    fi
fi

NGINX_CONF="/etc/nginx/conf.d/offer-hunter.conf"
if [ -d "/etc/nginx/sites-available" ]; then
    NGINX_CONF="/etc/nginx/sites-available/offer-hunter"
fi

cat > "$NGINX_CONF" << NGX
server {
    listen 80;
    server_name _;

    # 所有请求都代理到 FastAPI（它自己 serve 前端 + API）
    location / {
        proxy_pass http://127.0.0.1:${SERVER_PORT};
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_read_timeout 120s;
        proxy_buffering off;
    }
}
NGX

if [ -d "/etc/nginx/sites-enabled" ]; then
    ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/offer-hunter
fi

rm -f /etc/nginx/sites-enabled/default 2>/dev/null || true
rm -f /etc/nginx/conf.d/default.conf 2>/dev/null || true

nginx -t && systemctl restart nginx
echo "  Nginx 配置完成"

# ---- Step 7: 验证 ----
echo "[7/7] 验证..."
sleep 2
echo "  后端健康检查:"
curl -s "http://127.0.0.1:${SERVER_PORT}/health" && echo ""
echo "  前端检查:"
curl -s -o /dev/null -w "  HTTP %{http_code}" http://127.0.0.1/ && echo ""

echo ""
echo "=== 部署完成 ==="
echo ""
echo "后续操作:"
echo "1. 检查 .env 配置:  vim $APP_DIR/.env"
echo "2. 查看后端日志:   journalctl -u offer-hunter -f"
echo "3. 重启服务:       systemctl restart offer-hunter"
echo "4. 防火墙放行 80:  firewall-cmd --add-port=80/tcp --permanent"
