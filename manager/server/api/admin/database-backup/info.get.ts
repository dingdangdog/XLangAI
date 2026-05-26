import prisma from "../../../lib/prisma";
import { getDatabaseBackupInfo } from "../../../utils/databaseBackup";

export default defineEventHandler(() => getDatabaseBackupInfo(prisma));
