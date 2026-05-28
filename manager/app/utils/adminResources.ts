/** 与 server/utils/adminResource.ts 中 slug 一致 */
export const ADMIN_RESOURCE_SLUGS = [
  "languages",
  "llm-service-configs",
  "stt-service-configs",
  "translate-service-configs",
  "object-storage-configs",
  "sms-service-configs",
  "system-settings",
  "tts-service-configs",
  "voice-roles",
  "prompt-templates",
  "membership-tiers",
  "users",
  "user-usage",
  "conversations",
  "messages",
  "users-backup",
  "conversations-backup",
  "messages-backup",
  "user-usage-backup",
] as const;

export type AdminResourceSlug = (typeof ADMIN_RESOURCE_SLUGS)[number];

const READ_ONLY = new Set<string>([
  "users-backup",
  "conversations-backup",
  "messages-backup",
  "user-usage-backup",
]);

export function isAdminReadOnly(slug: string): boolean {
  return READ_ONLY.has(slug);
}

export function adminResourceTitle(slug: string): string {
  const map: Record<string, string> = {
    languages: "语言",
    "llm-service-configs": "LLM 服务配置",
    "stt-service-configs": "STT 服务配置",
    "translate-service-configs": "翻译服务配置",
    "object-storage-configs": "对象存储 / 图床",
    "sms-service-configs": "短信服务配置",
    "system-settings": "系统变量",
    "tts-service-configs": "TTS 服务配置",
    "voice-roles": "语音角色",
    "prompt-templates": "提示词模板",
    "membership-tiers": "会员等级",
    users: "用户",
    "user-usage": "用户用量",
    conversations: "会话",
    messages: "消息",
    "users-backup": "用户注销备份",
    "conversations-backup": "会话备份",
    "messages-backup": "消息备份",
    "user-usage-backup": "用量备份",
  };
  return map[slug] ?? slug;
}
