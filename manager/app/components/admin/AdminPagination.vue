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
    class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-border px-4 py-3 text-sm text-muted"
  >
    <span>共 {{ total }} 条</span>
    <div class="flex flex-wrap items-center gap-3">
      <label class="flex items-center gap-2">
        每页
        <select
          v-model.number="pageSize"
          class="rounded-lg border border-border bg-surface px-2 py-1 text-foreground"
        >
          <option v-for="s in pageSizes" :key="s" :value="s">{{ s }}</option>
        </select>
      </label>
      <div class="flex items-center gap-1">
        <AdminButton size="sm" :disabled="page <= 1" @click="go(page - 1)">上一页</AdminButton>
        <span class="min-w-[4rem] text-center text-foreground">{{ page }} / {{ totalPages }}</span>
        <AdminButton size="sm" :disabled="page >= totalPages" @click="go(page + 1)">
          下一页
        </AdminButton>
      </div>
    </div>
  </div>
</template>
