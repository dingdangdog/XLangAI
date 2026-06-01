<script setup lang="ts">
import AdminServiceConfigCard from "~/components/admin/config/AdminServiceConfigCard.vue";

const props = withDefaults(
  defineProps<{
    items: Record<string, unknown>[];
    selectedId?: string | null;
    loading?: boolean;
    total?: number;
    activeLabel?: string;
    inactiveLabel?: string;
    getTitle: (row: Record<string, unknown>) => string;
    getSubtitle: (row: Record<string, unknown>) => string;
    getMeta?: (row: Record<string, unknown>) => string;
    getBadge?: (row: Record<string, unknown>) => { text: string; variant?: "default" | "success" | "warning" | "danger" | "muted" };
  }>(),
  {
    selectedId: null,
    loading: false,
    total: 0,
    activeLabel: "active",
    inactiveLabel: "inactive",
  },
);

const pageModel = defineModel<number>("page", { default: 1 });
const pageSizeModel = defineModel<number>("pageSize", { default: 20 });

const emit = defineEmits<{
  create: [];
  select: [row: Record<string, unknown>];
}>();

const activeCount = computed(
  () => props.items.filter((r) => String(r.status) === "active").length,
);
const inactiveCount = computed(() => Math.max(0, props.total - activeCount.value));
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col gap-3">
    <div class="grid gap-3 sm:grid-cols-3">
      <AdminServiceConfigStat :label="$t('serviceConfig.statsTotal')" :value="total" />
      <AdminServiceConfigStat :label="$t('serviceConfig.statsActive')" :value="activeCount" tone="success"
        :hint="$t('serviceConfig.statsActiveHint')" />
      <AdminServiceConfigStat :label="$t('serviceConfig.statsInactive')" :value="inactiveCount" tone="default" />
    </div>

    <div class="flex min-h-0 flex-1 gap-4">
      <aside class="flex w-full shrink-0 flex-col gap-3 lg:w-80">
        <div class="flex items-center justify-between gap-2">
          <h3 class="text-sm font-medium text-foreground">{{ $t("serviceConfig.configList") }}</h3>
          <AdminButton variant="primary" size="sm" @click="emit('create')">
            {{ $t("common.create") }}
          </AdminButton>
        </div>

        <div v-if="loading && !items.length" class="space-y-2">
          <AdminSkeleton v-for="i in 4" :key="i" class="h-16 rounded-xl" />
        </div>

        <div v-else-if="!items.length" class="rounded-xl border border-dashed border-border px-4 py-8 text-center">
          <p class="text-sm text-muted">{{ $t("serviceConfig.emptyList") }}</p>
          <AdminButton class="mt-3" variant="secondary" size="sm" @click="emit('create')">
            {{ $t("serviceConfig.createFirst") }}
          </AdminButton>
        </div>

        <div v-else class="min-h-0 flex-1 space-y-2 overflow-y-auto pr-1">
          <AdminServiceConfigCard v-for="row in items" :key="String(row.id)" :name="getTitle(row)"
            :subtitle="getSubtitle(row)" :meta="getMeta?.(row)"
            :badge="getBadge?.(row)?.text ?? (String(row.status) === 'active' ? activeLabel : inactiveLabel)"
            :badge-variant="getBadge?.(row)?.variant ?? (String(row.status) === 'active' ? 'success' : 'muted')"
            :active="String(row.status) === 'active'" :selected="selectedId === String(row.id)"
            @select="emit('select', row)" />
        </div>

        <AdminPagination v-if="total > pageSizeModel" v-model:page="pageModel" v-model:page-size="pageSizeModel"
          :total="total" class="shrink-0" />
      </aside>

      <main class="min-h-0 min-w-0 flex-1">
        <slot />
      </main>
    </div>
  </div>
</template>
