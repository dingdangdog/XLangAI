<script setup lang="ts">
import { vendorTheme } from "~/utils/vendorTheme";

defineProps<{
  vendorId: string;
  title: string;
  subtitle: string;
  model: string;
  active: boolean;
  meta?: string;
}>();

const emit = defineEmits<{ click: [] }>();
</script>

<template>
  <button
    type="button"
    class="group relative flex w-full flex-col overflow-hidden rounded-2xl border border-border/70 bg-surface p-4 text-left shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-primary-300/50 hover:shadow-lg hover:shadow-primary-500/5"
    @click="emit('click')"
  >
    <div
      class="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-gradient-to-br opacity-[0.07] blur-2xl transition group-hover:opacity-[0.12]"
      :class="vendorTheme(vendorId).gradient"
    />

    <div class="flex items-start gap-3">
      <div
        class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br text-sm font-bold text-white shadow-sm"
        :class="vendorTheme(vendorId).gradient"
      >
        {{ vendorTheme(vendorId).abbr }}
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <h3 class="truncate font-medium text-foreground">{{ title }}</h3>
          <span
            class="inline-flex h-2 w-2 shrink-0 rounded-full"
            :class="active ? 'bg-primary-500 shadow-[0_0_0_3px_rgba(16,185,129,0.2)]' : 'bg-border'"
          />
        </div>
        <p class="mt-0.5 truncate text-xs text-muted">{{ subtitle }}</p>
      </div>
    </div>

    <div class="mt-4 flex items-end justify-between gap-2 border-t border-border/50 pt-3">
      <p class="truncate font-mono text-xs text-foreground/80">{{ model }}</p>
      <p v-if="meta" class="shrink-0 text-[11px] tabular-nums text-muted">{{ meta }}</p>
    </div>
  </button>
</template>
