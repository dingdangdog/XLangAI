import { adminUpdateHandler } from "~~/server/utils/adminCrudHandlers";
import type { ResourceSlug } from "~~/server/utils/adminResource";

const SLUG: ResourceSlug = "read-aloud-categories";

export default defineEventHandler((event) => adminUpdateHandler(event, SLUG));
