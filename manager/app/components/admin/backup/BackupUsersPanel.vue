<script setup lang="ts">
const { t } = useI18n();

const API = "/api/admin/users-backup";
const toast = useToast();

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const filterBatch = ref("");

async function load() {
  loading.value = true;
  try {
    const q: Record<string, string | number | boolean> = {
      page: page.value,
      pageSize: pageSize.value,
    };
    if (filterBatch.value.trim()) q.backupBatch = filterBatch.value.trim();
    const res = await $fetch<{ items: Record<string, unknown>[]; total: number }>(API, { query: q });
    list.value = res.items;
    total.value = res.total;
  } catch (e) {
    toast.error(t("toast.loadFailed"));
    console.error(e);
  } finally {
    loading.value = false;
  }
}

watch([page, pageSize, filterBatch], () => void load(), { immediate: true });
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col">
    <div class="mb-4 flex shrink-0 justify-end">
      <div class="w-56">
        <AdminInput
          v-model="filterBatch"
          :placeholder="$t('common.filterByBackupBatch')"
        />
      </div>
    </div>
    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("fields.originalUserId") }}</AdminTh>
          <AdminTh>{{ $t("fields.phone") }}</AdminTh>
          <AdminTh>{{ $t("fields.email") }}</AdminTh>
          <AdminTh>{{ $t("fields.nickname") }}</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh>{{ $t("fields.batch") }}</AdminTh>
          <AdminTh>{{ $t("fields.backupTime") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id) + String(row.backupBatch)">
          <AdminTd>{{ row.id }}</AdminTd>
          <AdminTd>{{ row.phone ?? $t("common.emDash") }}</AdminTd>
          <AdminTd>{{ row.email ?? $t("common.emDash") }}</AdminTd>
          <AdminTd>{{ row.nickname ?? $t("common.emDash") }}</AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd>{{ row.backupBatch }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.cancelledAt) }}</AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>
  </div>
</template>
