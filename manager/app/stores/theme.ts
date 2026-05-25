import { defineStore } from "pinia";

type ThemeMode = "light" | "dark";

export const useThemeStore = defineStore("theme", () => {
  const mode = ref<ThemeMode>("light");
  const isDark = computed(() => mode.value === "dark");

  const applyToDocument = () => {
    if (!import.meta.client) return;
    const html = document.documentElement;
    if (mode.value === "dark") {
      html.classList.add("dark");
    } else {
      html.classList.remove("dark");
    }
  };

  const persist = () => {
    if (!import.meta.client) return;
    localStorage.setItem("manager-theme-mode", mode.value);
    const cookie = useCookie<ThemeMode>("managerThemeMode", {
      default: () => "light",
      sameSite: "lax",
      maxAge: 60 * 60 * 24 * 365,
    });
    cookie.value = mode.value;
  };

  const init = () => {
    if (!import.meta.client) return;
    const cookie = useCookie<ThemeMode>("managerThemeMode", { default: () => "light" });
    const stored = localStorage.getItem("manager-theme-mode") as ThemeMode | null;
    const initial =
      stored === "light" || stored === "dark"
        ? stored
        : cookie.value === "dark"
          ? "dark"
          : "light";
    mode.value = initial;
    applyToDocument();
    persist();
  };

  const setMode = (next: ThemeMode) => {
    if (mode.value === next) return;
    mode.value = next;
    applyToDocument();
    persist();
  };

  const toggle = () => {
    setMode(mode.value === "light" ? "dark" : "light");
  };

  return { mode, isDark, init, setMode, toggle };
});
