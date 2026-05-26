<script setup lang="ts">
const { t } = useI18n();
const localePath = useLocalePath();

const { data: stats, pending } = await useFetch("/api/admin/stats");

const cards = computed(() => {
  const s = stats.value;
  if (!s) return [];
  return [
    { label: t("pages.dashboard.users"), value: s.users, to: localePath("/manage/users") },
    { label: t("pages.dashboard.languages"), value: s.languages, to: localePath("/manage/languages") },
    {
      label: t("pages.dashboard.llmConfigs"),
      value: s.llmConfigs,
      to: localePath("/manage/llm-service-configs"),
    },
    {
      label: t("pages.dashboard.sttConfigs"),
      value: s.sttConfigs,
      to: localePath("/manage/stt-service-configs"),
    },
    {
      label: t("pages.dashboard.ttsConfigs"),
      value: s.ttsConfigs,
      to: localePath("/manage/tts-service-configs"),
    },
    {
      label: t("pages.dashboard.voiceRoles"),
      value: s.voiceRoles,
      to: localePath("/manage/voice-roles"),
    },
    {
      label: t("pages.dashboard.promptTemplates"),
      value: s.promptTemplates,
      to: localePath("/manage/prompt-templates"),
    },
    { label: t("pages.dashboard.tiers"), value: s.tiers, to: localePath("/manage/membership-tiers") },
    {
      label: t("pages.dashboard.conversations"),
      value: s.conversations,
      to: localePath("/manage/conversations"),
    },
    { label: t("pages.dashboard.messages"), value: s.messages, to: localePath("/manage/messages") },
    {
      label: t("pages.dashboard.serverStore"),
      value: t("common.config"),
      to: localePath("/manage/server-store"),
    },
  ];
});
</script>

<template>
  <div class="min-h-0 flex-1 overflow-y-auto">
    <AdminPageHeader
      :title="$t('pages.dashboard.title')"
      :description="$t('pages.dashboard.description')"
    />
    <AdminSkeleton v-if="pending" :rows="6" />
    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <NuxtLink
        v-for="c in cards"
        :key="c.label"
        :to="c.to"
        class="block rounded-xl border border-border bg-surface p-4 transition-colors hover:border-primary-300 hover:bg-surface-muted"
      >
        <div class="text-sm text-muted">{{ c.label }}</div>
        <div class="mt-2 text-2xl font-bold text-foreground">{{ c.value }}</div>
      </NuxtLink>
    </div>
  </div>
</template>
