import {
  configObjectFromSchemaFields,
  getConfigSchema,
  parseConfigObject,
  schemaFieldsFromConfigObject,
  stringifyConfigObject,
} from "~/utils/serviceConfigSchemas";
import {
  LLM_VENDOR_PRESET_CUSTOM_OPENAI,
  LLM_VENDOR_PRESETS,
  getLlmVendorPreset,
  isOpenAiCompatPreset,
  resolveLlmVendorPresetId,
  type LlmVendorPreset,
} from "~/utils/llmVendorPresets";
import { llmProtocolToFamily } from "~/utils/serverServiceCatalog";

export function useLlmConfigEditor() {
  const { t } = useI18n();
  const { serviceUsageTodayLine } = useUsageDisplay();
  const requestFetch = useRequestFetch();

  const API = "/api/admin/llm-service-configs";
  const api = useAdminResourceApi(API);
  const toast = useToast();
  const { confirm } = useConfirm();
  const { probe, probing, lastResult, clearProbe } = useServiceConfigProbe(API);

  const page = ref(1);
  const pageSize = ref(24);
  const total = ref(0);
  const list = ref<Record<string, unknown>[]>([]);
  const loading = ref(false);

  const drawerOpen = ref(false);
  const setupPhase = ref<"vendor" | "connect">("vendor");
  const editorMode = ref<"create" | "edit">("create");
  const showAdvanced = ref(false);
  const saving = ref(false);
  const deleting = ref(false);

  const form = reactive({
    id: "",
    code: "",
    name: "",
    vendorPresetId: "deepseek",
    storedProtocol: "deepseek",
    baseUrl: "",
    apiKey: "",
    modelCode: "",
    config: "{}",
    status: "active",
    sortOrder: 0,
    remark: "",
  });

  const schemaFields = ref<Record<string, string | number | boolean>>({});
  const snapshot = ref("");
  const modelOptions = ref<{ value: string; label: string }[]>([]);
  const modelsLoading = ref(false);
  const modelsError = ref<string | null>(null);
  const modelsManualOnly = ref(false);

  const activePreset = computed(() => getLlmVendorPreset(form.vendorPresetId));
  const isCustomOpenAi = computed(() => form.vendorPresetId === LLM_VENDOR_PRESET_CUSTOM_OPENAI);
  const isAzurePreset = computed(() => form.vendorPresetId === "azure_openai");
  const baseUrlEditable = computed(() => {
    const p = activePreset.value;
    if (!p || p.isCustom || isAzurePreset.value) return true;
    if (p.id === "ollama") return true;
    return !p.baseUrl;
  });
  const baseUrlReadonly = computed(() => !baseUrlEditable.value && Boolean(form.baseUrl));

  const vendorGridItems = computed(() =>
    LLM_VENDOR_PRESETS.map((p) => ({
      id: p.id,
      label: t(p.labelKey),
      description: p.descriptionKey ? t(p.descriptionKey) : undefined,
      group: t(p.groupKey),
    })),
  );

  const configSchema = computed(() => {
    const family = llmProtocolToFamily(form.storedProtocol);
    const flavor = form.storedProtocol === "azure_openai" ? "azure" : "generic";
    return getConfigSchema("llm", family, {
      llmOpenAiFlavor: flavor,
      llmStoredProtocol: form.storedProtocol,
    });
  });

  const setupStepNumber = computed(() => (setupPhase.value === "vendor" ? 1 : 2));
  const drawerTitle = computed(() =>
    editorMode.value === "create"
      ? setupPhase.value === "vendor"
        ? t("agentHub.addLlm")
        : t("agentHub.connectLlm")
      : form.name || t("agentHub.editLlm"),
  );
  const drawerSubtitle = computed(() => {
    if (setupPhase.value === "vendor") return t("agentHub.pickVendorSubtitle");
    const preset = activePreset.value;
    return preset ? t(preset.labelKey) : "";
  });

  const dirty = computed(() => serializeFormSnapshot() !== snapshot.value);
  const activeCount = computed(() => list.value.filter((r) => String(r.status) === "active").length);

  function autoLlmCode(protocol: string) {
    return `${(protocol || "openai").trim().replace(/[^a-zA-Z0-9_-]/g, "_")}-${Date.now()}`;
  }

  function serializeFormSnapshot() {
    return JSON.stringify({ ...form, schemaFields: schemaFields.value, showAdvanced: showAdvanced.value });
  }

  function syncSchemaFromConfig() {
    schemaFields.value = schemaFieldsFromConfigObject(
      configSchema.value,
      parseConfigObject(form.config),
    );
  }

  function syncConfigFromSchema() {
    const base = parseConfigObject(form.config);
    const fromSchema = configObjectFromSchemaFields(configSchema.value, schemaFields.value);
    form.config = stringifyConfigObject({ ...base, ...fromSchema });
  }

  function syncStoredFromPreset() {
    const preset = activePreset.value;
    if (preset) form.storedProtocol = preset.storedProtocol;
  }

  function applyVendorPreset(preset: LlmVendorPreset, keepApiKey = false) {
    form.vendorPresetId = preset.id;
    form.storedProtocol = preset.storedProtocol;
    if (!keepApiKey && (editorMode.value === "create" || !form.name.trim())) {
      form.name = t(preset.labelKey);
    }
    if (preset.isCustom) {
      if (editorMode.value === "create") form.baseUrl = "";
    } else if (preset.id === "azure_openai" && editorMode.value === "create") {
      form.baseUrl = "";
    } else {
      form.baseUrl = preset.baseUrl;
    }
    if (editorMode.value === "create" || !form.modelCode.trim()) {
      form.modelCode = preset.defaultModel;
    }
    syncSchemaFromConfig();
  }

  function resetForm() {
    form.id = "";
    form.code = "";
    form.vendorPresetId = "deepseek";
    applyVendorPreset(getLlmVendorPreset("deepseek")!);
    form.apiKey = "";
    form.config = "{}";
    form.status = "active";
    form.sortOrder = 0;
    form.remark = "";
    showAdvanced.value = false;
    modelOptions.value = [];
    modelsError.value = null;
    snapshot.value = serializeFormSnapshot();
    clearProbe();
  }

  function loadFormFromRow(row: Record<string, unknown>) {
    const stored = String(row.protocol ?? "openai");
    form.id = String(row.id ?? "");
    form.code = String(row.code ?? "");
    form.name = String(row.name ?? "");
    form.vendorPresetId = resolveLlmVendorPresetId({
      protocol: stored,
      baseUrl: row.baseUrl != null ? String(row.baseUrl) : null,
    });
    syncStoredFromPreset();
    form.baseUrl = String(row.baseUrl ?? "");
    form.apiKey = String(row.apiKey ?? "");
    form.modelCode = String(row.modelCode ?? "");
    form.config = row.config != null ? String(row.config) : "{}";
    form.status = String(row.status ?? "active");
    form.sortOrder = Number(row.sortOrder ?? 0);
    form.remark = String(row.remark ?? "");
    syncSchemaFromConfig();
    modelOptions.value = form.modelCode ? [{ value: form.modelCode, label: form.modelCode }] : [];
    snapshot.value = serializeFormSnapshot();
    clearProbe();
  }

  function vendorIdForRow(row: Record<string, unknown>) {
    return resolveLlmVendorPresetId({
      protocol: String(row.protocol ?? ""),
      baseUrl: row.baseUrl != null ? String(row.baseUrl) : null,
    });
  }

  function vendorLabelForRow(row: Record<string, unknown>) {
    const id = vendorIdForRow(row);
    const preset = getLlmVendorPreset(id);
    if (preset && id !== LLM_VENDOR_PRESET_CUSTOM_OPENAI) return t(preset.labelKey);
    return String(row.name ?? "");
  }

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

  watch(
    () => form.vendorPresetId,
    (id) => {
      const preset = getLlmVendorPreset(id);
      if (preset) applyVendorPreset(preset, true);
      modelOptions.value = [];
      modelsError.value = null;
    },
  );

  function openCreate() {
    editorMode.value = "create";
    setupPhase.value = "vendor";
    resetForm();
    drawerOpen.value = true;
  }

  function openEdit(row: Record<string, unknown>) {
    editorMode.value = "edit";
    setupPhase.value = "connect";
    loadFormFromRow(row);
    drawerOpen.value = true;
  }

  function closeDrawer() {
    drawerOpen.value = false;
  }

  function goToConnect() {
    const preset = activePreset.value;
    if (!preset) return;
    applyVendorPreset(preset, true);
    setupPhase.value = "connect";
  }

  function backToVendor() {
    if (editorMode.value === "create") setupPhase.value = "vendor";
  }

  async function fetchModels() {
    syncStoredFromPreset();
    const preset = activePreset.value;
    if (preset && !isOpenAiCompatPreset(preset)) {
      modelsManualOnly.value = true;
      return;
    }
    if (!form.apiKey.trim()) {
      toast.warning(t("serverCatalog.llm.apiKeyRequiredForModels"));
      return;
    }
    modelsLoading.value = true;
    modelsError.value = null;
    try {
      const res = await requestFetch<{
        models: { id: string }[];
        manualOnly?: boolean;
        error?: string;
      }>(`${API}/list-models`, {
        method: "POST",
        body: {
          configId: form.id || undefined,
          vendorPresetId: form.vendorPresetId,
          protocol: form.storedProtocol,
          baseUrl: form.baseUrl,
          apiKey: form.apiKey,
        },
      });
      modelsManualOnly.value = Boolean(res.manualOnly);
      modelOptions.value = (res.models ?? []).map((m) => ({ value: m.id, label: m.id }));
      if (res.error) modelsError.value = res.error;
      if (modelOptions.value.length && !form.modelCode.trim()) {
        form.modelCode = modelOptions.value[0]!.value;
      }
    } catch {
      toast.error(t("serverCatalog.llm.modelsFetchFailed"));
    } finally {
      modelsLoading.value = false;
    }
  }

  async function submit() {
    if (!form.name.trim() || !form.modelCode.trim()) {
      toast.warning(t("validation.fillNameAndModel"));
      return;
    }
    if ((isCustomOpenAi.value || isAzurePreset.value) && !form.baseUrl.trim()) {
      toast.warning(t("serverCatalog.llm.baseUrlRequired"));
      return;
    }
    if (!form.apiKey.trim()) {
      toast.warning(t("serverCatalog.llm.apiKeyRequired"));
      return;
    }
    syncStoredFromPreset();
    syncConfigFromSchema();
    let configStr = (form.config ?? "").trim();
    try {
      JSON.parse(configStr);
    } catch {
      toast.error(t("validation.invalidJson", { field: t("fields.extJson") }));
      return;
    }
    const protocol = form.storedProtocol;
    const body = {
      code: editorMode.value === "create" ? autoLlmCode(protocol) : form.code.trim(),
      name: form.name.trim(),
      protocol,
      baseUrl: form.baseUrl.trim() || null,
      apiKey: form.apiKey.trim() || null,
      modelCode: form.modelCode.trim(),
      config: configStr || "{}",
      status: form.status,
      sortOrder: form.sortOrder,
      remark: form.remark.trim() || null,
    };
    saving.value = true;
    try {
      if (editorMode.value === "create") {
        await api.create(body);
        toast.success(t("toast.created"));
      } else {
        await api.update(form.id, { id: form.id, ...body });
        toast.success(t("toast.saved"));
      }
      drawerOpen.value = false;
      await load();
    } catch (e) {
      toast.error(t("toast.saveFailed"));
      console.error(e);
    } finally {
      saving.value = false;
    }
  }

  async function removeCurrent() {
    if (!form.id) return;
    const ok = await confirm({
      message: t("confirm.deleteLlmConfig"),
      danger: true,
      confirmLabel: t("common.delete"),
    });
    if (!ok) return;
    deleting.value = true;
    try {
      await api.remove(form.id);
      toast.success(t("toast.deleted"));
      drawerOpen.value = false;
      await load();
    } catch (e) {
      toast.error(t("toast.deleteFailed"));
    } finally {
      deleting.value = false;
    }
  }

  async function runProbe() {
    if (!form.id) {
      toast.warning(t("serviceConfig.probeNeedSave"));
      return;
    }
    syncStoredFromPreset();
    syncConfigFromSchema();
    try {
      const result = await probe(form.id, {
        protocol: form.storedProtocol,
        baseUrl: form.baseUrl,
        apiKey: form.apiKey,
        modelCode: form.modelCode,
        config: form.config,
      });
      if (result.ok) toast.success(t("serviceConfig.probeOk"));
      else toast.error(result.message);
    } catch {
      toast.error(t("serviceConfig.probeFailed"));
    }
  }

  const { activateRow, activatingId } = useActivateConfigRow({
    api,
    getList: () => list.value,
    exclusivity: "multi-active",
    reload: load,
  });

  async function toggleActive(row: Record<string, unknown>) {
    if (String(row.status) !== "active") {
      await activateRow(row);
    } else {
      await api.update(String(row.id), { id: String(row.id), status: "inactive" });
      toast.success(t("agentHub.disabled"));
      await load();
    }
  }

  return {
    t,
    list,
    loading,
    total,
    page,
    pageSize,
    activeCount,
    drawerOpen,
    setupPhase,
    setupStepNumber,
    editorMode,
    showAdvanced,
    form,
    schemaFields,
    configSchema,
    dirty,
    saving,
    deleting,
    probing,
    lastResult,
    activatingId,
    activePreset,
    isCustomOpenAi,
    isAzurePreset,
    baseUrlEditable,
    baseUrlReadonly,
    vendorGridItems,
    modelOptions,
    modelsLoading,
    modelsError,
    modelsManualOnly,
    drawerTitle,
    drawerSubtitle,
    serviceUsageTodayLine,
    vendorIdForRow,
    vendorLabelForRow,
    openCreate,
    openEdit,
    closeDrawer,
    goToConnect,
    backToVendor,
    fetchModels,
    submit,
    removeCurrent,
    runProbe,
    toggleActive,
    syncConfigFromSchema,
  };
}
