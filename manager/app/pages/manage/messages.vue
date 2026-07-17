<script setup lang="ts">
const { t } = useI18n();
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
    toast.error(t("toast.loadFailed"));
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
    toast.warning(t("validation.fillConversationId"));
    return;
  }
  if (!form.role.trim()) {
    toast.warning(t("validation.selectRole"));
    return;
  }
  let meta: string | null = null;
  const rawMeta = (form.metadata ?? "").trim();
  if (rawMeta) {
    try {
      JSON.parse(rawMeta);
      meta = rawMeta;
    } catch {
      toast.error(t("validation.invalidJson", { field: "metadata" }));
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
    message: t("confirm.softDeleteMessage"),
    danger: true,
    confirmLabel: t("common.softDelete"),
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success(t("toast.markedDeleted"));
    await load();
  } catch (e) {
    toast.error(t("toast.operationFailed"));
    console.error(e);
  }
}

const roleOptions = [
  { value: "user", label: "user" },
  { value: "assistant", label: "assistant" },
  { value: "system", label: "system" },
];

const aiStatusLabels = computed<Record<string, string>>(() => ({
  success: t("status.aiSuccess"),
  quota_exceeded: t("status.quotaExceeded"),
  failed: t("status.aiFailed"),
}));

function aiStatusLabel(row: Record<string, unknown>): string {
  const raw = parseAiInteractionStatus(row);
  if (!raw) return t("common.emDash");
  return aiStatusLabels.value[raw] ?? raw;
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
        :title="$t('pages.messages.title')"
        :description="$t('pages.messages.description')"
      >
        <template #actions>
          <AdminCheckbox v-model="showDeleted" :label="$t('common.includeDeleted')" />
          <div class="w-full sm:w-56">
            <AdminInput
              v-model="filterConversationId"
              :placeholder="$t('pages.messages.filterConversationId')"
            />
          </div>
          <AdminButton variant="primary" class="w-full sm:w-auto" @click="openCreate">
            {{ $t("common.create") }}
          </AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh width="88px">{{ $t("fields.role") }}</AdminTh>
          <AdminTh width="96px">{{ $t("fields.aiStatus") }}</AdminTh>
          <AdminTh>{{ $t("fields.content") }}</AdminTh>
          <AdminTh>{{ $t("fields.conversationId") }}</AdminTh>
          <AdminTh>{{ $t("fields.audio") }}</AdminTh>
          <AdminTh width="96px">{{ $t("fields.durationMs") }}</AdminTh>
          <AdminTh>{{ $t("common.deletedAt") }}</AdminTh>
          <AdminTh>{{ $t("common.createdAt") }}</AdminTh>
          <AdminTh width="140px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd><AdminBadge>{{ row.role }}</AdminBadge></AdminTd>
          <AdminTd>{{ aiStatusLabel(row) }}</AdminTd>
          <AdminTd>
            <AdminCellText :title="String(row.content ?? '')">
              {{ row.content ?? t("common.emDash") }}
            </AdminCellText>
          </AdminTd>
          <AdminTd>
            <AdminCellText :title="String(row.conversationId ?? '')">
              {{ row.conversationId }}
            </AdminCellText>
          </AdminTd>
          <AdminTd>
            <AdminCellText :title="row.audioUrl != null ? String(row.audioUrl) : undefined">
              {{ row.audioUrl ?? t("common.emDash") }}
            </AdminCellText>
          </AdminTd>
          <AdminTd>{{ row.durationMs ?? t("common.emDash") }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.deletedAt) }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.createdAt) }}</AdminTd>
          <AdminTd align="right">
            <AdminButton variant="link" @click="openEdit(row)">{{ $t("common.edit") }}</AdminButton>
            <AdminButton
              variant="link"
              class="!text-danger-600"
              :disabled="!!row.deletedAt"
              @click="removeRow(row)"
            >
              {{ $t("common.softDelete") }}
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
            :title="String(row.content ?? $t('common.emDash')).slice(0, 80)"
            :subtitle="String(row.conversationId ?? '')"
          >
            <template #badge>
              <AdminBadge>{{ row.role }}</AdminBadge>
            </template>
            <template #menu>
              <AdminOverflowMenu
                :actions="[
                  { label: $t('common.edit'), onClick: () => openEdit(row) },
                  {
                    label: $t('common.softDelete'),
                    danger: true,
                    disabled: !!row.deletedAt,
                    onClick: () => removeRow(row),
                  },
                ]"
              />
            </template>
            <AdminMobileMeta :label="$t('fields.aiStatus')">
              {{ aiStatusLabel(row) }}
            </AdminMobileMeta>
            <AdminMobileMeta :label="$t('fields.durationMs')">
              {{ row.durationMs ?? $t("common.emDash") }}
            </AdminMobileMeta>
            <AdminMobileMeta :label="$t('common.createdAt')">
              {{ formatDateTime(row.createdAt) }}
            </AdminMobileMeta>
          </AdminMobileCard>
        </template>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? $t('pages.messages.createDialog') : $t('pages.messages.editDialog')"
      width="lg"
    >
      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField :label="$t('fields.conversationId')" required>
        <AdminInput v-model="form.conversationId" :disabled="dialogMode === 'edit'" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.role')" required>
        <AdminSelect v-model="form.role" :options="roleOptions" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.body')" required>
        <AdminInput v-model="form.content" type="textarea" :rows="6" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.audioUrl')">
        <AdminInput v-model="form.audioUrl" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.originalAudioUrl')">
        <AdminInput v-model="form.originalAudioUrl" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.sttText')">
        <AdminInput v-model="form.sttText" type="textarea" :rows="2" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.durationMs')">
        <AdminInput v-model="form.durationMs" type="number" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.metadataJson')">
        <AdminInput v-model="form.metadata" type="textarea" :rows="3" class="font-mono text-sm" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">{{ $t("common.save") }}</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
