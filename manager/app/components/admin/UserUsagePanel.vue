<script setup lang="ts">
const props = defineProps<{
  userId: string;
  userLabel?: string;
}>();

const { t } = useI18n();
const { usageCountCharsLine, formatAudioBytes } = useUsageDisplay();
const api = useAdminResourceApi("/api/admin/user-usage");
const toast = useToast();

const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);

async function load() {
  if (!props.userId) return;
  loading.value = true;
  try {
    const res = await api.list({
      page: page.value,
      pageSize: pageSize.value,
      userId: props.userId,
    });
    list.value = res.items;
    total.value = res.total;
  } catch (error) {
    toast.error(t("toast.loadUsageFailed"));
    console.error(error);
  } finally {
    loading.value = false;
  }
}

watch(
  () => props.userId,
  () => {
    page.value = 1;
    void load();
  },
  { immediate: true },
);
watch([page, pageSize], () => void load());
</script>

<template>
  <div>
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <p class="text-sm text-muted">
        {{ $t("usage.userLabel") }}
        <span class="font-medium text-foreground">{{ userLabel || userId }}</span>
        · {{ $t("usage.detailIntro") }}
      </p>
      <AdminBadge variant="muted">{{ $t("pages.users.usageReadOnly") }}</AdminBadge>
    </div>

    <AdminTable :loading="loading">
      <template #head>
        <AdminTh width="120px">{{ $t("fields.date") }}</AdminTh>
        <AdminTh width="80px">{{ $t("fields.calls") }}</AdminTh>
        <AdminTh width="80px">{{ $t("fields.token") }}</AdminTh>
        <AdminTh width="100px">{{ $t("fields.translate") }}</AdminTh>
        <AdminTh width="100px">{{ $t("fields.tts") }}</AdminTh>
        <AdminTh width="100px">{{ $t("fields.stt") }}</AdminTh>
        <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>
      </template>
      <AdminTr v-for="row in list" :key="String(row.id)">
        <AdminTd>{{ formatDate(row.date) }}</AdminTd>
        <AdminTd>{{ row.usageCount }}</AdminTd>
        <AdminTd>{{ row.tokenCount }}</AdminTd>
        <AdminTd class="text-xs">
          {{ usageCountCharsLine(row.translateCount, row.translateChars) }}
        </AdminTd>
        <AdminTd class="text-xs">
          {{ usageCountCharsLine(row.ttsCount, row.ttsChars) }}
        </AdminTd>
        <AdminTd class="text-xs">
          {{
            $t("usage.sttLine", {
              count: Number(row.sttCount ?? 0),
              size: formatAudioBytes(String(row.sttAudioBytes ?? 0)),
            })
          }}
        </AdminTd>
        <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
      </AdminTr>
      <template #mobile>
        <p v-if="!list.length && !loading" class="py-12 text-center text-sm text-muted">
          {{ $t("table.noData") }}
        </p>
        <AdminMobileCard
          v-for="row in list"
          :key="String(row.id)"
          :title="formatDate(row.date)"
          :subtitle="`${$t('fields.calls')} ${row.usageCount} · ${$t('fields.token')} ${row.tokenCount}`"
        >
          <AdminMobileMeta :label="$t('fields.translate')">
            {{ usageCountCharsLine(row.translateCount, row.translateChars) }}
          </AdminMobileMeta>
          <AdminMobileMeta :label="$t('fields.tts')">
            {{ usageCountCharsLine(row.ttsCount, row.ttsChars) }}
          </AdminMobileMeta>
          <AdminMobileMeta :label="$t('fields.stt')">
            {{
              $t("usage.sttLine", {
                count: Number(row.sttCount ?? 0),
                size: formatAudioBytes(String(row.sttAudioBytes ?? 0)),
              })
            }}
          </AdminMobileMeta>
        </AdminMobileCard>
      </template>
    </AdminTable>
    <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
  </div>
</template>
