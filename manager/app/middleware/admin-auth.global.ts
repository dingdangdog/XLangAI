import { buildLoginRedirect, LOGIN_PATH } from "~/utils/authRedirect";

/** 除 /login 外，所有页面需运营管理员登录（参考 jimily admin 中间件） */
export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === LOGIN_PATH) return;

  const userStore = useUserStore();
  if (userStore.user?.isManagerAdmin) return;

  const me = await userStore.fetchUser();
  if (me?.isManagerAdmin) return;

  return navigateTo(buildLoginRedirect(to.fullPath));
});
