import { generateVoiceRolePreview } from "~~/server/lib/voicePreview/generate";

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id?.trim()) {
    throw createError({ statusCode: 400, message: "缺少语音角色 id" });
  }
  try {
    return await generateVoiceRolePreview(id);
  } catch (e) {
    const msg = e instanceof Error ? e.message : "生成试听失败";
    throw createError({ statusCode: 502, message: msg });
  }
});
