import { adminDeleteHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "conversations";

export default defineEventHandler((event) => adminDeleteHandler(event, SLUG));
