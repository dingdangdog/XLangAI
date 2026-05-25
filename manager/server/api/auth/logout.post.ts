export default defineEventHandler((event) => {
  deleteCookie(event, "Authorization", { path: "/" });
  return { ok: true };
});
