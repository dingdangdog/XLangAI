<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string;
    variant?: "info" | "warning";
  }>(),
  { variant: "info" }
);

const styles = {
  info: "border-primary-200 bg-primary-50 text-primary-900 dark:border-primary-800 dark:bg-primary-950/30 dark:text-primary-100",
  warning: "border-amber-200 bg-amber-50 text-amber-900 dark:border-amber-800 dark:bg-amber-950/30 dark:text-amber-100",
};
</script>

<template>
  <div class="mb-4 rounded-xl border px-4 py-3 text-sm" :class="styles[variant]">
    <p v-if="title" class="mb-1 font-semibold">{{ title }}</p>
    <div class="text-muted [&_code]:rounded [&_code]:bg-surface-muted [&_code]:px-1">
      <slot />
    </div>
  </div>
</template>
