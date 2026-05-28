import { adminDeleteHandler } from "~~/server/utils/adminCrudHandlers";
import type { ResourceSlug } from "~~/server/utils/adminResource";

const SLUG: ResourceSlug = "conversations";

export default defineEventHandler((event) => adminDeleteHandler(event, SLUG));
