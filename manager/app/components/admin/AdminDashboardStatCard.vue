<script setup lang="ts">
import type { Component } from "vue";

defineProps<{
  label: string;
  value: string | number;
  to?: string;
  icon?: Component;
}>();
</script>

<template>
  <component
    :is="to ? 'NuxtLink' : 'div'"
    :to="to"
    class="flex min-w-0 items-center gap-1.5 rounded-md border border-border px-2 py-1.5 text-sm transition-colors"
    :class="to ? 'hover:border-primary-300 hover:bg-surface-muted' : ''"
  >
    <component
      v-if="icon"
      :is="icon"
      class="h-3.5 w-3.5 shrink-0 text-muted"
      aria-hidden="true"
    />
    <span class="min-w-0 flex-1 truncate text-muted">{{ label }}</span>
    <span class="shrink-0 tabular-nums text-foreground">{{ value }}</span>
  </component>
</template>
