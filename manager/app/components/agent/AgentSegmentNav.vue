<script setup lang="ts">
import type { Component } from "vue";

export type SegmentItem = { key: string; label: string; icon?: Component };

const props = defineProps<{ items: SegmentItem[]; active: string }>();
const emit = defineEmits<{ change: [key: string] }>();

const navRef = ref<HTMLElement | null>(null);

watch(
  () => props.active,
  async () => {
    await nextTick();
    const el = navRef.value?.querySelector<HTMLElement>("[data-active='true']");
    el?.scrollIntoView({ inline: "center", block: "nearest", behavior: "smooth" });
  },
  { immediate: true },
);
</script>

<template>
  <nav
    ref="navRef"
    class="flex gap-1 overflow-x-auto pb-0.5 [-ms-overflow-style:none] [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
  >
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      class="inline-flex shrink-0 items-center gap-2 rounded-full px-3 py-1.5 text-sm font-medium transition-all sm:px-4 sm:py-2"
      :data-active="active === item.key ? 'true' : undefined"
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
