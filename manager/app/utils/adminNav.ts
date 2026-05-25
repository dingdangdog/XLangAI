import type { Component } from "vue";
import {
  ArchiveBoxIcon,
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
  label: string;
  icon: Component;
}

export interface AdminNavGroup {
  title: string;
  items: AdminNavItem[];
}

export const ADMIN_NAV_GROUPS: AdminNavGroup[] = [
  {
    title: "概览",
    items: [{ to: "/", label: "数据概览", icon: ChartBarIcon }],
  },
  {
    title: "业务管理",
    items: [
      { to: "/manage/users", label: "用户", icon: UsersIcon },
      { to: "/manage/conversations", label: "会话", icon: ChatBubbleLeftIcon },
      { to: "/manage/messages", label: "消息", icon: ChatBubbleLeftIcon },
      { to: "/manage/membership-tiers", label: "会员管理", icon: UserGroupIcon },
      { to: "/manage/languages", label: "语言管理", icon: GlobeAltIcon },
      { to: "/manage/prompt-templates", label: "系统提示词", icon: DocumentTextIcon },
    ],
  },
  {
    title: "AI 与语音",
    items: [
      { to: "/manage/llm-service-configs", label: "LLM 服务配置", icon: CpuChipIcon },
      { to: "/manage/stt-service-configs", label: "STT 服务配置", icon: ServerIcon },
      { to: "/manage/tts-service-configs", label: "TTS 服务配置", icon: MicrophoneIcon },
      { to: "/manage/voice-roles", label: "TTS 语音角色", icon: SparklesIcon },
      { to: "/manage/translate-service-configs", label: "翻译服务配置", icon: GlobeAltIcon },
      { to: "/manage/object-storage-configs", label: "对象存储 / 图床", icon: CloudArrowUpIcon },
    ],
  },
  {
    title: "系统管理",
    items: [
      { to: "/manage/server-store", label: "服务器商店", icon: ServerIcon },
      { to: "/manage/system-settings", label: "系统变量", icon: Cog6ToothIcon },
      { to: "/manage/backups", label: "备份归档", icon: ArchiveBoxIcon },
    ],
  },
];

/** 旧备份子路由重定向到统一页时，侧栏仍高亮「备份归档」 */
export const BACKUP_LEGACY_PATHS = [
  "/manage/users-backup",
  "/manage/conversations-backup",
  "/manage/messages-backup",
  "/manage/user-usage-backup",
] as const;
