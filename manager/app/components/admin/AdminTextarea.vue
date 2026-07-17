<script setup lang="ts">
const model = defineModel<string>({ required: true });

withDefaults(
  defineProps<{
    rows?: number;
    placeholder?: string;
    disabled?: boolean;
  }>(),
  { rows: 3, disabled: false },
);
</script>

<template>
  <AdminInput
    v-model="model"
    type="textarea"
    :rows="rows"
    :placeholder="placeholder"
    :disabled="disabled"
  />
</template>
