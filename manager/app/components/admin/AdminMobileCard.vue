<script setup lang="ts">
defineProps<{
  title: string;
  subtitle?: string;
}>();
</script>

<template>
  <article
    class="rounded-xl border border-border bg-surface p-3 shadow-sm active:bg-surface-muted/40"
  >
    <div class="flex items-start justify-between gap-2">
      <div class="min-w-0 flex-1">
        <div class="flex flex-wrap items-center gap-2">
          <h3 class="truncate text-sm font-semibold text-foreground">{{ title }}</h3>
          <slot name="badge" />
        </div>
        <p v-if="subtitle" class="mt-0.5 truncate text-xs text-muted">{{ subtitle }}</p>
        <slot name="header-extra" />
      </div>
      <div v-if="$slots.menu" class="shrink-0">
        <slot name="menu" />
      </div>
    </div>

    <div v-if="$slots.default" class="mt-2.5 space-y-1.5 border-t border-border/70 pt-2.5">
      <slot />
    </div>

    <div
      v-if="$slots.footer"
      class="mt-2.5 flex flex-wrap items-center gap-x-3 gap-y-1 border-t border-border/70 pt-2"
    >
      <slot name="footer" />
    </div>
  </article>
</template>
