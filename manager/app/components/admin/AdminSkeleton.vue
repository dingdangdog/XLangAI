<script setup lang="ts">
withDefaults(defineProps<{ rows?: number }>(), { rows: 6 });
</script>

<template>
  <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
    <div
      v-for="i in rows"
      :key="i"
      class="h-24 animate-pulse rounded-xl border border-border bg-surface-muted"
    />
  </div>
</template>
