#!/bin/sh
set -eu

shutdown() {
  if [ -f /tmp/go.pid ]; then
    kill -TERM "$(cat /tmp/go.pid)" 2>/dev/null || true
    wait "$(cat /tmp/go.pid)" 2>/dev/null || true
  fi
  if [ -f /tmp/node.pid ]; then
    kill -TERM "$(cat /tmp/node.pid)" 2>/dev/null || true
    wait "$(cat /tmp/node.pid)" 2>/dev/null || true
  fi
}

trap shutdown INT TERM

run_as_app() {
  if [ "$(id -u)" = "0" ]; then
    su-exec "$APP_USER" "$@"
  else
    "$@"
  fi
}

prepare_runtime_dirs() {
  mkdir -p "$AUDIO_DIR" "$AVATAR_DIR"
  if [ "$(id -u)" = "0" ]; then
    chown -R "$APP_USER:$APP_USER" /app/storage 2>/dev/null || true
  fi
}

seed_bundled_preview_audio() {
  src="/app/bootstrap-storage/audio"
  dst="$AUDIO_DIR"
  if [ ! -d "$src" ]; then
    return
  fi
  mkdir -p "$dst"
  copied=0
  for file in "$src"/*; do
    if [ ! -f "$file" ]; then
      continue
    fi
    name="$(basename "$file")"
    if [ ! -f "$dst/$name" ]; then
      cp "$file" "$dst/$name"
      copied=$((copied + 1))
    fi
  done
  if [ "$copied" -gt 0 ]; then
    echo "[entrypoint] 已初始化内置试听音频 ${copied} 个到 ${dst}"
  fi
  if [ "$(id -u)" = "0" ]; then
    chown -R "$APP_USER:$APP_USER" "$dst" 2>/dev/null || true
  fi
}

prepare_runtime_dirs
seed_bundled_preview_audio

cd /app/manager

(
  run_as_app node .output/server/index.mjs
) &
echo $! >/tmp/node.pid

echo "[entrypoint] 等待运营后台就绪 (:${PORT})，数据库迁移在此阶段完成…"
ready=0
i=0
while [ "$i" -lt 120 ]; do
  if wget -q -O /dev/null "http://127.0.0.1:${PORT}/" 2>/dev/null; then
    ready=1
    break
  fi
  if ! kill -0 "$(cat /tmp/node.pid)" 2>/dev/null; then
    echo "[entrypoint] 运营后台进程已退出，请检查日志"
    exit 1
  fi
  i=$((i + 1))
  sleep 1
done

if [ "$ready" -ne 1 ]; then
  echo "[entrypoint] 运营后台在 120s 内未就绪"
  shutdown
  exit 1
fi

echo "[entrypoint] 启动 Go API (:${XLANGAI_SERVER_PORT})"
(
  run_as_app /app/bin/xlangai-server
) &
echo $! >/tmp/go.pid

echo "[entrypoint] 服务已启动 — manager :${PORT}，api :${XLANGAI_SERVER_PORT}"

wait "$(cat /tmp/node.pid)"
