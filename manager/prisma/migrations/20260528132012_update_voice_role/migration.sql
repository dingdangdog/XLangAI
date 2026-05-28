-- AlterTable
ALTER TABLE "sys_voice_roles" ADD COLUMN     "llm_service_config_id" VARCHAR(36),
ADD COLUMN     "role_prompt" TEXT,
ADD COLUMN     "synthesis_type" VARCHAR(30) NOT NULL DEFAULT 'tts';

-- CreateIndex
CREATE INDEX "sys_voice_roles_synthesis_type_idx" ON "sys_voice_roles"("synthesis_type");

-- CreateIndex
CREATE INDEX "sys_voice_roles_llm_service_config_id_idx" ON "sys_voice_roles"("llm_service_config_id");
