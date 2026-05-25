<script setup lang="ts">
import { MoonIcon, SunIcon } from "@heroicons/vue/24/outline";
import { resolveAuthRedirect } from "~/utils/authRedirect";

definePageMeta({
  layout: "blank",
  middleware: ["guest"],
});

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

function validate() {
  errors.username = undefined;
  errors.password = undefined;

  if (!form.username.trim()) {
    errors.username = "请输入手机号或邮箱";
  }
  if (!form.password) {
    errors.password = "请输入密码";
  } else if (form.password.length < 6) {
    errors.password = "密码至少 6 位";
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
    toast.success("登录成功");
    await navigateTo(resolveAuthRedirect(route.query.callbackUrl));
  } catch (e: unknown) {
    const err = e as {
      data?: { message?: string; statusMessage?: string };
      message?: string;
      statusMessage?: string;
    };
    toast.error(
      err.data?.message ||
        err.data?.statusMessage ||
        err.message ||
        err.statusMessage ||
        "登录失败",
    );
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
        <h1 class="text-2xl font-bold text-foreground">小浪AI 运营后台</h1>
        <p class="mt-2 text-sm text-muted">请使用管理员账号登录</p>
      </div>

      <form class="space-y-5" @submit.prevent="submit">
        <AdminFormField label="账号" :error="errors.username" required>
          <AdminInput
            v-model="form.username"
            autocomplete="username"
            placeholder="手机号或邮箱"
          />
        </AdminFormField>

        <AdminFormField label="密码" :error="errors.password" required>
          <AdminInput
            v-model="form.password"
            type="password"
            autocomplete="current-password"
            placeholder="请输入密码"
            revealable
          />
        </AdminFormField>

        <AdminButton type="submit" variant="primary" class="w-full" :loading="loading">
          登录
        </AdminButton>
      </form>

      <div class="mt-6 flex items-center justify-center">
        <button
          type="button"
          class="inline-flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-muted transition-colors hover:bg-surface-muted hover:text-foreground"
          @click="themeStore.toggle()"
        >
          <SunIcon v-if="themeStore.isDark" class="h-5 w-5" />
          <MoonIcon v-else class="h-5 w-5" />
          <span>{{ themeStore.isDark ? "浅色模式" : "深色模式" }}</span>
        </button>
      </div>
    </div>
  </div>
</template>
