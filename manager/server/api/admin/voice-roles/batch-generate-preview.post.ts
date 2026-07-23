import { batchGenerateVoiceRolePreviews } from "~~/server/lib/voicePreview/batchGenerate";

export default defineEventHandler(async (event) => {
  const body = await readBody<{
    ids?: string[];
    onlyMissing?: boolean;
  }>(event).catch(() => ({} as Record<string, unknown>));

  try {
    return await batchGenerateVoiceRolePreviews({
      ids: Array.isArray(body?.ids) ? body.ids.map((id) => String(id)) : undefined,
      onlyMissing: body?.onlyMissing !== false,
    });
  } catch (e) {
    if (e && typeof e === "object" && "statusCode" in e) throw e;
    const msg = e instanceof Error ? e.message : "批量生成试听失败";
    throw createError({ statusCode: 502, message: msg });
  }
});
