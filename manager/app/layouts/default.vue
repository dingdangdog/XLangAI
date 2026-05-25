<script setup lang="ts">
import {
  ArrowRightOnRectangleIcon,
  Bars3Icon,
  MoonIcon,
  SunIcon,
  XMarkIcon,
} from "@heroicons/vue/24/outline";
import { ADMIN_NAV_GROUPS, BACKUP_LEGACY_PATHS } from "~/utils/adminNav";

const route = useRoute();
const router = useRouter();
const themeStore = useThemeStore();
const userStore = useUserStore();
const toast = useToast();
const mobileMenuOpen = ref(false);
const logoutLoading = ref(false);

async function logout() {
  logoutLoading.value = true;
  try {
    await $fetch("/api/auth/logout", { method: "POST", credentials: "include" });
    userStore.clearUser();
    toast.success("已退出登录");
    await router.push("/login");
  } catch {
    userStore.clearUser();
    await router.push("/login");
  } finally {
    logoutLoading.value = false;
  }
}

function isActive(path: string) {
  if (path === "/") return route.path === "/";
  if (path === "/manage/backups") {
    return (
      route.path === path ||
      route.path.startsWith(path + "/") ||
      BACKUP_LEGACY_PATHS.includes(route.path as (typeof BACKUP_LEGACY_PATHS)[number])
    );
  }
  return route.path === path || route.path.startsWith(path + "/");
}

function navClick() {
  mobileMenuOpen.value = false;
}

watch(
  () => route.path,
  () => {
    mobileMenuOpen.value = false;
  }
);
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-background text-foreground">
    <!-- 桌面侧栏 -->
    <aside class="hidden h-full w-64 shrink-0 flex-col border-r border-border bg-surface md:flex">
      <div class="flex shrink-0 items-center gap-3 border-b border-border p-5">
        <img src="/favicon.ico" alt="" class="h-10 w-10 rounded-full" />
        <div>
          <h2 class="text-lg font-bold text-foreground">小浪AI</h2>
          <p class="text-xs text-muted">运营后台</p>
        </div>
      </div>

      <nav class="min-h-0 flex-1 overflow-y-auto p-5">
        <div class="space-y-4">
          <template v-for="(g, gi) in ADMIN_NAV_GROUPS" :key="g.title">
            <div v-if="gi > 0" class="border-t border-border pt-4" />
            <div>
              <p v-if="g.title !== '概览'" class="mb-1.5 px-3 text-xs font-medium uppercase tracking-wider text-muted">
                {{ g.title }}
              </p>
              <ul class="space-y-1">
                <li v-for="item in g.items" :key="item.to">
                  <NuxtLink :to="item.to"
                    class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors" :class="isActive(item.to)
                      ? 'bg-primary-50 text-primary-600 dark:bg-primary-950/40 dark:text-primary-400'
                      : 'text-muted hover:bg-surface-muted'
                      ">
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
        <p class="truncate px-3 pb-1 text-xs text-muted" :title="userStore.displayName">
          {{ userStore.displayName }}
        </p>
        <button type="button"
          class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-muted transition-colors hover:bg-surface-muted"
          @click="themeStore.toggle()">
          <SunIcon v-if="themeStore.isDark" class="h-5 w-5" />
          <MoonIcon v-else class="h-5 w-5" />
          <span>{{ themeStore.isDark ? "浅色模式" : "深色模式" }}</span>
        </button>
        <button type="button"
          class="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-muted transition-colors hover:bg-surface-muted disabled:opacity-50"
          :disabled="logoutLoading" @click="logout">
          <ArrowRightOnRectangleIcon class="h-5 w-5" />
          <span>{{ logoutLoading ? "退出中…" : "退出登录" }}</span>
        </button>
      </div>
    </aside>

    <main class="flex min-w-0 flex-1 flex-col overflow-hidden">
      <!-- 移动顶栏 -->
      <header class="flex shrink-0 items-center justify-between border-b border-border bg-surface px-3 py-2 md:hidden">
        <h2 class="text-base font-semibold">小浪AI 后台</h2>
        <button type="button" class="rounded-lg p-1.5 text-muted hover:bg-surface-muted" aria-label="打开菜单"
          @click="mobileMenuOpen = true">
          <Bars3Icon class="h-5 w-5" />
        </button>
      </header>

      <Transition name="fade">
        <div v-if="mobileMenuOpen" class="fixed inset-0 z-50 bg-black/50 md:hidden" @click="mobileMenuOpen = false" />
      </Transition>

      <Transition name="slide">
        <aside v-if="mobileMenuOpen"
          class="fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-border bg-surface md:hidden"
          @click.stop>
          <div class="flex items-center justify-between border-b border-border px-3 py-2">
            <span class="font-bold">菜单</span>
            <button type="button" class="rounded-lg p-1.5 text-muted hover:bg-surface-muted"
              @click="mobileMenuOpen = false">
              <XMarkIcon class="h-5 w-5" />
            </button>
          </div>
          <nav class="flex-1 overflow-y-auto p-3">
            <template v-for="(g, gi) in ADMIN_NAV_GROUPS" :key="'m-' + g.title">
              <div v-if="gi > 0" class="my-3 border-t border-border" />
              <p v-if="g.title !== '概览'" class="mb-1 px-2 text-xs font-medium uppercase text-muted">
                {{ g.title }}
              </p>
              <ul class="space-y-0.5">
                <li v-for="item in g.items" :key="item.to">
                  <NuxtLink :to="item.to" class="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm font-medium"
                    :class="isActive(item.to)
                      ? 'bg-primary-50 text-primary-600'
                      : 'text-muted hover:bg-surface-muted'
                      " @click="navClick">
                    <component :is="item.icon" class="h-4 w-4" />
                    {{ item.label }}
                  </NuxtLink>
                </li>
              </ul>
            </template>
          </nav>
          <div class="border-t border-border p-3">
            <button type="button"
              class="flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-sm text-muted hover:bg-surface-muted"
              @click="themeStore.toggle()">
              <SunIcon v-if="themeStore.isDark" class="h-4 w-4" />
              <MoonIcon v-else class="h-4 w-4" />
              {{ themeStore.isDark ? "浅色模式" : "深色模式" }}
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
