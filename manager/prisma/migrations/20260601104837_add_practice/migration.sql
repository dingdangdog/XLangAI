-- CreateTable
CREATE TABLE "sys_read_aloud_categories" (
    "id" TEXT NOT NULL,
    "code" VARCHAR(32) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "name_en" VARCHAR(100),
    "icon" VARCHAR(50),
    "description" VARCHAR(500),
    "description_en" VARCHAR(500),
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_read_aloud_categories_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "sys_read_aloud_category_locales" (
    "id" TEXT NOT NULL,
    "category_id" VARCHAR(36) NOT NULL,
    "language_id" VARCHAR(36) NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "description" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_read_aloud_category_locales_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "sys_read_aloud_vocabularies" (
    "id" TEXT NOT NULL,
    "category_id" VARCHAR(36) NOT NULL,
    "language_id" VARCHAR(36) NOT NULL,
    "word" VARCHAR(200) NOT NULL,
    "example_sentence" VARCHAR(500) NOT NULL,
    "voice_role_id" VARCHAR(36) NOT NULL,
    "word_audio_url" VARCHAR(500),
    "word_audio_local_filename" VARCHAR(200),
    "word_audio_generated_at" TIMESTAMPTZ(6),
    "sentence_audio_url" VARCHAR(500),
    "sentence_audio_local_filename" VARCHAR(200),
    "sentence_audio_generated_at" TIMESTAMPTZ(6),
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "status" VARCHAR(20) NOT NULL DEFAULT 'active',
    "remark" VARCHAR(500),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "sys_read_aloud_vocabularies_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "usr_read_aloud_sessions" (
    "id" TEXT NOT NULL,
    "user_id" VARCHAR(36) NOT NULL,
    "category_id" VARCHAR(36) NOT NULL,
    "language_id" VARCHAR(36) NOT NULL,
    "status" VARCHAR(20) NOT NULL DEFAULT 'in_progress',
    "total_items" INTEGER NOT NULL DEFAULT 0,
    "completed_items" INTEGER NOT NULL DEFAULT 0,
    "average_score" INTEGER,
    "started_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "completed_at" TIMESTAMPTZ(6),
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "usr_read_aloud_sessions_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "usr_read_aloud_attempts" (
    "id" TEXT NOT NULL,
    "session_id" VARCHAR(36) NOT NULL,
    "vocabulary_id" VARCHAR(36) NOT NULL,
    "part" VARCHAR(20) NOT NULL,
    "reference_text" VARCHAR(500) NOT NULL,
    "transcript" TEXT NOT NULL,
    "score" INTEGER NOT NULL,
    "match_detail" TEXT,
    "duration_ms" INTEGER,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "usr_read_aloud_attempts_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "sys_read_aloud_categories_code_key" ON "sys_read_aloud_categories"("code");

-- CreateIndex
CREATE INDEX "sys_read_aloud_categories_code_idx" ON "sys_read_aloud_categories"("code");

-- CreateIndex
CREATE INDEX "sys_read_aloud_categories_status_idx" ON "sys_read_aloud_categories"("status");

-- CreateIndex
CREATE INDEX "sys_read_aloud_category_locales_category_id_idx" ON "sys_read_aloud_category_locales"("category_id");

-- CreateIndex
CREATE INDEX "sys_read_aloud_category_locales_language_id_idx" ON "sys_read_aloud_category_locales"("language_id");

-- CreateIndex
CREATE UNIQUE INDEX "sys_read_aloud_category_locales_category_id_language_id_key" ON "sys_read_aloud_category_locales"("category_id", "language_id");

-- CreateIndex
CREATE INDEX "sys_read_aloud_vocabularies_category_id_idx" ON "sys_read_aloud_vocabularies"("category_id");

-- CreateIndex
CREATE INDEX "sys_read_aloud_vocabularies_language_id_idx" ON "sys_read_aloud_vocabularies"("language_id");

-- CreateIndex
CREATE INDEX "sys_read_aloud_vocabularies_category_id_language_id_idx" ON "sys_read_aloud_vocabularies"("category_id", "language_id");

-- CreateIndex
CREATE INDEX "sys_read_aloud_vocabularies_status_idx" ON "sys_read_aloud_vocabularies"("status");

-- CreateIndex
CREATE INDEX "usr_read_aloud_sessions_user_id_idx" ON "usr_read_aloud_sessions"("user_id");

-- CreateIndex
CREATE INDEX "usr_read_aloud_sessions_user_id_created_at_idx" ON "usr_read_aloud_sessions"("user_id", "created_at");

-- CreateIndex
CREATE INDEX "usr_read_aloud_sessions_status_idx" ON "usr_read_aloud_sessions"("status");

-- CreateIndex
CREATE INDEX "usr_read_aloud_attempts_session_id_idx" ON "usr_read_aloud_attempts"("session_id");

-- CreateIndex
CREATE INDEX "usr_read_aloud_attempts_vocabulary_id_idx" ON "usr_read_aloud_attempts"("vocabulary_id");
