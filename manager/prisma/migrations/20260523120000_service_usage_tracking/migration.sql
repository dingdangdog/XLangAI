-- 用户按日用量：补全 TTS / 翻译 / STT 细分指标
ALTER TABLE "usr_user_usage" ADD COLUMN IF NOT EXISTS "translate_count" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "usr_user_usage" ADD COLUMN IF NOT EXISTS "translate_chars" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "usr_user_usage" ADD COLUMN IF NOT EXISTS "tts_count" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "usr_user_usage" ADD COLUMN IF NOT EXISTS "tts_chars" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "usr_user_usage" ADD COLUMN IF NOT EXISTS "stt_count" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "usr_user_usage" ADD COLUMN IF NOT EXISTS "stt_audio_bytes" BIGINT NOT NULL DEFAULT 0;

-- 服务配置维度：按 UTC 日 + 类型 + config_id 聚合
CREATE TABLE IF NOT EXISTS "sys_service_usage_daily" (
    "id" VARCHAR(36) NOT NULL,
    "date" DATE NOT NULL,
    "service_type" VARCHAR(20) NOT NULL,
    "config_id" VARCHAR(36) NOT NULL,
    "request_count" INTEGER NOT NULL DEFAULT 0,
    "unit_count" BIGINT NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_service_usage_daily_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX IF NOT EXISTS "sys_service_usage_daily_date_service_type_config_id_key"
    ON "sys_service_usage_daily"("date", "service_type", "config_id");

CREATE INDEX IF NOT EXISTS "sys_service_usage_daily_service_type_config_id_date_idx"
    ON "sys_service_usage_daily"("service_type", "config_id", "date");
