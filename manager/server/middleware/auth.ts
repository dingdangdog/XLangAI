import { getManagerAuthPayload } from "../utils/jwt";

function normalizePathname(pathname: string): string {
  if (pathname.length <= 1) return pathname;
  return pathname.replace(/\/+$/, "") || "/";
}

function isUnderApiBase(pathname: string, base: string): boolean {
  return pathname === base || pathname.startsWith(`${base}/`);
}

export default defineEventHandler((event) => {
  if (event.method === "OPTIONS") return;

  const pathname = normalizePathname(getRequestURL(event).pathname);
  if (!isUnderApiBase(pathname, "/api/admin")) return;

  const payload = getManagerAuthPayload(event);
  if (!payload?.id || !payload.isManagerAdmin) {
    throw createError({ statusCode: 401, message: "未登录或登录失效" });
  }

  event.context.managerAuth = payload;
});
