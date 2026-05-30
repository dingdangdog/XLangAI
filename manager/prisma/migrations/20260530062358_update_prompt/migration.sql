-- AlterTable
ALTER TABLE "bak_usr_conversations" ADD COLUMN     "scenario_code" VARCHAR(32);

-- AlterTable
ALTER TABLE "usr_conversations" ADD COLUMN     "scenario_code" VARCHAR(32);

-- CreateTable
CREATE TABLE "sys_practice_scenarios" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(32) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "name_en" VARCHAR(100),
    "icon" VARCHAR(50),
    "description" VARCHAR(500),
    "description_en" VARCHAR(500),
    "prompt_template_id" VARCHAR(36),
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_practice_scenarios_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "sys_practice_scenarios_code_key" ON "sys_practice_scenarios"("code");

-- CreateIndex
CREATE INDEX "sys_practice_scenarios_code_idx" ON "sys_practice_scenarios"("code");

-- CreateIndex
CREATE INDEX "sys_practice_scenarios_status_idx" ON "sys_practice_scenarios"("status");

-- CreateIndex
CREATE INDEX "usr_conversations_scenario_code_idx" ON "usr_conversations"("scenario_code");
