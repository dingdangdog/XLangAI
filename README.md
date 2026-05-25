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
  <a href="https://xlangai.com">官网</a> ·
  <a href="https://xlangai.com/servers">服务器商店</a> ·
  <a href="https://github.com/DingDangDog/XLangAI">GitHub 源码</a> ·
  <a href="https://xlangai.com/download">下载客户端</a>
</p>

<p align="center">
  <sub>📱 iOS 客户端 · 一次性 $2 买断 &nbsp;|&nbsp; 🏪 官网服务器商店 &nbsp;|&nbsp; 🔧 开源服务端（官方不提供正式托管）</sub>
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

### 生态三角

| 端 | 说明 | 入口 |
| --- | --- | --- |
| **客户端** | 闭源 | [xlangai.com/download](https://xlangai.com/download) |
| **官网 · 服务器商店** | 浏览社区公开服务器，复制地址到 App | [xlangai.com/servers](https://xlangai.com/servers) |
| **开源服务端**（本目录） | Docker 部署 API + 运营后台 | 本仓库 |

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

在宿主机或独立容器中创建 PostgreSQL 数据库（示例库名 `xlangai`），并记下连接串。

Windows / macOS 容器访问宿主机 Postgres 时，将 `DATABASE_URL` 主机改为 `host.docker.internal`（Compose 已配置 `extra_hosts`）。

### 2. 配置环境

在 `docker/` 目录创建 `.env`，至少覆盖数据库、JWT 密钥与首次管理员账号：

```env
DATABASE_URL="postgresql://postgres:your-password@host.docker.internal:5432/xlangai?schema=public"
JWT_SECRET="请替换为足够长的随机字符串"

NUXT_MANAGER_AUTH_SECRET="请替换为另一个足够长的随机字符串"
# 首次部署：运营管理员（写入后建议关闭初始化并删除密码变量）
NUXT_MANAGER_ADMIN_USERNAME="admin@example.com"
NUXT_MANAGER_ADMIN_PASSWORD="your-strong-password"
NUXT_MANAGER_ADMIN_NICKNAME="管理员"
```

> `JWT_SECRET` 用于 App 用户会话；`NUXT_MANAGER_AUTH_SECRET` 用于运营后台登录会话。生产环境建议分别设置。

### 3. 启动

```bash
cd servers
docker-compose up -d
```

首次启动时，`manager` 会先执行 Prisma 迁移和种子数据；完成后 `entrypoint.sh` 再启动 Go API。

| 服务         | 地址                                                                             |
| ------------ | -------------------------------------------------------------------------------- |
| 运营后台     | [http://localhost:3312](http://localhost:3312)                                   |
| Go API       | [http://localhost:8080](http://localhost:8080)                                   |
| API 健康检查 | [http://localhost:8080/api/v1/languages](http://localhost:8080/api/v1/languages) |

---

## Docker环境变量

变量按职责分组。**Nuxt 应用配置**走 `runtimeConfig`，环境变量必须以 `NUXT_` / `NUXT_PUBLIC_` 开头；**Node/Nitro/Prisma/Go** 使用各自生态约定名。

### 1. Nuxt / Nitro（Node 约定，不用 `NUXT_` 前缀）

| 变量                       | 默认   | 说明                               |
| -------------------------- | ------ | ---------------------------------- |
| `MANAGER_HOST_PORT`        | `3312` | 映射到容器内 `PORT`                |
| `XLANGAI_SERVER_HOST_PORT` | `8080` | 映射到容器内 `XLANGAI_SERVER_PORT` |

### 2. Nuxt runtimeConfig — manager 私有（`NUXT_MANAGER_*`）

| 变量                                 | 默认     | 说明                                     |
| ------------------------------------ | -------- | ---------------------------------------- |
| `NUXT_MANAGER_DATABASE_AUTO_MIGRATE` | `true`   | 启动时执行 Prisma 迁移                   |
| `NUXT_MANAGER_AUTH_SECRET`           | —        | 运营后台 JWT 签名密钥（**生产必改**）    |
| `NUXT_MANAGER_AUTO_SEED`             | `true`   | 业务种子数据                             |
| `NUXT_MANAGER_TEST_ACCOUNT_SEED`     | `false`  | 联调测试账号（`13800138000` / `123456`） |
| `NUXT_MANAGER_ADMIN_USERNAME`        | —        | 首次运营管理员登录名                     |
| `NUXT_MANAGER_ADMIN_PASSWORD`        | —        | 明文密码（≥6 位），入库 bcrypt           |
| `NUXT_MANAGER_ADMIN_NICKNAME`        | `管理员` | 管理员昵称                               |
| `NUXT_MANAGER_ADMIN_SEED`            | `true`   | 设为 `false` 关闭管理员自动初始化        |

### 3. Nuxt runtimeConfig — 公开（`NUXT_PUBLIC_*`）

| 变量                            | 默认                  | 说明                  |
| ------------------------------- | --------------------- | --------------------- |
| `NUXT_PUBLIC_OFFICIAL_HOME_URL` | `https://xlangai.com` | 官网 / 服务器商店地址 |

### 5. 共享基础设施（Prisma / 跨服务，非 `NUXT_` 前缀）

| 变量                | 默认                           | 说明                                                     |
| ------------------- | ------------------------------ | -------------------------------------------------------- |
| `DATABASE_URL`      | —                              | PostgreSQL 连接串；manager 与 Go **共用**（Prisma 约定） |
| `AUDIO_DIR`         | `/app/storage/audio`           | manager 与 Go 共用音频目录                               |
| `AVATAR_DIR`        | `/app/storage/avatars`         | 用户头像（Go）                                           |
| `BUNDLED_AUDIO_DIR` | `/app/bootstrap-storage/audio` | 内置试听音频（Go fallback）                              |

---

## 运营管理员（首次部署）

1. 在 Compose 或 `docker/.env` 中**同时**设置 `NUXT_MANAGER_ADMIN_USERNAME` 与 `NUXT_MANAGER_ADMIN_PASSWORD`。
2. 配置 `NUXT_MANAGER_AUTH_SECRET`。
3. 首次启动后使用上述账号登录 [http://localhost:3312](http://localhost:3312)。
4. 初始化完成后：设置 `NUXT_MANAGER_ADMIN_SEED=false`，并**移除** `NUXT_MANAGER_ADMIN_PASSWORD` 环境变量。

未配置管理员时不会自动生成随机密码，需自行创建。

---

## 生产上线检查

- 确认 `DATABASE_URL` 指向生产数据库，且数据库只允许可信网络访问。
- 替换 `JWT_SECRET` 与 `NUXT_MANAGER_AUTH_SECRET`，不要使用镜像默认值。
- 首次管理员初始化完成后，关闭 `NUXT_MANAGER_ADMIN_SEED` 并移除 `NUXT_MANAGER_ADMIN_PASSWORD`。
- 将 `NUXT_MANAGER_TEST_ACCOUNT_SEED` 保持为 `false`，避免创建联调测试账号。
- 为运营后台和 API 配置 HTTPS；如暴露在公网，建议在反向代理层增加访问控制与限流。
- 将 `storage/` 挂载到持久卷，并纳入备份策略。

---

## 安全提示

- 切勿将生产 `JWT_SECRET`、服务商 API Key、管理员密码写入镜像或提交版本库。
- 生产务必修改默认密钥，并关闭 `NUXT_MANAGER_TEST_ACCOUNT_SEED`。

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
