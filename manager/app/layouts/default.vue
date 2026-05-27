<script setup lang="ts">
import {
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  MoonIcon,
  SunIcon,
  XMarkIcon,
} from "@heroicons/vue/24/outline";
import { AI_SETTINGS_LEGACY_PATHS, BACKUP_LEGACY_PATHS, SYSTEM_SETTINGS_LEGACY_PATHS } from "~/utils/adminNav";

const route = useRoute();
const router = useRouter();
const { t, locale, locales } = useI18n();
const localePath = useLocalePath();
const { groups: navGroups } = useAdminNav();
const themeStore = useThemeStore();
const userStore = useUserStore();
const toast = useToast();
const mobileMenuOpen = ref(false);
const logoutLoading = ref(false);

useHead(() => ({
  title: t("app.title"),
  htmlAttrs: {
    lang: locales.value.find((item) => item.code === locale.value)?.language ?? locale.value,
  },
}));

const backupLegacyPaths = computed(() =>
  BACKUP_LEGACY_PATHS.map((path) => localePath(path)),
);

const aiSettingsLegacyPaths = computed(() =>
  AI_SETTINGS_LEGACY_PATHS.map((path) => localePath(path)),
);

const systemSettingsLegacyPaths = computed(() =>
  SYSTEM_SETTINGS_LEGACY_PATHS.map((path) => localePath(path)),
);

async function logout() {
  logoutLoading.value = true;
  try {
    await $fetch("/api/auth/logout", { method: "POST", credentials: "include" });
    userStore.clearUser();
    toast.success(t("toast.logoutSuccess"));
    await router.push(localePath("/login"));
  } catch {
    userStore.clearUser();
    await router.push(localePath("/login"));
  } finally {
    logoutLoading.value = false;
  }
}

function isActive(path: string) {
  const localizedPath = localePath(path);
  if (path === "/") {
    return route.path === localizedPath;
  }
  if (path === "/manage/backups") {
    return (
      route.path === localizedPath ||
      route.path.startsWith(`${localizedPath}/`) ||
      backupLegacyPaths.value.includes(route.path)
    );
  }
  if (path === "/manage/ai-settings") {
    return (
      route.path === localizedPath ||
      route.path.startsWith(`${localizedPath}/`) ||
      aiSettingsLegacyPaths.value.includes(route.path)
    );
  }
  if (path === "/manage/system-settings") {
    return (
      route.path === localizedPath ||
      route.path.startsWith(`${localizedPath}/`) ||
      systemSettingsLegacyPaths.value.includes(route.path)
    );
  }
  return route.path === localizedPath || route.path.startsWith(`${localizedPath}/`);
}

function navClick() {
  mobileMenuOpen.value = false;
}

watch(
  () => route.path,
  () => {
    mobileMenuOpen.value = false;
  },
);
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-background text-foreground">
    <aside class="hidden h-full w-64 shrink-0 flex-col border-r border-border bg-surface md:flex">
      <div class="flex shrink-0 items-center gap-3 border-b border-border p-5">
        <img src="/favicon.ico" alt="" class="h-10 w-10 rounded-full" />
        <div>
          <h2 class="text-lg font-bold text-foreground">{{ $t("common.appName") }}</h2>
          <p class="text-xs text-muted">{{ $t("common.adminConsole") }}</p>
        </div>
      </div>

      <nav class="min-h-0 flex-1 overflow-y-auto p-5">
        <div class="space-y-4">
          <template v-for="(g, gi) in navGroups" :key="g.titleKey">
            <div v-if="gi > 0" class="border-t border-border pt-4" />
            <div>
              <p
                v-if="g.titleKey !== 'nav.groups.overview'"
                class="mb-1.5 px-3 text-xs font-medium uppercase tracking-wider text-muted"
              >
                {{ g.title }}
              </p>
              <ul class="space-y-1">
                <li v-for="item in g.items" :key="item.to">
                  <NuxtLink
                    :to="localePath(item.to)"
                    class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors"
                    :class="
                      isActive(item.to)
                        ? 'bg-primary-50 text-primary-600 dark:bg-primary-950/40 dark:text-primary-400'
                        : 'text-muted hover:bg-surface-muted'
                    "
                  >
                    <component :is="item.icon" class="h-5 w-5 shrink-0" />
                    <span>{{ item.label }}</span>
                  </NuxtLink>
                </li>
              </ul>
            </div>
          </template>
        </div>
      </nav>

      <div class="shrink-0 space-y-1 border-t border-border p-5">
        <p
          class="truncate px-3 pb-1 text-xs text-muted"
          :title="userStore.displayName || $t('common.adminFallback')"
        >
          {{ userStore.displayName || $t("common.adminFallback") }}
        </p>
        <AdminLanguageSwitcher />
        <button
          type="button"
          class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-muted transition-colors hover:bg-surface-muted"
          @click="themeStore.toggle()"
        >
          <SunIcon v-if="themeStore.isDark" class="h-5 w-5" />
          <MoonIcon v-else class="h-5 w-5" />
          <span>{{ themeStore.isDark ? $t("common.lightMode") : $t("common.darkMode") }}</span>
        </button>
        <button
          type="button"
          class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-muted transition-colors hover:bg-surface-muted disabled:opacity-50"
          :disabled="logoutLoading"
          @click="logout"
        >
          <ArrowRightOnRectangleIcon class="h-5 w-5" />
          <span>{{ logoutLoading ? $t("common.loggingOut") : $t("common.logout") }}</span>
        </button>
      </div>
    </aside>

    <main class="flex min-w-0 flex-1 flex-col overflow-hidden">
      <header class="flex shrink-0 items-center justify-between border-b border-border bg-surface px-3 py-2 md:hidden">
        <h2 class="text-base font-semibold">{{ $t("common.adminConsoleShort") }}</h2>
        <div class="flex items-center gap-1">
          <div class="w-36">
            <AdminLanguageSwitcher />
          </div>
          <button
            type="button"
            class="rounded-lg p-1.5 text-muted hover:bg-surface-muted"
            :aria-label="$t('layout.menu.open')"
            @click="mobileMenuOpen = true"
          >
            <Bars3Icon class="h-5 w-5" />
          </button>
        </div>
      </header>

      <Transition name="fade">
        <div v-if="mobileMenuOpen" class="fixed inset-0 z-50 bg-black/50 md:hidden" @click="mobileMenuOpen = false" />
      </Transition>

      <Transition name="slide">
        <aside
          v-if="mobileMenuOpen"
          class="fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-border bg-surface md:hidden"
          @click.stop
        >
          <div class="flex items-center justify-between border-b border-border px-3 py-2">
            <span class="font-bold">{{ $t("common.menu") }}</span>
            <button type="button" class="rounded-lg p-1.5 text-muted hover:bg-surface-muted" @click="mobileMenuOpen = false">
              <XMarkIcon class="h-5 w-5" />
            </button>
          </div>
          <nav class="flex-1 overflow-y-auto p-3">
            <template v-for="(g, gi) in navGroups" :key="'m-' + g.titleKey">
              <div v-if="gi > 0" class="my-3 border-t border-border" />
              <p
                v-if="g.titleKey !== 'nav.groups.overview'"
                class="mb-1 px-2 text-xs font-medium uppercase text-muted"
              >
                {{ g.title }}
              </p>
              <ul class="space-y-0.5">
                <li v-for="item in g.items" :key="item.to">
                  <NuxtLink
                    :to="localePath(item.to)"
                    class="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm font-medium"
                    :class="
                      isActive(item.to)
                        ? 'bg-primary-50 text-primary-600'
                        : 'text-muted hover:bg-surface-muted'
                    "
                    @click="navClick"
                  >
                    <component :is="item.icon" class="h-4 w-4" />
                    {{ item.label }}
                  </NuxtLink>
                </li>
              </ul>
            </template>
          </nav>
          <div class="space-y-1 border-t border-border p-3">
            <AdminLanguageSwitcher />
            <button
              type="button"
              class="flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-sm text-muted hover:bg-surface-muted"
              @click="themeStore.toggle()"
            >
              <SunIcon v-if="themeStore.isDark" class="h-4 w-4" />
              <MoonIcon v-else class="h-4 w-4" />
              {{ themeStore.isDark ? $t("common.lightMode") : $t("common.darkMode") }}
            </button>
          </div>
        </aside>
      </Transition>

      <div class="flex min-h-0 flex-1 flex-col overflow-hidden p-2 md:p-4">
        <slot />
      </div>
    </main>

    <AdminToastHost />
    <AdminConfirmHost />
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active {
  transition: transform 0.3s ease-out;
}

.slide-leave-active {
  transition: transform 0.3s ease-in;
}

.slide-enter-from,
.slide-leave-to {
  transform: translateX(-100%);
}
</style>
