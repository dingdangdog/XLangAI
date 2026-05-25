<script setup lang="ts">
defineProps<{
  loading?: boolean;
}>();
</script>

<template>
  <div class="relative min-h-0 flex-1 overflow-auto">
    <div
      v-if="loading"
      class="absolute inset-0 z-10 flex items-center justify-center bg-surface/60"
    >
      <span
        class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600 border-t-transparent"
      />
    </div>
    <table class="w-full min-w-[640px] text-left text-sm">
      <thead class="sticky top-0 z-[1] bg-surface">
        <tr class="border-b border-border text-xs uppercase tracking-wide text-muted">
          <slot name="head" />
        </tr>
      </thead>
      <tbody class="divide-y divide-border">
        <slot />
      </tbody>
    </table>
    <div v-if="!$slots.default && !loading" class="py-12 text-center text-sm text-muted">
      暂无数据
    </div>
  </div>
</template>
