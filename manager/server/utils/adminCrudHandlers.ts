import type { H3Event } from "h3";
import prisma from "../lib/prisma";
import { getPagination } from "./pagination";
import {
  assertResource,
  getDelegate,
  isReadOnlyResource,
  RESOURCE_META,
  type ResourceSlug,
} from "./adminResource";
import { prepareVoiceRoleWrite, stripVoiceRoleVirtualFields } from "./voiceRoleAdmin";
import { attachUserListFields, prepareUserAdminWriteData, redactUserRecord } from "./userAdmin";
import { attachServiceConfigUsageFields } from "./serviceUsageAdmin";
import { prepareObjectStorageServiceConfigWrite } from "./objectStorageServiceConfigAdmin";
import { prepareSmsServiceConfigWrite } from "./smsServiceConfigAdmin";
import { prepareLlmServiceConfigWrite } from "./llmServiceConfigAdmin";
import { prepareTtsServiceConfigWrite } from "./ttsServiceConfigAdmin";
import { prepareTranslateServiceConfigWrite } from "./translateServiceConfigAdmin";
import { prepareSttServiceConfigWrite } from "./sttServiceConfigAdmin";
import { assertEditableSystemSettingKey, prepareSystemSettingWrite } from "./systemSettingsAdmin";
import { INTERNAL_SYSTEM_SETTING_KEYS } from "./systemSettingsKeys";
import { jsonSafe } from "./jsonSafe";
import { listVoiceRolesForAdmin } from "./voiceRoleList";

async function attachVoiceRoleListFields(
  rows: Record<string, unknown>[],
): Promise<Record<string, unknown>[]> {
  const ttsIds = [
    ...new Set(
      rows
        .map((r) => r.ttsServiceConfigId)
        .filter((x): x is string => typeof x === "string" && x.length > 0),
    ),
  ];
  const llmIds = [
    ...new Set(
      rows
        .map((r) => r.llmServiceConfigId)
        .filter((x): x is string => typeof x === "string" && x.length > 0),
    ),
  ];
  const langIds = [
    ...new Set(
      rows
        .map((r) => r.languageId)
        .filter((x): x is string => typeof x === "string" && x.length > 0),
    ),
  ];
  const [ttsList, llmList, langList] = await Promise.all([
    ttsIds.length
      ? prisma.ttsServiceConfig.findMany({ where: { id: { in: ttsIds } } })
      : Promise.resolve([]),
    llmIds.length
      ? prisma.sysLlmServiceConfig.findMany({ where: { id: { in: llmIds } } })
      : Promise.resolve([]),
    langIds.length
      ? prisma.language.findMany({ where: { id: { in: langIds } } })
      : Promise.resolve([]),
  ]);
  const ttsById = new Map(ttsList.map((t) => [t.id, t]));
  const llmById = new Map(llmList.map((l) => [l.id, l]));
  const langById = new Map(langList.map((l) => [l.id, l]));
  return rows.map((r) => {
    const tid = r.ttsServiceConfigId;
    const llid = r.llmServiceConfigId;
    const lid = r.languageId;
    const t = typeof tid === "string" ? ttsById.get(tid) : undefined;
    const llm = typeof llid === "string" ? llmById.get(llid) : undefined;
    const l = typeof lid === "string" ? langById.get(lid) : undefined;
    const languageLabel = l
      ? [l.code, l.name].filter((x) => x != null && String(x).length > 0).join(" · ")
      : "";
    const ttsConfigLabel = t ? t.name : "";
    const llmConfigLabel = llm ? llm.name : "";
    return {
      ...r,
      languageLabel,
      ttsConfigLabel,
      llmConfigLabel,
    };
  });
}

function omitReadonly(data: Record<string, unknown>) {
  const next = { ...data };
  delete next.id;
  delete next.createdAt;
  delete next.updatedAt;
  delete next.cancelledAt;
  return next;
}

const SOFT_DELETE: Record<string, boolean> = {
  users: true,
  conversations: true,
  messages: true,
};

export async function adminListHandler(event: H3Event, resource: ResourceSlug) {
  assertResource(resource);
  const meta = RESOURCE_META[resource];
  const delegate = getDelegate(prisma as unknown as Record<string, unknown>, resource);
  const { skip, take, page, pageSize } = getPagination(event);
  const q = getQuery(event);

  const where: Record<string, unknown> = {};
  if (meta.softListFilter && q.includeDeleted !== "1" && q.includeDeleted !== "true") {
    where.deletedAt = null;
  }
  if (resource === "messages" && q.conversationId) {
    where.conversationId = String(q.conversationId);
  }
  if (resource === "conversations" && q.userId) {
    where.userId = String(q.userId);
  }
  if (resource === "user-usage" && q.userId) {
    where.userId = String(q.userId);
  }
  if (
    (resource === "users-backup" ||
      resource === "conversations-backup" ||
      resource === "messages-backup" ||
      resource === "user-usage-backup") &&
    q.backupBatch
  ) {
    where.backupBatch = String(q.backupBatch);
  }
  if (resource === "system-settings") {
    where.key = { notIn: [...INTERNAL_SYSTEM_SETTING_KEYS] };
  }

  const [rawItems, total] =
    resource === "voice-roles"
      ? await Promise.all([
          listVoiceRolesForAdmin({ skip, take }),
          prisma.voiceRole.count({ where }),
        ])
      : await Promise.all([
          delegate.findMany({
            where,
            skip,
            take,
            orderBy: meta.orderBy,
          }),
          delegate.count({ where }),
        ]);

  let items = rawItems as Record<string, unknown>[];
  if (resource === "voice-roles") {
    items = await attachVoiceRoleListFields(items);
  }
  if (resource === "users") {
    items = await attachUserListFields(items);
  }
  if (resource === "llm-service-configs") {
    items = await attachServiceConfigUsageFields(items, "llm");
  }
  if (resource === "tts-service-configs") {
    items = await attachServiceConfigUsageFields(items, "tts");
  }
  if (resource === "translate-service-configs") {
    items = await attachServiceConfigUsageFields(items, "translate");
  }

  return jsonSafe({ items, total, page, pageSize });
}

export async function adminCreateHandler(event: H3Event, resource: ResourceSlug) {
  assertResource(resource);
  if (isReadOnlyResource(resource)) {
    throw createError({ statusCode: 405, message: "Read-only" });
  }
  const body = (await readBody(event)) as Record<string, unknown> | null;
  if (!body || typeof body !== "object") {
    throw createError({ statusCode: 400, message: "Invalid body" });
  }
  const delegate = getDelegate(prisma as unknown as Record<string, unknown>, resource);
  let data = omitReadonly(body);
  if (resource === "users") {
    data = await prepareUserAdminWriteData(data, "create");
  }
  if (resource === "translate-service-configs") {
    data = await prepareTranslateServiceConfigWrite(data);
  }
  if (resource === "stt-service-configs") {
    data = await prepareSttServiceConfigWrite(data);
  }
  if (resource === "object-storage-configs") {
    data = await prepareObjectStorageServiceConfigWrite(data);
  }
  if (resource === "sms-service-configs") {
    data = await prepareSmsServiceConfigWrite(data);
  }
  if (resource === "llm-service-configs") {
    data = await prepareLlmServiceConfigWrite(data);
  }
  if (resource === "tts-service-configs") {
    data = await prepareTtsServiceConfigWrite(data);
  }
  if (resource === "voice-roles") {
    data = prepareVoiceRoleWrite(data);
  }
  if (resource === "system-settings") {
    data = prepareSystemSettingWrite(data, "create");
  }
  const created = (await delegate.create({ data })) as Record<string, unknown>;
  const row = resource === "users" ? redactUserRecord(created) : created;
  return jsonSafe(row);
}

export async function adminUpdateHandler(event: H3Event, resource: ResourceSlug) {
  const id = getRouterParam(event, "id");
  assertResource(resource);
  if (!id) {
    throw createError({ statusCode: 400, message: "Missing id" });
  }
  if (isReadOnlyResource(resource)) {
    throw createError({ statusCode: 405, message: "Read-only" });
  }
  const body = (await readBody(event)) as Record<string, unknown> | null;
  if (!body || typeof body !== "object") {
    throw createError({ statusCode: 400, message: "Invalid body" });
  }
  const delegate = getDelegate(prisma as unknown as Record<string, unknown>, resource);
  let data = omitReadonly(body);
  if (resource === "voice-roles") {
    data = prepareVoiceRoleWrite(data);
  }
  if (resource === "users") {
    data = await prepareUserAdminWriteData(data, "update");
  }
  if (resource === "translate-service-configs") {
    data = await prepareTranslateServiceConfigWrite(data, id);
  }
  if (resource === "stt-service-configs") {
    data = await prepareSttServiceConfigWrite(data, id);
  }
  if (resource === "object-storage-configs") {
    data = await prepareObjectStorageServiceConfigWrite(data, id);
  }
  if (resource === "sms-service-configs") {
    data = await prepareSmsServiceConfigWrite(data, id);
  }
  if (resource === "llm-service-configs") {
    data = await prepareLlmServiceConfigWrite(data);
  }
  if (resource === "tts-service-configs") {
    data = await prepareTtsServiceConfigWrite(data);
  }
  if (resource === "system-settings") {
    const existing = (await prisma.sysSystemSetting.findUnique({ where: { id } })) as {
      key?: string;
    } | null;
    if (existing?.key) {
      assertEditableSystemSettingKey(existing.key);
    }
    data = prepareSystemSettingWrite(data, "update", existing?.key);
  }
  const updated = (await delegate.update({
    where: { id },
    data,
  })) as Record<string, unknown>;
  const row = resource === "users" ? redactUserRecord(updated) : updated;
  return jsonSafe(row);
}

export async function adminDeleteHandler(event: H3Event, resource: ResourceSlug) {
  const id = getRouterParam(event, "id");
  assertResource(resource);
  if (!id) {
    throw createError({ statusCode: 400, message: "Missing id" });
  }
  if (isReadOnlyResource(resource)) {
    throw createError({ statusCode: 405, message: "Read-only" });
  }
  const delegate = getDelegate(prisma as unknown as Record<string, unknown>, resource);
  if (resource === "system-settings") {
    const existing = (await prisma.sysSystemSetting.findUnique({ where: { id } })) as {
      key?: string;
    } | null;
    if (existing?.key) {
      assertEditableSystemSettingKey(existing.key);
    }
  }
  if (SOFT_DELETE[resource]) {
    const row = await delegate.update({
      where: { id },
      data: { deletedAt: new Date() },
    });
    return jsonSafe(resource === "users" ? redactUserRecord(row as Record<string, unknown>) : row);
  }
  const row = await delegate.delete({ where: { id } });
  return jsonSafe(resource === "users" ? redactUserRecord(row as Record<string, unknown>) : row);
}
