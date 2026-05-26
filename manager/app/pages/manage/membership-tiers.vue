<script setup lang="ts">
const { t } = useI18n();
const API = "/api/admin/membership-tiers";
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
  dailyLimit: null as number | null,
  monthlyLimit: null as number | null,
  features: "{}",
  status: "active",
  sortOrder: 0,
  remark: "",
});

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.dailyLimit = null;
  form.monthlyLimit = null;
  form.features = "{}";
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
  form.dailyLimit = row.dailyLimit != null ? Number(row.dailyLimit) : null;
  form.monthlyLimit = row.monthlyLimit != null ? Number(row.monthlyLimit) : null;
  form.features = row.features != null ? String(row.features) : "{}";
  form.status = String(row.status ?? "active");
  form.sortOrder = Number(row.sortOrder ?? 0);
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
}

async function submit() {
  if (!form.code.trim() || !form.name.trim()) {
    toast.warning(t("validation.fillTierCodeAndName"));
    return;
  }
  let feat = (form.features ?? "").trim();
  if (feat) {
    try {
      JSON.parse(feat);
    } catch {
      toast.error(t("validation.invalidJson", { field: "features" }));
      return;
    }
  } else {
    feat = "{}";
  }
  const body = {
    code: form.code.trim(),
    name: form.name.trim(),
    dailyLimit: form.dailyLimit,
    monthlyLimit: form.monthlyLimit,
    features: feat,
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
    message: t("confirm.deleteTier"),
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
      <AdminPageHeader title="会员等级" description="定义套餐与用量上限；features 为 JSON 扩展字段。">
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh width="100px">编码</AdminTh>
          <AdminTh>{{ $t("common.name") }}</AdminTh>
          <AdminTh width="88px">日限额</AdminTh>
          <AdminTh width="88px">月限额</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh width="72px">{{ $t("common.sort") }}</AdminTh>
          <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>
          <AdminTh width="200px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap>{{ row.code }}</AdminTd>
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.dailyLimit ?? t("common.emDash") }}</AdminTd>
          <AdminTd>{{ row.monthlyLimit ?? t("common.emDash") }}</AdminTd>
          <AdminTd>
            <AdminBadge>{{ row.status }}</AdminBadge>
          </AdminTd>
          <AdminTd>{{ row.sortOrder }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
          <AdminTd align="right" class="whitespace-nowrap">
            <AdminButton v-if="String(row.status) !== 'active'" variant="link"
              :loading="activatingId === String(row.id)" @click="activateRow(row)">
              启用
            </AdminButton>
            <AdminButton variant="link" @click="openEdit(row)">{{ $t("common.edit") }}</AdminButton>
            <AdminButton variant="link" class="!text-danger-600" @click="removeRow(row)">
              删除
            </AdminButton>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新建会员等级' : '编辑会员等级'" width="lg">
      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField label="编码" required>
        <AdminInput v-model="form.code" :disabled="dialogMode === 'edit'" placeholder="如 free、pro" />
      </AdminFormField>
      <AdminFormField :label="$t('common.name')" required>
        <AdminInput v-model="form.name" />
      </AdminFormField>
      <AdminFormField label="日限额" hint="留空表示不限制">
        <AdminInput v-model="form.dailyLimit" type="number" placeholder="空为不限制" />
      </AdminFormField>
      <AdminFormField label="月限额" hint="留空表示不限制">
        <AdminInput v-model="form.monthlyLimit" type="number" placeholder="空为不限制" />
      </AdminFormField>
      <AdminFormField label="权益 JSON">
        <AdminInput v-model="form.features" type="textarea" :rows="4" class="font-mono text-sm" />
      </AdminFormField>
      <AdminFormField :label="$t('common.status')">
        <AdminSelect v-model="form.status" :options="statusOptions" />
      </AdminFormField>
      <AdminFormField :label="$t('common.sort')">
        <AdminInput v-model="form.sortOrder" type="number" />
      </AdminFormField>
      <AdminFormField :label="$t('common.remark')">
        <AdminInput v-model="form.remark" type="textarea" :rows="2" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">{{ $t("common.save") }}</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
