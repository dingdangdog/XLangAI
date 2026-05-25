import { adminListHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "user-usage-backup";

export default defineEventHandler((event) => adminListHandler(event, SLUG));
