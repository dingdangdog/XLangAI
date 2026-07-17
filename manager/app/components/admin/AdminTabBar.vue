<script setup lang="ts">
import type { Component } from "vue";

export interface AdminTabItem {
  key: string;
  label: string;
  icon?: Component;
}

const props = defineProps<{
  tabs: AdminTabItem[];
  activeTab: string;
}>();

const emit = defineEmits<{
  change: [key: string];
}>();

const navRef = ref<HTMLElement | null>(null);

watch(
  () => props.activeTab,
  async () => {
    await nextTick();
    const el = navRef.value?.querySelector<HTMLElement>("[data-active='true']");
    el?.scrollIntoView({ inline: "center", block: "nearest", behavior: "smooth" });
  },
  { immediate: true },
);
</script>

<template>
  <div
    ref="navRef"
    class="flex gap-1 overflow-x-auto rounded-lg border border-border bg-surface-muted p-1 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
    role="tablist"
  >
    <button
      v-for="tab in tabs"
      :key="tab.key"
      type="button"
      role="tab"
      class="inline-flex shrink-0 items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
      :data-active="activeTab === tab.key ? 'true' : undefined"
      :class="
        activeTab === tab.key
          ? 'bg-surface text-foreground shadow-sm'
          : 'text-muted hover:text-foreground'
      "
      :aria-selected="activeTab === tab.key"
      @click="emit('change', tab.key)"
    >
      <component :is="tab.icon" v-if="tab.icon" class="h-4 w-4 shrink-0" />
      {{ tab.label }}
    </button>
  </div>
</template>
