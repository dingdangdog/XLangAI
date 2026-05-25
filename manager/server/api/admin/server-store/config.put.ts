import { getServerStoreAdminState, saveLocalServerStoreConfig } from "../../../utils/serverStoreConfig";

export default defineEventHandler(async (event) => {
  const body = (await readBody(event)) as Record<string, unknown> | null;
  if (!body || typeof body !== "object" || Array.isArray(body)) {
    throw createError({ statusCode: 400, message: "Invalid body" });
  }

  await saveLocalServerStoreConfig(body);
  return getServerStoreAdminState();
});
