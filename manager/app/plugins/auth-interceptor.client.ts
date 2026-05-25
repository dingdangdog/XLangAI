import { buildLoginRedirect } from "~/utils/authRedirect";

export default defineNuxtPlugin(() => {
  if (!import.meta.client) return;

  const router = useRouter();
  const userStore = useUserStore();

  globalThis.$fetch = $fetch.create({
    onResponseError({ response }) {
      if (response.status !== 401) return;
      userStore.clearUser();
      const fullPath = router.currentRoute.value.fullPath;
      if (fullPath.startsWith("/login")) return;
      void navigateTo(buildLoginRedirect(fullPath));
    },
  });
});
