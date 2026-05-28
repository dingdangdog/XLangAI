<script setup lang="ts">
withDefaults(
  defineProps<{
    /** 子内容占满剩余高度并在内部自行滚动，而非整页滚动 */
    fill?: boolean;
  }>(),
  { fill: false },
);
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col gap-2">
    <div v-if="$slots.header" class="shrink-0 space-y-4">
      <slot name="header"></slot>
    </div>
    <div class="min-h-0 flex-1" :class="fill ? 'flex flex-col overflow-hidden' : 'overflow-y-auto'">
      <slot></slot>
    </div>
  </div>
</template>
