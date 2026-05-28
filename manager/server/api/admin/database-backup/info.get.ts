import prisma from "~~/server/lib/prisma";
import { getDatabaseBackupInfo } from "../../../utils/databaseBackup";

export default defineEventHandler(() => getDatabaseBackupInfo(prisma));
