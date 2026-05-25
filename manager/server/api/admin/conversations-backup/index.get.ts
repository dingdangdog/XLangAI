import { adminListHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "conversations-backup";

export default defineEventHandler((event) => adminListHandler(event, SLUG));
