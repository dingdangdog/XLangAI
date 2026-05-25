import { resolveAuthRedirect } from "~/utils/authRedirect";

export default defineNuxtRouteMiddleware(async (to) => {
  const userStore = useUserStore();
  if (userStore.user?.isManagerAdmin) {
    return navigateTo(resolveAuthRedirect(to.query.callbackUrl));
  }

  const me = await userStore.fetchUser();
  if (me?.isManagerAdmin) {
    return navigateTo(resolveAuthRedirect(to.query.callbackUrl));
  }
});
