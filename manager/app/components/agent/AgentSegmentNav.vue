<script setup lang="ts">
import type { Component } from "vue";

export type SegmentItem = { key: string; label: string; icon?: Component };

defineProps<{ items: SegmentItem[]; active: string }>();
const emit = defineEmits<{ change: [key: string] }>();
</script>

<template>
  <nav class="flex gap-1 overflow-x-auto pb-0.5 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      class="inline-flex shrink-0 items-center gap-2 rounded-full px-4 py-2 text-sm font-medium transition-all"
      :class="
        active === item.key
          ? 'bg-primary-600 text-white shadow-md shadow-primary-600/25'
          : 'text-muted hover:bg-surface-muted hover:text-foreground'
      "
      @click="emit('change', item.key)"
    >
      <component :is="item.icon" v-if="item.icon" class="h-4 w-4" />
      {{ item.label }}
    </button>
  </nav>
</template>
