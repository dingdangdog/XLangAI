<script setup lang="ts">
import { autoAdminCode } from "~/utils/autoAdminCode";
import { ArrowLeftIcon } from "@heroicons/vue/24/outline";
import {
  emptyLocaleMapByLanguageId,
  localeLabelsToMap,
  mapToLocaleLabels,
  resolveCategoryDisplayName,
  type CategoryLocaleEntry,
  type CategoryLocaleLabel,
} from "~/utils/readAloudCategoryLocale";
import AdminServiceConfigCard from "~/components/admin/config/AdminServiceConfigCard.vue";

const { t } = useI18n();
const route = useRoute();
const localePath = useLocalePath();
const isNarrow = useIsNarrow();

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
/** 窄屏：是否进入词汇钻取页 */
const mobileShowVocab = ref(!!String(route.query.categoryId ?? "").trim());

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
  if (isNarrow.value) mobileShowVocab.value = true;
  void navigateTo({
    path: localePath("/manage/read-aloud"),
    query: { categoryId: selectedCategoryId.value ?? undefined },
  });
}

function backToCategories() {
  mobileShowVocab.value = false;
}

watch(isNarrow, (narrow) => {
  if (!narrow) mobileShowVocab.value = false;
});

watch([catPage, catPageSize], () => void loadCategories(), { immediate: true });

// —— 词汇（右侧） ——
const vocabPage = ref(1);
const vocabPageSize = ref(20);
const vocabTotal = ref(0);
const vocabList = ref<Record<string, unknown>[]>([]);
const vocabLoading = ref(false);
const filterLanguageId = ref("");
const audioGeneratingId = ref("");
const selectedVocabIds = ref<string[]>([]);
const batchAudioDialogVisible = ref(false);
const batchAudioRunning = ref(false);
const batchAudioMode = ref<"missing" | "all" | "selected">("missing");

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
    const visible = new Set(res.items.map((r) => String(r.id)));
    selectedVocabIds.value = selectedVocabIds.value.filter((id) => visible.has(id));
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
const translatingTitles = ref(false);

const categoryForm = reactive({
  id: "",
  code: "",
  icon: "",
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
  categoryForm.icon = "";
  categoryForm.status = "active";
  categoryForm.sortOrder = 0;
  categoryForm.remark = "";
  syncCategoryLocales();
}

function findFilledSourceLocale():
  | { languageId: string; code: string; title: string }
  | null {
  for (const lang of activeLanguages.value) {
    const title = categoryLocales.value[lang.id]?.name?.trim() ?? "";
    if (title) return { languageId: lang.id, code: lang.code, title };
  }
  return null;
}

async function translateCategoryTitles() {
  const source = findFilledSourceLocale();
  if (!source) {
    toast.warning(t("pages.readAloud.translateTitlesNoSource"));
    return;
  }
  translatingTitles.value = true;
  try {
    const res = await $fetch<{
      titlesByLanguageId: Record<string, string>;
    }>("/api/admin/read-aloud-categories/translate-titles", {
      method: "POST",
      body: {
        sourceLanguageId: source.languageId,
        title: source.title,
      },
    });
    const map = res.titlesByLanguageId ?? {};
    let applied = 0;
    for (const lang of activeLanguages.value) {
      const translated = map[lang.id]?.trim();
      if (!translated) continue;
      if (lang.id === source.languageId) continue;
      if (categoryLocales.value[lang.id]) {
        categoryLocales.value[lang.id].name = translated;
        applied += 1;
      }
    }
    toast.success(t("pages.readAloud.translateTitlesDone", { count: applied }));
  } catch (e) {
    toast.error(t("pages.readAloud.translateTitlesFailed"));
    console.error(e);
  } finally {
    translatingTitles.value = false;
  }
}

function seedLegacyNameIntoFirstLang(fallbackName: string) {
  const name = fallbackName.trim();
  if (!name) return;
  const firstId = activeLanguages.value[0]?.id;
  if (!firstId || !categoryLocales.value[firstId]) return;
  if (!categoryLocales.value[firstId].name.trim()) {
    categoryLocales.value[firstId].name = name;
  }
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
  categoryForm.icon = String(target.icon ?? "");
  categoryForm.status = String(target.status ?? "active");
  categoryForm.sortOrder = Number(target.sortOrder ?? 0);
  categoryForm.remark = String(target.remark ?? "");
  await loadActiveLanguages();
  const labels = target.localeLabels as CategoryLocaleLabel[] | undefined;
  syncCategoryLocales(localeLabelsToMap(labels));
  seedLegacyNameIntoFirstLang(String(target.name ?? ""));
  categoryDialogVisible.value = true;
}

async function submitCategory() {
  const localeLabels = mapToLocaleLabels(
    activeLanguages.value.map((l) => l.id),
    categoryLocales.value,
  );
  if (!localeLabels.length) {
    toast.warning(t("pages.readAloud.fillOneLocaleName"));
    return;
  }
  const primaryName = localeLabels[0].name;
  const body = {
    code:
      categoryDialogMode.value === "create"
        ? autoAdminCode(primaryName)
        : categoryForm.code,
    name: primaryName,
    localeLabels,
    icon: categoryForm.icon.trim() || null,
    description: null,
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

const allVocabSelectedOnPage = computed(() => {
  if (!vocabList.value.length) return false;
  return vocabList.value.every((r) => selectedVocabIds.value.includes(String(r.id)));
});

function toggleSelectAllVocab(checked: boolean) {
  if (checked) {
    const ids = new Set(selectedVocabIds.value);
    for (const r of vocabList.value) ids.add(String(r.id));
    selectedVocabIds.value = [...ids];
  } else {
    const pageIds = new Set(vocabList.value.map((r) => String(r.id)));
    selectedVocabIds.value = selectedVocabIds.value.filter((id) => !pageIds.has(id));
  }
}

function toggleVocabSelected(id: string, checked: boolean) {
  if (checked) {
    if (!selectedVocabIds.value.includes(id)) {
      selectedVocabIds.value = [...selectedVocabIds.value, id];
    }
  } else {
    selectedVocabIds.value = selectedVocabIds.value.filter((x) => x !== id);
  }
}

function openBatchAudio() {
  if (!selectedCategoryId.value) {
    toast.warning(t("pages.readAloud.selectCategoryFirst"));
    return;
  }
  batchAudioMode.value = selectedVocabIds.value.length > 0 ? "selected" : "missing";
  batchAudioDialogVisible.value = true;
}

async function runBatchAudio() {
  if (!selectedCategoryId.value) {
    toast.warning(t("pages.readAloud.selectCategoryFirst"));
    return;
  }
  if (batchAudioMode.value === "selected" && selectedVocabIds.value.length === 0) {
    toast.warning(t("pages.readAloudVocabularies.batchAudioNothingSelected"));
    return;
  }
  if (batchAudioMode.value === "selected" && selectedVocabIds.value.length > 30) {
    toast.warning(t("pages.readAloudVocabularies.batchAudioTooManySelected"));
    return;
  }

  batchAudioRunning.value = true;
  try {
    const body: Record<string, unknown> = {
      part: "both",
      onlyMissing: batchAudioMode.value === "missing",
    };
    if (batchAudioMode.value === "selected") {
      body.ids = [...selectedVocabIds.value];
      // 勾选模式下仍尊重「仅缺失」以外的两种：selected 用 onlyMissing=false 覆盖已有
      body.onlyMissing = false;
    } else {
      body.categoryId = selectedCategoryId.value;
      if (filterLanguageId.value) body.languageId = filterLanguageId.value;
    }

    const res = await $fetch<{
      ok: number;
      failed: Array<{ id: string; word?: string; error: string }>;
      processed: number;
      remaining: number;
      totalMatched: number;
    }>("/api/admin/read-aloud-vocabularies/batch-generate-audio", {
      method: "POST",
      body,
      timeout: 600_000,
    });

    const failCount = res.failed?.length ?? 0;
    if (failCount === 0 && (res.ok ?? 0) === 0 && (res.totalMatched ?? 0) === 0) {
      toast.warning(t("pages.readAloudVocabularies.batchAudioNothingToDo"));
    } else if (failCount === 0) {
      toast.success(
        t("pages.readAloudVocabularies.batchAudioDone", {
          ok: res.ok ?? 0,
          remaining: res.remaining ?? 0,
        }),
      );
    } else {
      toast.warning(
        t("pages.readAloudVocabularies.batchAudioPartial", {
          ok: res.ok ?? 0,
          failed: failCount,
          remaining: res.remaining ?? 0,
        }),
      );
      console.warn("[batch-generate-audio] failures", res.failed);
    }

    batchAudioDialogVisible.value = false;
    selectedVocabIds.value = [];
    await loadVocabularies();
  } catch (e) {
    toast.error(t("pages.readAloudVocabularies.batchAudioFailed"));
    console.error(e);
  } finally {
    batchAudioRunning.value = false;
  }
}

const readAloudAudioEl = ref<HTMLAudioElement | null>(null);

function readAloudAudioPlayUrl(audioUrl: unknown, localFilename: unknown): string | null {
  const raw = String(audioUrl ?? "").trim();
  if (!raw) return null;
  if (raw.startsWith("http://") || raw.startsWith("https://")) return raw;

  const local = String(localFilename ?? "").trim();
  if (local) {
    return `/api/admin/preview-audio/${encodeURIComponent(local)}`;
  }

  const match = raw.match(/\/api\/v1\/audio\/([^/?#]+)/);
  if (match?.[1]) {
    return `/api/admin/preview-audio/${encodeURIComponent(match[1])}`;
  }

  return null;
}

function hasReadAloudAudio(row: Record<string, unknown>, part: "word" | "sentence") {
  const url = part === "word" ? row.wordAudioUrl : row.sentenceAudioUrl;
  return !!String(url ?? "").trim();
}

function playReadAloudAudio(row: Record<string, unknown>, part: "word" | "sentence") {
  const url =
    part === "word"
      ? readAloudAudioPlayUrl(row.wordAudioUrl, row.wordAudioLocalFilename)
      : readAloudAudioPlayUrl(row.sentenceAudioUrl, row.sentenceAudioLocalFilename);
  if (!url) {
    toast.warning(t("pages.readAloudVocabularies.audioNotGenerated"));
    return;
  }
  let el = readAloudAudioEl.value;
  if (!el) {
    el = new Audio();
    readAloudAudioEl.value = el;
  }
  el.src = url;
  void el.play().catch((e) => {
    toast.error(t("toast.playFailed"));
    console.error(e);
  });
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
    const localized = resolveCategoryDisplayName(row, filterLanguageId.value);
    if (localized) return localized;
  }
  return categoryTitle(row);
}

onMounted(() => {
  void loadOptions();
  void loadCategories();
});
</script>

<template>
  <AdminListPage fill>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.readAloud.title')"
        :description="$t('pages.readAloud.description')"
      />
    </template>

    <div class="flex min-h-0 flex-1 flex-col">
      <div
        class="grid min-h-0 flex-1 gap-4 overflow-hidden lg:grid-cols-[minmax(260px,300px)_minmax(0,1fr)]"
      >
      <!-- 左侧：场景卡片 -->
      <aside
        class="min-h-0 flex-col overflow-hidden rounded-xl border border-border bg-surface shadow-sm"
        :class="isNarrow && mobileShowVocab ? 'hidden' : 'flex'"
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
        class="min-h-0 min-w-0 flex-col overflow-hidden rounded-xl border border-border bg-surface shadow-sm"
        :class="isNarrow && !mobileShowVocab ? 'hidden' : 'flex'"
      >
        <div
          v-if="!selectedCategory"
          class="flex flex-1 flex-col items-center justify-center px-6 py-16 text-center"
        >
          <p class="text-sm font-medium text-foreground">{{ $t("pages.readAloud.selectCategoryTitle") }}</p>
          <p class="mt-2 max-w-sm text-sm text-muted">{{ $t("pages.readAloud.selectCategoryHint") }}</p>
        </div>

        <template v-else>
          <header class="shrink-0 border-b border-border px-3 py-3 sm:px-4 md:px-5">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div class="flex min-w-0 flex-1 items-start gap-2">
                <button
                  v-if="isNarrow"
                  type="button"
                  class="mt-0.5 rounded-lg p-1.5 text-muted hover:bg-surface-muted lg:hidden"
                  :aria-label="$t('common.back')"
                  @click="backToCategories"
                >
                  <ArrowLeftIcon class="h-5 w-5" />
                </button>
                <div class="min-w-0 flex-1">
                  <h2 class="truncate text-base font-semibold text-foreground">
                    {{ selectedCategory ? categoryCardTitle(selectedCategory) : "" }}
                  </h2>
                </div>
              </div>
              <div class="flex w-full flex-wrap gap-2 sm:w-auto">
                <AdminButton size="sm" variant="secondary" @click="openCategoryEdit()">
                  {{ $t("pages.readAloud.editCategory") }}
                </AdminButton>
                <AdminButton size="sm" variant="secondary" class="!text-danger-600" @click="removeCategory()">
                  {{ $t("pages.readAloud.deleteCategory") }}
                </AdminButton>
                <AdminButton size="sm" @click="openLlmBatch">
                  {{ $t("pages.readAloudVocabularies.llmBatch") }}
                </AdminButton>
                <AdminButton size="sm" variant="secondary" @click="openBatchAudio">
                  {{ $t("pages.readAloudVocabularies.batchAudio") }}
                </AdminButton>
                <AdminButton size="sm" variant="primary" @click="openVocabCreate">
                  {{ $t("pages.readAloud.addVocabulary") }}
                </AdminButton>
              </div>
            </div>

            <div class="mt-3 flex flex-wrap items-end gap-3">
              <AdminFormField :label="$t('pages.readAloudVocabularies.filterLanguage')" class="w-full min-w-0 sm:min-w-[200px] sm:w-auto">
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
                <AdminTh width="40px">
                  <input
                    type="checkbox"
                    class="rounded border-border"
                    :checked="allVocabSelectedOnPage"
                    :disabled="!vocabList.length"
                    @change="toggleSelectAllVocab(($event.target as HTMLInputElement).checked)"
                  />
                </AdminTh>
                <AdminTh>{{ $t("pages.readAloudVocabularies.word") }}</AdminTh>
                <AdminTh>{{ $t("pages.readAloudVocabularies.exampleSentence") }}</AdminTh>
                <AdminTh>{{ $t("pages.readAloudVocabularies.language") }}</AdminTh>
                <AdminTh>{{ $t("pages.readAloudVocabularies.audio") }}</AdminTh>
                <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
                <AdminTh width="280px" align="right">{{ $t("common.actions") }}</AdminTh>
              </template>
              <AdminTr v-for="row in vocabList" :key="String(row.id)">
                <AdminTd>
                  <input
                    type="checkbox"
                    class="rounded border-border"
                    :checked="selectedVocabIds.includes(String(row.id))"
                    @change="toggleVocabSelected(String(row.id), ($event.target as HTMLInputElement).checked)"
                  />
                </AdminTd>
                <AdminTd>{{ row.word }}</AdminTd>
                <AdminTd>
                  <AdminCellText :title="String(row.exampleSentence ?? '')">
                    {{ row.exampleSentence }}
                  </AdminCellText>
                </AdminTd>
                <AdminTd>{{ labelById(languageOptions, row.languageId) }}</AdminTd>
                <AdminTd>
                  <div class="flex flex-col gap-1.5 text-xs">
                    <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                      <span class="text-muted">{{ $t("pages.readAloudVocabularies.word") }}:</span>
                      <span>{{
                        hasReadAloudAudio(row, "word") ? $t("common.generated") : $t("common.notGenerated")
                      }}</span>
                      <AdminButton
                        v-if="hasReadAloudAudio(row, 'word')"
                        variant="link"
                        size="sm"
                        class="!px-0"
                        @click="playReadAloudAudio(row, 'word')"
                      >
                        {{ $t("common.play") }}
                      </AdminButton>
                    </div>
                    <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                      <span class="text-muted">{{ $t("pages.readAloudVocabularies.exampleSentence") }}:</span>
                      <span>{{
                        hasReadAloudAudio(row, "sentence")
                          ? $t("common.generated")
                          : $t("common.notGenerated")
                      }}</span>
                      <AdminButton
                        v-if="hasReadAloudAudio(row, 'sentence')"
                        variant="link"
                        size="sm"
                        class="!px-0"
                        @click="playReadAloudAudio(row, 'sentence')"
                      >
                        {{ $t("common.play") }}
                      </AdminButton>
                    </div>
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
              <template #mobile>
                <p v-if="!vocabList.length && !vocabLoading" class="py-12 text-center text-sm text-muted">
                  {{ $t("table.noData") }}
                </p>
                <AdminMobileCard
                  v-for="row in vocabList"
                  :key="String(row.id)"
                  :title="String(row.word ?? '')"
                  :subtitle="labelById(languageOptions, row.languageId)"
                >
                  <template #badge>
                    <AdminBadge>{{ row.status }}</AdminBadge>
                  </template>
                  <template #menu>
                    <AdminOverflowMenu
                      :actions="[
                        {
                          label: $t('pages.readAloudVocabularies.generateBoth'),
                          onClick: () => generateAudio(row, 'both'),
                        },
                        { label: $t('common.edit'), onClick: () => openVocabEdit(row) },
                        {
                          label: $t('common.delete'),
                          danger: true,
                          onClick: () => removeVocab(row),
                        },
                      ]"
                    />
                  </template>
                  <AdminMobileMeta :label="$t('pages.readAloudVocabularies.exampleSentence')">
                    {{ row.exampleSentence || $t("common.emDash") }}
                  </AdminMobileMeta>
                  <AdminMobileMeta :label="$t('pages.readAloudVocabularies.audio')">
                    {{
                      [
                        hasReadAloudAudio(row, "word") ? $t("pages.readAloudVocabularies.word") : null,
                        hasReadAloudAudio(row, "sentence")
                          ? $t("pages.readAloudVocabularies.exampleSentence")
                          : null,
                      ]
                        .filter(Boolean)
                        .join(" · ") || $t("common.notGenerated")
                    }}
                  </AdminMobileMeta>
                  <template #footer>
                    <AdminButton
                      v-if="hasReadAloudAudio(row, 'word')"
                      variant="link"
                      size="sm"
                      class="!px-0"
                      @click="playReadAloudAudio(row, 'word')"
                    >
                      {{ $t("common.play") }} · {{ $t("pages.readAloudVocabularies.word") }}
                    </AdminButton>
                    <AdminButton
                      v-if="hasReadAloudAudio(row, 'sentence')"
                      variant="link"
                      size="sm"
                      class="!px-0"
                      @click="playReadAloudAudio(row, 'sentence')"
                    >
                      {{ $t("common.play") }} · {{ $t("pages.readAloudVocabularies.exampleSentence") }}
                    </AdminButton>
                  </template>
                </AdminMobileCard>
              </template>
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
    </div>

    <!-- 场景对话框 -->
    <AdminDialog
      v-model="categoryDialogVisible"
      :title="
        categoryDialogMode === 'create'
          ? $t('pages.readAloudCategories.createDialog')
          : $t('pages.readAloudCategories.editDialog')
      "
      width="lg"
    >
      <div class="space-y-4">
        <div v-if="activeLanguagesLoading" class="py-6">
          <AdminSkeleton class="h-24 rounded-lg" />
        </div>
        <p v-else-if="!activeLanguages.length" class="text-sm text-muted">
          {{ $t("pages.readAloud.noActiveLanguages") }}
          <AdminButton to="/manage/languages" variant="link" size="sm" class="ml-1">
            {{ $t("nav.items.languages") }}
          </AdminButton>
        </p>
        <div v-else class="space-y-2">
          <div class="flex flex-wrap items-center gap-2">
            <AdminButton
              variant="secondary"
              :loading="translatingTitles"
              :disabled="activeLanguagesLoading"
              @click="translateCategoryTitles"
            >
              {{ $t("pages.readAloud.translateTitles") }}
            </AdminButton>
            <span class="text-xs text-muted">{{ $t("pages.readAloud.translateTitlesHint") }}</span>
          </div>
          <div class="overflow-x-auto rounded-lg border border-border">
          <table class="min-w-full text-sm">
            <thead>
              <tr class="border-b border-border bg-surface-muted/40 text-left text-xs text-muted">
                <th class="px-3 py-2 font-medium">{{ $t("pages.readAloudVocabularies.language") }}</th>
                <th class="px-3 py-2 font-medium">{{ $t("pages.readAloud.sceneTitle") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="lang in activeLanguages"
                :key="lang.id"
                class="border-b border-border last:border-b-0"
              >
                <td class="whitespace-nowrap px-3 py-3 align-middle text-foreground">{{ lang.label }}</td>
                <td class="px-3 py-2 align-middle">
                  <AdminInput
                    v-if="categoryLocales[lang.id]"
                    v-model="categoryLocales[lang.id].name"
                  />
                </td>
              </tr>
            </tbody>
          </table>
          </div>
        </div>

        <AdminFormField :label="$t('pages.readAloudCategories.icon')">
          <AdminInput v-model="categoryForm.icon" />
        </AdminFormField>
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
      width="lg"
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

    <!-- 批量生成音频 -->
    <AdminDialog
      v-model="batchAudioDialogVisible"
      :title="$t('pages.readAloudVocabularies.batchAudioDialog')"
      size="md"
    >
      <div class="space-y-4">
        <p class="text-sm text-muted">
          {{ $t("pages.readAloudVocabularies.batchAudioHint") }}
        </p>
        <div class="space-y-2">
          <label class="flex cursor-pointer items-start gap-2 text-sm">
            <input v-model="batchAudioMode" type="radio" value="missing" class="mt-1" />
            <span>{{ $t("pages.readAloudVocabularies.batchAudioModeMissing") }}</span>
          </label>
          <label class="flex cursor-pointer items-start gap-2 text-sm">
            <input v-model="batchAudioMode" type="radio" value="all" class="mt-1" />
            <span>{{ $t("pages.readAloudVocabularies.batchAudioModeAll") }}</span>
          </label>
          <label class="flex cursor-pointer items-start gap-2 text-sm">
            <input
              v-model="batchAudioMode"
              type="radio"
              value="selected"
              class="mt-1"
              :disabled="selectedVocabIds.length === 0"
            />
            <span>
              {{
                $t("pages.readAloudVocabularies.batchAudioModeSelected", {
                  count: selectedVocabIds.length,
                })
              }}
            </span>
          </label>
        </div>
        <p v-if="batchAudioRunning" class="text-sm text-muted">
          {{ $t("pages.readAloudVocabularies.batchAudioRunning") }}
        </p>
      </div>
      <template #footer>
        <AdminButton :disabled="batchAudioRunning" @click="batchAudioDialogVisible = false">
          {{ $t("common.cancel") }}
        </AdminButton>
        <AdminButton variant="primary" :loading="batchAudioRunning" @click="runBatchAudio">
          {{ $t("pages.readAloudVocabularies.batchAudioRun") }}
        </AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
