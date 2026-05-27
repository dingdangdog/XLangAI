/** 清除前端登录态（Cookie + UserStore），与 jimily clearAuthStorage 一致 */
export function clearAuthStorage() {
  useUserStore().clearUser();
  const authCookie = useCookie("Authorization");
  authCookie.value = null;
  try {
    $fetch("/api/auth/logout", { method: "POST", credentials: "include" }).catch(() => {});
  } catch {
    // 忽略
  }
}
