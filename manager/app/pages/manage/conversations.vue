<script setup lang="ts">
const { t } = useI18n();
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
    toast.error(t("toast.loadRefsFailed"));
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
    toast.error(t("toast.loadFailed"));
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
  title: "",
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
  form.title = t("pages.conversations.defaultTitle");
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
  form.title = String(row.title ?? t("pages.conversations.defaultTitle"));
  form.status = String(row.status ?? "active");
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
  void loadRefs();
}

async function submit() {
  if (!form.userId.trim()) {
    toast.warning(t("validation.fillUserId"));
    return;
  }
  if (!form.languageId.trim()) {
    toast.warning(t("validation.selectTargetLanguage"));
    return;
  }
  const body = {
    userId: form.userId.trim(),
    languageId: form.languageId.trim(),
    voiceRoleId: form.voiceRoleId.trim() || null,
    llmConfigId: form.llmConfigId.trim() || null,
    promptId: form.promptId.trim() || null,
    title: form.title.trim() || t("pages.conversations.defaultTitle"),
    status: form.status,
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
    message: t("confirm.softDeleteConversation"),
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

const conversationStatusOptions = [
  { value: "active", label: "active" },
  { value: "archived", label: "archived" },
  { value: "inactive", label: "inactive" },
];

const optionalSelect = (opts: Opt[]) => [
  { value: "", label: t("common.optional") },
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
        :title="$t('pages.conversations.title')"
        :description="$t('pages.conversations.description')"
      >
        <template #actions>
          <AdminCheckbox v-model="showDeleted" :label="$t('common.includeDeleted')" />
          <div class="w-56">
            <AdminInput v-model="filterUserId" :placeholder="$t('pages.conversations.filterUserId')" />
          </div>
          <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("fields.title") }}</AdminTh>
          <AdminTh>{{ $t("fields.userId") }}</AdminTh>
          <AdminTh>{{ $t("fields.languageId") }}</AdminTh>
          <AdminTh>{{ $t("fields.voiceRole") }}</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh>{{ $t("common.deletedAt") }}</AdminTh>
          <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>
          <AdminTh width="140px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd>{{ row.title }}</AdminTd>
          <AdminTd>{{ row.userId }}</AdminTd>
          <AdminTd>{{ row.languageId }}</AdminTd>
          <AdminTd>{{ row.voiceRoleId ?? t("common.emDash") }}</AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.deletedAt) }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
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
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? $t('pages.conversations.createDialog') : $t('pages.conversations.editDialog')"
      width="lg"
    >
      <AdminSkeleton v-if="optionsLoading" :rows="6" />
      <template v-else>
        <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
          <AdminInput v-model="form.id" disabled />
        </AdminFormField>
        <AdminFormField :label="$t('fields.userId')" required>
          <AdminInput v-model="form.userId" placeholder="UUID" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.language')" required>
          <AdminSelect
            v-model="form.languageId"
            :options="langSelectOptions"
            :placeholder="$t('pages.conversations.selectLanguage')"
          />
        </AdminFormField>
        <AdminFormField :label="$t('fields.voiceRole')">
          <AdminSelect v-model="form.voiceRoleId" :options="voiceSelectOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.llmConfig')">
          <AdminSelect v-model="form.llmConfigId" :options="aiSelectOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.promptTemplate')">
          <AdminSelect v-model="form.promptId" :options="promptSelectOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.title')">
          <AdminInput v-model="form.title" />
        </AdminFormField>
        <AdminFormField :label="$t('common.status')">
          <AdminSelect v-model="form.status" :options="conversationStatusOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('common.remark')">
          <AdminInput v-model="form.remark" type="textarea" :rows="2" />
        </AdminFormField>
      </template>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">{{ $t("common.save") }}</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
