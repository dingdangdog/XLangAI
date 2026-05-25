<script setup lang="ts">
const API = "/api/admin/prompt-templates";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

type LangOpt = { id: string; code: string; name: string };

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const langOptions = ref<LangOpt[]>([]);
const optionsLoading = ref(false);

async function loadLangs() {
  optionsLoading.value = true;
  try {
    const lr = await $fetch<{ items: Record<string, unknown>[] }>("/api/admin/languages", {
      query: { page: 1, pageSize: 500 },
    });
    langOptions.value = lr.items
      .filter((r) => String(r.status ?? "active") === "active")
      .map((r) => ({
        id: String(r.id),
        code: String(r.code ?? ""),
        name: String(r.name ?? ""),
      }));
  } catch (e) {
    toast.error("加载语言列表失败");
    console.error(e);
  } finally {
    optionsLoading.value = false;
  }
}

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
  content: "",
  variables: "",
  languageId: "",
  status: "active",
  sortOrder: 0,
  remark: "",
});

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.content = "";
  form.variables = "";
  form.languageId = "";
  form.status = "active";
  form.sortOrder = 0;
  form.remark = "";
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
  void loadLangs();
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  form.code = String(row.code ?? "");
  form.name = String(row.name ?? "");
  form.content = String(row.content ?? "");
  form.variables = row.variables != null ? String(row.variables) : "";
  form.languageId = String(row.languageId ?? "");
  form.status = String(row.status ?? "active");
  form.sortOrder = Number(row.sortOrder ?? 0);
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
  void loadLangs();
}

async function submit() {
  if (!form.code.trim() || !form.name.trim()) {
    toast.warning("请填写模板编码与名称");
    return;
  }
  if (!form.content.trim()) {
    toast.warning("请填写模板内容");
    return;
  }
  let variables: string | null = null;
  const rawVars = (form.variables ?? "").trim();
  if (rawVars) {
    try {
      JSON.parse(rawVars);
      variables = rawVars;
    } catch {
      toast.error("variables 须为合法 JSON（如数组）");
      return;
    }
  }
  const body = {
    code: form.code.trim(),
    name: form.name.trim(),
    content: form.content,
    variables,
    languageId: form.languageId || null,
    status: form.status,
    sortOrder: form.sortOrder,
    remark: form.remark.trim() || null,
  };
  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await api.create(body);
      toast.success("已创建");
    } else {
      await api.update(form.id, { id: form.id, ...body });
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
    message: "确认删除该提示词模板？",
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

const langSelectOptions = computed(() => [
  { value: "", label: "不绑定则全局" },
  ...langOptions.value.map((l) => ({ value: l.id, label: `${l.code} · ${l.name}` })),
]);

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
        title="提示词模板"
        description="系统提示词等业务模板；code 唯一，供服务端按编码取用。"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">新建</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>编码</AdminTh>
          <AdminTh>名称</AdminTh>
          <AdminTh>语言 ID</AdminTh>
          <AdminTh width="88px">状态</AdminTh>
          <AdminTh width="72px">排序</AdminTh>
          <AdminTh>更新时间</AdminTh>
          <AdminTh width="200px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap>{{ row.code }}</AdminTd>
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.languageId ?? "—" }}</AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd>{{ row.sortOrder }}</AdminTd>
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
      :title="dialogMode === 'create' ? '新建模板' : '编辑模板'"
      width="lg"
    >
      <AdminSkeleton v-if="optionsLoading" :rows="4" />
      <template v-else>
        <AdminFormField v-if="dialogMode === 'edit'" label="ID">
          <AdminInput v-model="form.id" disabled />
        </AdminFormField>
        <AdminFormField label="编码" required>
          <AdminInput v-model="form.code" :disabled="dialogMode === 'edit'" />
        </AdminFormField>
        <AdminFormField label="名称" required>
          <AdminInput v-model="form.name" />
        </AdminFormField>
        <AdminFormField label="语言">
          <AdminSelect v-model="form.languageId" :options="langSelectOptions" />
        </AdminFormField>
        <AdminFormField label="内容" required>
          <AdminInput v-model="form.content" type="textarea" :rows="8" class="font-mono text-sm" />
        </AdminFormField>
        <AdminFormField label="variables" hint='JSON，如 ["name"] 或留空'>
          <AdminInput
            v-model="form.variables"
            type="textarea"
            :rows="3"
            class="font-mono text-sm"
          />
        </AdminFormField>
        <AdminFormField label="状态">
          <AdminSelect v-model="form.status" :options="statusOptions" />
        </AdminFormField>
        <AdminFormField label="排序">
          <AdminInput v-model="form.sortOrder" type="number" />
        </AdminFormField>
        <AdminFormField label="备注">
          <AdminInput v-model="form.remark" type="textarea" :rows="2" />
        </AdminFormField>
      </template>
      <template #footer>
        <AdminButton @click="dialogVisible = false">取消</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">保存</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
