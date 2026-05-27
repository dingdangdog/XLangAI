import { useRouteBaseName } from "#i18n";
import { getAuthHeaders } from "~/utils/api";

export default defineNuxtPlugin(() => {
  if (!import.meta.client) return;

  const router = useRouter();
  const userStore = useUserStore();
  const routeBaseName = useRouteBaseName();
  const { buildLoginRedirect } = useAuthRedirect();

  globalThis.$fetch = $fetch.create({
    onRequest({ options }) {
      options.credentials = options.credentials ?? "include";
      options.headers = {
        ...(options.headers as Record<string, string> | undefined),
        ...getAuthHeaders(),
      };
    },
    onResponseError({ response }) {
      if (response.status !== 401) return;
      userStore.clearUser();
      const route = router.currentRoute.value;
      if (routeBaseName(route) === "login") return;
      void navigateTo(buildLoginRedirect(route.fullPath));
    },
  });
});
