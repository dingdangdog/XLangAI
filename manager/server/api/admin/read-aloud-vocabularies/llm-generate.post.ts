import { generateReadAloudVocabularyWithLlm } from "~~/server/lib/readAloudVocabulary/llmGenerate";

export default defineEventHandler(async (event) => {
  const body = await readBody<{
    llmServiceConfigId?: string;
    categoryId?: string;
    languageId?: string;
    count?: number;
    extraInstructions?: string | null;
  }>(event);

  try {
    return await generateReadAloudVocabularyWithLlm({
      llmServiceConfigId: String(body?.llmServiceConfigId ?? ""),
      categoryId: String(body?.categoryId ?? ""),
      languageId: String(body?.languageId ?? ""),
      count: body?.count,
      extraInstructions: body?.extraInstructions ?? null,
    });
  } catch (e) {
    if (e && typeof e === "object" && "statusCode" in e) throw e;
    const msg = e instanceof Error ? e.message : "LLM 生成失败";
    throw createError({ statusCode: 502, message: msg });
  }
});
