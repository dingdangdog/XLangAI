import { getServerStoreAdminState } from "~~/server/utils/serverStoreConfig";

export default defineEventHandler(() => getServerStoreAdminState());
