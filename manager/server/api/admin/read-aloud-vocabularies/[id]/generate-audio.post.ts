import { generateReadAloudVocabularyAudio } from "~~/server/lib/readAloudAudio/generate";

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id?.trim()) {
    throw createError({ statusCode: 400, message: "缺少词汇 id" });
  }
  const body = await readBody<{ part?: string }>(event).catch(() => ({}));
  const rawPart = (body?.part ?? "both").trim().toLowerCase();
  const part =
    rawPart === "word" || rawPart === "sentence" || rawPart === "both"
      ? rawPart
      : "both";
  try {
    return await generateReadAloudVocabularyAudio(id, part);
  } catch (e) {
    const msg = e instanceof Error ? e.message : "生成音频失败";
    throw createError({ statusCode: 502, message: msg });
  }
});
