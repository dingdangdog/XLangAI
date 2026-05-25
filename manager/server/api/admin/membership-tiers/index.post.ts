import { adminCreateHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "membership-tiers";

export default defineEventHandler((event) => adminCreateHandler(event, SLUG));
