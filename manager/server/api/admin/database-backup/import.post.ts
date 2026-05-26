import { createError } from "h3";
import prisma from "../../../lib/prisma";
import {
  importDatabaseBackup,
  parseDatabaseBackupPayload,
  type DbBackupImportMode,
} from "../../../utils/databaseBackup";

type ImportBody = {
  mode?: DbBackupImportMode;
  backup?: unknown;
};

export default defineEventHandler(async (event) => {
  const contentType = getHeader(event, "content-type") ?? "";
  let mode: DbBackupImportMode = "merge";
  let backupRaw: unknown;

  if (contentType.includes("multipart/form-data")) {
    const parts = await readMultipartFormData(event);
    if (!parts?.length) {
      throw createError({ statusCode: 400, message: "未收到上传文件" });
    }

    const filePart = parts.find((part) => part.name === "file" && part.data?.length);
    if (!filePart) {
      throw createError({ statusCode: 400, message: "请上传 JSON 备份文件" });
    }

    const modePart = parts.find((part) => part.name === "mode");
    if (modePart?.data?.length) {
      mode = String(modePart.data).trim() as DbBackupImportMode;
    }

    try {
      backupRaw = JSON.parse(filePart.data.toString("utf8"));
    } catch {
      throw createError({ statusCode: 400, message: "备份文件不是有效的 JSON" });
    }
  } else {
    const body = (await readBody(event)) as ImportBody;
    mode = body?.mode ?? "merge";
    backupRaw = body?.backup;
  }

  if (backupRaw === undefined || backupRaw === null) {
    throw createError({ statusCode: 400, message: "缺少 backup 数据" });
  }

  // 先校验格式，便于返回明确错误
  parseDatabaseBackupPayload(backupRaw);

  return importDatabaseBackup(prisma, backupRaw, mode);
});
