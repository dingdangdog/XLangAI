/*
  Warnings:

  - A unique constraint covering the columns `[apple_sub]` on the table `usr_users` will be added. If there are existing duplicate values, this will fail.
  - A unique constraint covering the columns `[google_sub]` on the table `usr_users` will be added. If there are existing duplicate values, this will fail.

*/
-- AlterTable
ALTER TABLE "usr_users" ADD COLUMN     "apple_sub" VARCHAR(255),
ADD COLUMN     "google_sub" VARCHAR(255);

-- CreateTable
CREATE TABLE "sys_translate_service_configs" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(100) NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "protocol" VARCHAR(50) NOT NULL DEFAULT 'openai',
    "base_url" VARCHAR(512),
    "api_key" VARCHAR(500),
    "model_code" VARCHAR(100) NOT NULL,
    "config" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'inactive',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_translate_service_configs_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "sys_object_storage_configs" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(100) NOT NULL,
    "name" VARCHAR(200) NOT NULL,
    "provider" VARCHAR(50) NOT NULL DEFAULT 'local',
    "base_url" VARCHAR(512),
    "public_base_url" VARCHAR(512),
    "api_key" VARCHAR(500),
    "secret_key" VARCHAR(500),
    "bucket" VARCHAR(128),
    "region" VARCHAR(64),
    "config" TEXT,
    "status" VARCHAR(20) NOT NULL DEFAULT 'inactive',
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_object_storage_configs_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "sys_translate_service_configs_code_key" ON "sys_translate_service_configs"("code");

-- CreateIndex
CREATE INDEX "sys_translate_service_configs_protocol_idx" ON "sys_translate_service_configs"("protocol");

-- CreateIndex
CREATE INDEX "sys_translate_service_configs_status_idx" ON "sys_translate_service_configs"("status");

-- CreateIndex
CREATE UNIQUE INDEX "sys_object_storage_configs_code_key" ON "sys_object_storage_configs"("code");

-- CreateIndex
CREATE INDEX "sys_object_storage_configs_provider_idx" ON "sys_object_storage_configs"("provider");

-- CreateIndex
CREATE INDEX "sys_object_storage_configs_status_idx" ON "sys_object_storage_configs"("status");

-- CreateIndex
CREATE UNIQUE INDEX "usr_users_apple_sub_key" ON "usr_users"("apple_sub");

-- CreateIndex
CREATE UNIQUE INDEX "usr_users_google_sub_key" ON "usr_users"("google_sub");
