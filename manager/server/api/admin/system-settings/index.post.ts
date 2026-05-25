import { adminCreateHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "system-settings";

export default defineEventHandler((event) => adminCreateHandler(event, SLUG));
