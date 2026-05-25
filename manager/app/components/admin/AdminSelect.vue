<script setup lang="ts">
export interface SelectOption {
  value: string;
  label: string;
}

const model = defineModel<string>({ required: true });

defineProps<{
  options: SelectOption[];
  placeholder?: string;
  disabled?: boolean;
}>();

const selectClass =
  "w-full rounded-lg border border-border bg-surface-muted/50 px-3 py-2 text-sm text-foreground focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 disabled:cursor-not-allowed disabled:opacity-60";
</script>

<template>
  <select v-model="model" :disabled="disabled" :class="selectClass">
    <option v-if="placeholder" value="" disabled>{{ placeholder }}</option>
    <option v-for="o in options" :key="o.value" :value="o.value">{{ o.label }}</option>
  </select>
</template>
