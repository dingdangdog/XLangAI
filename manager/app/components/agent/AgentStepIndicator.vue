<script setup lang="ts">
const props = defineProps<{
  current: number;
  total: number;
}>();
</script>

<template>
  <div class="flex items-center gap-2">
    <div class="flex gap-1.5">
      <span
        v-for="i in total"
        :key="i"
        class="h-1.5 rounded-full transition-all duration-300"
        :class="i <= current ? 'w-6 bg-primary-500' : 'w-1.5 bg-border'"
      />
    </div>
    <span class="text-xs text-muted">{{ current }} / {{ total }}</span>
  </div>
</template>
