-- CreateTable
CREATE TABLE "sys_sms_service_configs" (
    "id" VARCHAR(36) NOT NULL,
    "code" VARCHAR(100) NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "provider" VARCHAR(50) NOT NULL DEFAULT 'aliyun',
    "api_key" VARCHAR(500),
    "secret_key" VARCHAR(500),
    "region" VARCHAR(64),
    "sign_name" VARCHAR(100),
    "template_code" VARCHAR(100),
    "config" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'inactive',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_sms_service_configs_pkey" PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "sys_sms_service_configs_code_key" ON "sys_sms_service_configs"("code");
CREATE INDEX "sys_sms_service_configs_provider_idx" ON "sys_sms_service_configs"("provider");
CREATE INDEX "sys_sms_service_configs_status_idx" ON "sys_sms_service_configs"("status");
