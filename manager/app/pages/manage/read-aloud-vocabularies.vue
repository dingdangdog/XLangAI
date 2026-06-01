<script setup lang="ts">
const { t } = useI18n();
const API = "/api/admin/read-aloud-vocabularies";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

type Opt = { id: string; label: string };

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const categoryOptions = ref<Opt[]>([]);
const languageOptions = ref<Opt[]>([]);
const voiceOptions = ref<Opt[]>([]);
const optionsLoading = ref(false);
const filterCategoryId = ref("");
const filterLanguageId = ref("");
const audioGeneratingId = ref("");

const llmOptions = ref<Opt[]>([]);
const llmDialogVisible = ref(false);
const llmGenerating = ref(false);
const llmImporting = ref(false);
const llmPreviewItems = ref<Array<{ word: string; exampleSentence: string; selected: boolean }>>([]);

const llmForm = reactive({
  llmServiceConfigId: "",
  categoryId: "",
  languageId: "",
  voiceRoleId: "",
  count: 10,
  extraInstructions: "",
});

async function loadOptions() {
  optionsLoading.value = true;
  try {
    const [cats, langs, voices, llms] = await Promise.all([
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/read-aloud-categories", {
        query: { page: 1, pageSize: 200 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/languages", {
        query: { page: 1, pageSize: 200 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/voice-roles", {
        query: { page: 1, pageSize: 500 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/llm-service-configs", {
        query: { page: 1, pageSize: 200 },
      }),
    ]);
    categoryOptions.value = cats.items.map((r) => ({
      id: String(r.id),
      label: `${r.code} · ${r.name}`,
    }));
    languageOptions.value = langs.items.map((r) => ({
      id: String(r.id),
      label: `${r.code} · ${r.name}`,
    }));
    voiceOptions.value = voices.items
      .filter((r) => (r.synthesisType ?? "tts") === "tts")
      .map((r) => ({
        id: String(r.id),
        label: `${r.name} (${r.voiceCode})`,
      }));
    llmOptions.value = llms.items
      .filter((r) => r.status === "active")
      .map((r) => ({
        id: String(r.id),
        label: `${r.code} · ${r.name} (${r.modelCode})`,
      }));
  } catch (e) {
    toast.error(t("toast.loadOptionsFailed"));
    console.error(e);
  } finally {
    optionsLoading.value = false;
  }
}

async function load() {
  loading.value = true;
  try {
    const query: Record<string, unknown> = { page: page.value, pageSize: pageSize.value };
    if (filterCategoryId.value) query.categoryId = filterCategoryId.value;
    if (filterLanguageId.value) query.languageId = filterLanguageId.value;
    const res = await api.list(query);
    list.value = res.items;
    total.value = res.total;
  } catch (e) {
    toast.error(t("toast.loadFailed"));
    console.error(e);
  } finally {
    loading.value = false;
  }
}

watch([page, pageSize, filterCategoryId, filterLanguageId], () => void load(), { immediate: true });

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

const form = reactive({
  id: "",
  categoryId: "",
  languageId: "",
  word: "",
  exampleSentence: "",
  voiceRoleId: "",
  status: "active",
  sortOrder: 0,
  remark: "",
});

function resetForm() {
  form.id = "";
  form.categoryId = filterCategoryId.value || "";
  form.languageId = filterLanguageId.value || "";
  form.word = "";
  form.exampleSentence = "";
  form.voiceRoleId = "";
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
  form.categoryId = String(row.categoryId ?? "");
  form.languageId = String(row.languageId ?? "");
  form.word = String(row.word ?? "");
  form.exampleSentence = String(row.exampleSentence ?? "");
  form.voiceRoleId = String(row.voiceRoleId ?? "");
  form.status = String(row.status ?? "active");
  form.sortOrder = Number(row.sortOrder ?? 0);
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
}

async function submit() {
  if (!form.categoryId || !form.languageId || !form.word.trim() || !form.exampleSentence.trim() || !form.voiceRoleId) {
    toast.warning(t("pages.readAloudVocabularies.fillRequired"));
    return;
  }
  const body = {
    categoryId: form.categoryId,
    languageId: form.languageId,
    word: form.word.trim(),
    exampleSentence: form.exampleSentence.trim(),
    voiceRoleId: form.voiceRoleId,
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
    message: t("pages.readAloudVocabularies.deleteConfirm"),
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

function openLlmBatch() {
  llmForm.llmServiceConfigId = llmOptions.value[0]?.id ?? "";
  llmForm.categoryId = filterCategoryId.value || categoryOptions.value[0]?.id || "";
  llmForm.languageId = filterLanguageId.value || languageOptions.value[0]?.id || "";
  llmForm.voiceRoleId = voiceOptions.value[0]?.id ?? "";
  llmForm.count = 10;
  llmForm.extraInstructions = "";
  llmPreviewItems.value = [];
  llmDialogVisible.value = true;
}

async function runLlmGenerate() {
  if (!llmForm.llmServiceConfigId || !llmForm.categoryId || !llmForm.languageId) {
    toast.warning(t("pages.readAloudVocabularies.llmFillRequired"));
    return;
  }
  llmGenerating.value = true;
  try {
    const res = await $fetch<{
      items: Array<{ word: string; exampleSentence: string }>;
    }>("/api/admin/read-aloud-vocabularies/llm-generate", {
      method: "POST",
      body: {
        llmServiceConfigId: llmForm.llmServiceConfigId,
        categoryId: llmForm.categoryId,
        languageId: llmForm.languageId,
        count: llmForm.count,
        extraInstructions: llmForm.extraInstructions.trim() || null,
      },
    });
    llmPreviewItems.value = (res.items ?? []).map((it) => ({
      word: it.word,
      exampleSentence: it.exampleSentence,
      selected: true,
    }));
    if (llmPreviewItems.value.length === 0) {
      toast.warning(t("pages.readAloudVocabularies.llmEmptyResult"));
    } else {
      toast.success(t("pages.readAloudVocabularies.llmGenerated", { count: llmPreviewItems.value.length }));
    }
  } catch (e) {
    toast.error(t("pages.readAloudVocabularies.llmGenerateFailed"));
    console.error(e);
  } finally {
    llmGenerating.value = false;
  }
}

async function importLlmPreview() {
  if (!llmForm.categoryId || !llmForm.languageId || !llmForm.voiceRoleId) {
    toast.warning(t("pages.readAloudVocabularies.llmImportFillRequired"));
    return;
  }
  const selected = llmPreviewItems.value.filter((it) => it.selected && it.word.trim() && it.exampleSentence.trim());
  if (selected.length === 0) {
    toast.warning(t("pages.readAloudVocabularies.llmNothingSelected"));
    return;
  }
  llmImporting.value = true;
  try {
    const res = await $fetch<{ created: number }>("/api/admin/read-aloud-vocabularies/batch-create", {
      method: "POST",
      body: {
        categoryId: llmForm.categoryId,
        languageId: llmForm.languageId,
        voiceRoleId: llmForm.voiceRoleId,
        items: selected.map((it) => ({
          word: it.word.trim(),
          exampleSentence: it.exampleSentence.trim(),
        })),
        remark: "LLM batch import",
      },
    });
    toast.success(t("pages.readAloudVocabularies.llmImported", { count: res.created ?? selected.length }));
    llmDialogVisible.value = false;
    filterCategoryId.value = llmForm.categoryId;
    filterLanguageId.value = llmForm.languageId;
    await load();
  } catch (e) {
    toast.error(t("pages.readAloudVocabularies.llmImportFailed"));
    console.error(e);
  } finally {
    llmImporting.value = false;
  }
}

const llmSelectOptions = computed(() =>
  llmOptions.value.map((o) => ({ value: o.id, label: o.label })),
);

const llmSelectedCount = computed(
  () => llmPreviewItems.value.filter((it) => it.selected).length,
);

async function generateAudio(row: Record<string, unknown>, part: "word" | "sentence" | "both") {
  const id = String(row.id ?? "");
  if (!id) return;
  audioGeneratingId.value = `${id}:${part}`;
  try {
    await $fetch(`/api/admin/read-aloud-vocabularies/${id}/generate-audio`, {
      method: "POST",
      body: { part },
    });
    toast.success(t("pages.readAloudVocabularies.audioGenerated"));
    await load();
  } catch (e) {
    toast.error(t("pages.readAloudVocabularies.audioGenerateFailed"));
    console.error(e);
  } finally {
    audioGeneratingId.value = "";
  }
}

const statusOptions = [
  { value: "active", label: "active" },
  { value: "inactive", label: "inactive" },
];

const categorySelectOptions = computed(() => [
  { value: "", label: t("common.none") },
  ...categoryOptions.value.map((o) => ({ value: o.id, label: o.label })),
]);

const languageSelectOptions = computed(() => [
  { value: "", label: t("common.none") },
  ...languageOptions.value.map((o) => ({ value: o.id, label: o.label })),
]);

const voiceSelectOptions = computed(() =>
  voiceOptions.value.map((o) => ({ value: o.id, label: o.label })),
);

const labelById = (opts: Opt[], id: unknown) => {
  const key = String(id ?? "").trim();
  if (!key) return t("common.emDash");
  return opts.find((o) => o.id === key)?.label ?? key;
};

onMounted(() => void loadOptions());
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.readAloudVocabularies.title')"
        :description="$t('pages.readAloudVocabularies.description')"
      >
        <template #actions>
          <AdminButton to="/manage/read-aloud-categories">{{ $t("pages.readAloudCategories.title") }}</AdminButton>
          <AdminButton @click="openLlmBatch">{{ $t("pages.readAloudVocabularies.llmBatch") }}</AdminButton>
          <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel class="mb-4">
      <div class="flex flex-wrap gap-4 p-4">
        <AdminFormField :label="$t('pages.readAloudVocabularies.filterCategory')" class="min-w-[220px]">
          <AdminSelect v-model="filterCategoryId" :options="categorySelectOptions" :loading="optionsLoading" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudVocabularies.filterLanguage')" class="min-w-[220px]">
          <AdminSelect v-model="filterLanguageId" :options="languageSelectOptions" :loading="optionsLoading" />
        </AdminFormField>
      </div>
    </AdminPanel>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("pages.readAloudVocabularies.word") }}</AdminTh>
          <AdminTh>{{ $t("pages.readAloudVocabularies.exampleSentence") }}</AdminTh>
          <AdminTh>{{ $t("pages.readAloudVocabularies.category") }}</AdminTh>
          <AdminTh>{{ $t("pages.readAloudVocabularies.language") }}</AdminTh>
          <AdminTh>{{ $t("pages.readAloudVocabularies.audio") }}</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh width="240px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd>{{ row.word }}</AdminTd>
          <AdminTd><AdminCellText :title="String(row.exampleSentence ?? '')">{{ row.exampleSentence }}</AdminCellText></AdminTd>
          <AdminTd>{{ labelById(categoryOptions, row.categoryId) }}</AdminTd>
          <AdminTd>{{ labelById(languageOptions, row.languageId) }}</AdminTd>
          <AdminTd class="text-xs">
            <div>{{ row.wordAudioUrl ? $t("common.generated") : $t("common.notGenerated") }} / {{ row.sentenceAudioUrl ? $t("common.generated") : $t("common.notGenerated") }}</div>
          </AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd align="right" class="whitespace-nowrap">
            <AdminButton
              variant="link"
              :loading="audioGeneratingId === `${row.id}:both`"
              @click="generateAudio(row, 'both')"
            >
              {{ $t("pages.readAloudVocabularies.generateBoth") }}
            </AdminButton>
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
      :title="dialogMode === 'create' ? $t('pages.readAloudVocabularies.createDialog') : $t('pages.readAloudVocabularies.editDialog')"
      size="lg"
    >
      <div class="space-y-4">
        <AdminFormField :label="$t('pages.readAloudVocabularies.category')" required>
          <AdminSelect v-model="form.categoryId" :options="categorySelectOptions.filter((o) => o.value)" :loading="optionsLoading" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudVocabularies.language')" required>
          <AdminSelect v-model="form.languageId" :options="languageSelectOptions.filter((o) => o.value)" :loading="optionsLoading" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudVocabularies.word')" required>
          <AdminInput v-model="form.word" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudVocabularies.exampleSentence')" required>
          <AdminTextarea v-model="form.exampleSentence" rows="2" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudVocabularies.voiceRole')" required>
          <AdminSelect v-model="form.voiceRoleId" :options="voiceSelectOptions" :loading="optionsLoading" />
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

    <AdminDialog
      v-model="llmDialogVisible"
      :title="$t('pages.readAloudVocabularies.llmBatchDialog')"
      size="xl"
    >
      <div class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ $t("pages.readAloudVocabularies.llmBatchHint") }}
        </p>
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField :label="$t('pages.readAloudVocabularies.llmConfig')" required>
            <AdminSelect v-model="llmForm.llmServiceConfigId" :options="llmSelectOptions" :loading="optionsLoading" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.readAloudVocabularies.llmCount')" required>
            <AdminInput v-model.number="llmForm.count" type="number" min="1" max="30" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.readAloudVocabularies.category')" required>
            <AdminSelect v-model="llmForm.categoryId" :options="categorySelectOptions.filter((o) => o.value)" :loading="optionsLoading" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.readAloudVocabularies.language')" required>
            <AdminSelect v-model="llmForm.languageId" :options="languageSelectOptions.filter((o) => o.value)" :loading="optionsLoading" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.readAloudVocabularies.voiceRole')" required class="md:col-span-2">
            <AdminSelect v-model="llmForm.voiceRoleId" :options="voiceSelectOptions" :loading="optionsLoading" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.readAloudVocabularies.llmExtraInstructions')" class="md:col-span-2">
            <AdminTextarea v-model="llmForm.extraInstructions" rows="2" :placeholder="$t('pages.readAloudVocabularies.llmExtraPlaceholder')" />
          </AdminFormField>
        </div>

        <div class="flex flex-wrap gap-2">
          <AdminButton variant="primary" :loading="llmGenerating" @click="runLlmGenerate">
            {{ $t("pages.readAloudVocabularies.llmGeneratePreview") }}
          </AdminButton>
          <span v-if="llmPreviewItems.length" class="self-center text-sm text-gray-500">
            {{ $t("pages.readAloudVocabularies.llmSelectedCount", { count: llmSelectedCount }) }}
          </span>
        </div>

        <div v-if="llmPreviewItems.length" class="max-h-80 overflow-auto rounded border border-gray-200 dark:border-gray-700">
          <table class="min-w-full text-sm">
            <thead class="sticky top-0 bg-gray-50 dark:bg-gray-900">
              <tr>
                <th class="px-3 py-2 text-left">{{ $t("common.actions") }}</th>
                <th class="px-3 py-2 text-left">{{ $t("pages.readAloudVocabularies.word") }}</th>
                <th class="px-3 py-2 text-left">{{ $t("pages.readAloudVocabularies.exampleSentence") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, idx) in llmPreviewItems" :key="idx" class="border-t border-gray-100 dark:border-gray-800">
                <td class="px-3 py-2 align-top">
                  <input v-model="row.selected" type="checkbox" class="rounded" />
                </td>
                <td class="px-3 py-2 align-top">
                  <AdminInput v-model="row.word" />
                </td>
                <td class="px-3 py-2 align-top">
                  <AdminTextarea v-model="row.exampleSentence" rows="2" />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <template #footer>
        <AdminButton @click="llmDialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton
          variant="primary"
          :loading="llmImporting"
          :disabled="llmPreviewItems.length === 0"
          @click="importLlmPreview"
        >
          {{ $t("pages.readAloudVocabularies.llmImportSelected") }}
        </AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
