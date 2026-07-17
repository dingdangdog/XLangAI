<script setup lang="ts">
const { t } = useI18n();

const API = "/api/admin/conversations-backup";
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
      <div class="w-full sm:w-56">
        <AdminInput
          v-model="filterBatch"
          :placeholder="$t('common.filterByBackupBatch')"
        />
      </div>
    </div>
    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("fields.originalConversationId") }}</AdminTh>
          <AdminTh>{{ $t("fields.title") }}</AdminTh>
          <AdminTh>{{ $t("fields.userId") }}</AdminTh>
          <AdminTh>{{ $t("fields.languageId") }}</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh>{{ $t("fields.batch") }}</AdminTh>
          <AdminTh>{{ $t("fields.backupTime") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id) + String(row.backupBatch)">
          <AdminTd>{{ row.id }}</AdminTd>
          <AdminTd>{{ row.title }}</AdminTd>
          <AdminTd>{{ row.userId }}</AdminTd>
          <AdminTd>{{ row.languageId }}</AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd>{{ row.backupBatch }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.cancelledAt) }}</AdminTd>
        </AdminTr>
        <template #mobile>
          <p v-if="!list.length && !loading" class="py-12 text-center text-sm text-muted">
            {{ $t("table.noData") }}
          </p>
          <AdminMobileCard
            v-for="row in list"
            :key="String(row.id) + String(row.backupBatch)"
            :title="String(row.title ?? row.id)"
            :subtitle="String(row.userId ?? '')"
          >
            <template #badge>
              <AdminBadge>{{ row.status }}</AdminBadge>
            </template>
            <AdminMobileMeta :label="$t('fields.originalConversationId')">
              {{ row.id }}
            </AdminMobileMeta>
            <AdminMobileMeta :label="$t('fields.languageId')">{{ row.languageId }}</AdminMobileMeta>
            <AdminMobileMeta :label="$t('fields.batch')">{{ row.backupBatch }}</AdminMobileMeta>
            <AdminMobileMeta :label="$t('fields.backupTime')">
              {{ formatDateTime(row.cancelledAt) }}
            </AdminMobileMeta>
          </AdminMobileCard>
        </template>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>
  </div>
</template>
