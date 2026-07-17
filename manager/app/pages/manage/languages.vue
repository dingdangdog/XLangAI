<script setup lang="ts">
const { t } = useI18n();

const API = "/api/admin/languages";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

const page = ref(1);
const pageSize = ref(20);
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
  code: "",
  name: "",
  nameNative: "",
  previewSampleText: "",
  sortOrder: 0,
  status: "active",
  remark: "",
});

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.nameNative = "";
  form.previewSampleText = "";
  form.sortOrder = 0;
  form.status = "active";
  form.remark = "";
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  form.code = String(row.code ?? "");
  form.name = String(row.name ?? "");
  form.nameNative = String(row.nameNative ?? "");
  form.previewSampleText = String(row.previewSampleText ?? "");
  form.sortOrder = Number(row.sortOrder ?? 0);
  form.status = String(row.status ?? "active");
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
}

function payload() {
  return {
    code: form.code.trim(),
    name: form.name.trim(),
    nameNative: form.nameNative.trim() || null,
    previewSampleText: form.previewSampleText.trim() || null,
    sortOrder: form.sortOrder,
    status: form.status,
    remark: form.remark.trim() || null,
  };
}

async function submit() {
  if (!form.code.trim() || !form.name.trim()) {
    toast.warning(t("validation.fillLanguageCodeAndName"));
    return;
  }
  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await api.create(payload());
      toast.success(t("toast.created"));
    } else {
      await api.update(form.id, { id: form.id, ...payload() });
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
    message: t("confirm.deleteLanguage"),
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

const statusOptions = [
  { value: "active", label: "active" },
  { value: "inactive", label: "inactive" },
];

const { activateRow, activatingId } = useActivateConfigRow({
  api,
  getList: () => list.value,
  exclusivity: "multi-active",
  reload: load,
});
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.languages.title')"
        :description="$t('pages.languages.description')"
      >
        <template #actions>
          <AdminButton variant="primary" class="w-full sm:w-auto" @click="openCreate">
            {{ $t("common.create") }}
          </AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh width="100px">{{ $t("common.code") }}</AdminTh>
          <AdminTh>{{ $t("common.name") }}</AdminTh>
          <AdminTh>{{ $t("fields.localName") }}</AdminTh>
          <AdminTh width="80px">{{ $t("common.sort") }}</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh>{{ $t("common.remark") }}</AdminTh>
          <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>
          <AdminTh width="200px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap>{{ row.code }}</AdminTd>
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.nameNative ?? $t("common.emDash") }}</AdminTd>
          <AdminTd>{{ row.sortOrder }}</AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd>{{ row.remark ?? $t("common.emDash") }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
          <AdminTd align="right" class="whitespace-nowrap">
            <AdminButton
              v-if="String(row.status) !== 'active'"
              variant="link"
              :loading="activatingId === String(row.id)"
              @click="activateRow(row)"
            >
              {{ $t("common.enable") }}
            </AdminButton>
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
            :title="String(row.name ?? '')"
            :subtitle="String(row.code ?? '')"
          >
            <template #badge>
              <AdminBadge>{{ row.status }}</AdminBadge>
            </template>
            <template #menu>
              <AdminOverflowMenu
                :actions="[
                  ...(String(row.status) !== 'active'
                    ? [{ label: $t('common.enable'), onClick: () => activateRow(row) }]
                    : []),
                  { label: $t('common.edit'), onClick: () => openEdit(row) },
                  { label: $t('common.delete'), danger: true, onClick: () => removeRow(row) },
                ]"
              />
            </template>
            <AdminMobileMeta :label="$t('fields.localName')">
              {{ row.nameNative ?? $t("common.emDash") }}
            </AdminMobileMeta>
            <AdminMobileMeta :label="$t('common.sort')">{{ row.sortOrder }}</AdminMobileMeta>
            <AdminMobileMeta :label="$t('common.updatedAt')">
              {{ formatDateTime(row.updatedAt) }}
            </AdminMobileMeta>
          </AdminMobileCard>
        </template>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="
        dialogMode === 'create'
          ? $t('pages.languages.createDialog')
          : $t('pages.languages.editDialog')
      "
    >
      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField :label="$t('common.code')" required>
        <AdminInput
          v-model="form.code"
          :disabled="dialogMode === 'edit'"
          :placeholder="$t('pages.languages.codePlaceholder')"
        />
      </AdminFormField>
      <AdminFormField :label="$t('common.name')" required>
        <AdminInput v-model="form.name" :placeholder="$t('pages.languages.namePlaceholder')" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.localName')">
        <AdminInput v-model="form.nameNative" :placeholder="$t('pages.languages.nativeNamePlaceholder')" />
      </AdminFormField>
      <AdminFormField
        :label="$t('fields.previewSampleText')"
        :hint="$t('pages.languages.previewHint')"
      >
        <AdminInput
          v-model="form.previewSampleText"
          type="textarea"
          :rows="2"
          :placeholder="$t('pages.languages.previewPlaceholder')"
        />
      </AdminFormField>
      <AdminFormField :label="$t('common.sort')">
        <AdminInput v-model="form.sortOrder" type="number" />
      </AdminFormField>
      <AdminFormField :label="$t('common.status')">
        <AdminSelect v-model="form.status" :options="statusOptions" />
      </AdminFormField>
      <AdminFormField :label="$t('common.remark')">
        <AdminInput v-model="form.remark" type="textarea" :rows="2" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">
          {{ $t("common.save") }}
        </AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
