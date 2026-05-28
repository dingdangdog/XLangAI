/** 系统变量左侧菜单分组（与 server/utils/systemSettingsKeys.ts 键顺序一致） */

export const SYSTEM_SETTING_GROUPS = [
  {
    id: "auth",
    labelKey: "pages.systemSettings.groups.auth",
    descriptionKey: "pages.systemSettings.groups.authDesc",
    keys: [
      "auth.password.enabled",
      "auth.password.register_enabled",
      "auth.sms.enabled",
      "auth.sms.register_enabled",
      "auth.google.enabled",
      "auth.google.register_enabled",
      "auth.apple.enabled",
      "auth.apple.register_enabled",
    ],
  },
  {
    id: "media",
    labelKey: "pages.systemSettings.groups.media",
    descriptionKey: "pages.systemSettings.groups.mediaDesc",
    keys: [
      "media.user_recording.storage",
      "media.assistant_tts.storage",
      "media.avatar.storage",
    ],
  },
] as const;

export type SystemSettingGroupId = (typeof SYSTEM_SETTING_GROUPS)[number]["id"];

export function allSystemSettingKeys(): string[] {
  return SYSTEM_SETTING_GROUPS.flatMap((g) => [...g.keys]);
}
