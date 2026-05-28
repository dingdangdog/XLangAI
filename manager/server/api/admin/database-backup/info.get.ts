import prisma from "~~/server/lib/prisma";
import { getDatabaseBackupInfo } from "~~/server/utils/databaseBackup";

export default defineEventHandler(() => getDatabaseBackupInfo(prisma));
