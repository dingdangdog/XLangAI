import { createError } from "h3";
import type { AdminListOrderBy } from "./adminListOrderBy";
import {
  orderByCancelledDesc,
  orderByCreatedDesc,
  orderByKeyAsc,
  orderBySortOrderCode,
  orderByVoiceRole,
  orderByUsageDateDesc,
} from "./adminListOrderBy";

export type ResourceSlug = (typeof RESOURCE_SLUGS)[number];

export const RESOURCE_SLUGS = [
  "languages",
  "llm-service-configs",
  "stt-service-configs",
  "translate-service-configs",
  "object-storage-configs",
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

const READ_ONLY = new Set<string>([
  "users-backup",
  "conversations-backup",
  "messages-backup",
  "user-usage-backup",
]);

type DelegateKey =
  | "language"
  | "sysLlmServiceConfig"
  | "sysSttServiceConfig"
  | "sysTranslateServiceConfig"
  | "sysObjectStorageConfig"
  | "sysSystemSetting"
  | "ttsServiceConfig"
  | "voiceRole"
  | "promptTemplate"
  | "membershipTier"
  | "user"
  | "userUsage"
  | "conversation"
  | "message"
  | "usersBackup"
  | "conversationsBackup"
  | "messagesBackup"
  | "userUsageBackup";

export const RESOURCE_META: Record<
  ResourceSlug,
  {
    delegate: DelegateKey;
    softListFilter: boolean;
    orderBy: AdminListOrderBy;
  }
> = {
  languages: {
    delegate: "language",
    softListFilter: false,
    orderBy: orderBySortOrderCode(),
  },
  "llm-service-configs": {
    delegate: "sysLlmServiceConfig",
    softListFilter: false,
    orderBy: orderBySortOrderCode(),
  },
  "stt-service-configs": {
    delegate: "sysSttServiceConfig",
    softListFilter: false,
    orderBy: orderBySortOrderCode(),
  },
  "translate-service-configs": {
    delegate: "sysTranslateServiceConfig",
    softListFilter: false,
    orderBy: orderBySortOrderCode(),
  },
  "object-storage-configs": {
    delegate: "sysObjectStorageConfig",
    softListFilter: false,
    orderBy: orderBySortOrderCode(),
  },
  "system-settings": {
    delegate: "sysSystemSetting",
    softListFilter: false,
    orderBy: orderByKeyAsc(),
  },
  "tts-service-configs": {
    delegate: "ttsServiceConfig",
    softListFilter: false,
    orderBy: orderBySortOrderCode(),
  },
  "voice-roles": {
    delegate: "voiceRole",
    softListFilter: false,
    orderBy: orderByVoiceRole(),
  },
  "prompt-templates": {
    delegate: "promptTemplate",
    softListFilter: false,
    orderBy: orderBySortOrderCode(),
  },
  "membership-tiers": {
    delegate: "membershipTier",
    softListFilter: false,
    orderBy: orderBySortOrderCode(),
  },
  users: {
    delegate: "user",
    softListFilter: true,
    orderBy: orderByCreatedDesc(),
  },
  "user-usage": {
    delegate: "userUsage",
    softListFilter: false,
    orderBy: orderByUsageDateDesc(),
  },
  conversations: {
    delegate: "conversation",
    softListFilter: true,
    orderBy: orderByCreatedDesc(),
  },
  messages: {
    delegate: "message",
    softListFilter: true,
    orderBy: orderByCreatedDesc(),
  },
  "users-backup": {
    delegate: "usersBackup",
    softListFilter: false,
    orderBy: orderByCancelledDesc(),
  },
  "conversations-backup": {
    delegate: "conversationsBackup",
    softListFilter: false,
    orderBy: orderByCancelledDesc(),
  },
  "messages-backup": {
    delegate: "messagesBackup",
    softListFilter: false,
    orderBy: orderByCancelledDesc(),
  },
  "user-usage-backup": {
    delegate: "userUsageBackup",
    softListFilter: false,
    orderBy: orderByCancelledDesc(),
  },
};

export function assertResource(slug: string | undefined): asserts slug is ResourceSlug {
  if (!slug || !(RESOURCE_SLUGS as readonly string[]).includes(slug)) {
    throw createError({ statusCode: 404, message: "Unknown resource" });
  }
}

export function isReadOnlyResource(slug: ResourceSlug): boolean {
  return READ_ONLY.has(slug);
}

export function getDelegate(prisma: Record<string, unknown>, slug: ResourceSlug) {
  const key = RESOURCE_META[slug].delegate;
  const d = prisma[key];
  if (!d || typeof d !== "object") {
    throw createError({ statusCode: 500, message: "Prisma delegate missing" });
  }
  return d as {
    findMany: (args: object) => Promise<unknown[]>;
    count: (args: object) => Promise<number>;
    create: (args: { data: object }) => Promise<unknown>;
    update: (args: { where: object; data: object }) => Promise<unknown>;
    delete: (args: { where: object }) => Promise<unknown>;
  };
}
