import { adminListHandler } from "~~/server/utils/adminCrudHandlers";
import type { ResourceSlug } from "~~/server/utils/adminResource";

const SLUG: ResourceSlug = "messages";

export default defineEventHandler((event) => adminListHandler(event, SLUG));
