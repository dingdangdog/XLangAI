<script setup lang="ts">
const props = defineProps<{
  userId: string;
  userLabel?: string;
}>();

const { t } = useI18n();
const { usageCountCharsLine, formatAudioBytes } = useUsageDisplay();

const API = "/api/admin/user-usage";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

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
  } catch (e) {
    toast.error(t("toast.loadUsageFailed"));
    console.error(e);
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

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

const form = reactive({
  id: "",
  date: "",
  usageCount: 0,
  tokenCount: 0,
});

function resetForm() {
  form.id = "";
  form.date = new Date().toISOString().slice(0, 10);
  form.usageCount = 0;
  form.tokenCount = 0;
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  const d = row.date;
  form.date =
    typeof d === "string"
      ? d.slice(0, 10)
      : d instanceof Date
        ? d.toISOString().slice(0, 10)
        : String(d ?? "").slice(0, 10);
  form.usageCount = Number(row.usageCount ?? 0);
  form.tokenCount = Number(row.tokenCount ?? 0);
  dialogVisible.value = true;
}

async function submit() {
  if (!form.date) {
    toast.warning(t("validation.selectDate"));
    return;
  }
  const body = {
    userId: props.userId,
    date: new Date(`${form.date}T12:00:00.000Z`).toISOString(),
    usageCount: form.usageCount,
    tokenCount: form.tokenCount,
  };
  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await api.create(body);
      toast.success(t("toast.created"));
    } else {
      await api.update(form.id, { id: form.id, ...body });
      toast.success(t("toast.saved"));
    }
    dialogVisible.value = false;
    await load();
    emit("changed");
  } catch (e) {
    toast.error(t("toast.saveFailed"));
    console.error(e);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: t("confirm.deleteUsageRecord"),
    danger: true,
    confirmLabel: t("common.delete"),
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success(t("toast.deleted"));
    await load();
    emit("changed");
  } catch (e) {
    toast.error(t("toast.deleteFailed"));
    console.error(e);
  }
}

const emit = defineEmits<{ changed: [] }>();
</script>

<template>
  <div>
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <p class="text-sm text-muted">
        {{ $t("usage.userLabel") }}
        <span class="font-medium text-foreground">{{ userLabel || userId }}</span>
        · {{ $t("usage.detailIntro") }}
      </p>
      <AdminButton variant="primary" size="sm" @click="openCreate">
        {{ $t("pages.userUsage.createRecord") }}
      </AdminButton>
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
        <AdminTh width="140px" align="right">{{ $t("common.actions") }}</AdminTh>
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
        <AdminTd align="right">
          <AdminButton variant="link" @click="openEdit(row)">{{ $t("common.edit") }}</AdminButton>
          <AdminButton variant="link" class="!text-danger-600" @click="removeRow(row)">
            {{ $t("common.delete") }}
          </AdminButton>
        </AdminTd>
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
            <template #menu>
              <AdminOverflowMenu
                :actions="[
                  { label: $t('common.edit'), onClick: () => openEdit(row) },
                  { label: $t('common.delete'), danger: true, onClick: () => removeRow(row) },
                ]"
              />
            </template>
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

    <AdminDialog
      v-model="dialogVisible"
      :title="
        dialogMode === 'create'
          ? $t('pages.userUsage.createDialog')
          : $t('pages.userUsage.editDialog')
      "
    >
      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField :label="$t('fields.date')" required>
        <AdminInput
          v-model="form.date"
          type="text"
          placeholder="YYYY-MM-DD"
          :disabled="dialogMode === 'edit'"
        />
      </AdminFormField>
      <AdminFormField :label="$t('fields.usageCount')">
        <AdminInput v-model="form.usageCount" type="number" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.token')">
        <AdminInput v-model="form.tokenCount" type="number" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">
          {{ $t("common.save") }}
        </AdminButton>
      </template>
    </AdminDialog>
  </div>
</template>
