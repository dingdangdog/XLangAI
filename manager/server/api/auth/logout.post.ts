export default defineEventHandler((event) => {
  deleteCookie(event, "Authorization");
  return { ok: true };
});
