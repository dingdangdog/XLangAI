import prisma from "~~/server/lib/prisma";
import {
  buildBackupDownloadFilename,
  exportDatabaseBackup,
} from "~~/server/utils/databaseBackup";

export default defineEventHandler(async (event) => {
  const backup = await exportDatabaseBackup(prisma);
  const filename = buildBackupDownloadFilename(backup.exportedAt);

  setHeader(event, "Content-Type", "application/json; charset=utf-8");
  setHeader(event, "Content-Disposition", `attachment; filename="${filename}"`);

  return backup;
});
