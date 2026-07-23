<script setup lang="ts">
withDefaults(
  defineProps<{
    /**
     * 默认 true：页头固定，内容区占满剩余高度（配合 AdminPanel + AdminTable + AdminPagination：
     * 仅表格滚动，分页与操作栏不跟着滚）。
     * 仪表盘等需要整页滚动的长内容页设为 false。
     */
    fill?: boolean;
  }>(),
  { fill: true },
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
