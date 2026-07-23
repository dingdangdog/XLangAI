import {
  batchGenerateReadAloudVocabularyAudio,
  type BatchGenerateAudioInput,
} from "~~/server/lib/readAloudAudio/batchGenerate";
import type { ReadAloudAudioPart } from "~~/server/lib/readAloudAudio/generate";

export default defineEventHandler(async (event) => {
  const body = await readBody<{
    ids?: string[];
    categoryId?: string;
    languageId?: string;
    onlyMissing?: boolean;
    part?: string;
  }>(event).catch(() => ({} as Record<string, unknown>));

  const rawPart = String(body?.part ?? "both")
    .trim()
    .toLowerCase();
  const part: ReadAloudAudioPart =
    rawPart === "word" || rawPart === "sentence" || rawPart === "both"
      ? rawPart
      : "both";

  const input: BatchGenerateAudioInput = {
    ids: Array.isArray(body?.ids) ? body.ids.map((id) => String(id)) : undefined,
    categoryId: body?.categoryId != null ? String(body.categoryId) : undefined,
    languageId: body?.languageId != null ? String(body.languageId) : undefined,
    onlyMissing: body?.onlyMissing !== false,
    part,
  };

  try {
    return await batchGenerateReadAloudVocabularyAudio(input);
  } catch (e) {
    if (e && typeof e === "object" && "statusCode" in e) throw e;
    const msg = e instanceof Error ? e.message : "批量生成音频失败";
    throw createError({ statusCode: 502, message: msg });
  }
});
