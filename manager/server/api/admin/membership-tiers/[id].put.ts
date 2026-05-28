import { adminUpdateHandler } from "~~/server/utils/adminCrudHandlers";
import type { ResourceSlug } from "~~/server/utils/adminResource";

const SLUG: ResourceSlug = "membership-tiers";

export default defineEventHandler((event) => adminUpdateHandler(event, SLUG));
