<script setup lang="ts">
const API = "/api/admin/messages-backup";
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
    toast.error("加载失败");
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
        <AdminInput v-model="filterBatch" placeholder="按 backupBatch 筛选" />
      </div>
    </div>
    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>原消息 ID</AdminTh>
          <AdminTh width="88px">角色</AdminTh>
          <AdminTh>内容</AdminTh>
          <AdminTh>会话 ID</AdminTh>
          <AdminTh>批次</AdminTh>
          <AdminTh>原创建时间</AdminTh>
          <AdminTh>备份时间</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id) + String(row.backupBatch)">
          <AdminTd>{{ row.id }}</AdminTd>
          <AdminTd><AdminBadge>{{ row.role }}</AdminBadge></AdminTd>
          <AdminTd>{{ row.content }}</AdminTd>
          <AdminTd>{{ row.conversationId }}</AdminTd>
          <AdminTd>{{ row.backupBatch }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.createdAt) }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.cancelledAt) }}</AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>
  </div>
</template>
