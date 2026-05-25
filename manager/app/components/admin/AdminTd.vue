<script setup lang="ts">
defineProps<{
  align?: "left" | "right" | "center";
  nowrap?: boolean;
}>();
</script>

<template>
  <td
    class="px-4 py-3 text-foreground"
    :class="{
      'text-right': align === 'right',
      'text-center': align === 'center',
      'whitespace-nowrap': nowrap,
      'max-w-xs truncate': !$slots.default,
    }"
  >
    <slot />
  </td>
</template>
