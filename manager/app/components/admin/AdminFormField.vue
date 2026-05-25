<script setup lang="ts">
defineProps<{
  label: string;
  required?: boolean;
  hint?: string;
  error?: string;
}>();
</script>

<template>
  <div class="mb-4">
    <label class="mb-1.5 block text-sm font-medium text-foreground">
      {{ label }}
      <span v-if="required" class="text-danger-500">*</span>
    </label>
    <slot />
    <p v-if="error" class="mt-1 text-xs text-danger-500">{{ error }}</p>
    <p v-else-if="hint" class="mt-1 text-xs text-muted">{{ hint }}</p>
  </div>
</template>
