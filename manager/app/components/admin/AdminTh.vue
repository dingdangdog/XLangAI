<script setup lang="ts">
defineProps<{
  width?: string;
  align?: "left" | "right" | "center";
}>();
</script>

<template>
  <th
    class="px-4 py-3 font-medium"
    :class="{
      'text-right': align === 'right',
      'text-center': align === 'center',
    }"
    :style="width ? { width } : undefined"
  >
    <slot />
  </th>
</template>
