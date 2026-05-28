<script setup lang="ts">
defineProps<{
  label: string;
  value: string | number;
  hint?: string;
  tone?: "default" | "success" | "warning" | "danger";
}>();

const toneClass = {
  default: "border-border bg-surface",
  success: "border-primary-300/60 bg-primary-50/50 dark:bg-primary-950/20",
  warning: "border-amber-300/60 bg-amber-50/50 dark:bg-amber-950/20",
  danger: "border-danger-300/60 bg-danger-50/40 dark:bg-danger-950/20",
};
</script>

<template>
  <div
    class="rounded-xl border px-3 py-2.5"
    :class="toneClass[tone ?? 'default']"
  >
    <div class="text-xs text-muted">{{ label }}</div>
    <div class="mt-0.5 text-lg font-semibold tabular-nums text-foreground">{{ value }}</div>
    <p v-if="hint" class="mt-1 text-xs text-muted">{{ hint }}</p>
  </div>
</template>
