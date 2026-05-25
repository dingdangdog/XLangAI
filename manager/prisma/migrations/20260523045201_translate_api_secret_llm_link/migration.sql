-- AlterTable
ALTER TABLE "sys_translate_service_configs" ADD COLUMN     "api_secret" VARCHAR(500),
ADD COLUMN     "llm_config_id" VARCHAR(36);

-- CreateIndex
CREATE INDEX "sys_translate_service_configs_llm_config_id_idx" ON "sys_translate_service_configs"("llm_config_id");
