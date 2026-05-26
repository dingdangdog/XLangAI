import type { Component } from "vue";
import {
  ArchiveBoxIcon,
  ArrowDownTrayIcon,
  ChatBubbleLeftIcon,
  ChartBarIcon,
  CloudArrowUpIcon,
  Cog6ToothIcon,
  CpuChipIcon,
  DocumentTextIcon,
  GlobeAltIcon,
  MicrophoneIcon,
  ServerIcon,
  SparklesIcon,
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
    ],
  },
  {
    titleKey: "nav.groups.aiVoice",
    items: [
      { to: "/manage/llm-service-configs", labelKey: "nav.items.llmConfigs", icon: CpuChipIcon },
      { to: "/manage/stt-service-configs", labelKey: "nav.items.sttConfigs", icon: ServerIcon },
      { to: "/manage/tts-service-configs", labelKey: "nav.items.ttsConfigs", icon: MicrophoneIcon },
      { to: "/manage/voice-roles", labelKey: "nav.items.voiceRoles", icon: SparklesIcon },
      { to: "/manage/translate-service-configs", labelKey: "nav.items.translateConfigs", icon: GlobeAltIcon },
      { to: "/manage/object-storage-configs", labelKey: "nav.items.objectStorage", icon: CloudArrowUpIcon },
    ],
  },
  {
    titleKey: "nav.groups.system",
    items: [
      { to: "/manage/server-store", labelKey: "nav.items.serverStore", icon: ServerIcon },
      { to: "/manage/system-settings", labelKey: "nav.items.systemSettings", icon: Cog6ToothIcon },
      { to: "/manage/data-backup", labelKey: "nav.items.dataBackup", icon: ArrowDownTrayIcon },
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
