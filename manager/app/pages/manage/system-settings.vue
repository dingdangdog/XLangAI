<script setup lang="ts">
const { t, te } = useI18n();
const API = "/api/admin/system-settings";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

const page = ref(1);
const pageSize = ref(50);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);

async function load() {
  loading.value = true;
  try {
    const res = await api.list({ page: page.value, pageSize: pageSize.value });
    list.value = res.items;
    total.value = res.total;
  } catch (e) {
    toast.error(t("toast.loadFailed"));
    console.error(e);
  } finally {
    loading.value = false;
  }
}

watch([page, pageSize], () => void load(), { immediate: true });

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

const form = reactive({
  id: "",
  key: "",
  value: "",
  valueType: "bool",
  description: "",
});

function labelForKey(k: string) {
  const i18nKey = `pages.systemSettings.keys.${k}`;
  return te(i18nKey) ? t(i18nKey) : k;
}

function storageOptionsForKey(k: string) {
  if (k === "media.user_recording.storage") return storageOptionsAll.value;
  return storageOptionsNoClient.value;
}

const storageOptionsAll = computed(() => [
  { value: "client", label: t("pages.systemSettings.storageClient") },
  { value: "server", label: t("pages.systemSettings.storageServer") },
  { value: "cloud", label: t("pages.systemSettings.storageCloud") },
]);

const storageOptionsNoClient = computed(() =>
  storageOptionsAll.value.filter((o) => o.value !== "client"),
);

function resetForm() {
  form.id = "";
  form.key = "";
  form.value = "true";
  form.valueType = "bool";
  form.description = "";
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  form.key = String(row.key ?? "");
  form.value = String(row.value ?? "");
  form.valueType = String(row.valueType ?? "string");
  form.description = String(row.description ?? "");
  dialogVisible.value = true;
}

function payload() {
  return {
    key: form.key.trim(),
    value: String(form.value).trim(),
    valueType: form.valueType,
    description: form.description.trim() || null,
  };
}

async function submit() {
  if (dialogMode.value === "create" && !form.key.trim()) {
    toast.warning(t("validation.fillKey"));
    return;
  }
  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await api.create(payload());
      toast.success(t("toast.created"));
    } else {
      await api.update(form.id, { id: form.id, ...payload(), key: form.key });
      toast.success(t("toast.saved"));
    }
    dialogVisible.value = false;
    await load();
  } catch (e) {
    toast.error(t("toast.saveFailed"));
    console.error(e);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: t("confirm.deleteSystemSetting", { key: String(row.key ?? "") }),
    danger: true,
    confirmLabel: t("common.delete"),
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success(t("toast.deleted"));
    await load();
  } catch (e) {
    toast.error(t("toast.deleteFailed"));
    console.error(e);
  }
}

const valueTypeOptions = [
  { value: "bool", label: "bool" },
  { value: "string", label: "string" },
  { value: "json", label: "json" },
];

const boolValueOptions = computed(() => [
  { value: "true", label: t("status.trueOn") },
  { value: "false", label: t("status.falseOff") },
]);
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.systemSettings.title')"
        :description="$t('pages.systemSettings.description')"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>Key</AdminTh>
          <AdminTh>{{ $t("common.description") }}</AdminTh>
          <AdminTh>{{ $t("common.value") }}</AdminTh>
          <AdminTh width="72px">{{ $t("common.type") }}</AdminTh>
          <AdminTh>{{ $t("common.remark") }}</AdminTh>
          <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>
          <AdminTh width="140px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap><code class="text-xs">{{ row.key }}</code></AdminTd>
          <AdminTd>{{ labelForKey(String(row.key ?? '')) }}</AdminTd>
          <AdminTd>
            <span v-if="row.valueType === 'bool'">
              {{ row.value === 'true' ? $t('status.on') : $t('status.off') }}
            </span>
            <span v-else>{{ row.value }}</span>
          </AdminTd>
          <AdminTd>{{ row.valueType }}</AdminTd>
          <AdminTd>{{ row.description ?? t("common.emDash") }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
          <AdminTd align="right" nowrap>
            <AdminButton variant="link" @click="openEdit(row)">{{ $t("common.edit") }}</AdminButton>
            <AdminButton variant="link" class="!text-danger-600" @click="removeRow(row)">
              {{ $t("common.delete") }}
            </AdminButton>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? $t('pages.systemSettings.createDialog') : $t('pages.systemSettings.editDialog')"
    >
      <AdminFormField label="Key" required>
        <AdminInput
          v-model="form.key"
          :disabled="dialogMode === 'edit'"
          :placeholder="$t('pages.systemSettings.keyPlaceholder')"
        />
      </AdminFormField>
      <AdminFormField :label="$t('common.type')">
        <AdminSelect v-model="form.valueType" :options="valueTypeOptions" :disabled="dialogMode === 'edit'" />
      </AdminFormField>
      <AdminFormField v-if="form.valueType === 'bool'" :label="$t('common.value')">
        <AdminSelect v-model="form.value" :options="boolValueOptions" />
      </AdminFormField>
      <AdminFormField v-else-if="form.key.startsWith('media.')" :label="$t('common.value')">
        <AdminSelect v-model="form.value" :options="storageOptionsForKey(form.key)" />
      </AdminFormField>
      <AdminFormField v-else :label="$t('common.value')">
        <AdminInput v-model="form.value" />
      </AdminFormField>
      <AdminFormField :label="$t('common.remark')">
        <AdminInput v-model="form.description" type="textarea" :rows="2" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">{{ $t("common.save") }}</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
