import { adminListHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "system-settings";

export default defineEventHandler((event) => adminListHandler(event, SLUG));
