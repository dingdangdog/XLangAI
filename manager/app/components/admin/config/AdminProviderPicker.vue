<script setup lang="ts">
export type ProviderOption = {
  value: string;
  label: string;
  description?: string;
  group?: string;
};

defineProps<{
  options: ProviderOption[];
  modelValue: string;
}>();

const emit = defineEmits<{ "update:modelValue": [value: string] }>();
</script>

<template>
  <div class="grid gap-2 sm:grid-cols-2 xl:grid-cols-3">
    <button
      v-for="opt in options"
      :key="opt.value"
      type="button"
      class="rounded-xl border px-3 py-3 text-left transition-colors"
      :class="
        modelValue === opt.value
          ? 'border-primary-400 bg-primary-50/70 dark:border-primary-500/60 dark:bg-primary-950/30'
          : 'border-border bg-surface hover:border-primary-200 hover:bg-surface-muted/50'
      "
      @click="emit('update:modelValue', opt.value)"
    >
      <div class="text-sm font-medium text-foreground">{{ opt.label }}</div>
      <p v-if="opt.description" class="mt-1 text-xs text-muted">{{ opt.description }}</p>
      <p v-if="opt.group" class="mt-2 text-[11px] uppercase tracking-wide text-muted">{{ opt.group }}</p>
    </button>
  </div>
</template>
