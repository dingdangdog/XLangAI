import { useRouteBaseName } from "#i18n";

export default defineNuxtPlugin(() => {
  if (!import.meta.client) return;

  const router = useRouter();
  const userStore = useUserStore();
  const routeBaseName = useRouteBaseName();
  const { buildLoginRedirect } = useAuthRedirect();

  globalThis.$fetch = $fetch.create({
    onResponseError({ response }) {
      if (response.status !== 401) return;
      userStore.clearUser();
      const route = router.currentRoute.value;
      if (routeBaseName(route) === "login") return;
      void navigateTo(buildLoginRedirect(route.fullPath));
    },
  });
});
