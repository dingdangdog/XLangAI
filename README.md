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
  <strong>XLangAI（小浪AI） · 服务端与运营后台</strong> — 面向多语种口语练习场景的完整后端解决方案
</p>

<p align="center">
  <strong>语言</strong>：简体中文 · <a href="README.en.md">English</a>
</p>

<p align="center">
  <a href="#功能特性">功能特性</a> ·
  <a href="#架构">架构</a> ·
  <a href="#快速开始">快速开始</a> ·
  <a href="#本地开发">本地开发</a> ·
  <a href="#环境变量">环境变量</a> ·
  <a href="#api-概览">API 概览</a>
</p>

---

## 关于

**XLangAI**（英文品牌 *XlangAI*，意为 *all language AI*）是一款多语种 AI 口语练习产品。本目录包含可独立部署的**服务端与运营后台**：Go API 为客户端提供鉴权、对话、语音、计费与媒体接口，Nuxt 运营后台负责语种、模型、会员、提示词与系统配置。

| 子目录                 | 说明                                               | 技术栈                                 |
| ---------------------- | -------------------------------------------------- | -------------------------------------- |
| [`server/`](server/)   | 面向 App 的 Go API：鉴权、对话、语音、计费、媒体等 | Go 1.26 · Gin · GORM · PostgreSQL      |
| [`manager/`](manager/) | 运营后台：配置中心、用户与对话管理、数据备份       | Nuxt 4 · Vue 3 · Prisma 7 · PostgreSQL |

> 本目录**不包含** Flutter 客户端与官网（`client`、`home` 位于 XLangAI 主仓库的其他目录）。只部署 `servers/` 即可支撑 App API 与运营后台。

### 阅读路径

- **只想部署**：从 [快速开始](#快速开始) 开始，按顺序配置数据库、密钥和管理员账号。
- **本地联调**：先启动 `manager` 完成 Prisma 迁移与种子数据，再启动 `server`。
- **排查配置**：优先查看 [环境变量](#环境变量) 与 [生产上线检查](#生产上线检查)。

**开源协议**：[MIT](LICENSE)  
**作者**：GT · DingDangDog

---

## 功能特性

### Go API（`server`）

- **用户与鉴权**：注册/登录、短信验证码、Google / Apple 第三方登录、JWT 会话
- **多语种对话**：创建会话、文本/语音对话、消息历史、翻译接口
- **语音链路**：STT（多协议可配置）→ LLM → TTS；内置 ffmpeg 响度归一化
- **会员与计费**：会员档位、应用内购校验（Apple / Google Play）
- **媒体与存储**：头像上传、音频试听、对象存储预签名（R2 / S3 / 七牛 / 阿里云 OSS / 本地）
- **可观测性**：Gin 访问日志 + 统一 `[api]` 错误日志（见 [`design/server-logging.md`](../design/server-logging.md)）

### 运营后台（`manager`）

- **服务配置**：LLM / STT / TTS / 翻译 / 对象存储 多服务商管理
- **内容与运营**：语种、音色角色、提示词模板、会员档位、系统设置
- **用户域**：用户列表、对话与消息、用量统计、注销备份
- **运维**：数据库迁移（Nitro 启动时自动执行）、种子数据、备份导出、服务器商店同步

### 部署

- **单镜像**：根目录 `Dockerfile` 构建 Nuxt + Go 一体化镜像
- **Docker Compose**：一键启动，挂载 `storage` 持久卷
- **分体部署**：`server` 与 `manager` 可分别本地或容器运行，共用同一 PostgreSQL

---

## 架构

```mermaid
flowchart TB
  subgraph clients [客户端]
    App[Flutter App]
    Browser[运营人员浏览器]
  end

  subgraph servers [本仓库 servers]
    API[Go API :8080]
    MGR[Nuxt Manager :3312]
    DB[(PostgreSQL)]
    Vol[/storage 音频·头像/]
  end

  subgraph external [外部服务 可配置]
    LLM[LLM 服务商]
    STT[STT 服务商]
    TTS[TTS 服务商]
    OSS[对象存储]
  end

  App --> API
  Browser --> MGR
  MGR --> DB
  API --> DB
  MGR -.配置下发.-> API
  API --> LLM & STT & TTS
  API --> OSS
  API --> Vol
  MGR --> Vol
```

---

## 前置要求

| 场景         | 依赖                                                          |
| ------------ | ------------------------------------------------------------- |
| Docker 部署  | Docker 24+、Docker Compose v2、PostgreSQL 15+（宿主机或容器） |
| 本地开发 API | Go 1.26+、PostgreSQL、可选 Redis                              |
| 本地开发后台 | Node.js 22+、pnpm 11+、PostgreSQL                             |
| 语音处理     | `ffmpeg`（Docker 镜像已内置；本地 STT/TTS 转码需自行安装）    |

---

## 快速开始

### 1. 准备数据库

在宿主机或独立容器中创建 PostgreSQL 数据库（示例库名 `wlltalk`），并记下连接串。

Windows / macOS 容器访问宿主机 Postgres 时，将 `DATABASE_URL` 主机改为 `host.docker.internal`（Compose 已配置 `extra_hosts`）。

### 2. 配置环境

在 `docker/` 目录创建 `.env`，至少覆盖数据库、JWT 密钥与首次管理员账号：

```env
DATABASE_URL=postgresql://postgres:your-password@host.docker.internal:5432/wlltalk?schema=public
JWT_SECRET=请替换为足够长的随机字符串
NUXT_MANAGER_AUTH_SECRET=请替换为另一个足够长的随机字符串

# 首次部署：运营管理员（写入后建议关闭初始化并删除密码变量）
NUXT_MANAGER_ADMIN_USERNAME=admin@example.com
NUXT_MANAGER_ADMIN_PASSWORD=your-strong-password
NUXT_MANAGER_ADMIN_NICKNAME=管理员
```

> `JWT_SECRET` 用于 App 用户会话；`NUXT_MANAGER_AUTH_SECRET` 用于运营后台登录会话。生产环境建议分别设置。

### 3. 启动

```bash
cd servers
mkdir -p docker/data/storage/audio docker/data/storage/avatars

docker compose -f docker/docker-compose.yml up -d --build
```

首次启动时，`manager` 会先执行 Prisma 迁移和种子数据；完成后 `entrypoint.sh` 再启动 Go API。

| 服务         | 地址                                                                             |
| ------------ | -------------------------------------------------------------------------------- |
| 运营后台     | [http://localhost:3312](http://localhost:3312)                                   |
| Go API       | [http://localhost:8080](http://localhost:8080)                                   |
| API 健康检查 | [http://localhost:8080/api/v1/languages](http://localhost:8080/api/v1/languages) |

### 4. 仅构建 / 导出镜像

```bash
cd servers
docker build -t xlangai:latest .

# 离线分发
docker save -o xlangai.tar xlangai:latest
docker load -i xlangai.tar
```

> `manager/Dockerfile` 为历史独立镜像，已弃用；请使用根目录 `Dockerfile` 与 `docker/docker-compose.yml`。

---

## 本地开发

`server` 与 `manager` 共用同一 PostgreSQL 实例（Prisma 迁移由 manager 负责）。

推荐先启动 `manager`，确认数据库迁移和初始数据完成后，再启动 `server`。

### 运营后台

```bash
cd servers/manager
cp env .env   # 按需修改 DATABASE_URL、管理员账号等
pnpm install
pnpm run postinstall
pnpm dev      # 默认 http://localhost:3312
```

数据库迁移（开发环境）：

```bash
pnpm db:migrate
# 或生产式：npx prisma migrate deploy
```

### Go API

```bash
cd servers/server
cp env .env   # 配置 DATABASE_URL、JWT_SECRET 等
go mod download
go run .
# 默认 http://localhost:8080
```

本地同时运行 manager 与 Go API 时，在 `manager/.env` 中设置 `AUDIO_DIR=../server/storage/audio`（与 Go 侧音频目录一致）。后台试听由 manager 路由 `/api/admin/preview-audio/:filename` 播放。

---

## 环境变量

变量按职责分组。**Nuxt 应用配置**走 `runtimeConfig`，环境变量必须以 `NUXT_` / `NUXT_PUBLIC_` 开头；**Node/Nitro/Prisma/Go** 使用各自生态约定名。

### 1. 容器运行时

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `TZ` | `Asia/Shanghai` | 时区 |
| `APP_USER` | `xlangai` | entrypoint 降权运行用户 |
| `NODE_ENV` | `production` | Node 运行模式 |

### 2. Nuxt / Nitro（Node 约定，不用 `NUXT_` 前缀）

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `PORT` | `3312` | 运营后台监听端口（Nitro 默认 host 为 `0.0.0.0`，无需另设 `NITRO_HOST`） |

Compose 宿主机映射（仅改对外端口）：

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `MANAGER_HOST_PORT` | `3312` | 映射到容器内 `PORT` |
| `XLANGAI_SERVER_HOST_PORT` | `8080` | 映射到容器内 `XLANGAI_SERVER_PORT` |

### 3. Nuxt runtimeConfig — manager 私有（`NUXT_MANAGER_*`）

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `NUXT_MANAGER_DATABASE_AUTO_MIGRATE` | `true` | 启动时执行 Prisma 迁移 |
| `NUXT_MANAGER_AUTH_SECRET` | — | 运营后台 JWT 签名密钥（**生产必改**） |
| `NUXT_MANAGER_AUTO_SEED` | `true` | 业务种子数据 |
| `NUXT_MANAGER_TEST_ACCOUNT_SEED` | `false` | 联调测试账号（`13800138000` / `123456`） |
| `NUXT_MANAGER_ADMIN_USERNAME` | — | 首次运营管理员登录名 |
| `NUXT_MANAGER_ADMIN_PASSWORD` | — | 明文密码（≥6 位），入库 bcrypt |
| `NUXT_MANAGER_ADMIN_NICKNAME` | `管理员` | 管理员昵称 |
| `NUXT_MANAGER_ADMIN_SEED` | `true` | 设为 `false` 关闭管理员自动初始化 |

### 4. Nuxt runtimeConfig — 公开（`NUXT_PUBLIC_*`）

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `NUXT_PUBLIC_OFFICIAL_HOME_URL` | `https://xlangai.com` | 官网 / 服务器商店地址 |

### 5. 共享基础设施（Prisma / 跨服务，非 `NUXT_` 前缀）

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `DATABASE_URL` | — | PostgreSQL 连接串；manager 与 Go **共用**（Prisma 约定） |
| `DATABASE_BOOTSTRAP_URL` | — | 可选；空库时先连 postgres 库建库 |
| `DATABASE_SCHEMA` | `public` | 可选；迁移目标 schema |
| `AUDIO_DIR` | `/app/storage/audio` | manager 与 Go 共用音频目录 |
| `AVATAR_DIR` | `/app/storage/avatars` | 用户头像（Go） |
| `BUNDLED_AUDIO_DIR` | `/app/bootstrap-storage/audio` | 内置试听音频（Go fallback） |

### 6. Go API（`XLANGAI_*` / `JWT_*`）

| 变量 | 默认 | 说明 |
| --- | --- | --- |
| `XLANGAI_SERVER_PORT` | `8080` | Go API 监听端口（单容器内与 `PORT` 分离） |
| `JWT_SECRET` | — | App 用户 JWT 签名密钥（**生产必改**） |
| `REDIS_URL` | — | 可选 Redis |
| `GIN_MODE` | `release` | Gin 运行模式 |
| `XLANGAI_VERBOSE_LOGS` | `0` | `1` 开启详细错误上下文 |
| `XLANGAI_FFMPEG_PATH` | `ffmpeg` | ffmpeg 可执行路径 |
| `XLANGAI_STT_LANGUAGE_MODE` | — | `auto` / `target` |
| `XLANGAI_TTS_LOUDNESS_NORM` | 开启 | TTS 响度归一化 |
| `XLANGAI_APPLE_*` / `XLANGAI_GOOGLE_*` | — | 内购与 OAuth 校验 |

完整 Docker 说明见主仓库 [`design/DOCKER.md`](../design/DOCKER.md)。

---

## 运营管理员（首次部署）

1. 在 Compose 或 `docker/.env` 中**同时**设置 `NUXT_MANAGER_ADMIN_USERNAME` 与 `NUXT_MANAGER_ADMIN_PASSWORD`。
2. 配置 `NUXT_MANAGER_AUTH_SECRET`。
3. 首次启动后使用上述账号登录 [http://localhost:3312](http://localhost:3312)。
4. 初始化完成后：设置 `NUXT_MANAGER_ADMIN_SEED=false`，并**移除** `NUXT_MANAGER_ADMIN_PASSWORD` 环境变量。

未配置管理员时不会自动生成随机密码，需自行创建。

---

## API 概览

基础路径：`/api/v1`

| 类型   | 路径示例                             | 说明           |
| ------ | ------------------------------------ | -------------- |
| 公开   | `GET /languages`                     | 语种列表       |
| 公开   | `GET /public/settings`               | 客户端公开配置 |
| 公开   | `POST /auth/login`                   | 登录           |
| 公开   | `POST /auth/login/google` · `/apple` | 第三方登录     |
| 需鉴权 | `GET /users/me`                      | 当前用户       |
| 需鉴权 | `POST /conversations/:id/chat`       | 文本对话       |
| 需鉴权 | `POST /conversations/:id/voice`      | 语音对话       |
| 需鉴权 | `POST /billing/verify`               | 内购校验       |
| 静态   | `GET /audio/:filename`               | 音频试听       |

路由定义见 [`server/internal/router/router.go`](server/internal/router/router.go)。

---

## 数据卷

Compose 默认挂载（相对 `docker/` 目录）：

```yaml
volumes:
  - ./data/storage:/app/storage    # 音频、头像
```

首次部署前建议创建目录：

```bash
mkdir -p docker/data/storage/audio docker/data/storage/avatars
```

---

## 其他部署方式

| 方式               | 说明                                                                     |
| ------------------ | ------------------------------------------------------------------------ |
| **Docker 单镜像**  | 推荐；见上文 Quick Start                                                 |
| **分体容器**       | 分别构建 `server` 二进制与 `manager` Nuxt output，自行编排进程与反向代理 |
| **裸机 / systemd** | `go build` + `node .output/server/index.mjs`，Nginx 分流 3312 / 8080     |
| **Kubernetes**     | 以单 Pod 双容器或双 Deployment 方式部署，共享 PVC 与 `DATABASE_URL`      |

生产环境请为运营后台与 API 配置 HTTPS，并限制 Postgres 与 Redis 的网络访问。

---

## 生产上线检查

- 确认 `DATABASE_URL` 指向生产数据库，且数据库只允许可信网络访问。
- 替换 `JWT_SECRET` 与 `NUXT_MANAGER_AUTH_SECRET`，不要使用镜像默认值。
- 首次管理员初始化完成后，关闭 `NUXT_MANAGER_ADMIN_SEED` 并移除 `NUXT_MANAGER_ADMIN_PASSWORD`。
- 将 `NUXT_MANAGER_TEST_ACCOUNT_SEED` 保持为 `false`，避免创建联调测试账号。
- 为运营后台和 API 配置 HTTPS；如暴露在公网，建议在反向代理层增加访问控制与限流。
- 将 `storage/` 挂载到持久卷，并纳入备份策略。

---

## 相关项目

XLangAI 主仓库中的其他子项目（本仓库未包含）：

| 目录      | 说明                                |
| --------- | ----------------------------------- |
| `client/` | Flutter 口语练习客户端              |
| `home/`   | Nuxt 官网（独立数据库 `alai_home`） |
| `design/` | 架构与部署设计文档                  |

---

## 安全提示

- 切勿将生产 `JWT_SECRET`、服务商 API Key、管理员密码写入镜像或提交版本库。
- 生产务必修改默认密钥，并关闭 `NUXT_MANAGER_TEST_ACCOUNT_SEED`。
- Apple IAP 私钥建议通过只读卷挂载，例如 `/run/secrets/apple-iap.p8`。
- Prisma 迁移脚本需由运维在受控环境执行；**请勿**在 CI 中自动生成 migration SQL 提交本仓库（项目约定）。

---

## 参与贡献

欢迎 Issue 与 Pull Request。提交前请：

1. 在本地分别验证 `server` 与 `manager` 可正常启动；
2. 若修改数据库结构，请自行运行 `prisma migrate dev` 并附带 migration 文件；
3. 勿提交 `.env`、密钥与 `storage/` 用户数据。

---

## 许可证

本项目采用 [MIT License](LICENSE) 开源。

Copyright © 2026 **GT**, **DingDangDog**
