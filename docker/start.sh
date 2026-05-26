#!/bin/bash

DEPLOY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UPDATE_SCRIPT="${DEPLOY_DIR}/update.sh"

echo "=========================================="
echo "          CI/CD 部署辅助脚本启动器          "
echo "=========================================="
echo "本目录应包含: docker-compose.yml、update.sh、.env"
echo "首次部署请将 compose 中镜像 tag 设为当前 Release（如 v0.0.1）"
echo "=========================================="

if [ ! -f "$UPDATE_SCRIPT" ]; then
    echo "错误: 未找到 update.sh"
    exit 1
fi

if [ ! -f "${DEPLOY_DIR}/docker-compose.yml" ] && [ ! -f "${DEPLOY_DIR}/docker-compose.yaml" ]; then
    echo "错误: 未找到 docker-compose.yml / docker-compose.yaml"
    exit 1
fi

chmod +x "$UPDATE_SCRIPT"

echo "正在执行首次同步测试..."
bash "$UPDATE_SCRIPT"
echo "完成，请查看 ${DEPLOY_DIR}/update.log"

CRON_JOB="*/10 * * * * /bin/bash ${UPDATE_SCRIPT}"
if (crontab -l 2>/dev/null | grep -F "$UPDATE_SCRIPT") >/dev/null 2>&1; then
    echo "提示: 定时任务（每 10 分钟）已存在，跳过。"
else
    echo "正在写入 crontab（每 10 分钟轮询）..."
    (crontab -l 2>/dev/null; echo "$CRON_JOB") | crontab -
    echo "定时任务已添加。"
fi

echo "=========================================="
echo "查看日志: tail -f ${DEPLOY_DIR}/update.log"
echo "=========================================="
