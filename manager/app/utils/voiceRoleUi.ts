/** 与 server/utils/voiceRoleAdmin.ts 一致：列表接口附加字段，非表列 */
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

export function stripVoiceRoleVirtualFields(
  row: Record<string, unknown>,
): Record<string, unknown> {
  const o = { ...row };
  for (const k of VOICE_ROLE_VIRTUAL_FIELDS) {
    delete o[k];
  }
  return o;
}

export const VOICE_ROLE_VIRTUAL_FIELD_SET = new Set<string>(VOICE_ROLE_VIRTUAL_FIELDS);
