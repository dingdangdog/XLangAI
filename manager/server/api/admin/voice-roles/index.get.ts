import { adminListHandler } from "../../../utils/adminCrudHandlers";
import type { ResourceSlug } from "../../../utils/adminResource";

const SLUG: ResourceSlug = "voice-roles";

export default defineEventHandler((event) => adminListHandler(event, SLUG));
