<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: "default" | "success" | "warning" | "danger" | "muted";
  }>(),
  { variant: "default" }
);

const styles = {
  default: "bg-primary-50 text-primary-700 dark:bg-primary-950/40 dark:text-primary-300",
  success: "bg-primary-50 text-primary-700 dark:bg-primary-950/40 dark:text-primary-300",
  warning: "bg-amber-50 text-amber-700 dark:bg-amber-950/30 dark:text-amber-300",
  danger: "bg-danger-50 text-danger-600 dark:bg-danger-50/10 dark:text-danger-500",
  muted: "bg-surface-muted text-muted",
};
</script>

<template>
  <span class="inline-flex rounded-md px-2 py-0.5 text-xs font-medium" :class="styles[variant]">
    <slot />
  </span>
</template>
