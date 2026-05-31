-- CreateTable
CREATE TABLE "sys_scenario_opening_lines" (
    "id" TEXT NOT NULL,
    "scenario_code" VARCHAR(32) NOT NULL,
    "language_code" VARCHAR(10) NOT NULL,
    "template" VARCHAR(500) NOT NULL,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_scenario_opening_lines_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "sys_scenario_opening_lines_scenario_code_idx" ON "sys_scenario_opening_lines"("scenario_code");

-- CreateIndex
CREATE INDEX "sys_scenario_opening_lines_language_code_idx" ON "sys_scenario_opening_lines"("language_code");

-- CreateIndex
CREATE INDEX "sys_scenario_opening_lines_status_idx" ON "sys_scenario_opening_lines"("status");

-- CreateIndex
CREATE UNIQUE INDEX "sys_scenario_opening_lines_scenario_code_language_code_key" ON "sys_scenario_opening_lines"("scenario_code", "language_code");
