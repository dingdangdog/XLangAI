import { adminCreateHandler } from "~~/server/utils/adminCrudHandlers";
import type { ResourceSlug } from "~~/server/utils/adminResource";

const SLUG: ResourceSlug = "voice-roles";

export default defineEventHandler((event) => adminCreateHandler(event, SLUG));
