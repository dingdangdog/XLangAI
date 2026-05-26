import { createError } from "h3";
import type { AppPrismaClient } from "../lib/prisma";
import { jsonSafe } from "./jsonSafe";

export const DB_BACKUP_FORMAT = "xlangai-manager-db-backup";
export const DB_BACKUP_VERSION = 1;

/** 导出/导入顺序：先系统配置，再用户域，最后注销归档表 */
export const DB_BACKUP_TABLES = [
  { key: "languages", delegate: "language" },
  { key: "membershipTiers", delegate: "membershipTier" },
  { key: "sysLlmServiceConfigs", delegate: "sysLlmServiceConfig" },
  { key: "sysSttServiceConfigs", delegate: "sysSttServiceConfig" },
  { key: "sysTranslateServiceConfigs", delegate: "sysTranslateServiceConfig" },
  { key: "sysObjectStorageConfigs", delegate: "sysObjectStorageConfig" },
  { key: "sysSystemSettings", delegate: "sysSystemSetting" },
  { key: "ttsServiceConfigs", delegate: "ttsServiceConfig" },
  { key: "billingProducts", delegate: "billingProduct" },
  { key: "voiceRoles", delegate: "voiceRole" },
  { key: "promptTemplates", delegate: "promptTemplate" },
  { key: "users", delegate: "user" },
  { key: "storeTransactions", delegate: "storeTransaction" },
  { key: "userUsage", delegate: "userUsage" },
  { key: "conversations", delegate: "conversation" },
  { key: "messages", delegate: "message" },
  { key: "sysServiceUsageDaily", delegate: "sysServiceUsageDaily" },
  { key: "usersBackup", delegate: "usersBackup" },
  { key: "conversationsBackup", delegate: "conversationsBackup" },
  { key: "messagesBackup", delegate: "messagesBackup" },
  { key: "userUsageBackup", delegate: "userUsageBackup" },
] as const;

export type DbBackupTableKey = (typeof DB_BACKUP_TABLES)[number]["key"];
export type DbBackupImportMode = "merge" | "replace";

export type DatabaseBackupPayload = {
  format: typeof DB_BACKUP_FORMAT;
  version: typeof DB_BACKUP_VERSION;
  exportedAt: string;
  app: string;
  tables: Partial<Record<DbBackupTableKey, Record<string, unknown>[]>>;
  counts: Partial<Record<DbBackupTableKey, number>>;
};

const BIGINT_FIELDS = new Set(["tokenBalance", "tokenGrant", "sttAudioBytes", "unitCount"]);
const DATE_ONLY_FIELDS = new Set(["date"]);

type TableDelegate = {
  findMany: (args?: object) => Promise<unknown[]>;
  count: (args?: object) => Promise<number>;
  deleteMany: (args?: object) => Promise<unknown>;
  createMany: (args: { data: object[]; skipDuplicates?: boolean }) => Promise<unknown>;
  upsert: (args: { where: object; create: object; update: object }) => Promise<unknown>;
};

function getTableDelegate(db: AppPrismaClient, delegateName: string): TableDelegate {
  const delegate = (db as Record<string, unknown>)[delegateName];
  if (!delegate || typeof delegate !== "object") {
    throw createError({ statusCode: 500, message: `Prisma delegate missing: ${delegateName}` });
  }
  return delegate as TableDelegate;
}

function reviveRow(row: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(row)) {
    if (value === null || value === undefined) {
      out[key] = value;
      continue;
    }
    if (BIGINT_FIELDS.has(key)) {
      out[key] = typeof value === "bigint" ? value : BigInt(String(value));
      continue;
    }
    if (DATE_ONLY_FIELDS.has(key) && typeof value === "string") {
      out[key] = new Date(value);
      continue;
    }
    out[key] = value;
  }
  return out;
}

function reviveRows(rows: unknown[]): Record<string, unknown>[] {
  return rows.map((row) => {
    if (!row || typeof row !== "object" || Array.isArray(row)) {
      throw createError({ statusCode: 400, message: "备份文件中存在无效的行数据" });
    }
    return reviveRow(row as Record<string, unknown>);
  });
}

export function parseDatabaseBackupPayload(raw: unknown): DatabaseBackupPayload {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    throw createError({ statusCode: 400, message: "备份文件必须是 JSON 对象" });
  }
  const obj = raw as Record<string, unknown>;
  if (obj.format !== DB_BACKUP_FORMAT) {
    throw createError({ statusCode: 400, message: "不支持的备份格式" });
  }
  if (obj.version !== DB_BACKUP_VERSION) {
    throw createError({
      statusCode: 400,
      message: `不支持的备份版本：${String(obj.version)}，当前仅支持 v${DB_BACKUP_VERSION}`,
    });
  }
  if (!obj.tables || typeof obj.tables !== "object" || Array.isArray(obj.tables)) {
    throw createError({ statusCode: 400, message: "备份文件缺少 tables 字段" });
  }
  return obj as DatabaseBackupPayload;
}

export async function getDatabaseBackupInfo(db: AppPrismaClient) {
  const tables: Partial<Record<DbBackupTableKey, number>> = {};
  let totalRows = 0;

  for (const spec of DB_BACKUP_TABLES) {
    const delegate = getTableDelegate(db, spec.delegate);
    const count = await delegate.count();
    tables[spec.key] = count;
    totalRows += count;
  }

  return {
    format: DB_BACKUP_FORMAT,
    version: DB_BACKUP_VERSION,
    totalRows,
    tables,
  };
}

export async function exportDatabaseBackup(db: AppPrismaClient) {
  const tables: Partial<Record<DbBackupTableKey, unknown[]>> = {};
  const counts: Partial<Record<DbBackupTableKey, number>> = {};

  for (const spec of DB_BACKUP_TABLES) {
    const delegate = getTableDelegate(db, spec.delegate);
    const rows = (await delegate.findMany()) as Record<string, unknown>[];
    tables[spec.key] = rows;
    counts[spec.key] = rows.length;
  }

  const payload: DatabaseBackupPayload = {
    format: DB_BACKUP_FORMAT,
    version: DB_BACKUP_VERSION,
    exportedAt: new Date().toISOString(),
    app: "xlangai-manager",
    tables,
    counts,
  };

  return jsonSafe(payload);
}

const CREATE_CHUNK_SIZE = 200;

async function createManyInChunks(delegate: TableDelegate, rows: Record<string, unknown>[]) {
  for (let i = 0; i < rows.length; i += CREATE_CHUNK_SIZE) {
    await delegate.createMany({ data: rows.slice(i, i + CREATE_CHUNK_SIZE) });
  }
}

async function mergeTableRows(delegate: TableDelegate, rows: Record<string, unknown>[]) {
  for (const row of rows) {
    const id = row.id;
    if (typeof id !== "string" || !id) {
      throw createError({ statusCode: 400, message: "备份行缺少有效 id，无法合并导入" });
    }
    await delegate.upsert({
      where: { id },
      create: row,
      update: row,
    });
  }
}

async function replaceAllTables(db: AppPrismaClient) {
  for (const spec of [...DB_BACKUP_TABLES].reverse()) {
    const delegate = getTableDelegate(db, spec.delegate);
    await delegate.deleteMany({});
  }
}

async function importTables(
  db: AppPrismaClient,
  payload: DatabaseBackupPayload,
  mode: DbBackupImportMode,
) {
  const imported: Partial<Record<DbBackupTableKey, number>> = {};

  if (mode === "replace") {
    await replaceAllTables(db);
  }

  for (const spec of DB_BACKUP_TABLES) {
    const rawRows = payload.tables[spec.key];
    if (!rawRows) {
      imported[spec.key] = 0;
      continue;
    }
    if (!Array.isArray(rawRows)) {
      throw createError({ statusCode: 400, message: `表 ${spec.key} 的数据格式无效` });
    }

    const rows = reviveRows(rawRows);
    const delegate = getTableDelegate(db, spec.delegate);

    if (rows.length === 0) {
      imported[spec.key] = 0;
      continue;
    }

    if (mode === "merge") {
      await mergeTableRows(delegate, rows);
    } else {
      await createManyInChunks(delegate, rows);
    }

    imported[spec.key] = rows.length;
  }

  return imported;
}

export async function importDatabaseBackup(
  db: AppPrismaClient,
  raw: unknown,
  mode: DbBackupImportMode,
) {
  if (mode !== "merge" && mode !== "replace") {
    throw createError({ statusCode: 400, message: "mode 必须是 merge 或 replace" });
  }

  const payload = parseDatabaseBackupPayload(raw);

  const imported = await db.$transaction(async (tx) => {
    return importTables(tx as AppPrismaClient, payload, mode);
  });

  const totalImported = Object.values(imported).reduce((sum, n) => sum + (n ?? 0), 0);

  return {
    mode,
    imported,
    totalImported,
    backupExportedAt: payload.exportedAt,
  };
}

export function buildBackupDownloadFilename(exportedAt?: string) {
  const stamp = (exportedAt ? new Date(exportedAt) : new Date())
    .toISOString()
    .replace(/[:.]/g, "-")
    .slice(0, 19);
  return `xlangai-db-backup-${stamp}.json`;
}
