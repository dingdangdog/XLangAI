<script setup lang="ts">
const page = defineModel<number>("page", { required: true });
const pageSize = defineModel<number>("pageSize", { required: true });

const props = defineProps<{ total: number }>();

const pageSizes = [10, 20, 50, 100];

const totalPages = computed(() =>
  Math.max(1, Math.ceil(props.total / pageSize.value))
);

function go(p: number) {
  page.value = Math.min(Math.max(1, p), totalPages.value);
}
</script>

<template>
  <div
    class="flex shrink-0 flex-col gap-2 border-t border-border px-3 py-2.5 text-sm text-muted sm:flex-row sm:flex-wrap sm:items-center sm:justify-between sm:gap-3 sm:px-4 sm:py-3"
  >
    <span class="text-xs sm:text-sm">{{ $t("pagination.total", { total }) }}</span>
    <div class="flex flex-wrap items-center justify-between gap-2 sm:justify-end sm:gap-3">
      <label class="flex items-center gap-2 text-xs sm:text-sm">
        {{ $t("pagination.perPage") }}
        <select
          v-model.number="pageSize"
          class="rounded-lg border border-border bg-surface px-2 py-1 text-foreground"
        >
          <option v-for="s in pageSizes" :key="s" :value="s">{{ s }}</option>
        </select>
      </label>
      <div class="flex items-center gap-1">
        <AdminButton size="sm" :disabled="page <= 1" @click="go(page - 1)">
          {{ $t("pagination.prev") }}
        </AdminButton>
        <span class="min-w-[3.5rem] text-center text-xs text-foreground sm:min-w-[4rem] sm:text-sm">
          {{ page }} / {{ totalPages }}
        </span>
        <AdminButton size="sm" :disabled="page >= totalPages" @click="go(page + 1)">
          {{ $t("pagination.next") }}
        </AdminButton>
      </div>
    </div>
  </div>
</template>
