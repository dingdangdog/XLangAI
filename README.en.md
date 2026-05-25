<div align="center">

<img src="logo/logo_256x256.png" alt="XLangAI" width="128" height="128" />
</div>

<h1 align="center">XLangAI Servers</h1>

<p align="center">
  <img alt="version" src="https://img.shields.io/badge/version-v0.0.1-blue" />
  <img alt="license" src="https://img.shields.io/badge/license-MIT-yellow.svg" />
  <img alt="go" src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" />
  <img alt="nuxt" src="https://img.shields.io/badge/Nuxt-4-00DC82?logo=nuxt.js&logoColor=white" />
  <img alt="postgresql" src="https://img.shields.io/badge/PostgreSQL-15+-4169E1?logo=postgresql&logoColor=white" />
  <img alt="docker" src="https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white" />
</p>

<p align="center">
  <strong>XLangAI · API Server and Admin Console</strong> — a complete backend solution for multilingual AI speaking practice
</p>

<p align="center">
  <strong>Language</strong>: <a href="README.md">简体中文</a> · English
</p>

<p align="center">
  <a href="#features">Features</a> ·
  <a href="#architecture">Architecture</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="#local-development">Local Development</a> ·
  <a href="#environment-variables">Environment Variables</a> ·
  <a href="#api-overview">API Overview</a>
</p>

---

## About

**XLangAI** means *all language AI*. This directory contains the independently deployable **API server and admin console**. The Go API provides authentication, conversation, speech, billing, and media endpoints for the client app. The Nuxt admin console manages languages, AI providers, membership tiers, prompt templates, and system settings.

| Directory              | Description                                                                     | Stack                                  |
| ---------------------- | ------------------------------------------------------------------------------- | -------------------------------------- |
| [`server/`](server/)   | Go API for the mobile app: auth, conversation, speech, billing, media, and more | Go 1.26 · Gin · GORM · PostgreSQL      |
| [`manager/`](manager/) | Admin console: configuration center, users, conversations, and backups          | Nuxt 4 · Vue 3 · Prisma 7 · PostgreSQL |

> This directory does **not** include the Flutter client or the website. `client` and `home` live in other directories of the main XLangAI repository. Deploying `servers/` is enough to support the app API and admin console.

### Reading Guide

- **Deploy only**: start from [Quick Start](#quick-start), then configure the database, secrets, and the first admin account.
- **Local integration**: start `manager` first so Prisma migrations and seed data are applied, then start `server`.
- **Troubleshooting configuration**: check [Environment Variables](#environment-variables) and [Production Checklist](#production-checklist).

**License**: [MIT](LICENSE)  
**Authors**: GT · DingDangDog

---

## Features

### Go API (`server`)

- **Users and authentication**: sign-up/sign-in, SMS verification codes, Google / Apple login, JWT sessions
- **Multilingual conversations**: create conversations, text/voice chat, message history, translation APIs
- **Speech pipeline**: STT (configurable providers) -> LLM -> TTS; ffmpeg loudness normalization included
- **Membership and billing**: membership tiers and in-app purchase verification for Apple / Google Play
- **Media and storage**: avatar upload, audio previews, object-storage presigned URLs (R2 / S3 / Qiniu / Alibaba Cloud OSS / local)
- **Observability**: Gin access logs and unified `[api]` error logs; see [`design/server-logging.md`](../design/server-logging.md)

### Admin Console (`manager`)

- **Service configuration**: multiple providers for LLM / STT / TTS / translation / object storage
- **Content and operations**: languages, voices, prompt templates, membership tiers, and system settings
- **User domain**: user list, conversations, messages, usage statistics, and account-deletion backups
- **Operations**: database migrations at Nitro startup, seed data, backup export, and server-store sync

### Deployment

- **Single image**: root `Dockerfile` builds Nuxt + Go into one runtime image
- **Docker Compose**: one-command startup with a persistent `storage` volume
- **Split deployment**: `server` and `manager` can run separately while sharing the same PostgreSQL database

---

## Architecture

```mermaid
flowchart TB
  subgraph clients [Clients]
    App[Flutter App]
    Browser[Admin Browser]
  end

  subgraph servers [servers directory]
    API[Go API :8080]
    MGR[Nuxt Manager :3312]
    DB[(PostgreSQL)]
    Vol[/storage audio avatars/]
  end

  subgraph external [Configurable external services]
    LLM[LLM Providers]
    STT[STT Providers]
    TTS[TTS Providers]
    OSS[Object Storage]
  end

  App --> API
  Browser --> MGR
  MGR --> DB
  API --> DB
  MGR -.config.-> API
  API --> LLM & STT & TTS
  API --> OSS
  API --> Vol
  MGR --> Vol
```

**Startup order for the unified Docker image**: `entrypoint.sh` starts Nuxt first so Prisma migrations and seed data run, waits until the admin console is ready, then starts the Go API.

---

## Directory Structure

```text
servers/
├── server/                 # Go API source
│   ├── main.go
│   ├── config/             # Environment and config loading
│   └── internal/           # handler · router · ai · billing · media …
├── manager/                # Nuxt admin console
│   ├── app/                # Pages and components
│   ├── server/             # Nitro APIs and startup plugins, including DB migration
│   ├── prisma/             # Schema and migrations
│   └── storage/audio/      # Bundled audio previews copied into the image
├── docker/
│   ├── docker-compose.yml  # Recommended deployment entry point
│   └── entrypoint.sh       # Two-process startup script
├── logo/                   # Brand icons
├── Dockerfile              # Unified image build
├── LICENSE                 # MIT
├── README.md               # Default Simplified Chinese documentation
└── README.en.md            # English documentation
```

---

## Requirements

| Scenario                | Dependencies                                                                                        |
| ----------------------- | --------------------------------------------------------------------------------------------------- |
| Docker deployment       | Docker 24+, Docker Compose v2, PostgreSQL 15+ on host or in a container                             |
| Local API development   | Go 1.26+, PostgreSQL, optional Redis                                                                |
| Local admin development | Node.js 22+, pnpm 11+, PostgreSQL                                                                   |
| Speech processing       | `ffmpeg`; it is included in the Docker image, but must be installed locally for STT/TTS transcoding |

---

## Quick Start

### 1. Prepare PostgreSQL

Create a PostgreSQL database on the host or in a separate container. The examples use `wlltalk` as the database name.

When a Windows or macOS container connects to PostgreSQL on the host, use `host.docker.internal` in `DATABASE_URL`. The Compose file already configures `extra_hosts`.

### 2. Configure Environment

Create `docker/.env` and at least override the database URL, JWT secrets, and first admin account:

```env
DATABASE_URL=postgresql://postgres:your-password@host.docker.internal:5432/wlltalk?schema=public
JWT_SECRET=replace-with-a-long-random-string
NUXT_MANAGER_AUTH_SECRET=replace-with-another-long-random-string

# First deployment: admin account. Disable bootstrap and remove the password variable after it is created.
NUXT_MANAGER_ADMIN_USERNAME=admin@example.com
NUXT_MANAGER_ADMIN_PASSWORD=your-strong-password
NUXT_MANAGER_ADMIN_NICKNAME=Admin
```

> `JWT_SECRET` signs app-user sessions. `NUXT_MANAGER_AUTH_SECRET` signs admin-console sessions. Use separate values in production.

### 3. Start

```bash
cd servers
mkdir -p docker/data/storage/audio docker/data/storage/avatars

docker compose -f docker/docker-compose.yml up -d --build
```

On the first startup, `manager` runs Prisma migrations and seed data first. After that, `entrypoint.sh` starts the Go API.

| Service          | URL                                                                              |
| ---------------- | -------------------------------------------------------------------------------- |
| Admin console    | [http://localhost:3312](http://localhost:3312)                                   |
| Go API           | [http://localhost:8080](http://localhost:8080)                                   |
| API health check | [http://localhost:8080/api/v1/languages](http://localhost:8080/api/v1/languages) |

### 4. Build or Export the Image Only

```bash
cd servers
docker build -t xlangai:latest .

# Offline distribution
docker save -o xlangai.tar xlangai:latest
docker load -i xlangai.tar
```

> `manager/Dockerfile` is a deprecated historical standalone image. Use the root `Dockerfile` and `docker/docker-compose.yml`.

---

## Local Development

`server` and `manager` share the same PostgreSQL instance. Prisma migrations are owned by `manager`.

Start `manager` first, confirm migrations and initial data are ready, then start `server`.

### Admin Console

```bash
cd servers/manager
cp env .env   # Update DATABASE_URL, admin account, and other values as needed.
pnpm install
pnpm run postinstall
pnpm dev      # Defaults to http://localhost:3312
```

Database migration in development:

```bash
pnpm db:migrate
# Or production-style:
npx prisma migrate deploy
```

### Go API

```bash
cd servers/server
cp env .env   # Configure DATABASE_URL, JWT_SECRET, and other values.
go mod download
go run .
# Defaults to http://localhost:8080
```

When running manager and the Go API locally, set `AUDIO_DIR=../server/storage/audio` in `manager/.env` so both services use the same directory. Admin preview playback is served by manager at `/api/admin/preview-audio/:filename`.

---

## Environment Variables

Variables are grouped by responsibility. **Nuxt app config** goes through `runtimeConfig` and must use the `NUXT_` / `NUXT_PUBLIC_` prefix. **Node/Nitro/Prisma/Go** use their own ecosystem names.

### 1. Container Runtime

| Variable | Default | Description |
| --- | --- | --- |
| `TZ` | `Asia/Shanghai` | Time zone |
| `APP_USER` | `xlangai` | User for entrypoint privilege drop |
| `NODE_ENV` | `production` | Node runtime mode |

### 2. Nuxt / Nitro (Node conventions, no `NUXT_` prefix)

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `3312` | Admin console listen port (Nitro default host is `0.0.0.0`; no `NITRO_HOST` needed) |

Compose host port mapping (external ports only):

| Variable | Default | Description |
| --- | --- | --- |
| `MANAGER_HOST_PORT` | `3312` | Maps to container `PORT` |
| `XLANGAI_SERVER_HOST_PORT` | `8080` | Maps to container `XLANGAI_SERVER_PORT` |

### 3. Nuxt runtimeConfig — manager private (`NUXT_MANAGER_*`)

| Variable | Default | Description |
| --- | --- | --- |
| `NUXT_MANAGER_DATABASE_AUTO_MIGRATE` | `true` | Run Prisma migrations on startup |
| `NUXT_MANAGER_AUTH_SECRET` | — | Admin-console JWT secret (**change in production**) |
| `NUXT_MANAGER_AUTO_SEED` | `true` | Business seed data |
| `NUXT_MANAGER_TEST_ACCOUNT_SEED` | `false` | Integration test account (`13800138000` / `123456`) |
| `NUXT_MANAGER_ADMIN_USERNAME` | — | First admin login ID |
| `NUXT_MANAGER_ADMIN_PASSWORD` | — | Plaintext password, bcrypt-hashed before storage |
| `NUXT_MANAGER_ADMIN_NICKNAME` | `Admin` | Admin display name |
| `NUXT_MANAGER_ADMIN_SEED` | `true` | Set to `false` to disable admin bootstrap |

### 4. Nuxt runtimeConfig — public (`NUXT_PUBLIC_*`)

| Variable | Default | Description |
| --- | --- | --- |
| `NUXT_PUBLIC_OFFICIAL_HOME_URL` | `https://xlangai.com` | Official site / server store URL |

### 5. Shared Infrastructure (Prisma / cross-service, no `NUXT_` prefix)

| Variable | Default | Description |
| --- | --- | --- |
| `DATABASE_URL` | — | PostgreSQL connection string shared by manager and Go (Prisma convention) |
| `DATABASE_BOOTSTRAP_URL` | — | Optional; connect to `postgres` DB first to create target DB |
| `DATABASE_SCHEMA` | `public` | Optional migration target schema |
| `AUDIO_DIR` | `/app/storage/audio` | Shared audio directory for manager and Go |
| `AVATAR_DIR` | `/app/storage/avatars` | User avatars (Go) |
| `BUNDLED_AUDIO_DIR` | `/app/bootstrap-storage/audio` | Bundled preview audio (Go fallback) |

### 6. Go API (`XLANGAI_*` / `JWT_*`)

| Variable | Default | Description |
| --- | --- | --- |
| `XLANGAI_SERVER_PORT` | `8080` | Go API listen port (separate from `PORT` in the unified container) |
| `JWT_SECRET` | — | App user JWT secret (**change in production**) |
| `REDIS_URL` | — | Optional Redis |
| `GIN_MODE` | `release` | Gin runtime mode |
| `XLANGAI_VERBOSE_LOGS` | `0` | Set to `1` for verbose error context |
| `XLANGAI_FFMPEG_PATH` | `ffmpeg` | ffmpeg executable path |
| `XLANGAI_STT_LANGUAGE_MODE` | — | `auto` / `target` |
| `XLANGAI_TTS_LOUDNESS_NORM` | enabled | TTS loudness normalization |
| `XLANGAI_APPLE_*` / `XLANGAI_GOOGLE_*` | — | IAP and OAuth verification |

For the full Docker guide, see [`design/DOCKER.md`](../design/DOCKER.md).

---

## First Admin Account

1. Set both `NUXT_MANAGER_ADMIN_USERNAME` and `NUXT_MANAGER_ADMIN_PASSWORD` in Compose or `docker/.env`.
2. Configure `NUXT_MANAGER_AUTH_SECRET`.
3. After the first startup, sign in at [http://localhost:3312](http://localhost:3312).
4. After initialization, set `NUXT_MANAGER_ADMIN_SEED=false` and remove `NUXT_MANAGER_ADMIN_PASSWORD`.

If no admin credentials are configured, the system will not generate a random password. Create an admin account manually.

---

## API Overview

Base path: `/api/v1`

| Type          | Example Path                         | Description            |
| ------------- | ------------------------------------ | ---------------------- |
| Public        | `GET /languages`                     | Language list          |
| Public        | `GET /public/settings`               | Public client settings |
| Public        | `POST /auth/login`                   | Login                  |
| Public        | `POST /auth/login/google` · `/apple` | Third-party login      |
| Auth required | `GET /users/me`                      | Current user           |
| Auth required | `POST /conversations/:id/chat`       | Text chat              |
| Auth required | `POST /conversations/:id/voice`      | Voice chat             |
| Auth required | `POST /billing/verify`               | IAP verification       |
| Static        | `GET /audio/:filename`               | Audio preview          |

Route definitions live in [`server/internal/router/router.go`](server/internal/router/router.go).

---

## Volumes

Compose mounts the following paths relative to `docker/`:

```yaml
volumes:
  - ./data/storage:/app/storage    # Audio and avatars
```

Create directories before the first deployment:

```bash
mkdir -p docker/data/storage/audio docker/data/storage/avatars
```

---

## Other Deployment Options

| Option                   | Description                                                                                                      |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------- |
| **Docker single image**  | Recommended; see Quick Start above                                                                               |
| **Split containers**     | Build the `server` binary and `manager` Nuxt output separately, then manage processes and reverse proxy yourself |
| **Bare metal / systemd** | `go build` + `node .output/server/index.mjs`, with Nginx routing ports 3312 / 8080                               |
| **Kubernetes**           | Run as a single Pod with two containers or as two Deployments sharing PVC and `DATABASE_URL`                     |

Use HTTPS for both the admin console and API in production. Restrict network access to PostgreSQL and Redis.

---

## Production Checklist

- Make sure `DATABASE_URL` points to the production database and the database only accepts trusted network traffic.
- Replace `JWT_SECRET` and `NUXT_MANAGER_AUTH_SECRET`; do not use image defaults.
- After the first admin account is created, disable `NUXT_MANAGER_ADMIN_SEED` and remove `NUXT_MANAGER_ADMIN_PASSWORD`.
- Keep `NUXT_MANAGER_TEST_ACCOUNT_SEED=false` to avoid creating the integration test account.
- Configure HTTPS for the admin console and API. If exposed publicly, add access control and rate limiting at the reverse proxy layer.
- Mount `storage/` as a persistent volume and include it in backups.

---

## Related Projects

Other XLangAI subprojects in the main repository, not included in this directory:

| Directory | Description                                              |
| --------- | -------------------------------------------------------- |
| `client/` | Flutter speaking-practice client                         |
| `home/`   | Nuxt website, using the independent `alai_home` database |
| `design/` | Architecture and deployment design documents             |

---

## Security Notes

- Never bake production `JWT_SECRET`, provider API keys, or admin passwords into the image or commit them to the repository.
- Always replace default secrets in production and keep `NUXT_MANAGER_TEST_ACCOUNT_SEED` disabled.
- Mount the Apple IAP private key as a read-only volume, for example `/run/secrets/apple-iap.p8`.
- Prisma migration scripts must be generated by operators in a controlled environment. **Do not** automatically generate migration SQL in CI and commit it to this repository.

---

## Contributing

Issues and pull requests are welcome. Before submitting:

1. Verify that both `server` and `manager` can start locally.
2. If you change the database schema, run `prisma migrate dev` yourself and include the migration files.
3. Do not commit `.env`, secrets, or user data under `storage/`.

---

## License

This project is released under the [MIT License](LICENSE).

Copyright © 2026 **GT**, **DingDangDog**
