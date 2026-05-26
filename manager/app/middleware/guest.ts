export default defineNuxtRouteMiddleware(async (to) => {
  const localePath = useLocalePath();
  const { resolveAuthRedirect } = useAuthRedirect();
  const userStore = useUserStore();
  const redirectTarget = localePath(resolveAuthRedirect(to.query.callbackUrl));

  if (userStore.user?.isManagerAdmin) {
    return navigateTo(redirectTarget);
  }

  const me = await userStore.fetchUser();
  if (me?.isManagerAdmin) {
    return navigateTo(redirectTarget);
  }
});
