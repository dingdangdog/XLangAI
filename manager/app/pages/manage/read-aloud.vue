<script setup lang="ts">
import { autoAdminCode } from "~/utils/autoAdminCode";
import {
  emptyLocaleMapByLanguageId,
  localeLabelsToMap,
  mapToLocaleLabels,
  resolveCategoryDisplayName,
  type CategoryLocaleEntry,
  type CategoryLocaleLabel,
} from "~/utils/readAloudCategoryLocale";

const { t } = useI18n();
const route = useRoute();
const localePath = useLocalePath();

const categoriesApi = useAdminResourceApi("/api/admin/read-aloud-categories");
const vocabApi = useAdminResourceApi("/api/admin/read-aloud-vocabularies");
const toast = useToast();
const { confirm } = useConfirm();

type Opt = { id: string; label: string };
type LangOpt = { id: string; code: string; label: string };

// —— 场景（左侧） ——
const catPage = ref(1);
const catPageSize = ref(50);
const catTotal = ref(0);
const categories = ref<Record<string, unknown>[]>([]);
const categoriesLoading = ref(false);
const selectedCategoryId = ref<string | null>(
  String(route.query.categoryId ?? "").trim() || null,
);

const selectedCategory = computed(() =>
  categories.value.find((r) => String(r.id) === selectedCategoryId.value) ?? null,
);

async function loadCategories() {
  categoriesLoading.value = true;
  try {
    const res = await categoriesApi.list({ page: catPage.value, pageSize: catPageSize.value });
    categories.value = res.items;
    catTotal.value = res.total;
    if (selectedCategoryId.value && !categories.value.some((r) => String(r.id) === selectedCategoryId.value)) {
      selectedCategoryId.value = categories.value[0] ? String(categories.value[0].id) : null;
    }
    if (!selectedCategoryId.value && categories.value.length) {
      selectedCategoryId.value = String(categories.value[0].id);
    }
  } catch (e) {
    toast.error(t("toast.loadFailed"));
    console.error(e);
  } finally {
    categoriesLoading.value = false;
  }
}

function selectCategory(row: Record<string, unknown>) {
  selectedCategoryId.value = String(row.id);
  vocabPage.value = 1;
  void navigateTo({
    path: localePath("/manage/read-aloud"),
    query: { categoryId: selectedCategoryId.value ?? undefined },
  });
}

watch([catPage, catPageSize], () => void loadCategories(), { immediate: true });

// —— 词汇（右侧） ——
const vocabPage = ref(1);
const vocabPageSize = ref(20);
const vocabTotal = ref(0);
const vocabList = ref<Record<string, unknown>[]>([]);
const vocabLoading = ref(false);
const filterLanguageId = ref("");
const audioGeneratingId = ref("");

const languageOptions = ref<LangOpt[]>([]);
/** 场景表单用：每次打开对话框从「语言管理」重新拉取 active 语言 */
const activeLanguages = ref<LangOpt[]>([]);
const activeLanguagesLoading = ref(false);
const voiceOptions = ref<Opt[]>([]);
const llmOptions = ref<Opt[]>([]);
const optionsLoading = ref(false);

async function loadOptions() {
  optionsLoading.value = true;
  try {
    const [langs, voices, llms] = await Promise.all([
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
    const active = langs.items.filter((r) => String(r.status ?? "active") === "active");
    languageOptions.value = active.map((r) => ({
      id: String(r.id),
      code: String(r.code ?? "")
        .trim()
        .toLowerCase(),
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
    if (!filterLanguageId.value && languageOptions.value.length) {
      filterLanguageId.value = languageOptions.value[0].id;
    }
  } catch (e) {
    toast.error(t("toast.loadOptionsFailed"));
    console.error(e);
  } finally {
    optionsLoading.value = false;
  }
}

async function loadVocabularies() {
  if (!selectedCategoryId.value) {
    vocabList.value = [];
    vocabTotal.value = 0;
    return;
  }
  vocabLoading.value = true;
  try {
    const query: Record<string, unknown> = {
      page: vocabPage.value,
      pageSize: vocabPageSize.value,
      categoryId: selectedCategoryId.value,
    };
    if (filterLanguageId.value) query.languageId = filterLanguageId.value;
    const res = await vocabApi.list(query);
    vocabList.value = res.items;
    vocabTotal.value = res.total;
  } catch (e) {
    toast.error(t("toast.loadFailed"));
    console.error(e);
  } finally {
    vocabLoading.value = false;
  }
}

watch([vocabPage, vocabPageSize, selectedCategoryId, filterLanguageId], () => void loadVocabularies());

// —— 场景表单 ——
const categoryDialogVisible = ref(false);
const categoryDialogMode = ref<"create" | "edit">("create");
const categorySaving = ref(false);

const categoryForm = reactive({
  id: "",
  code: "",
  name: "",
  icon: "",
  description: "",
  status: "active",
  sortOrder: 0,
  remark: "",
});

const categoryLocales = ref<Record<string, CategoryLocaleEntry>>({});

async function loadActiveLanguages() {
  activeLanguagesLoading.value = true;
  try {
    const res = await $fetch<{ items: Record<string, unknown>[] }>("/api/admin/languages", {
      query: { page: 1, pageSize: 200 },
    });
    activeLanguages.value = res.items
      .filter((r) => String(r.status ?? "active") === "active")
      .map((r) => ({
        id: String(r.id),
        code: String(r.code ?? "")
          .trim()
          .toLowerCase(),
        label: `${r.code} · ${r.name}`,
      }));
  } catch (e) {
    toast.error(t("toast.loadOptionsFailed"));
    console.error(e);
  } finally {
    activeLanguagesLoading.value = false;
  }
}

function syncCategoryLocales(merge?: Record<string, CategoryLocaleEntry>) {
  const ids = activeLanguages.value.map((l) => l.id);
  const next = emptyLocaleMapByLanguageId(ids);
  for (const id of ids) {
    if (merge?.[id]) {
      next[id] = { name: merge[id].name ?? "", description: merge[id].description ?? "" };
    }
  }
  categoryLocales.value = next;
}

function resetCategoryForm() {
  categoryForm.id = "";
  categoryForm.code = "";
  categoryForm.name = "";
  categoryForm.icon = "";
  categoryForm.description = "";
  categoryForm.status = "active";
  categoryForm.sortOrder = 0;
  categoryForm.remark = "";
  syncCategoryLocales();
}

async function openCategoryCreate() {
  categoryDialogMode.value = "create";
  resetCategoryForm();
  await loadActiveLanguages();
  syncCategoryLocales();
  categoryDialogVisible.value = true;
}

async function openCategoryEdit(row?: Record<string, unknown>) {
  const target = row ?? selectedCategory.value;
  if (!target) return;
  categoryDialogMode.value = "edit";
  categoryForm.id = String(target.id ?? "");
  categoryForm.code = String(target.code ?? "");
  categoryForm.name = String(target.name ?? "");
  categoryForm.icon = String(target.icon ?? "");
  categoryForm.description = String(target.description ?? "");
  categoryForm.status = String(target.status ?? "active");
  categoryForm.sortOrder = Number(target.sortOrder ?? 0);
  categoryForm.remark = String(target.remark ?? "");
  await loadActiveLanguages();
  const labels = target.localeLabels as CategoryLocaleLabel[] | undefined;
  syncCategoryLocales(localeLabelsToMap(labels));
  categoryDialogVisible.value = true;
}

async function submitCategory() {
  if (!categoryForm.name.trim()) {
    toast.warning(t("validation.fillName"));
    return;
  }
  const localeLabels = mapToLocaleLabels(
    activeLanguages.value.map((l) => l.id),
    categoryLocales.value,
  );
  const body = {
    code:
      categoryDialogMode.value === "create"
        ? autoAdminCode(categoryForm.name)
        : categoryForm.code,
    name: categoryForm.name.trim(),
    localeLabels,
    icon: categoryForm.icon.trim() || null,
    description: categoryForm.description.trim() || null,
    status: categoryForm.status,
    sortOrder: categoryForm.sortOrder,
    remark: categoryForm.remark.trim() || null,
  };
  categorySaving.value = true;
  try {
    if (categoryDialogMode.value === "create") {
      await categoriesApi.create(body);
      toast.success(t("toast.created"));
      await loadCategories();
    } else {
      await categoriesApi.update(categoryForm.id, { id: categoryForm.id, ...body });
      toast.success(t("toast.saved"));
      await loadCategories();
    }
    categoryDialogVisible.value = false;
  } catch (e) {
    toast.error(t("toast.saveFailed"));
    console.error(e);
  } finally {
    categorySaving.value = false;
  }
}

async function removeCategory(row?: Record<string, unknown>) {
  const target = row ?? selectedCategory.value;
  if (!target) return;
  const ok = await confirm({
    message: t("pages.readAloudCategories.deleteConfirm"),
    danger: true,
    confirmLabel: t("common.delete"),
  });
  if (!ok) return;
  const id = String(target.id);
  try {
    await categoriesApi.remove(id);
    toast.success(t("toast.deleted"));
    if (selectedCategoryId.value === id) {
      selectedCategoryId.value = null;
    }
    await loadCategories();
  } catch (e) {
    toast.error(t("toast.deleteFailed"));
    console.error(e);
  }
}

// —— 词汇表单 ——
const vocabDialogVisible = ref(false);
const vocabDialogMode = ref<"create" | "edit">("create");
const vocabSaving = ref(false);

const vocabForm = reactive({
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

function resetVocabForm() {
  vocabForm.id = "";
  vocabForm.categoryId = selectedCategoryId.value || "";
  vocabForm.languageId = filterLanguageId.value || "";
  vocabForm.word = "";
  vocabForm.exampleSentence = "";
  vocabForm.voiceRoleId = voiceOptions.value[0]?.id ?? "";
  vocabForm.status = "active";
  vocabForm.sortOrder = 0;
  vocabForm.remark = "";
}

function openVocabCreate() {
  if (!selectedCategoryId.value) {
    toast.warning(t("pages.readAloud.selectCategoryFirst"));
    return;
  }
  vocabDialogMode.value = "create";
  resetVocabForm();
  vocabDialogVisible.value = true;
}

function openVocabEdit(row: Record<string, unknown>) {
  vocabDialogMode.value = "edit";
  vocabForm.id = String(row.id ?? "");
  vocabForm.categoryId = String(row.categoryId ?? "");
  vocabForm.languageId = String(row.languageId ?? "");
  vocabForm.word = String(row.word ?? "");
  vocabForm.exampleSentence = String(row.exampleSentence ?? "");
  vocabForm.voiceRoleId = String(row.voiceRoleId ?? "");
  vocabForm.status = String(row.status ?? "active");
  vocabForm.sortOrder = Number(row.sortOrder ?? 0);
  vocabForm.remark = String(row.remark ?? "");
  vocabDialogVisible.value = true;
}

async function submitVocab() {
  if (
    !vocabForm.categoryId ||
    !vocabForm.languageId ||
    !vocabForm.word.trim() ||
    !vocabForm.exampleSentence.trim() ||
    !vocabForm.voiceRoleId
  ) {
    toast.warning(t("pages.readAloudVocabularies.fillRequired"));
    return;
  }
  const body = {
    categoryId: vocabForm.categoryId,
    languageId: vocabForm.languageId,
    word: vocabForm.word.trim(),
    exampleSentence: vocabForm.exampleSentence.trim(),
    voiceRoleId: vocabForm.voiceRoleId,
    status: vocabForm.status,
    sortOrder: vocabForm.sortOrder,
    remark: vocabForm.remark.trim() || null,
  };
  vocabSaving.value = true;
  try {
    if (vocabDialogMode.value === "create") {
      await vocabApi.create(body);
      toast.success(t("toast.created"));
    } else {
      await vocabApi.update(vocabForm.id, { id: vocabForm.id, ...body });
      toast.success(t("toast.saved"));
    }
    vocabDialogVisible.value = false;
    filterLanguageId.value = vocabForm.languageId;
    await loadVocabularies();
  } catch (e) {
    toast.error(t("toast.saveFailed"));
    console.error(e);
  } finally {
    vocabSaving.value = false;
  }
}

async function removeVocab(row: Record<string, unknown>) {
  const ok = await confirm({
    message: t("pages.readAloudVocabularies.deleteConfirm"),
    danger: true,
    confirmLabel: t("common.delete"),
  });
  if (!ok) return;
  try {
    await vocabApi.remove(String(row.id));
    toast.success(t("toast.deleted"));
    await loadVocabularies();
  } catch (e) {
    toast.error(t("toast.deleteFailed"));
    console.error(e);
  }
}

// —— LLM 批量 ——
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

function openLlmBatch() {
  if (!selectedCategoryId.value) {
    toast.warning(t("pages.readAloud.selectCategoryFirst"));
    return;
  }
  llmForm.llmServiceConfigId = llmOptions.value[0]?.id ?? "";
  llmForm.categoryId = selectedCategoryId.value;
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
    filterLanguageId.value = llmForm.languageId;
    await loadVocabularies();
  } catch (e) {
    toast.error(t("pages.readAloudVocabularies.llmImportFailed"));
    console.error(e);
  } finally {
    llmImporting.value = false;
  }
}

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
    await loadVocabularies();
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

const languageSelectOptions = computed(() =>
  languageOptions.value.map((o) => ({ value: o.id, label: o.label })),
);

const voiceSelectOptions = computed(() =>
  voiceOptions.value.map((o) => ({ value: o.id, label: o.label })),
);

const llmSelectOptions = computed(() =>
  llmOptions.value.map((o) => ({ value: o.id, label: o.label })),
);

const llmSelectedCount = computed(
  () => llmPreviewItems.value.filter((it) => it.selected).length,
);

const labelById = (opts: Opt[], id: unknown) => {
  const key = String(id ?? "").trim();
  if (!key) return t("common.emDash");
  return opts.find((o) => o.id === key)?.label ?? key;
};

function categoryTitle(row: Record<string, unknown>) {
  return String(row.name ?? row.code ?? "");
}

function categorySubtitle(row: Record<string, unknown>) {
  if (filterLanguageId.value) {
    const name = resolveCategoryDisplayName(row, filterLanguageId.value);
    if (name !== String(row.name ?? "")) return name;
  }
  const desc = row.description ? String(row.description) : "";
  return desc || "";
}

function categoryCardTitle(row: Record<string, unknown>) {
  if (filterLanguageId.value) {
    return resolveCategoryDisplayName(row, filterLanguageId.value);
  }
  return categoryTitle(row);
}

onMounted(() => void loadOptions());
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.readAloud.title')"
        :description="$t('pages.readAloud.description')"
      />
    </template>

    <div
      class="grid min-h-0 flex-1 gap-4 overflow-hidden lg:grid-cols-[minmax(260px,300px)_minmax(0,1fr)]"
    >
      <!-- 左侧：场景卡片 -->
      <aside
        class="flex min-h-[320px] flex-col overflow-hidden rounded-xl border border-border bg-surface shadow-sm lg:min-h-0"
      >
        <header class="shrink-0 border-b border-border px-4 py-3">
          <div class="flex items-center justify-between gap-2">
            <div>
              <h2 class="text-sm font-semibold text-foreground">
                {{ $t("pages.readAloud.categoryList") }}
              </h2>
              <!-- <p class="mt-0.5 text-xs text-muted">{{ $t("pages.readAloud.categoryListHint") }}</p> -->
            </div>
            <AdminButton variant="primary" size="sm" @click="openCategoryCreate">
              {{ $t("common.create") }}
            </AdminButton>
          </div>
        </header>

        <div v-if="categoriesLoading && !categories.length" class="space-y-2 p-3">
          <AdminSkeleton v-for="i in 5" :key="i" class="h-16 rounded-xl" />
        </div>

        <div
          v-else-if="!categories.length"
          class="flex flex-1 flex-col items-center justify-center px-4 py-10 text-center"
        >
          <p class="text-sm text-muted">{{ $t("pages.readAloud.emptyCategories") }}</p>
          <AdminButton class="mt-3" variant="secondary" size="sm" @click="openCategoryCreate">
            {{ $t("pages.readAloud.createFirstCategory") }}
          </AdminButton>
        </div>

        <div v-else class="min-h-0 flex-1 space-y-2 overflow-y-auto p-3">
          <AdminServiceConfigCard
            v-for="row in categories"
            :key="String(row.id)"
            :name="categoryCardTitle(row)"
            :subtitle="categorySubtitle(row)"
            :meta="row.description ? String(row.description) : undefined"
            :badge="String(row.status)"
            :badge-variant="String(row.status) === 'active' ? 'success' : 'muted'"
            :active="String(row.status) === 'active'"
            :selected="selectedCategoryId === String(row.id)"
            @select="selectCategory(row)"
          />
        </div>

        <AdminPagination
          v-if="catTotal > catPageSize"
          v-model:page="catPage"
          v-model:page-size="catPageSize"
          :total="catTotal"
          class="shrink-0 border-t border-border px-3 py-2"
        />
      </aside>

      <!-- 右侧：词汇列表 -->
      <main
        class="flex min-h-[320px] min-w-0 flex-col overflow-hidden rounded-xl border border-border bg-surface shadow-sm lg:min-h-0"
      >
        <div
          v-if="!selectedCategory"
          class="flex flex-1 flex-col items-center justify-center px-6 py-16 text-center"
        >
          <p class="text-sm font-medium text-foreground">{{ $t("pages.readAloud.selectCategoryTitle") }}</p>
          <p class="mt-2 max-w-sm text-sm text-muted">{{ $t("pages.readAloud.selectCategoryHint") }}</p>
        </div>

        <template v-else>
          <header class="shrink-0 border-b border-border px-4 py-3 md:px-5">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div class="min-w-0 flex-1">
                <h2 class="truncate text-base font-semibold text-foreground">
                  {{ selectedCategory ? categoryCardTitle(selectedCategory) : "" }}
                </h2>
              </div>
              <div class="flex flex-wrap gap-2">
                <AdminButton size="sm" variant="secondary" @click="openCategoryEdit()">
                  {{ $t("pages.readAloud.editCategory") }}
                </AdminButton>
                <AdminButton size="sm" variant="secondary" class="!text-danger-600" @click="removeCategory()">
                  {{ $t("pages.readAloud.deleteCategory") }}
                </AdminButton>
                <AdminButton size="sm" @click="openLlmBatch">
                  {{ $t("pages.readAloudVocabularies.llmBatch") }}
                </AdminButton>
                <AdminButton size="sm" variant="primary" @click="openVocabCreate">
                  {{ $t("pages.readAloud.addVocabulary") }}
                </AdminButton>
              </div>
            </div>

            <div class="mt-3 flex flex-wrap items-end gap-3">
              <AdminFormField :label="$t('pages.readAloudVocabularies.filterLanguage')" class="min-w-[200px]">
                <AdminSelect
                  v-model="filterLanguageId"
                  :options="languageSelectOptions"
                  :loading="optionsLoading"
                />
              </AdminFormField>
            </div>
          </header>

          <div class="min-h-0 flex-1 overflow-auto">
            <AdminTable :loading="vocabLoading" class="border-0 shadow-none">
              <template #head>
                <AdminTh>{{ $t("pages.readAloudVocabularies.word") }}</AdminTh>
                <AdminTh>{{ $t("pages.readAloudVocabularies.exampleSentence") }}</AdminTh>
                <AdminTh>{{ $t("pages.readAloudVocabularies.language") }}</AdminTh>
                <AdminTh>{{ $t("pages.readAloudVocabularies.audio") }}</AdminTh>
                <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
                <AdminTh width="220px" align="right">{{ $t("common.actions") }}</AdminTh>
              </template>
              <AdminTr v-for="row in vocabList" :key="String(row.id)">
                <AdminTd>{{ row.word }}</AdminTd>
                <AdminTd>
                  <AdminCellText :title="String(row.exampleSentence ?? '')">
                    {{ row.exampleSentence }}
                  </AdminCellText>
                </AdminTd>
                <AdminTd>{{ labelById(languageOptions, row.languageId) }}</AdminTd>
                <AdminTd class="text-xs">
                  <div>
                    {{ row.wordAudioUrl ? $t("common.generated") : $t("common.notGenerated") }} /
                    {{ row.sentenceAudioUrl ? $t("common.generated") : $t("common.notGenerated") }}
                  </div>
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
                  <AdminButton variant="link" @click="openVocabEdit(row)">{{ $t("common.edit") }}</AdminButton>
                  <AdminButton variant="link" class="!text-danger-600" @click="removeVocab(row)">
                    {{ $t("common.delete") }}
                  </AdminButton>
                </AdminTd>
              </AdminTr>
            </AdminTable>
          </div>

          <AdminPagination
            v-model:page="vocabPage"
            v-model:page-size="vocabPageSize"
            :total="vocabTotal"
            class="shrink-0 border-t border-border"
          />
        </template>
      </main>
    </div>

    <!-- 场景对话框 -->
    <AdminDialog
      v-model="categoryDialogVisible"
      :title="
        categoryDialogMode === 'create'
          ? $t('pages.readAloudCategories.createDialog')
          : $t('pages.readAloudCategories.editDialog')
      "
      size="lg"
    >
      <div class="space-y-4">
        <AdminFormField :label="$t('common.name')" required>
          <AdminInput v-model="categoryForm.name" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudCategories.icon')">
          <AdminInput v-model="categoryForm.icon" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudCategories.description')">
          <AdminTextarea v-model="categoryForm.description" rows="2" />
        </AdminFormField>

        <div class="rounded-lg border border-border">
          <div class="flex flex-wrap items-center justify-between gap-2 border-b border-border px-3 py-2">
            <p class="text-xs text-muted">{{ $t("pages.readAloud.localesTableHint") }}</p>
            <AdminButton to="/manage/languages" variant="link" size="sm">
              {{ $t("nav.items.languages") }}
            </AdminButton>
          </div>
          <div v-if="activeLanguagesLoading" class="p-4">
            <AdminSkeleton class="h-20 rounded-lg" />
          </div>
          <p v-else-if="!activeLanguages.length" class="p-4 text-sm text-muted">
            {{ $t("pages.readAloud.noActiveLanguages") }}
          </p>
          <div v-else class="overflow-x-auto">
            <table class="min-w-full text-sm">
              <thead>
                <tr class="border-b border-border bg-surface-muted/40 text-left text-xs text-muted">
                  <th class="px-3 py-2 font-medium">{{ $t("pages.readAloudVocabularies.language") }}</th>
                  <th class="px-3 py-2 font-medium">{{ $t("common.name") }}</th>
                  <th class="px-3 py-2 font-medium">{{ $t("pages.readAloudCategories.description") }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="lang in activeLanguages"
                  :key="lang.id"
                  class="border-b border-border last:border-b-0"
                >
                  <td class="whitespace-nowrap px-3 py-2 align-middle text-foreground">{{ lang.label }}</td>
                  <td class="px-3 py-2 align-middle">
                    <AdminInput
                      v-if="categoryLocales[lang.id]"
                      v-model="categoryLocales[lang.id].name"
                      :placeholder="categoryForm.name"
                    />
                  </td>
                  <td class="px-3 py-2 align-middle">
                    <AdminInput
                      v-if="categoryLocales[lang.id]"
                      v-model="categoryLocales[lang.id].description"
                      :placeholder="categoryForm.description"
                    />
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <AdminFormField :label="$t('common.status')">
          <AdminSelect v-model="categoryForm.status" :options="statusOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('common.sort')">
          <AdminInput v-model.number="categoryForm.sortOrder" type="number" />
        </AdminFormField>
        <AdminFormField :label="$t('common.remark')">
          <AdminTextarea v-model="categoryForm.remark" rows="2" />
        </AdminFormField>
      </div>
      <template #footer>
        <AdminButton @click="categoryDialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="categorySaving" @click="submitCategory">
          {{ $t("common.save") }}
        </AdminButton>
      </template>
    </AdminDialog>

    <!-- 词汇对话框 -->
    <AdminDialog
      v-model="vocabDialogVisible"
      :title="
        vocabDialogMode === 'create'
          ? $t('pages.readAloudVocabularies.createDialog')
          : $t('pages.readAloudVocabularies.editDialog')
      "
      size="lg"
    >
      <div class="space-y-4">
        <AdminFormField :label="$t('pages.readAloudVocabularies.language')" required>
          <AdminSelect
            v-model="vocabForm.languageId"
            :options="languageSelectOptions"
            :loading="optionsLoading"
          />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudVocabularies.word')" required>
          <AdminInput v-model="vocabForm.word" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudVocabularies.exampleSentence')" required>
          <AdminTextarea v-model="vocabForm.exampleSentence" rows="2" />
        </AdminFormField>
        <AdminFormField :label="$t('pages.readAloudVocabularies.voiceRole')" required>
          <AdminSelect v-model="vocabForm.voiceRoleId" :options="voiceSelectOptions" :loading="optionsLoading" />
        </AdminFormField>
        <AdminFormField :label="$t('common.status')">
          <AdminSelect v-model="vocabForm.status" :options="statusOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('common.sort')">
          <AdminInput v-model.number="vocabForm.sortOrder" type="number" />
        </AdminFormField>
        <AdminFormField :label="$t('common.remark')">
          <AdminTextarea v-model="vocabForm.remark" rows="2" />
        </AdminFormField>
      </div>
      <template #footer>
        <AdminButton @click="vocabDialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="vocabSaving" @click="submitVocab">
          {{ $t("common.save") }}
        </AdminButton>
      </template>
    </AdminDialog>

    <!-- LLM 批量对话框 -->
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
          <AdminFormField :label="$t('pages.readAloudVocabularies.language')" required>
            <AdminSelect v-model="llmForm.languageId" :options="languageSelectOptions" :loading="optionsLoading" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.readAloudVocabularies.voiceRole')" required>
            <AdminSelect v-model="llmForm.voiceRoleId" :options="voiceSelectOptions" :loading="optionsLoading" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.readAloudVocabularies.llmExtraInstructions')" class="md:col-span-2">
            <AdminTextarea
              v-model="llmForm.extraInstructions"
              rows="2"
              :placeholder="$t('pages.readAloudVocabularies.llmExtraPlaceholder')"
            />
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

        <div
          v-if="llmPreviewItems.length"
          class="max-h-80 overflow-auto rounded border border-gray-200 dark:border-gray-700"
        >
          <table class="min-w-full text-sm">
            <thead class="sticky top-0 bg-gray-50 dark:bg-gray-900">
              <tr>
                <th class="px-3 py-2 text-left">{{ $t("common.actions") }}</th>
                <th class="px-3 py-2 text-left">{{ $t("pages.readAloudVocabularies.word") }}</th>
                <th class="px-3 py-2 text-left">{{ $t("pages.readAloudVocabularies.exampleSentence") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, idx) in llmPreviewItems"
                :key="idx"
                class="border-t border-gray-100 dark:border-gray-800"
              >
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
