import { getServerStoreAdminState } from "../../../utils/serverStoreConfig";

export default defineEventHandler(() => getServerStoreAdminState());
