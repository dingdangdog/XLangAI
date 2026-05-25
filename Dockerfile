# XlangAI 统一镜像：Nuxt 运营后台 (manager) + Go API (server)
# 构建：docker build -t xlangai:latest .
# 运行：见 docker/docker-compose.yml
#
# 环境变量分组说明见 README.md「环境变量」章节。

# ── Go API ─────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS go-builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/xlangai-server .

# ── Nuxt manager ───────────────────────────────────────────────────────
FROM node:22.21.1-alpine3.22 AS manager-builder

WORKDIR /app

RUN apk add --no-cache openssl libc6-compat && corepack enable

COPY manager/package.json manager/pnpm-lock.yaml manager/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --ignore-scripts --registry=https://registry.npmmirror.com

COPY manager/ ./
RUN pnpm run postinstall && pnpm build

# ── 运行镜像 ───────────────────────────────────────────────────────────
FROM node:22.21.1-alpine3.22 AS runner

LABEL org.opencontainers.image.title="xlangai"
LABEL org.opencontainers.image.description="XlangAI manager (Nuxt) + API (Go) unified image"

RUN apk add --no-cache \
  openssl \
  libc6-compat \
  ca-certificates \
  tzdata \
  ffmpeg \
  tini \
  su-exec \
  wget \
  && addgroup --system --gid 1001 xlangai \
  && adduser --system --uid 1001 --ingroup xlangai xlangai

WORKDIR /app

COPY --from=go-builder --chown=xlangai:xlangai /out/xlangai-server /app/bin/xlangai-server
COPY --from=manager-builder --chown=xlangai:xlangai /app/.output /app/manager/.output
COPY --from=manager-builder --chown=xlangai:xlangai /app/prisma/generated /app/manager/prisma/generated
COPY --from=manager-builder --chown=xlangai:xlangai /app/prisma/migrations /app/manager/prisma/migrations
COPY --from=manager-builder --chown=xlangai:xlangai /app/storage/audio /app/bootstrap-storage/audio

COPY docker/entrypoint.sh /app/entrypoint.sh

RUN sed -i 's/\r$//' /app/entrypoint.sh \
  && chmod +x /app/entrypoint.sh \
  && mkdir -p /app/storage/audio /app/storage/avatars /app/bootstrap-storage/audio \
  && chown -R xlangai:xlangai /app/storage /app/bootstrap-storage

USER root

# ── 1. 容器运行时 ───────────────────────────────────────────────────────
ENV TZ="Asia/Shanghai"
ENV APP_USER="xlangai"
ENV NODE_ENV="production"

# ── 2. Nuxt / Nitro（只需 PORT；host 默认 0.0.0.0）──────────────────────
ENV PORT=3312

# ── 3. Nuxt runtimeConfig — manager 私有（NUXT_MANAGER_*）──────────────
ENV NUXT_MANAGER_DATABASE_AUTO_MIGRATE="true"
ENV NUXT_MANAGER_AUTH_SECRET="change-me-in-production"
ENV NUXT_MANAGER_AUTO_SEED="true"
ENV NUXT_MANAGER_TEST_ACCOUNT_SEED="false"
ENV NUXT_MANAGER_ADMIN_USERNAME="admin"
ENV NUXT_MANAGER_ADMIN_PASSWORD="123456"
ENV NUXT_MANAGER_ADMIN_NICKNAME="管理员"
ENV NUXT_MANAGER_ADMIN_SEED="true"

# ── 4. Nuxt runtimeConfig — 公开（NUXT_PUBLIC_*）──────────────────────
ENV NUXT_PUBLIC_OFFICIAL_HOME_URL="https://xlangai.com"

# ── 5. 共享基础设施（Prisma / 跨服务路径，非 NUXT_ 前缀）────────────────
ENV DATABASE_URL="postgresql://postgres:postgres@host.docker.internal:5432/xlangai?schema=public"
ENV AUDIO_DIR="/app/storage/audio"
ENV AVATAR_DIR="/app/storage/avatars"
ENV BUNDLED_AUDIO_DIR="/app/bootstrap-storage/audio"

# ── 6. Go API（XLANGAI_* / JWT_*）──────────────────────────────────────
ENV XLANGAI_SERVER_PORT=8080
ENV JWT_SECRET="change-me-in-production"

EXPOSE 3312 8080

VOLUME ["/app/storage"]

ENTRYPOINT ["/sbin/tini", "--", "/app/entrypoint.sh"]
