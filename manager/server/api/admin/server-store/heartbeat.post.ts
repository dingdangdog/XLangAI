import { sendServerStoreHeartbeat } from "../../../utils/serverStoreConfig";

export default defineEventHandler(() => sendServerStoreHeartbeat());
