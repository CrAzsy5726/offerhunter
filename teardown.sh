#!/bin/bash
# ============================================================
# 彻底删除旧版 Offer-Hunter（Go 版）
# 在服务器上以 root 执行: bash teardown.sh
# ============================================================
set -e

echo "=== 正在彻底删除旧版 Offer-Hunter ==="

# 1. 停止并禁用 systemd 服务
echo "[1/5] 停止 systemd 服务..."
systemctl stop offer-hunter 2>/dev/null || true
systemctl disable offer-hunter 2>/dev/null || true
rm -f /etc/systemd/system/offer-hunter.service
systemctl daemon-reload

# 2. 删除程序文件
echo "[2/5] 删除程序文件..."
rm -rf /opt/offer-hunter

# 3. 删除前端文件
echo "[3/5] 删除前端文件..."
rm -rf /var/www/offer-hunter

# 4. 清理 Nginx 配置
echo "[4/5] 清理 Nginx 配置..."
rm -f /etc/nginx/conf.d/offer-hunter.conf 2>/dev/null || true
rm -f /etc/nginx/sites-available/offer-hunter 2>/dev/null || true
rm -f /etc/nginx/sites-enabled/offer-hunter 2>/dev/null || true
# 重载 nginx（如果还在运行）
nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null || true

# 5. 验证清理结果
echo "[5/5] 验证..."
if [ -d "/opt/offer-hunter" ]; then
    echo "  ❌ /opt/offer-hunter 仍存在"
else
    echo "  ✅ /opt/offer-hunter 已删除"
fi
if systemctl is-active offer-hunter &>/dev/null; then
    echo "  ❌ offer-hunter 服务仍在运行"
else
    echo "  ✅ offer-hunter 服务已停止"
fi

echo ""
echo "=== 旧版已彻底删除 ==="
echo "可以开始部署新版: bash setup.sh"
