-- AlterTable
ALTER TABLE "usr_users" ADD COLUMN     "subscription_expires_at" TIMESTAMPTZ(6),
ADD COLUMN     "token_balance" BIGINT NOT NULL DEFAULT 0;

-- CreateTable
CREATE TABLE "sys_billing_products" (
    "id" VARCHAR(36) NOT NULL,
    "code" VARCHAR(64) NOT NULL,
    "kind" VARCHAR(32) NOT NULL,
    "ios_product_id" VARCHAR(200) NOT NULL,
    "android_product_id" VARCHAR(200) NOT NULL,
    "tier_code" VARCHAR(20),
    "token_grant" BIGINT NOT NULL DEFAULT 0,
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_billing_products_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "usr_store_transactions" (
    "id" VARCHAR(36) NOT NULL,
    "user_id" VARCHAR(36) NOT NULL,
    "platform" VARCHAR(16) NOT NULL,
    "store_transaction_id" VARCHAR(512) NOT NULL,
    "product_code" VARCHAR(64) NOT NULL,
    "raw_payload" TEXT,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "usr_store_transactions_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "sys_billing_products_code_key" ON "sys_billing_products"("code");

-- CreateIndex
CREATE INDEX "sys_billing_products_status_idx" ON "sys_billing_products"("status");

-- CreateIndex
CREATE INDEX "usr_store_transactions_user_id_idx" ON "usr_store_transactions"("user_id");

-- CreateIndex
CREATE UNIQUE INDEX "usr_store_transactions_platform_store_transaction_id_key" ON "usr_store_transactions"("platform", "store_transaction_id");
