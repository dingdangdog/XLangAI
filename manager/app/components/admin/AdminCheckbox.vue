<script setup lang="ts">
const model = defineModel<boolean>({ required: true });
defineProps<{ label: string }>();
</script>

<template>
  <label class="inline-flex cursor-pointer items-center gap-2 text-sm text-foreground">
    <input
      v-model="model"
      type="checkbox"
      class="h-4 w-4 rounded border-border text-primary-600 focus:ring-primary-500"
    />
    {{ label }}
  </label>
</template>
