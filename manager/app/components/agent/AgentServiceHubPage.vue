<script setup lang="ts">
export type HubTile = {
  vendorId: string;
  title: string;
  subtitle: string;
  model: string;
  active: boolean;
  meta?: string;
  row: Record<string, unknown>;
};

const page = defineModel<number>("page", { default: 1 });
const pageSize = defineModel<number>("pageSize", { default: 24 });

defineProps<{
  lead: string;
  activeCount: number;
  addLabel: string;
  emptyTitle: string;
  emptyBody: string;
  addFirstLabel: string;
  tiles: HubTile[];
  loading: boolean;
  total: number;
}>();

const emit = defineEmits<{ create: []; edit: [row: Record<string, unknown>] }>();
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="mb-6 flex flex-wrap items-center justify-between gap-4">
      <p class="text-sm text-muted">
        {{ lead }}
        <span v-if="activeCount" class="text-primary-600 dark:text-primary-400">
          · {{ $t("agentHub.activeCount", { count: activeCount }) }}
        </span>
      </p>
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-full bg-primary-600 px-5 py-2.5 text-sm font-medium text-white shadow-lg shadow-primary-600/25 transition hover:bg-primary-700 active:scale-[0.98]"
        @click="emit('create')"
      >
        <span class="text-lg leading-none">+</span>
        {{ addLabel }}
      </button>
    </div>

    <div v-if="loading && !tiles.length" class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <div v-for="i in 6" :key="i" class="h-36 animate-pulse rounded-2xl bg-surface-muted/80" />
    </div>

    <div
      v-else-if="!tiles.length"
      class="flex flex-1 flex-col items-center justify-center rounded-3xl border border-dashed border-border/80 bg-gradient-to-b from-surface to-surface-muted/20 px-8 py-16 text-center"
    >
      <div
        class="mb-5 flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-accent-500 text-2xl text-white shadow-lg shadow-primary-500/30"
      >
        ✦
      </div>
      <h3 class="text-lg font-semibold text-foreground">{{ emptyTitle }}</h3>
      <p class="mt-2 max-w-sm text-sm leading-relaxed text-muted">{{ emptyBody }}</p>
      <button
        type="button"
        class="mt-6 rounded-full bg-primary-600 px-6 py-2.5 text-sm font-medium text-white shadow-md hover:bg-primary-700"
        @click="emit('create')"
      >
        {{ addFirstLabel }}
      </button>
    </div>

    <div v-else class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <AgentConnectionTile
        v-for="tile in tiles"
        :key="String(tile.row.id)"
        :vendor-id="tile.vendorId"
        :title="tile.title"
        :subtitle="tile.subtitle"
        :model="tile.model"
        :active="tile.active"
        :meta="tile.meta"
        @click="emit('edit', tile.row)"
      />
    </div>

    <AdminPagination
      v-if="total > pageSize"
      v-model:page="page"
      v-model:page-size="pageSize"
      :total="total"
      class="mt-6 shrink-0 rounded-2xl border border-border/60 bg-surface"
    />
  </div>
</template>
