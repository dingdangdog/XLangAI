import { createError } from "h3";

/**
 * 语音角色管理列表接口附加字段（由 TTS / LLM / 语言表解析），不属于 voice_roles 表，更新时不得写入 Prisma。
 */
export const VOICE_ROLE_VIRTUAL_FIELDS = [
  "languageLabel",
  "ttsConfigLabel",
  "llmConfigLabel",
] as const;

export const VOICE_ROLE_SYNTHESIS_TYPES = [
  "tts",
  "native_audio_in_text",
  "native_audio_io",
] as const;

export type VoiceRoleSynthesisType = (typeof VOICE_ROLE_SYNTHESIS_TYPES)[number];

export function stripVoiceRoleVirtualFields(
  data: Record<string, unknown>,
): Record<string, unknown> {
  const next = { ...data };
  for (const k of VOICE_ROLE_VIRTUAL_FIELDS) {
    delete next[k];
  }
  return next;
}

function normalizeSynthesisType(raw: unknown): VoiceRoleSynthesisType {
  const s = String(raw ?? "tts").trim();
  if ((VOICE_ROLE_SYNTHESIS_TYPES as readonly string[]).includes(s)) {
    return s as VoiceRoleSynthesisType;
  }
  return "tts";
}

/** 创建/更新前校验并规范化 synthesisType 与关联字段 */
export function prepareVoiceRoleWrite(
  data: Record<string, unknown>,
): Record<string, unknown> {
  const next = stripVoiceRoleVirtualFields({ ...data });
  const st = normalizeSynthesisType(next.synthesisType);

  const voiceCode = String(next.voiceCode ?? "").trim();
  if (!voiceCode) {
    throw createError({ statusCode: 400, message: "voice_code is required" });
  }

  if (st === "tts") {
    const ttsId = String(next.ttsServiceConfigId ?? "").trim();
    if (!ttsId) {
      throw createError({
        statusCode: 400,
        message: "tts_service_config_id is required for synthesis type tts",
      });
    }
    next.synthesisType = st;
    next.ttsServiceConfigId = ttsId;
    next.llmServiceConfigId = null;
    return next;
  }

  const llmId = String(next.llmServiceConfigId ?? "").trim();
  if (!llmId) {
    throw createError({
      statusCode: 400,
      message: "llm_service_config_id is required for native synthesis types",
    });
  }
  next.synthesisType = st;
  next.llmServiceConfigId = llmId;
  next.ttsServiceConfigId = null;
  return next;
}
