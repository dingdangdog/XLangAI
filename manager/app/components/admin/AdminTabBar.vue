<script setup lang="ts">
import type { Component } from "vue";

export interface AdminTabItem {
  key: string;
  label: string;
  icon?: Component;
}

defineProps<{
  tabs: AdminTabItem[];
  activeTab: string;
}>();

const emit = defineEmits<{
  change: [key: string];
}>();
</script>

<template>
  <div
    class="flex flex-wrap gap-1 rounded-lg border border-border bg-surface-muted p-1"
    role="tablist"
  >
    <button
      v-for="tab in tabs"
      :key="tab.key"
      type="button"
      role="tab"
      class="inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
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
