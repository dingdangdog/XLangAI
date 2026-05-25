<script setup lang="ts">
const { data: stats, pending } = await useFetch("/api/admin/stats");

const cards = computed(() => {
  const s = stats.value;
  if (!s) return [];
  return [
    { label: "用户", value: s.users, to: "/manage/users" },
    { label: "语言", value: s.languages, to: "/manage/languages" },
    { label: "LLM 配置", value: s.llmConfigs, to: "/manage/llm-service-configs" },
    { label: "STT 配置", value: s.sttConfigs, to: "/manage/stt-service-configs" },
    { label: "TTS 配置", value: s.ttsConfigs, to: "/manage/tts-service-configs" },
    { label: "语音角色", value: s.voiceRoles, to: "/manage/voice-roles" },
    { label: "提示词模板", value: s.promptTemplates, to: "/manage/prompt-templates" },
    { label: "会员等级", value: s.tiers, to: "/manage/membership-tiers" },
    { label: "会话", value: s.conversations, to: "/manage/conversations" },
    { label: "消息", value: s.messages, to: "/manage/messages" },
    { label: "服务器商店", value: "配置", to: "/manage/server-store" },
  ];
});
</script>

<template>
  <div class="min-h-0 flex-1 overflow-y-auto">
    <AdminPageHeader title="数据概览" description="各业务实体数量统计，点击卡片进入管理" />
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
