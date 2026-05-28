<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string;
    lines?: 1 | 2;
  }>(),
  {
    lines: 2,
  },
);
</script>

<template>
  <span
    class="admin-cell-text"
    :class="lines === 1 ? 'admin-cell-text--1' : 'admin-cell-text--2'"
    :title="title?.trim() || undefined"
  >
    <slot />
  </span>
</template>

<style scoped>
.admin-cell-text {
  display: block;
  min-width: 0;
  max-width: 15rem;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.admin-cell-text--1 {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.admin-cell-text--2 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
</style>
