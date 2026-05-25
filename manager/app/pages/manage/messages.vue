<script setup lang="ts">
const API = "/api/admin/messages";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const showDeleted = ref(false);
const filterConversationId = ref("");

async function load() {
  loading.value = true;
  try {
    const q: Record<string, string | number | boolean> = {
      page: page.value,
      pageSize: pageSize.value,
    };
    if (showDeleted.value) q.includeDeleted = 1;
    if (filterConversationId.value.trim()) q.conversationId = filterConversationId.value.trim();
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

watch([page, pageSize, showDeleted, filterConversationId], () => void load(), { immediate: true });

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

const form = reactive({
  id: "",
  conversationId: "",
  role: "user",
  content: "",
  audioUrl: "",
  originalAudioUrl: "",
  sttText: "",
  durationMs: null as number | null,
  metadata: "",
});

function resetForm() {
  form.id = "";
  form.conversationId = "";
  form.role = "user";
  form.content = "";
  form.audioUrl = "";
  form.originalAudioUrl = "";
  form.sttText = "";
  form.durationMs = null;
  form.metadata = "";
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  form.conversationId = String(row.conversationId ?? "");
  form.role = String(row.role ?? "user");
  form.content = String(row.content ?? "");
  form.audioUrl = String(row.audioUrl ?? "");
  form.originalAudioUrl = String(row.originalAudioUrl ?? "");
  form.sttText = String(row.sttText ?? "");
  form.durationMs = row.durationMs != null ? Number(row.durationMs) : null;
  form.metadata = row.metadata != null ? String(row.metadata) : "";
  dialogVisible.value = true;
}

async function submit() {
  if (!form.conversationId.trim()) {
    toast.warning("请填写会话 ID");
    return;
  }
  if (!form.role.trim()) {
    toast.warning("请选择角色");
    return;
  }
  let meta: string | null = null;
  const rawMeta = (form.metadata ?? "").trim();
  if (rawMeta) {
    try {
      JSON.parse(rawMeta);
      meta = rawMeta;
    } catch {
      toast.error("metadata 须为合法 JSON");
      return;
    }
  }
  const body = {
    conversationId: form.conversationId.trim(),
    role: form.role.trim(),
    content: form.content,
    audioUrl: form.audioUrl.trim() || null,
    originalAudioUrl: form.originalAudioUrl.trim() || null,
    sttText: form.sttText.trim() || null,
    durationMs: form.durationMs,
    metadata: meta,
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
    message: "确认软删除该消息？",
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

const roleOptions = [
  { value: "user", label: "user" },
  { value: "assistant", label: "assistant" },
  { value: "system", label: "system" },
];

const aiStatusLabels: Record<string, string> = {
  success: "AI 成功",
  quota_exceeded: "额度用尽",
  failed: "AI 失败",
};

function aiStatusLabel(row: Record<string, unknown>): string {
  const raw = parseAiInteractionStatus(row);
  if (!raw) return "—";
  return aiStatusLabels[raw] ?? raw;
}

function parseAiInteractionStatus(row: Record<string, unknown>): string {
  const direct = row.aiInteractionStatus ?? row.ai_interaction_status;
  if (direct != null && String(direct).trim()) return String(direct).trim();
  const meta = row.metadata;
  if (meta == null) return "";
  try {
    const obj = typeof meta === "string" ? JSON.parse(meta) : meta;
    if (obj && typeof obj === "object" && "ai_interaction_status" in obj) {
      return String((obj as { ai_interaction_status?: unknown }).ai_interaction_status ?? "").trim();
    }
  } catch {
    /* ignore */
  }
  return "";
}
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        title="消息"
        description="消息归属会话；角色一般为 user / assistant / system；删除为软删除。"
      >
        <template #actions>
          <AdminCheckbox v-model="showDeleted" label="含已删除" />
          <div class="w-56">
            <AdminInput v-model="filterConversationId" placeholder="按会话 ID 筛选" />
          </div>
          <AdminButton variant="primary" @click="openCreate">新建</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh width="88px">角色</AdminTh>
          <AdminTh width="96px">AI 状态</AdminTh>
          <AdminTh>内容</AdminTh>
          <AdminTh>会话 ID</AdminTh>
          <AdminTh>音频</AdminTh>
          <AdminTh width="96px">时长 ms</AdminTh>
          <AdminTh>删除时间</AdminTh>
          <AdminTh>创建时间</AdminTh>
          <AdminTh width="140px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd><AdminBadge>{{ row.role }}</AdminBadge></AdminTd>
          <AdminTd>{{ aiStatusLabel(row) }}</AdminTd>
          <AdminTd>{{ row.content }}</AdminTd>
          <AdminTd>{{ row.conversationId }}</AdminTd>
          <AdminTd>{{ row.audioUrl ?? "—" }}</AdminTd>
          <AdminTd>{{ row.durationMs ?? "—" }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.deletedAt) }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.createdAt) }}</AdminTd>
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
      :title="dialogMode === 'create' ? '新建消息' : '编辑消息'"
      width="lg"
    >
      <AdminFormField v-if="dialogMode === 'edit'" label="ID">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField label="会话 ID" required>
        <AdminInput v-model="form.conversationId" :disabled="dialogMode === 'edit'" />
      </AdminFormField>
      <AdminFormField label="角色" required>
        <AdminSelect v-model="form.role" :options="roleOptions" />
      </AdminFormField>
      <AdminFormField label="正文" required>
        <AdminInput v-model="form.content" type="textarea" :rows="6" />
      </AdminFormField>
      <AdminFormField label="音频 URL">
        <AdminInput v-model="form.audioUrl" />
      </AdminFormField>
      <AdminFormField label="原始音频 URL">
        <AdminInput v-model="form.originalAudioUrl" />
      </AdminFormField>
      <AdminFormField label="STT 文本">
        <AdminInput v-model="form.sttText" type="textarea" :rows="2" />
      </AdminFormField>
      <AdminFormField label="时长 ms">
        <AdminInput v-model="form.durationMs" type="number" />
      </AdminFormField>
      <AdminFormField label="metadata JSON">
        <AdminInput v-model="form.metadata" type="textarea" :rows="3" class="font-mono text-sm" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">取消</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">保存</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
