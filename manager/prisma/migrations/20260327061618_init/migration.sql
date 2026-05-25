-- CreateTable
CREATE TABLE "sys_languages" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(10) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "name_native" VARCHAR(100),
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_languages_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "sys_llm_service_configs" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(100) NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "protocol" VARCHAR(50) NOT NULL DEFAULT 'openai',
    "base_url" VARCHAR(512),
    "api_key" VARCHAR(500),
    "model_code" VARCHAR(100) NOT NULL,
    "config" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_llm_service_configs_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "sys_stt_service_configs" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(100) NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "protocol" VARCHAR(50) NOT NULL DEFAULT 'openai',
    "base_url" VARCHAR(512),
    "api_key" VARCHAR(500),
    "model_code" VARCHAR(100) NOT NULL,
    "config" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_stt_service_configs_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "sys_tts_service_configs" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(100) NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "provider" VARCHAR(50) NOT NULL,
    "base_url" VARCHAR(512),
    "api_key" VARCHAR(500),
    "region" VARCHAR(64),
    "model_code" VARCHAR(100) NOT NULL,
    "config" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_tts_service_configs_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "sys_voice_roles" (
    "id" TEXT NOT NULL,
    "language_id" VARCHAR(36),
    "tts_service_config_id" VARCHAR(36),
    "voice_code" VARCHAR(50) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "gender" VARCHAR(20),
    "config" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_voice_roles_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "sys_prompt_templates" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(50) NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "content" TEXT NOT NULL,
    "variables" TEXT,
    "language_id" VARCHAR(36),
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_prompt_templates_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "sys_membership_tiers" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(20) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "daily_limit" INTEGER,
    "monthly_limit" INTEGER,
    "features" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_membership_tiers_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "usr_users" (
    "id" TEXT NOT NULL,
    "phone" VARCHAR(20),
    "email" VARCHAR(255),
    "password_hash" VARCHAR(255),
    "nickname" VARCHAR(100),
    "avatar_url" VARCHAR(500),
    "tier_id" VARCHAR(36),
    "language_id" VARCHAR(36),
    "settings" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "last_login_at" TIMESTAMPTZ(6),
    "remark" VARCHAR(500),
    "deleted_at" TIMESTAMPTZ(6),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "usr_users_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "usr_user_usage" (
    "id" TEXT NOT NULL,
    "user_id" VARCHAR(36) NOT NULL,
    "date" DATE NOT NULL,
    "usage_count" INTEGER NOT NULL DEFAULT 0,
    "token_count" INTEGER NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "usr_user_usage_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "usr_conversations" (
    "id" TEXT NOT NULL,
    "user_id" VARCHAR(36) NOT NULL,
    "language_id" VARCHAR(36) NOT NULL,
    "voice_role_id" VARCHAR(36),
    "llm_config_id" VARCHAR(36),
    "prompt_id" VARCHAR(36),
    "title" VARCHAR(200) DEFAULT '新对话',
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "remark" VARCHAR(500),
    "deleted_at" TIMESTAMPTZ(6),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "usr_conversations_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "usr_messages" (
    "id" TEXT NOT NULL,
    "conversation_id" VARCHAR(36) NOT NULL,
    "role" VARCHAR(20) NOT NULL,
    "content" TEXT NOT NULL,
    "audio_url" VARCHAR(500),
    "original_audio_url" VARCHAR(500),
    "stt_text" TEXT,
    "duration_ms" INTEGER,
    "metadata" TEXT,
    "deleted_at" TIMESTAMPTZ(6),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "usr_messages_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "bak_usr_users" (
    "id" VARCHAR(36) NOT NULL,
    "phone" VARCHAR(20),
    "email" VARCHAR(255),
    "password_hash" VARCHAR(255),
    "nickname" VARCHAR(100),
    "avatar_url" VARCHAR(500),
    "tier_id" VARCHAR(36),
    "language_id" VARCHAR(36),
    "settings" TEXT,
    "status" VARCHAR(20),
    "last_login_at" TIMESTAMPTZ(6),
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6),
    "updated_at" TIMESTAMPTZ(6),
    "cancelled_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "backup_batch" VARCHAR(36) NOT NULL,

    CONSTRAINT "bak_usr_users_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "bak_usr_conversations" (
    "id" VARCHAR(36) NOT NULL,
    "user_id" VARCHAR(36) NOT NULL,
    "language_id" VARCHAR(36) NOT NULL,
    "voice_role_id" VARCHAR(36),
    "llm_config_id" VARCHAR(36),
    "prompt_id" VARCHAR(36),
    "title" VARCHAR(200),
    "status" VARCHAR(20),
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6),
    "updated_at" TIMESTAMPTZ(6),
    "cancelled_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "backup_batch" VARCHAR(36) NOT NULL,

    CONSTRAINT "bak_usr_conversations_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "bak_usr_messages" (
    "id" VARCHAR(36) NOT NULL,
    "conversation_id" VARCHAR(36) NOT NULL,
    "role" VARCHAR(20) NOT NULL,
    "content" TEXT NOT NULL,
    "audio_url" VARCHAR(500),
    "original_audio_url" VARCHAR(500),
    "stt_text" TEXT,
    "duration_ms" INTEGER,
    "metadata" TEXT,
    "created_at" TIMESTAMPTZ(6) NOT NULL,
    "cancelled_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "backup_batch" VARCHAR(36) NOT NULL,

    CONSTRAINT "bak_usr_messages_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "bak_usr_user_usage" (
    "id" VARCHAR(36) NOT NULL,
    "user_id" VARCHAR(36) NOT NULL,
    "date" DATE NOT NULL,
    "usage_count" INTEGER NOT NULL DEFAULT 0,
    "token_count" INTEGER NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ(6),
    "updated_at" TIMESTAMPTZ(6),
    "cancelled_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "backup_batch" VARCHAR(36) NOT NULL,

    CONSTRAINT "bak_usr_user_usage_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "sys_languages_code_key" ON "sys_languages"("code");

-- CreateIndex
CREATE INDEX "sys_languages_code_idx" ON "sys_languages"("code");

-- CreateIndex
CREATE INDEX "sys_languages_status_idx" ON "sys_languages"("status");

-- CreateIndex
CREATE UNIQUE INDEX "sys_llm_service_configs_code_key" ON "sys_llm_service_configs"("code");

-- CreateIndex
CREATE INDEX "sys_llm_service_configs_protocol_idx" ON "sys_llm_service_configs"("protocol");

-- CreateIndex
CREATE INDEX "sys_llm_service_configs_status_idx" ON "sys_llm_service_configs"("status");

-- CreateIndex
CREATE UNIQUE INDEX "sys_stt_service_configs_code_key" ON "sys_stt_service_configs"("code");

-- CreateIndex
CREATE INDEX "sys_stt_service_configs_protocol_idx" ON "sys_stt_service_configs"("protocol");

-- CreateIndex
CREATE INDEX "sys_stt_service_configs_status_idx" ON "sys_stt_service_configs"("status");

-- CreateIndex
CREATE UNIQUE INDEX "sys_tts_service_configs_code_key" ON "sys_tts_service_configs"("code");

-- CreateIndex
CREATE INDEX "sys_tts_service_configs_provider_idx" ON "sys_tts_service_configs"("provider");

-- CreateIndex
CREATE INDEX "sys_tts_service_configs_status_idx" ON "sys_tts_service_configs"("status");

-- CreateIndex
CREATE INDEX "sys_voice_roles_language_id_idx" ON "sys_voice_roles"("language_id");

-- CreateIndex
CREATE INDEX "sys_voice_roles_tts_service_config_id_idx" ON "sys_voice_roles"("tts_service_config_id");

-- CreateIndex
CREATE INDEX "sys_voice_roles_status_idx" ON "sys_voice_roles"("status");

-- CreateIndex
CREATE UNIQUE INDEX "sys_prompt_templates_code_key" ON "sys_prompt_templates"("code");

-- CreateIndex
CREATE INDEX "sys_prompt_templates_code_idx" ON "sys_prompt_templates"("code");

-- CreateIndex
CREATE INDEX "sys_prompt_templates_status_idx" ON "sys_prompt_templates"("status");

-- CreateIndex
CREATE UNIQUE INDEX "sys_membership_tiers_code_key" ON "sys_membership_tiers"("code");

-- CreateIndex
CREATE INDEX "sys_membership_tiers_code_idx" ON "sys_membership_tiers"("code");

-- CreateIndex
CREATE INDEX "sys_membership_tiers_status_idx" ON "sys_membership_tiers"("status");

-- CreateIndex
CREATE INDEX "usr_users_tier_id_idx" ON "usr_users"("tier_id");

-- CreateIndex
CREATE INDEX "usr_users_status_idx" ON "usr_users"("status");

-- CreateIndex
CREATE INDEX "usr_users_phone_idx" ON "usr_users"("phone");

-- CreateIndex
CREATE INDEX "usr_users_email_idx" ON "usr_users"("email");

-- CreateIndex
CREATE INDEX "usr_user_usage_user_id_date_idx" ON "usr_user_usage"("user_id", "date");

-- CreateIndex
CREATE UNIQUE INDEX "usr_user_usage_user_id_date_key" ON "usr_user_usage"("user_id", "date");

-- CreateIndex
CREATE INDEX "usr_conversations_user_id_idx" ON "usr_conversations"("user_id");

-- CreateIndex
CREATE INDEX "usr_conversations_status_idx" ON "usr_conversations"("status");

-- CreateIndex
CREATE INDEX "usr_conversations_created_at_idx" ON "usr_conversations"("created_at");

-- CreateIndex
CREATE INDEX "usr_messages_conversation_id_idx" ON "usr_messages"("conversation_id");

-- CreateIndex
CREATE INDEX "usr_messages_created_at_idx" ON "usr_messages"("created_at");

-- CreateIndex
CREATE INDEX "bak_usr_users_backup_batch_idx" ON "bak_usr_users"("backup_batch");

-- CreateIndex
CREATE INDEX "bak_usr_conversations_backup_batch_idx" ON "bak_usr_conversations"("backup_batch");

-- CreateIndex
CREATE INDEX "bak_usr_messages_backup_batch_idx" ON "bak_usr_messages"("backup_batch");
