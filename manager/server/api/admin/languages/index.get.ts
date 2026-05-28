import { adminListHandler } from "~~/server/utils/adminCrudHandlers";
import type { ResourceSlug } from "~~/server/utils/adminResource";

const SLUG: ResourceSlug = "languages";

export default defineEventHandler((event) => adminListHandler(event, SLUG));
