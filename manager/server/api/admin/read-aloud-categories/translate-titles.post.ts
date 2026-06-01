import { translateReadAloudCategoryTitles } from "~~/server/lib/readAloudCategory/translateTitles";

export default defineEventHandler(async (event) => {
  const body = await readBody<{
    sourceLanguageId?: string;
    title?: string;
    llmServiceConfigId?: string;
  }>(event);

  try {
    return await translateReadAloudCategoryTitles({
      sourceLanguageId: String(body?.sourceLanguageId ?? ""),
      title: String(body?.title ?? ""),
      llmServiceConfigId: body?.llmServiceConfigId,
    });
  } catch (e) {
    if (e && typeof e === "object" && "statusCode" in e) throw e;
    const msg = e instanceof Error ? e.message : "翻译失败";
    throw createError({ statusCode: 502, message: msg });
  }
});
