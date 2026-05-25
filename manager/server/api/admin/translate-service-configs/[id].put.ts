import { adminUpdateHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "translate-service-configs";

export default defineEventHandler((event) => adminUpdateHandler(event, SLUG));
