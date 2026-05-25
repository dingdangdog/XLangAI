/**
 * 语音角色管理列表接口附加字段（由 TTS 配置 / 语言表解析），不属于 voice_roles 表，更新时不得写入 Prisma。
 */
export const VOICE_ROLE_VIRTUAL_FIELDS = [
  "languageLabel",
  "ttsConfigLabel",
] as const;

export function stripVoiceRoleVirtualFields(
  data: Record<string, unknown>,
): Record<string, unknown> {
  const next = { ...data };
  for (const k of VOICE_ROLE_VIRTUAL_FIELDS) {
    delete next[k];
  }
  return next;
}
