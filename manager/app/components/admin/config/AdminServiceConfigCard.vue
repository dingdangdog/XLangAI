<script setup lang="ts">
const props = defineProps<{
  name: string;
  subtitle?: string;
  badge?: string;
  badgeVariant?: "default" | "success" | "warning" | "danger" | "muted";
  selected?: boolean;
  active?: boolean;
  meta?: string;
}>();

const emit = defineEmits<{ select: [] }>();
</script>

<template>
  <button
    type="button"
    class="w-full rounded-xl border px-3 py-2.5 text-left transition-all"
    :class="
      selected
        ? 'border-primary-400 bg-primary-50/70 shadow-sm dark:border-primary-500/60 dark:bg-primary-950/30'
        : 'border-border bg-surface hover:border-primary-200 hover:bg-surface-muted/60'
    "
    @click="emit('select')"
  >
    <div class="flex items-start gap-2.5">
      <span
        class="mt-1.5 h-2 w-2 shrink-0 rounded-full"
        :class="active ? 'bg-primary-500 shadow-[0_0_0_3px_rgba(16,185,129,0.15)]' : 'bg-gray-300 dark:bg-gray-600'"
        aria-hidden="true"
      />
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="truncate font-medium text-foreground">{{ name }}</span>
          <AdminBadge v-if="badge" :variant="badgeVariant ?? 'default'" class="shrink-0">
            {{ badge }}
          </AdminBadge>
        </div>
        <p v-if="subtitle" class="mt-0.5 truncate text-xs text-muted">{{ subtitle }}</p>
        <p v-if="meta" class="mt-1 truncate text-[11px] text-muted">{{ meta }}</p>
      </div>
    </div>
  </button>
</template>
