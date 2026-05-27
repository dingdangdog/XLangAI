<div align="center">

<img src="logo/logo_256x256.png" alt="XLangAI" width="128" height="128" />
</div>

<h1 align="center">XLangAI Servers</h1>

<p align="center">
  <img alt="release" src="https://img.shields.io/github/v/release/dingdangdog/xlangai" />
  <img alt="license" src="https://img.shields.io/badge/license-MIT-yellow.svg" />
  <img alt="go" src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" />
  <img alt="nuxt" src="https://img.shields.io/badge/Nuxt-4-00DC82?logo=nuxt.js&logoColor=white" />
  <img alt="postgresql" src="https://img.shields.io/badge/PostgreSQL-15+-4169E1?logo=postgresql&logoColor=white" />
  <img alt="docker" src="https://img.shields.io/docker/pulls/dingdangdog/xlangai.svg" />
</p>

<p align="center">
  <strong>XLangAI · API Server and Admin Console</strong> — a complete backend solution for multilingual AI speaking practice
</p>

<p align="center">
  <strong>Language</strong>: <a href="README.md">简体中文</a> · English
</p>

<p align="center">
  <a href="https://xlangai.com">Website</a> ·
  <a href="https://xlangai.com/servers">Server store</a> ·
  <a href="https://github.com/DingDangDog/XLangAI">GitHub source</a> ·
  <a href="https://xlangai.com/download">Download client</a>
</p>

<p align="center">
  <sub>📱 iOS client &nbsp;|&nbsp; 🏪 Official server store &nbsp;|&nbsp; 🔧 Open-source backend (This Repository)</sub>
</p>

<p align="center">
  <a href="#features">Features</a> ·
  <a href="#architecture">Architecture</a> ·
  <a href="#requirements">Requirements</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="#docker-environment-variables">Docker Environment Variables</a> ·
  <a href="#production-checklist">Production Checklist</a>
</p>

---

## About

**XLangAI** (brand name *XlangAI*, meaning *all language AI*) is a multilingual AI speaking-practice product. This directory contains the independently deployable **API server and admin console**: the Go API provides authentication, conversation, speech, billing, and media endpoints for the client app; the Nuxt admin console manages languages, models, membership tiers, prompt templates, and system settings.

| Directory              | Description                                                                     | Stack                                  |
| ---------------------- | ------------------------------------------------------------------------------- | -------------------------------------- |
| [`server/`](server/)   | Go API for the mobile app: auth, conversation, speech, billing, media, and more | Go 1.26 · Gin · GORM · PostgreSQL      |
| [`manager/`](manager/) | Admin console: configuration center, users, conversations, and backups          | Nuxt 4 · Vue 3 · Prisma 7 · PostgreSQL |

> This directory does **not** include the Flutter client or the website (`client` and `home` live in other directories of the main XLangAI repository). Deploying `servers/` alone is enough to support the app API and admin console.

### Ecosystem

| Pillar                                   | Description                                                    | Link                                                 |
| ---------------------------------------- | -------------------------------------------------------------- | ---------------------------------------------------- |
| **Client**                               | Closed source                                                  | [xlangai.com/download](https://xlangai.com/download) |
| **Website · Server store**               | Browse public community servers; copy the address into the app | [xlangai.com/servers](https://xlangai.com/servers)   |
| **Open-source backend** (this directory) | Docker-deploy API + admin                                      | This repo                                            |

### Reading Guide

- **Deploy only**: start from [Quick Start](#quick-start), then configure the database, secrets, and the first admin account.
- **Local integration**: start `manager` first so Prisma migrations and seed data are applied, then start `server`.
- **Troubleshooting configuration**: check [Docker Environment Variables](#docker-environment-variables) and [Production Checklist](#production-checklist).

**License**: [MIT](LICENSE)  
**Authors**: GT · DingDangDog

---

## Features

### Go API (`server`)

- **Users and authentication**: sign-up/sign-in, SMS verification codes, Google / Apple login, JWT sessions
- **Multilingual conversations**: create conversations, text/voice chat, message history, translation APIs
- **Speech pipeline**: STT (configurable providers) → LLM → TTS; ffmpeg loudness normalization included
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

---

## Requirements

| Scenario                | Dependencies                                                                                       |
| ----------------------- | -------------------------------------------------------------------------------------------------- |
| Docker deployment       | Docker 24+, Docker Compose v2, PostgreSQL 15+ on host or in a container                            |
| Local API development   | Go 1.26+, PostgreSQL, optional Redis                                                               |
| Local admin development | Node.js 22+, pnpm 11+, PostgreSQL                                                                  |
| Speech processing       | `ffmpeg`; included in the Docker image; install locally for STT/TTS transcoding during development |

---

## Quick Start

### 1. Prepare PostgreSQL

Create a PostgreSQL database on the host or in a separate container (example database name: `xlangai`), and note the connection string.

When a Windows or macOS container connects to PostgreSQL on the host, set the `DATABASE_URL` host to `host.docker.internal` (Compose already configures `extra_hosts`).

### 2. Configure Environment

Create `docker/.env` and at least override the database URL, JWT secrets, and first admin account:

```env
DATABASE_URL="postgresql://postgres:your-password@host.docker.internal:5432/xlangai?schema=public"
JWT_SECRET="replace-with-a-long-random-string"

NUXT_MANAGER_AUTH_SECRET="replace-with-another-long-random-string"
# Keep development for HTTP / internal Docker; use production for public HTTPS deployment
NUXT_ENV="development"
# First deployment: admin account. Disable bootstrap and remove the password variable after it is created.
NUXT_MANAGER_ADMIN_USERNAME="admin@example.com"
NUXT_MANAGER_ADMIN_PASSWORD="your-strong-password"
NUXT_MANAGER_ADMIN_NICKNAME="Admin"
```

> `JWT_SECRET` signs app-user sessions. `NUXT_MANAGER_AUTH_SECRET` signs admin-console sessions. Use separate values in production.

### 3. Start

```bash
cd servers
docker-compose up -d
```

On the first startup, `manager` runs Prisma migrations and seed data first; after that, `entrypoint.sh` starts the Go API.

| Service          | URL                                                                              |
| ---------------- | -------------------------------------------------------------------------------- |
| Admin console    | [http://localhost:3312](http://localhost:3312)                                   |
| Go API           | [http://localhost:8080](http://localhost:8080)                                   |
| API health check | [http://localhost:8080/api/v1/languages](http://localhost:8080/api/v1/languages) |

---

## Docker Environment Variables

Variables are grouped by responsibility. **Nuxt app config** goes through `runtimeConfig` and must use the `NUXT_` / `NUXT_PUBLIC_` prefix; **Node/Nitro/Prisma/Go** use their own ecosystem names.

### 1. Nuxt / Nitro (Node conventions, no `NUXT_` prefix)

| Variable                   | Default | Description                             |
| -------------------------- | ------- | --------------------------------------- |
| `MANAGER_HOST_PORT`        | `3312`  | Maps to container `PORT`                |
| `XLANGAI_SERVER_HOST_PORT` | `8080`  | Maps to container `XLANGAI_SERVER_PORT` |

### 2. Nuxt runtimeConfig — global (`NUXT_*`)

| Variable   | Default       | Description                                                                                                                                          |
| ---------- | ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `NUXT_ENV` | `development` | Runtime environment. `development`: HTTP login works (login cookie without Secure). `production`: cookie uses Secure and **requires HTTPS to sign in** |

### 3. Nuxt runtimeConfig — manager private (`NUXT_MANAGER_*`)

| Variable                             | Default | Description                                                      |
| ------------------------------------ | ------- | ---------------------------------------------------------------- |
| `NUXT_MANAGER_DATABASE_AUTO_MIGRATE` | `true`  | Run Prisma migrations on startup                                 |
| `NUXT_MANAGER_AUTH_SECRET`           | —       | Admin-console JWT secret (**change in production**)              |
| `NUXT_MANAGER_AUTO_SEED`             | `true`  | Business seed data                                               |
| `NUXT_MANAGER_TEST_ACCOUNT_SEED`     | `false` | Integration test account (`13800138000` / `123456`)              |
| `NUXT_MANAGER_ADMIN_USERNAME`        | —       | First admin login ID                                             |
| `NUXT_MANAGER_ADMIN_PASSWORD`        | —       | Plaintext password (≥6 characters), bcrypt-hashed before storage |
| `NUXT_MANAGER_ADMIN_NICKNAME`        | `Admin` | Admin display name                                               |
| `NUXT_MANAGER_ADMIN_SEED`            | `true`  | Set to `false` to disable admin bootstrap                        |

### 4. Nuxt runtimeConfig — public (`NUXT_PUBLIC_*`)

| Variable                        | Default               | Description                      |
| ------------------------------- | --------------------- | -------------------------------- |
| `NUXT_PUBLIC_OFFICIAL_HOME_URL` | `https://xlangai.com` | Official site / server store URL |

### 5. Shared infrastructure (Prisma / cross-service, no `NUXT_` prefix)

| Variable            | Default                        | Description                                                               |
| ------------------- | ------------------------------ | ------------------------------------------------------------------------- |
| `DATABASE_URL`      | —                              | PostgreSQL connection string shared by manager and Go (Prisma convention) |
| `AUDIO_DIR`         | `/app/storage/audio`           | Shared audio directory for manager and Go                                 |
| `AVATAR_DIR`        | `/app/storage/avatars`         | User avatars (Go)                                                         |
| `BUNDLED_AUDIO_DIR` | `/app/bootstrap-storage/audio` | Bundled preview audio (Go fallback)                                       |

---

## First Admin Account

1. Set both `NUXT_MANAGER_ADMIN_USERNAME` and `NUXT_MANAGER_ADMIN_PASSWORD` in Compose or `docker/.env`.
2. Configure `NUXT_MANAGER_AUTH_SECRET`.
3. After the first startup, sign in at [http://localhost:3312](http://localhost:3312).
4. After initialization, set `NUXT_MANAGER_ADMIN_SEED=false` and remove `NUXT_MANAGER_ADMIN_PASSWORD`.

If no admin credentials are configured, the system will not generate a random password. Create an admin account manually.

---

## Production Checklist

- Make sure `DATABASE_URL` points to the production database and the database only accepts trusted network traffic.
- Replace `JWT_SECRET` and `NUXT_MANAGER_AUTH_SECRET`; do not use image defaults.
- After the first admin account is created, disable `NUXT_MANAGER_ADMIN_SEED` and remove `NUXT_MANAGER_ADMIN_PASSWORD`.
- Keep `NUXT_MANAGER_TEST_ACCOUNT_SEED=false` to avoid creating the integration test account.
- Configure HTTPS for the admin console and API. If exposed publicly, add access control and rate limiting at the reverse proxy layer.
- Set `NUXT_ENV=production` for public HTTPS deployment; keep `development` for internal or plain HTTP access.
- Mount `storage/` as a persistent volume and include it in backups.

---

## Security Notes

- Never bake production `JWT_SECRET`, provider API keys, or admin passwords into the image or commit them to the repository.
- Always replace default secrets in production and keep `NUXT_MANAGER_TEST_ACCOUNT_SEED` disabled.

---

## Contributing

Issues and pull requests are welcome. Before submitting:

1. Verify that both `server` and `manager` can start locally.
2. If you change the database schema, run `prisma migrate dev` yourself and include the migration files.
3. Do not commit `.env`, secrets, or user data under `storage/`.

---

## CI/CD (Docker Images & Auto-Deploy)

When you push a semantic version tag (e.g. `v0.0.2`), GitHub Actions builds multi-platform images, pushes them to Docker Hub (`dingdangdog/xlangai`), and creates a GitHub Release. On production servers, place `docker/update.sh` alongside `docker-compose.yml` and run `docker/start.sh` to register a cron job that polls releases and rolls deployments automatically.

### Stop auto-updates

`start.sh` and `update.sh` are not long-running daemons. `start.sh` adds a crontab entry that runs `update.sh` every 10 minutes. To **stop automatic version checks and upgrades**, remove that crontab line (replace the path with your actual `update.sh` path):

```bash
crontab -l
crontab -l | grep -v '/path/to/update.sh' | crontab -
crontab -l   # confirm no update.sh entry remains
```

> Removing the cron job does **not** stop Docker containers. To stop the XLangAI service, run `docker compose down` in your compose directory (volumes are kept by default).

See [docs/servers-cicd.md](../docs/servers-cicd.md) for the full CI/CD guide.

---

## License

This project is released under the [MIT License](LICENSE).

Copyright © 2026 **GT**, **DingDangDog**
