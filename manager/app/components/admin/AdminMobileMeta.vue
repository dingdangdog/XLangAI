<script setup lang="ts">
defineProps<{
  label: string;
}>();
</script>

<template>
  <div class="flex items-start justify-between gap-3 text-xs">
    <dt class="shrink-0 text-muted">{{ label }}</dt>
    <dd class="min-w-0 text-right text-foreground break-words">
      <slot />
    </dd>
  </div>
</template>
