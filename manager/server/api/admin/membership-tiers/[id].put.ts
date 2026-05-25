import { adminUpdateHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "membership-tiers";

export default defineEventHandler((event) => adminUpdateHandler(event, SLUG));
