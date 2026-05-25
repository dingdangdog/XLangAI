import { defineStore } from "pinia";

export type ManagerAuthUser = {
  id: string;
  username: string;
  nickname?: string | null;
  phone?: string | null;
  email?: string | null;
  isManagerAdmin: true;
};

export const useUserStore = defineStore("manager-user", () => {
  const user = ref<ManagerAuthUser | null>(null);

  const displayName = computed(() => user.value?.nickname?.trim() || user.value?.username || "管理员");

  function setUser(info: ManagerAuthUser | null) {
    user.value = info;
  }

  function clearUser() {
    user.value = null;
  }

  async function fetchUser(): Promise<ManagerAuthUser | null> {
    try {
      const headers: Record<string, string> = {};
      if (import.meta.server) {
        const cookie = useRequestHeaders(["cookie"]).cookie;
        if (cookie) headers.cookie = cookie;
      }

      const res = await $fetch<ManagerAuthUser>("/api/auth/me", {
        method: "GET",
        credentials: "include",
        headers,
      });
      setUser(res);
      return res;
    } catch {
      clearUser();
      return null;
    }
  }

  return {
    user,
    displayName,
    setUser,
    clearUser,
    fetchUser,
  };
});
