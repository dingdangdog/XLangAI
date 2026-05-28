import { sendServerStoreHeartbeat } from "~~/server/utils/serverStoreConfig";

export default defineEventHandler(() => sendServerStoreHeartbeat());
