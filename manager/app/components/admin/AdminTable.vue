<script setup lang="ts">
defineProps<{
  loading?: boolean;
}>();

const slots = useSlots();
const hasMobile = computed(() => !!slots.mobile);
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

    <!-- 桌面：表格；无 mobile 槽时手机也横向滚动兜底 -->
    <div :class="hasMobile ? 'hidden md:block' : ''">
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
        {{ $t("table.noData") }}
      </div>
    </div>

    <!-- 手机：卡片列表（空态由页面 #mobile 自行处理） -->
    <div v-if="hasMobile" class="space-y-2 p-2 md:hidden">
      <slot name="mobile" />
    </div>
  </div>
</template>
