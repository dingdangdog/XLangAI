import { batchCreateReadAloudVocabulary } from "~~/server/lib/readAloudVocabulary/batchCreate";

export default defineEventHandler(async (event) => {
  const body = await readBody<{
    categoryId?: string;
    languageId?: string;
    voiceRoleId?: string;
    items?: Array<{ word?: string; exampleSentence?: string }>;
    status?: string;
    remark?: string | null;
  }>(event);

  try {
    return await batchCreateReadAloudVocabulary({
      categoryId: String(body?.categoryId ?? ""),
      languageId: String(body?.languageId ?? ""),
      voiceRoleId: String(body?.voiceRoleId ?? ""),
      items: (body?.items ?? []).map((it) => ({
        word: String(it.word ?? ""),
        exampleSentence: String(it.exampleSentence ?? ""),
      })),
      status: body?.status,
      remark: body?.remark ?? "LLM batch import",
    });
  } catch (e) {
    if (e && typeof e === "object" && "statusCode" in e) throw e;
    const msg = e instanceof Error ? e.message : "批量导入失败";
    throw createError({ statusCode: 502, message: msg });
  }
});
