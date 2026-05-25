-- CreateTable
CREATE TABLE "sys_system_settings" (
    "id" TEXT NOT NULL,
    "key" VARCHAR(128) NOT NULL,
    "value" TEXT NOT NULL,
    "value_type" VARCHAR(20) NOT NULL DEFAULT 'string',
    "description" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_system_settings_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "sys_system_settings_key_key" ON "sys_system_settings"("key");

-- CreateIndex
CREATE INDEX "sys_system_settings_key_idx" ON "sys_system_settings"("key");
