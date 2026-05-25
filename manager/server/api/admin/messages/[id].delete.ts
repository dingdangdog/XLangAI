import { adminDeleteHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "messages";

export default defineEventHandler((event) => adminDeleteHandler(event, SLUG));
