<script setup lang="ts">
const { t } = useI18n();
const API = "/api/admin/read-aloud-categories";
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
  nameEn: "",
  icon: "",
  description: "",
  descriptionEn: "",
  status: "active",
  sortOrder: 0,
  remark: "",
});

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.nameEn = "";
  form.icon = "";
  form.description = "";
  form.descriptionEn = "";
  form.status = "active";
  form.sortOrder = 0;
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
  form.nameEn = String(row.nameEn ?? "");
  form.icon = String(row.icon ?? "");
  form.description = String(row.description ?? "");
  form.descriptionEn = String(row.descriptionEn ?? "");
  form.status = String(row.status ?? "active");
  form.sortOrder = Number(row.sortOrder ?? 0);
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
}

async function submit() {
  if (!form.code.trim() || !form.name.trim()) {
    toast.warning(t("validation.fillCodeAndName"));
    return;
  }
  const body = {
    code: form.code.trim(),
    name: form.name.trim(),
    nameEn: form.nameEn.trim() || null,
    icon: form.icon.trim() || null,
    description: form.description.trim() || null,
    descriptionEn: form.descriptionEn.trim() || null,
    status: form.status,
    sortOrder: form.sortOrder,
    remark: form.remark.trim() || null,
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
  } catch (e) {
    toast.error(t("toast.saveFailed"));
    console.error(e);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: t("pages.readAloudCategories.deleteConfirm"),
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
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.readAloudCategories.title')"
        :description="$t('pages.readAloudCategories.description')"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
          <AdminButton to="/manage/read-aloud-vocabularies">{{ $t("pages.readAloudVocabularies.title") }}</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("common.code") }}</AdminTh>
          <AdminTh>{{ $t("common.name") }}</AdminTh>
          <AdminTh>{{ $t("pages.readAloudCategories.nameEn") }}</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh width="72px">{{ $t("common.sort") }}</AdminTh>
          <AdminTh width="200px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap>{{ row.code }}</AdminTd>
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.nameEn ?? t("common.emDash") }}</AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd>{{ row.sortOrder }}</AdminTd>
          <AdminTd align="right" class="whitespace-nowrap">
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
      :title="dialogMode === 'create' ? $t('pages.readAloudCategories.createDialog') : $t('pages.readAloudCategories.editDialog')"
      size="lg"
    >
      <div class="space-y-4">
        <AdminFormField :label="$t('common.code')" required>
          <AdminInput v-model="form.code" :disabled="dialogMode === 'edit'" />
        </AdminFormField>
        <AdminFormField :label="$t('common.name')" required>
          <AdminInput v-model="form.name" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudCategories.nameEn')">
          <AdminInput v-model="form.nameEn" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudCategories.icon')">
          <AdminInput v-model="form.icon" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudCategories.description')">
          <AdminTextarea v-model="form.description" rows="2" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudCategories.descriptionEn')">
          <AdminTextarea v-model="form.descriptionEn" rows="2" />
        </AdminFormField>
        <AdminFormField :label="$t('common.status')">
          <AdminSelect v-model="form.status" :options="statusOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('common.sort')">
          <AdminInput v-model.number="form.sortOrder" type="number" />
        </AdminFormField>
        <AdminFormField :label="$t('common.remark')">
          <AdminTextarea v-model="form.remark" rows="2" />
        </AdminFormField>
      </div>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">{{ $t("common.save") }}</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
