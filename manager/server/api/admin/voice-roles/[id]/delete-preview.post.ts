import { deleteVoiceRolePreview } from "../../../../lib/voicePreview/delete";

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id?.trim()) {
    throw createError({ statusCode: 400, message: "缺少语音角色 id" });
  }
  try {
    return await deleteVoiceRolePreview(id);
  } catch (e) {
    const msg = e instanceof Error ? e.message : "删除试听失败";
    const statusCode =
      msg === "Voice role not found"
        ? 404
        : msg === "Voice role has no preview audio"
          ? 400
          : 502;
    throw createError({ statusCode, message: msg });
  }
});
