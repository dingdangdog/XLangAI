-- AlterTable
ALTER TABLE "sys_languages" ADD COLUMN     "preview_sample_text" VARCHAR(500);

-- AlterTable
ALTER TABLE "sys_voice_roles" ADD COLUMN     "preview_audio_url" VARCHAR(500),
ADD COLUMN     "preview_generated_at" TIMESTAMPTZ(6),
ADD COLUMN     "preview_local_filename" VARCHAR(200);
