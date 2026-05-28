<script setup lang="ts">
withDefaults(
  defineProps<{
    variant?: "primary" | "secondary" | "danger" | "ghost" | "link";
    size?: "sm" | "md";
    loading?: boolean;
    disabled?: boolean;
    type?: "button" | "submit";
  }>(),
  {
    variant: "secondary",
    size: "md",
    loading: false,
    disabled: false,
    type: "button",
  }
);

const base =
  "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-50";
const sizes = { sm: "px-2 py-0.5 text-xs", md: "px-4 py-1 text-sm" };
const variants = {
  primary: "bg-primary-600 text-white hover:bg-primary-700",
  secondary:
    "border border-border bg-transparent text-foreground hover:bg-surface-muted",
  danger: "bg-danger-600 text-white hover:bg-danger-500",
  ghost: "text-muted hover:bg-surface-muted hover:text-foreground",
  link: "text-primary-600 hover:text-primary-700 p-0",
};
</script>

<template>
  <button :type="type" :disabled="disabled || loading" :class="[base, sizes[size], variants[variant]]">
    <span v-if="loading" class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
    <slot />
  </button>
</template>
