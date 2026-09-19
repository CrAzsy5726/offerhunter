#!/bin/bash
# ============================================================
# Offer 捕手（Go 版）— 卸载脚本
# 在 Linux 服务器上以 root 执行: bash undeploy.sh
# ============================================================
set -euo pipefail

APP_DIR="/opt/offer-hunter"
SERVICE_NAME="offer-hunter"

echo "=== 卸载 Offer 捕手（Go 版）==="

# ---- Step 1: 停止并禁用 systemd 服务 ----
echo "[1/4] 停止 systemd 服务..."
systemctl stop "${SERVICE_NAME}" 2>/dev/null || true
systemctl disable "${SERVICE_NAME}" 2>/dev/null || true
rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
systemctl daemon-reload
echo "  服务已停止并移除"

# ---- Step 2: 删除程序文件 ----
echo "[2/4] 删除程序文件..."
rm -rf "${APP_DIR}"
echo "  已删除 ${APP_DIR}"

# ---- Step 3: 清理 Nginx 配置（如果存在） ----
echo "[3/4] 清理 Nginx 配置..."
rm -f /etc/nginx/conf.d/offer-hunter.conf 2>/dev/null || true
rm -f /etc/nginx/sites-available/offer-hunter 2>/dev/null || true
rm -f /etc/nginx/sites-enabled/offer-hunter 2>/dev/null || true
nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null || true
echo "  Nginx 配置已清理（如有）"

# ---- Step 4: 验证 ----
echo "[4/4] 验证..."
if [ -d "${APP_DIR}" ]; then
    echo "  ✗ ${APP_DIR} 仍存在"
else
    echo "  ✓ ${APP_DIR} 已删除"
fi
if systemctl is-active "${SERVICE_NAME}" &>/dev/null; then
    echo "  ✗ ${SERVICE_NAME} 仍在运行"
else
    echo "  ✓ ${SERVICE_NAME} 已停止"
fi

echo ""
echo "=== 卸载完成 ==="
