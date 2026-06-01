import type { Component } from "vue";
import {
  ArchiveBoxIcon,
  ChartBarIcon,
  ChatBubbleLeftIcon,
  Cog6ToothIcon,
  CpuChipIcon,
  DocumentTextIcon,
  GlobeAltIcon,
  UserGroupIcon,
  UsersIcon,
} from "@heroicons/vue/24/outline";

export interface AdminNavItem {
  to: string;
  labelKey: string;
  icon: Component;
}

export interface AdminNavGroup {
  titleKey: string;
  items: AdminNavItem[];
}

export const ADMIN_NAV_GROUPS: AdminNavGroup[] = [
  {
    titleKey: "nav.groups.overview",
    items: [{ to: "/", labelKey: "nav.items.dashboard", icon: ChartBarIcon }],
  },
  {
    titleKey: "nav.groups.business",
    items: [
      { to: "/manage/users", labelKey: "nav.items.users", icon: UsersIcon },
      { to: "/manage/conversations", labelKey: "nav.items.conversations", icon: ChatBubbleLeftIcon },
      { to: "/manage/messages", labelKey: "nav.items.messages", icon: ChatBubbleLeftIcon },
      { to: "/manage/membership-tiers", labelKey: "nav.items.membershipTiers", icon: UserGroupIcon },
      { to: "/manage/languages", labelKey: "nav.items.languages", icon: GlobeAltIcon },
      { to: "/manage/prompt-templates", labelKey: "nav.items.promptTemplates", icon: DocumentTextIcon },
      { to: "/manage/practice-scenarios", labelKey: "nav.items.practiceScenarios", icon: DocumentTextIcon },
      { to: "/manage/read-aloud-categories", labelKey: "nav.items.readAloudCategories", icon: DocumentTextIcon },
      { to: "/manage/read-aloud-vocabularies", labelKey: "nav.items.readAloudVocabularies", icon: DocumentTextIcon },
    ],
  },
  {
    titleKey: "nav.groups.system",
    items: [
      { to: "/manage/ai-settings", labelKey: "nav.items.aiSettings", icon: CpuChipIcon },
      { to: "/manage/system-settings", labelKey: "nav.items.systemSettings", icon: Cog6ToothIcon },
      { to: "/manage/backups", labelKey: "nav.items.deletionRecords", icon: ArchiveBoxIcon },
    ],
  },
];

/** 旧删除记录子路由重定向到统一页时，侧栏仍高亮对应菜单 */
export const BACKUP_LEGACY_PATHS = [
  "/manage/users-backup",
  "/manage/conversations-backup",
  "/manage/messages-backup",
  "/manage/user-usage-backup",
] as const;

/** 旧 AI 服务配置子路由重定向时，侧栏仍高亮 AI 设置菜单 */
export const AI_SETTINGS_LEGACY_PATHS = [
  "/manage/llm-service-configs",
  "/manage/stt-service-configs",
  "/manage/tts-service-configs",
  "/manage/translate-service-configs",
  "/manage/voice-roles",
  "/manage/object-storage-configs",
] as const;

/** 旧系统管理子路由重定向时，侧栏仍高亮系统设置菜单 */
export const SYSTEM_SETTINGS_LEGACY_PATHS = [
  "/manage/server-store",
  "/manage/data-backup",
  "/manage/sms-service-configs",
] as const;

export function useAdminNav() {
  const { t } = useI18n();

  const groups = computed(() =>
    ADMIN_NAV_GROUPS.map((group) => ({
      title: t(group.titleKey),
      titleKey: group.titleKey,
      items: group.items.map((item) => ({
        to: item.to,
        label: t(item.labelKey),
        icon: item.icon,
      })),
    })),
  );

  return { groups };
}
