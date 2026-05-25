<script setup lang="ts">
const API = "/api/admin/conversations";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

type Opt = { id: string; label: string };

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const showDeleted = ref(false);
const filterUserId = ref("");

const langOptions = ref<Opt[]>([]);
const voiceOptions = ref<Opt[]>([]);
const aiOptions = ref<Opt[]>([]);
const promptOptions = ref<Opt[]>([]);
const optionsLoading = ref(false);

async function loadRefs() {
  optionsLoading.value = true;
  try {
    const [lr, vr, ar, pr] = await Promise.all([
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/languages", {
        query: { page: 1, pageSize: 500 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/voice-roles", {
        query: { page: 1, pageSize: 500 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/llm-service-configs", {
        query: { page: 1, pageSize: 200 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/prompt-templates", {
        query: { page: 1, pageSize: 200 },
      }),
    ]);
    langOptions.value = lr.items.map((r) => ({
      id: String(r.id),
      label: `${r.code} · ${r.name}`,
    }));
    voiceOptions.value = vr.items.map((r) => ({
      id: String(r.id),
      label: String(r.name ?? r.id),
    }));
    aiOptions.value = ar.items.map((r) => ({
      id: String(r.id),
      label: `${r.code} · ${r.name}`,
    }));
    promptOptions.value = pr.items.map((r) => ({
      id: String(r.id),
      label: `${r.code} · ${r.name}`,
    }));
  } catch (e) {
    toast.error("加载下拉数据失败");
    console.error(e);
  } finally {
    optionsLoading.value = false;
  }
}

async function load() {
  loading.value = true;
  try {
    const q: Record<string, string | number | boolean> = {
      page: page.value,
      pageSize: pageSize.value,
    };
    if (showDeleted.value) q.includeDeleted = 1;
    if (filterUserId.value.trim()) q.userId = filterUserId.value.trim();
    const res = await api.list(q);
    list.value = res.items;
    total.value = res.total;
  } catch (e) {
    toast.error("加载失败");
    console.error(e);
  } finally {
    loading.value = false;
  }
}

watch([page, pageSize, showDeleted, filterUserId], () => void load(), { immediate: true });

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

const form = reactive({
  id: "",
  userId: "",
  languageId: "",
  voiceRoleId: "",
  llmConfigId: "",
  promptId: "",
  title: "新对话",
  status: "active",
  remark: "",
});

function resetForm() {
  form.id = "";
  form.userId = "";
  form.languageId = "";
  form.voiceRoleId = "";
  form.llmConfigId = "";
  form.promptId = "";
  form.title = "新对话";
  form.status = "active";
  form.remark = "";
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
  void loadRefs();
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  form.userId = String(row.userId ?? "");
  form.languageId = String(row.languageId ?? "");
  form.voiceRoleId = String(row.voiceRoleId ?? "");
  form.llmConfigId = String(row.llmConfigId ?? "");
  form.promptId = String(row.promptId ?? "");
  form.title = String(row.title ?? "新对话");
  form.status = String(row.status ?? "active");
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
  void loadRefs();
}

async function submit() {
  if (!form.userId.trim()) {
    toast.warning("请填写用户 ID");
    return;
  }
  if (!form.languageId.trim()) {
    toast.warning("请选择语言");
    return;
  }
  const body = {
    userId: form.userId.trim(),
    languageId: form.languageId.trim(),
    voiceRoleId: form.voiceRoleId.trim() || null,
    llmConfigId: form.llmConfigId.trim() || null,
    promptId: form.promptId.trim() || null,
    title: form.title.trim() || "新对话",
    status: form.status,
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
    message: "确认软删除该会话？",
    danger: true,
    confirmLabel: "软删",
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success("已标记删除");
    await load();
  } catch (e) {
    toast.error("操作失败");
    console.error(e);
  }
}

const conversationStatusOptions = [
  { value: "active", label: "active" },
  { value: "archived", label: "archived" },
  { value: "inactive", label: "inactive" },
];

const optionalSelect = (opts: Opt[]) => [
  { value: "", label: "（可选）" },
  ...opts.map((o) => ({ value: o.id, label: o.label })),
];

const langSelectOptions = computed(() => langOptions.value.map((l) => ({ value: l.id, label: l.label })));
const voiceSelectOptions = computed(() => optionalSelect(voiceOptions.value));
const aiSelectOptions = computed(() => optionalSelect(aiOptions.value));
const promptSelectOptions = computed(() => optionalSelect(promptOptions.value));
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        title="会话"
        description="会话绑定用户与语言，可选语音角色、LLM 配置与提示词模板；删除为软删除。"
      >
        <template #actions>
          <AdminCheckbox v-model="showDeleted" label="含已删除" />
          <div class="w-56">
            <AdminInput v-model="filterUserId" placeholder="按用户 ID 筛选" />
          </div>
          <AdminButton variant="primary" @click="openCreate">新建</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>标题</AdminTh>
          <AdminTh>用户 ID</AdminTh>
          <AdminTh>语言 ID</AdminTh>
          <AdminTh>语音角色</AdminTh>
          <AdminTh width="88px">状态</AdminTh>
          <AdminTh>删除时间</AdminTh>
          <AdminTh>更新时间</AdminTh>
          <AdminTh width="140px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd>{{ row.title }}</AdminTd>
          <AdminTd>{{ row.userId }}</AdminTd>
          <AdminTd>{{ row.languageId }}</AdminTd>
          <AdminTd>{{ row.voiceRoleId ?? "—" }}</AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.deletedAt) }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
          <AdminTd align="right">
            <AdminButton variant="link" @click="openEdit(row)">编辑</AdminButton>
            <AdminButton
              variant="link"
              class="!text-danger-600"
              :disabled="!!row.deletedAt"
              @click="removeRow(row)"
            >
              软删
            </AdminButton>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建会话' : '编辑会话'"
      width="lg"
    >
      <AdminSkeleton v-if="optionsLoading" :rows="6" />
      <template v-else>
        <AdminFormField v-if="dialogMode === 'edit'" label="ID">
          <AdminInput v-model="form.id" disabled />
        </AdminFormField>
        <AdminFormField label="用户 ID" required>
          <AdminInput v-model="form.userId" placeholder="UUID" />
        </AdminFormField>
        <AdminFormField label="语言" required>
          <AdminSelect v-model="form.languageId" :options="langSelectOptions" placeholder="选择语言" />
        </AdminFormField>
        <AdminFormField label="语音角色">
          <AdminSelect v-model="form.voiceRoleId" :options="voiceSelectOptions" />
        </AdminFormField>
        <AdminFormField label="LLM 配置">
          <AdminSelect v-model="form.llmConfigId" :options="aiSelectOptions" />
        </AdminFormField>
        <AdminFormField label="提示词模板">
          <AdminSelect v-model="form.promptId" :options="promptSelectOptions" />
        </AdminFormField>
        <AdminFormField label="标题">
          <AdminInput v-model="form.title" />
        </AdminFormField>
        <AdminFormField label="状态">
          <AdminSelect v-model="form.status" :options="conversationStatusOptions" />
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
