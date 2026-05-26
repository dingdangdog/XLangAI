#!/bin/bash
set -euo pipefail

# 与 docker-compose.yml / docker-compose.yaml 放在同一目录即可，无需额外配置
DEPLOY_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="dingdangdog/XLangAI"
LOG_FILE="${DEPLOY_DIR}/update.log"

if [ -f "${DEPLOY_DIR}/docker-compose.yml" ]; then
    COMPOSE_FILE="${DEPLOY_DIR}/docker-compose.yml"
elif [ -f "${DEPLOY_DIR}/docker-compose.yaml" ]; then
    COMPOSE_FILE="${DEPLOY_DIR}/docker-compose.yaml"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 错误: 未找到 docker-compose.yml 或 docker-compose.yaml" >> "$LOG_FILE"
    exit 1
fi

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$LOG_FILE"
}

fetch_latest_release_tag() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
        | head -1
}

get_image_line() {
    grep -E '^[[:space:]]*image:' "$COMPOSE_FILE" | head -1 | awk '{print $2}'
}

set_image_tag() {
    local image_name="$1"
    local new_tag="$2"
    sed -i -E "s|^([[:space:]]*image:[[:space:]]*)${image_name}:[^[:space:]]+|\1${image_name}:${new_tag}|" "$COMPOSE_FILE"
}

log "开始检查版本更新..."

LATEST_VERSION="$(fetch_latest_release_tag)"
if [ -z "$LATEST_VERSION" ]; then
    log "错误: 无法获取最新 Release，请检查仓库 ${REPO} 是否已发布 Release 及网络。"
    exit 1
fi

if [[ ! "$LATEST_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log "错误: 最新 Release 标签格式无效: ${LATEST_VERSION}（期望 vX.Y.Z）"
    exit 1
fi

IMAGE_LINE="$(get_image_line)"
if [ -z "$IMAGE_LINE" ]; then
    log "错误: 在 ${COMPOSE_FILE} 中未找到 image 行"
    exit 1
fi

IMAGE_NAME="${IMAGE_LINE%%:*}"
CURRENT_VERSION="${IMAGE_LINE##*:}"

if [ "$LATEST_VERSION" = "$CURRENT_VERSION" ]; then
    log "无更新: 当前版本 (${CURRENT_VERSION}) 已是最新。"
    exit 0
fi

log "发现新版本: ${LATEST_VERSION}（当前: ${CURRENT_VERSION}），开始更新..."
set_image_tag "$IMAGE_NAME" "$LATEST_VERSION"

cd "$DEPLOY_DIR"
log "正在拉取镜像 ${IMAGE_NAME}:${LATEST_VERSION} ..."
docker compose pull >> "$LOG_FILE" 2>&1

log "正在重启容器..."
docker compose up -d >> "$LOG_FILE" 2>&1

docker image prune -f >> /dev/null 2>&1
log "成功: 已升级至 ${LATEST_VERSION}"
