import { useRouteBaseName } from "#i18n";

/** 除登录页外，所有页面需运营管理员登录 */
export default defineNuxtRouteMiddleware(async (to) => {
  const routeBaseName = useRouteBaseName();
  if (routeBaseName(to) === "login") return;

  const { buildLoginRedirect } = useAuthRedirect();
  const userStore = useUserStore();
  if (userStore.user?.isManagerAdmin) return;

  const me = await userStore.fetchUser();
  if (me?.isManagerAdmin) return;

  return navigateTo(buildLoginRedirect(to.fullPath));
});
