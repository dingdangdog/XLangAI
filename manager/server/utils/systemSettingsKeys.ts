/** 与 server/internal/settings/keys.go 保持一致 */

export const SYSTEM_SETTING_KEYS = [
  "auth.password.enabled",
  "auth.password.register_enabled",
  "auth.sms.enabled",
  "auth.sms.register_enabled",
  "media.user_recording.storage",
  "media.assistant_tts.storage",
  "media.avatar.storage",
] as const;

export type SystemSettingKey = (typeof SYSTEM_SETTING_KEYS)[number];

export const VALUE_TYPES = ["bool", "string"] as const;

export const MEDIA_STORAGE_VALUES = ["client", "server", "cloud"] as const;

export const KEY_META: Record<
  SystemSettingKey,
  { label: string; valueType: "bool" | "string"; enumValues?: readonly string[] }
> = {
  "auth.password.enabled": { label: "账号密码登录", valueType: "bool" },
  "auth.password.register_enabled": { label: "账号密码注册", valueType: "bool" },
  "auth.sms.enabled": { label: "短信验证码登录", valueType: "bool" },
  "auth.sms.register_enabled": { label: "短信验证码注册", valueType: "bool" },
  "media.user_recording.storage": {
    label: "用户录音存储",
    valueType: "string",
    enumValues: MEDIA_STORAGE_VALUES,
  },
  "media.assistant_tts.storage": {
    label: "AI 回复音频存储",
    valueType: "string",
    enumValues: ["server", "cloud"],
  },
  "media.avatar.storage": {
    label: "头像存储",
    valueType: "string",
    enumValues: ["server", "cloud"],
  },
};

/** 由专用设置页管理，不在「系统变量」Tab 展示或编辑 */
export const INTERNAL_SYSTEM_SETTING_KEYS = ["official_server_store.config"] as const;

export type InternalSystemSettingKey = (typeof INTERNAL_SYSTEM_SETTING_KEYS)[number];

const KEY_SET = new Set<string>(SYSTEM_SETTING_KEYS);
const INTERNAL_KEY_SET = new Set<string>(INTERNAL_SYSTEM_SETTING_KEYS);

export function isKnownSystemSettingKey(key: string): key is SystemSettingKey {
  return KEY_SET.has(key);
}

export function isInternalSystemSettingKey(key: string): key is InternalSystemSettingKey {
  return INTERNAL_KEY_SET.has(key);
}
