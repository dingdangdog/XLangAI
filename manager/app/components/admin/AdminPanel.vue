<script setup lang="ts">
withDefaults(
  defineProps<{
    /** 占满剩余高度并在内部滚动（列表页 + AdminTable）；概览等长内容页设为 false */
    fill?: boolean;
  }>(),
  { fill: true },
);
</script>

<template>
  <div
    class="flex flex-col rounded-2xl border border-border bg-surface"
    :class="fill ? 'min-h-0 flex-1 overflow-hidden' : ''"
  >
    <slot></slot>
  </div>
</template>
