<script setup lang="ts">
defineProps<{
  label: string;
  hint?: string;
  required?: boolean;
}>();
</script>

<template>
  <div class="space-y-1.5">
    <label class="block text-sm font-medium text-foreground">
      {{ label }}
      <span v-if="required" class="text-primary-500">*</span>
    </label>
    <slot />
    <p v-if="hint" class="text-xs leading-relaxed text-muted">{{ hint }}</p>
  </div>
</template>
