/*
  Warnings:

  - The primary key for the `sys_service_usage_daily` table will be changed. If it partially fails, the table could be left without primary key constraint.

*/
-- AlterTable
ALTER TABLE "sys_service_usage_daily" DROP CONSTRAINT "sys_service_usage_daily_pkey",
ALTER COLUMN "id" SET DATA TYPE TEXT,
ADD CONSTRAINT "sys_service_usage_daily_pkey" PRIMARY KEY ("id");

-- AlterTable
ALTER TABLE "sys_system_settings" ADD COLUMN     "status" VARCHAR(20) NOT NULL DEFAULT 'active';

-- CreateTable
CREATE TABLE "sys_sms_service_configs" (
    "id" TEXT NOT NULL,
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

-- CreateIndex
CREATE UNIQUE INDEX "sys_sms_service_configs_code_key" ON "sys_sms_service_configs"("code");

-- CreateIndex
CREATE INDEX "sys_sms_service_configs_provider_idx" ON "sys_sms_service_configs"("provider");

-- CreateIndex
CREATE INDEX "sys_sms_service_configs_status_idx" ON "sys_sms_service_configs"("status");

-- CreateIndex
CREATE INDEX "sys_system_settings_status_idx" ON "sys_system_settings"("status");
