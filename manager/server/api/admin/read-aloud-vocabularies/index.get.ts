import { adminListHandler } from "~~/server/utils/adminCrudHandlers";
import type { ResourceSlug } from "~~/server/utils/adminResource";

const SLUG: ResourceSlug = "read-aloud-vocabularies";

export default defineEventHandler((event) => adminListHandler(event, SLUG));
