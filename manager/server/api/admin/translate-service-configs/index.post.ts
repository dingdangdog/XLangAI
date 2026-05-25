import { adminCreateHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "translate-service-configs";

export default defineEventHandler((event) => adminCreateHandler(event, SLUG));
