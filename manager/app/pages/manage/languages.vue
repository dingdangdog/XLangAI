<script setup lang="ts">
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
    toast.error("加载失败");
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
    toast.warning("请填写语言代码与名称");
    return;
  }
  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await api.create(payload());
      toast.success("已创建");
    } else {
      await api.update(form.id, { id: form.id, ...payload() });
      toast.success("已保存");
    }
    dialogVisible.value = false;
    await load();
  } catch (e) {
    toast.error("保存失败");
    console.error(e);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: "确认删除该语言？若仍被用户或会话引用可能导致数据不一致。",
    danger: true,
    confirmLabel: "删除",
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success("已删除");
    await load();
  } catch (e) {
    toast.error("删除失败");
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
        title="语言"
        description="管理客户端可选语言列表；code 需唯一（如 zh、en）。"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">新建</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh width="100px">代码</AdminTh>
          <AdminTh>名称</AdminTh>
          <AdminTh>本地名称</AdminTh>
          <AdminTh width="80px">排序</AdminTh>
          <AdminTh width="88px">状态</AdminTh>
          <AdminTh>备注</AdminTh>
          <AdminTh>更新时间</AdminTh>
          <AdminTh width="200px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap>{{ row.code }}</AdminTd>
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.nameNative ?? "—" }}</AdminTd>
          <AdminTd>{{ row.sortOrder }}</AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd>{{ row.remark ?? "—" }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
          <AdminTd align="right" class="whitespace-nowrap">
            <AdminButton
              v-if="String(row.status) !== 'active'"
              variant="link"
              :loading="activatingId === String(row.id)"
              @click="activateRow(row)"
            >
              启用
            </AdminButton>
            <AdminButton variant="link" @click="openEdit(row)">编辑</AdminButton>
            <AdminButton variant="link" class="!text-danger-600" @click="removeRow(row)">
              删除
            </AdminButton>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建语言' : '编辑语言'"
    >
      <AdminFormField v-if="dialogMode === 'edit'" label="ID">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField label="代码" required>
        <AdminInput
          v-model="form.code"
          :disabled="dialogMode === 'edit'"
          placeholder="如 zh、en"
        />
      </AdminFormField>
      <AdminFormField label="名称" required>
        <AdminInput v-model="form.name" placeholder="展示名称" />
      </AdminFormField>
      <AdminFormField label="本地名称">
        <AdminInput v-model="form.nameNative" placeholder="可选" />
      </AdminFormField>
      <AdminFormField
        label="试听文案模板"
        hint="支持 {name}、{{name}}、{{voice_role_name}}，将替换为语音角色显示名；留空则使用内置默认模板。"
      >
        <AdminInput
          v-model="form.previewSampleText"
          type="textarea"
          :rows="2"
          placeholder="如：你好，我是{name}"
        />
      </AdminFormField>
      <AdminFormField label="排序">
        <AdminInput v-model="form.sortOrder" type="number" />
      </AdminFormField>
      <AdminFormField label="状态">
        <AdminSelect v-model="form.status" :options="statusOptions" />
      </AdminFormField>
      <AdminFormField label="备注">
        <AdminInput v-model="form.remark" type="textarea" :rows="2" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">取消</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">保存</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
