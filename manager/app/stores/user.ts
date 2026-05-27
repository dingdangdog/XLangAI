import { defineStore } from "pinia";

export type ManagerAuthUser = {
  id: string;
  username: string;
  nickname?: string | null;
  phone?: string | null;
  email?: string | null;
  isManagerAdmin: true;
  token?: string;
};

export const useUserStore = defineStore("manager-user", () => {
  const user = ref<ManagerAuthUser | null>(null);

  const displayName = computed(
    () => user.value?.nickname?.trim() || user.value?.username || "",
  );

  function setUser(info: ManagerAuthUser | null) {
    user.value = info;
  }

  function clearUser() {
    user.value = null;
  }

  async function fetchUser(): Promise<ManagerAuthUser | null> {
    try {
      const res = await $fetch<ManagerAuthUser>("/api/auth/me", {
        method: "GET",
        headers: import.meta.server
          ? { Authorization: useCookie("Authorization").value || "" }
          : undefined,
        credentials: "include",
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
