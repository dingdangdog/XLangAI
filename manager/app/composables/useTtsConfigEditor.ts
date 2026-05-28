import {
  configObjectFromSchemaFields,
  getConfigSchema,
  parseConfigObject,
  schemaFieldsFromConfigObject,
  stringifyConfigObject,
} from "~/utils/serviceConfigSchemas";
import { TTS_DEFAULT_MODEL } from "~/utils/serverServiceCatalog";
import { getServicePreset, resolvePresetId } from "~/utils/serviceVendorPresets";

export function useTtsConfigEditor() {
  const hub = useAgentHubList("/api/admin/tts-service-configs");
  const { confirm } = useConfirm();
  const { serviceUsageMonthLine } = useUsageDisplay();
  const { probe, probing, lastResult, clearProbe } = useServiceConfigProbe("/api/admin/tts-service-configs");

  const form = reactive({
    id: "",
    code: "",
    name: "",
    vendorPresetId: "openai_rest",
    provider: "openai_rest",
    baseUrl: "",
    apiKey: "",
    region: "",
    modelCode: "tts-1",
    config: "{}",
    status: "active",
    sortOrder: 0,
    remark: "",
  });

  const probeVoiceCode = ref("alloy");
  const schemaFields = ref<Record<string, string | number | boolean>>({});

  const vendorGridItems = vendorGridFor("tts", hub.t);
  const activePreset = computed(() => getServicePreset("tts", form.vendorPresetId));
  const configSchema = computed(() => getConfigSchema("tts", form.provider));

  const regionRequired = computed(() =>
    ["azure_speech_rest", "aliyun_nls", "tencent_tts", "aws_polly"].includes(form.provider),
  );

  const drawerTitle = computed(() =>
    hub.editorMode.value === "create"
      ? hub.setupPhase.value === "vendor"
        ? hub.t("agentHub.addTts")
        : hub.t("agentHub.connectTts")
      : form.name || hub.t("agentHub.editTts"),
  );
  const drawerSubtitle = computed(() => {
    if (hub.setupPhase.value === "vendor") return hub.t("agentHub.pickVendorSubtitle");
    const p = activePreset.value;
    return p ? hub.t(p.labelKey) : "";
  });

  function serializeFormSnapshot() {
    return JSON.stringify({ ...form, schemaFields: schemaFields.value, probeVoiceCode: probeVoiceCode.value });
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
    const p = activePreset.value;
    if (p) form.provider = p.storedValue;
  }

  function applyPreset(keepApiKey = false) {
    const p = activePreset.value;
    if (!p) return;
    form.vendorPresetId = p.id;
    form.provider = p.storedValue;
    if (!keepApiKey && (hub.editorMode.value === "create" || !form.name.trim())) {
      form.name = hub.t(p.labelKey);
    }
    const def = TTS_DEFAULT_MODEL[p.storedValue] ?? p.defaultModel ?? "-";
    if (hub.editorMode.value === "create" || !form.modelCode.trim()) {
      form.modelCode = def;
    }
    if (p.region && (hub.editorMode.value === "create" || !form.region.trim())) {
      form.region = p.region;
    }
    probeVoiceCode.value = p.storedValue === "azure_speech_rest" ? "en-US-JennyNeural" : "alloy";
    syncSchemaFromConfig();
  }

  function resetForm() {
    form.id = "";
    form.code = "";
    form.vendorPresetId = "openai_rest";
    applyPreset();
    form.baseUrl = "";
    form.apiKey = "";
    form.config = "{}";
    form.status = "active";
    form.sortOrder = 0;
    form.remark = "";
    hub.snapshot.value = serializeFormSnapshot();
    clearProbe();
  }

  function loadFormFromRow(row: Record<string, unknown>) {
    const stored = String(row.provider ?? "openai_rest");
    form.id = String(row.id ?? "");
    form.code = String(row.code ?? "");
    form.name = String(row.name ?? "");
    form.vendorPresetId = resolvePresetId("tts", stored);
    syncStoredFromPreset();
    form.baseUrl = String(row.baseUrl ?? "");
    form.apiKey = String(row.apiKey ?? "");
    form.region = String(row.region ?? "");
    form.modelCode = String(row.modelCode ?? "-");
    form.config = row.config != null ? String(row.config) : "{}";
    form.status = String(row.status ?? "active");
    form.sortOrder = Number(row.sortOrder ?? 0);
    form.remark = String(row.remark ?? "");
    probeVoiceCode.value = form.provider === "azure_speech_rest" ? "en-US-JennyNeural" : "alloy";
    syncSchemaFromConfig();
    hub.snapshot.value = serializeFormSnapshot();
    clearProbe();
  }

  const dirty = computed(() => serializeFormSnapshot() !== hub.snapshot.value);

  const tiles = computed(() =>
    hub.list.value.map((row) => ({
      vendorId: resolvePresetId("tts", String(row.provider ?? "")),
      title: String(row.name ?? ""),
      subtitle: presetLabel("tts", resolvePresetId("tts", String(row.provider ?? "")), hub.t),
      model: String(row.modelCode ?? ""),
      active: String(row.status) === "active",
      meta: serviceUsageMonthLine(row.usage as Record<string, unknown> | undefined),
      row,
    })),
  );

  watch(
    () => form.vendorPresetId,
    () => applyPreset(true),
  );

  function openCreate() {
    hub.openCreate();
    resetForm();
  }

  function openEdit(row: Record<string, unknown>) {
    hub.openEdit(row);
    loadFormFromRow(row);
  }

  function goToConnect() {
    applyPreset(true);
    hub.goToConnect();
  }

  async function submit() {
    if (!form.name.trim()) {
      hub.toast.warning(hub.t("validation.fillName"));
      return;
    }
    if (regionRequired.value && !form.region.trim()) {
      hub.toast.warning(hub.t("validation.regionRequired"));
      return;
    }
    syncStoredFromPreset();
    syncConfigFromSchema();
    let configStr = (form.config ?? "").trim();
    try {
      JSON.parse(configStr || "{}");
    } catch {
      hub.toast.error(hub.t("validation.invalidJson", { field: hub.t("fields.extJson") }));
      return;
    }
    const body = {
      code: hub.editorMode.value === "create" ? hub.autoCode(form.provider) : form.code.trim(),
      name: form.name.trim(),
      provider: form.provider,
      baseUrl: form.baseUrl.trim() || null,
      apiKey: form.apiKey.trim() || null,
      region: form.region.trim() || null,
      modelCode: form.modelCode.trim() || "-",
      config: configStr || "{}",
      status: form.status,
      sortOrder: form.sortOrder,
      remark: form.remark.trim() || null,
    };
    hub.saving.value = true;
    try {
      if (hub.editorMode.value === "create") {
        await hub.api.create(body);
        hub.toast.success(hub.t("toast.created"));
      } else {
        await hub.api.update(form.id, { id: form.id, ...body });
        hub.toast.success(hub.t("toast.saved"));
      }
      hub.closeDrawer();
      await hub.load();
    } catch (e) {
      hub.toast.error(hub.t("toast.saveFailed"));
      console.error(e);
    } finally {
      hub.saving.value = false;
    }
  }

  async function removeCurrent() {
    if (!form.id) return;
    const ok = await confirm({
      message: hub.t("confirm.deleteTtsConfig"),
      danger: true,
      confirmLabel: hub.t("common.delete"),
    });
    if (!ok) return;
    hub.deleting.value = true;
    try {
      await hub.api.remove(form.id);
      hub.toast.success(hub.t("toast.deleted"));
      hub.closeDrawer();
      await hub.load();
    } catch {
      hub.toast.error(hub.t("toast.deleteFailed"));
    } finally {
      hub.deleting.value = false;
    }
  }

  async function runProbe() {
    if (!form.id) {
      hub.toast.warning(hub.t("serviceConfig.probeNeedSave"));
      return;
    }
    syncConfigFromSchema();
    try {
      const result = await probe(form.id, {
        provider: form.provider,
        baseUrl: form.baseUrl,
        apiKey: form.apiKey,
        region: form.region,
        modelCode: form.modelCode,
        config: form.config,
        voiceCode: probeVoiceCode.value,
      });
      if (result.ok) hub.toast.success(hub.t("serviceConfig.probeOk"));
      else hub.toast.error(result.message);
    } catch {
      hub.toast.error(hub.t("serviceConfig.probeFailed"));
    }
  }

  return {
    ...hub,
    form,
    dirty,
    tiles,
    vendorGridItems,
    activePreset,
    configSchema,
    schemaFields,
    regionRequired,
    probeVoiceCode,
    probing,
    lastResult,
    drawerTitle,
    drawerSubtitle,
    openCreate,
    openEdit,
    goToConnect,
    submit,
    removeCurrent,
    runProbe,
    syncConfigFromSchema,
  };
}
