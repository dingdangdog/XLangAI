<script setup lang="ts">
import { MoonIcon, SunIcon } from "@heroicons/vue/24/outline";

definePageMeta({
  layout: "blank",
  middleware: ["guest"],
});

const { t } = useI18n();
const localePath = useLocalePath();
const localeHead = useLocaleHead({ seo: true });
const { resolveAuthRedirect } = useAuthRedirect();
const { localizeAuthError } = useAuthApiError();
const route = useRoute();
const themeStore = useThemeStore();
const toast = useToast();
const userStore = useUserStore();

const form = reactive({
  username: "",
  password: "",
});

const loading = ref(false);
const errors = reactive<{ username?: string; password?: string }>({});

useHead(() => ({
  title: t("pages.login.title"),
  htmlAttrs: localeHead.value.htmlAttrs,
  link: localeHead.value.link,
  meta: localeHead.value.meta,
}));

function validate() {
  errors.username = undefined;
  errors.password = undefined;

  if (!form.username.trim()) {
    errors.username = t("validation.usernameRequired");
  }
  if (!form.password) {
    errors.password = t("validation.passwordRequired");
  } else if (form.password.length < 6) {
    errors.password = t("validation.passwordMin");
  }

  return !errors.username && !errors.password;
}

async function submit() {
  if (!validate()) return;

  loading.value = true;
  try {
    const user = await $fetch("/api/auth/login", {
      method: "POST",
      body: {
        username: form.username.trim(),
        password: form.password,
      },
      credentials: "include",
    });
    userStore.setUser(user);
    toast.success(t("toast.loginSuccess"));
    await navigateTo(localePath(resolveAuthRedirect(route.query.callbackUrl)));
  } catch (e: unknown) {
    toast.error(localizeAuthError(e));
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center p-4">
    <div class="w-full max-w-md rounded-2xl border border-border bg-surface p-8 shadow-lg">
      <div class="mb-8 text-center">
        <img src="/favicon.ico" alt="" class="mx-auto mb-4 h-14 w-14 rounded-full" />
        <h1 class="text-2xl font-bold text-foreground">{{ $t("pages.login.title") }}</h1>
        <p class="mt-2 text-sm text-muted">{{ $t("pages.login.subtitle") }}</p>
      </div>

      <form class="space-y-5" @submit.prevent="submit">
        <AdminFormField :label="$t('fields.account')" :error="errors.username" required>
          <AdminInput
            v-model="form.username"
            autocomplete="username"
            :placeholder="$t('pages.login.usernamePlaceholder')"
          />
        </AdminFormField>

        <AdminFormField :label="$t('fields.password')" :error="errors.password" required>
          <AdminInput
            v-model="form.password"
            type="password"
            autocomplete="current-password"
            :placeholder="$t('pages.login.passwordPlaceholder')"
            revealable
          />
        </AdminFormField>

        <AdminButton type="submit" variant="primary" class="w-full" :loading="loading">
          {{ $t("pages.login.submit") }}
        </AdminButton>
      </form>

      <div class="mt-6 flex flex-col items-center gap-3">
        <AdminLanguageSwitcher class="max-w-xs" />
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-muted transition-colors hover:bg-surface-muted hover:text-foreground"
          @click="themeStore.toggle()"
        >
          <SunIcon v-if="themeStore.isDark" class="h-5 w-5" />
          <MoonIcon v-else class="h-5 w-5" />
          <span>{{ themeStore.isDark ? $t("common.lightMode") : $t("common.darkMode") }}</span>
        </button>
      </div>
    </div>
  </div>
</template>
