import { getServicePreset, resolvePresetId } from "~/utils/serviceVendorPresets";

const CONFIG_HINTS: Record<string, string> = {
  openai: '{"system_prompt_template":"... %s ...","max_chars":12000}',
  azure_translator: '{"region":"eastasia","api_version":"3.0"}',
  deepl: '{"use_free_api":true}',
  google_translate: "{}",
  aws_translate: '{"region":"us-east-1"}',
  ibm_watson_translate: '{"region":"us-south","api_version":"2018-05-01"}',
  papago_translate: "{}",
  libretranslate: "{}",
  baidu_translate: "{}",
  youdao_translate: "{}",
  tencent_translate: '{"region":"ap-guangzhou"}',
  aliyun_translate: '{"region":"cn-hangzhou"}',
  xunfei_translate: '{"app_id":"your_app_id"}',
  volcengine_translate: '{"region":"cn-north-1"}',
};

const OPENAI_COMPAT = new Set([
  "openai",
  "azure_openai",
  "ollama",
  "deepseek",
  "openrouter",
  "groq",
  "together",
  "zhipu",
  "moonshot",
  "siliconflow",
  "nvidia_nim",
  "mistral",
]);

function isOpenAICompatible(protocol: string): boolean {
  return OPENAI_COMPAT.has(protocol.trim().toLowerCase());
}

function needsApiSecret(protocol: string): boolean {
  return [
    "tencent_translate",
    "baidu_translate",
    "aliyun_translate",
    "aws_translate",
    "youdao_translate",
    "papago_translate",
    "xunfei_translate",
    "volcengine_translate",
  ].includes(protocol);
}

function needsBaseUrl(protocol: string): boolean {
  return ["azure_translator", "deepl", "libretranslate", "ibm_watson_translate"].includes(protocol);
}

export function useTranslateConfigEditor() {
  const hub = useAgentHubList("/api/admin/translate-service-configs");
  const { confirm } = useConfirm();
  const { serviceUsageMonthLine, serviceUsageTodayLine } = useUsageDisplay();
  const llmApi = useAdminResourceApi("/api/admin/llm-service-configs");

  const llmConfigs = ref<Record<string, unknown>[]>([]);
  const useLlmLink = ref(false);

  const form = reactive({
    id: "",
    code: "",
    name: "",
    vendorPresetId: "openai",
    protocol: "openai",
    baseUrl: "",
    apiKey: "",
    apiSecret: "",
    modelCode: "gpt-4o-mini",
    llmConfigId: "",
    config: CONFIG_HINTS.openai ?? "{}",
    status: "inactive",
    sortOrder: 0,
    remark: "",
  });

  const vendorGridItems = vendorGridFor("translate", hub.t);
  const activePreset = computed(() => getServicePreset("translate", form.vendorPresetId));
  const isOpenAi = computed(() => form.protocol === "openai");

  const llmConfigOptions = computed(() =>
    llmConfigs.value.map((r) => ({
      value: String(r.id),
      label: `${r.name} · ${r.protocol} · ${r.modelCode}`,
    })),
  );

  const drawerTitle = computed(() =>
    hub.editorMode.value === "create"
      ? hub.setupPhase.value === "vendor"
        ? hub.t("agentHub.addTranslate")
        : hub.t("agentHub.connectTranslate")
      : form.name || hub.t("agentHub.editTranslate"),
  );
  const drawerSubtitle = computed(() => {
    if (hub.setupPhase.value === "vendor") return hub.t("agentHub.pickVendorSubtitle");
    const p = activePreset.value;
    return p ? hub.t(p.labelKey) : "";
  });

  function apiKeyLabel(protocol: string): string {
    switch (protocol) {
      case "tencent_translate":
        return "SecretId";
      case "baidu_translate":
        return "App ID";
      case "youdao_translate":
        return "App Key";
      case "aliyun_translate":
      case "aws_translate":
      case "volcengine_translate":
        return hub.t("fields.accessKeyId");
      case "papago_translate":
        return "Client ID";
      case "google_translate":
        return hub.t("fields.apiKey");
      case "deepl":
        return "Auth Key";
      case "azure_translator":
        return hub.t("pages.translateConfigs.subscriptionKey");
      case "ibm_watson_translate":
        return "IBM API Key";
      case "libretranslate":
        return `${hub.t("fields.apiKey")} (${hub.t("common.optional")})`;
      case "xunfei_translate":
        return "API Key";
      default:
        return hub.t("fields.apiKey");
    }
  }

  function apiSecretLabel(protocol: string): string {
    switch (protocol) {
      case "tencent_translate":
        return "SecretKey";
      case "baidu_translate":
        return hub.t("pages.translateConfigs.secretKey");
      case "youdao_translate":
        return "App Secret";
      case "aliyun_translate":
      case "aws_translate":
      case "volcengine_translate":
        return hub.t("fields.secretAccessKey");
      case "papago_translate":
        return "Client Secret";
      case "xunfei_translate":
        return "API Secret";
      default:
        return "API Secret";
    }
  }

  async function loadLlmConfigs() {
    try {
      const res = await llmApi.list({ page: 1, pageSize: 200 });
      llmConfigs.value = res.items.filter(
        (r) => String(r.status) === "active" && isOpenAICompatible(String(r.protocol ?? "openai")),
      );
    } catch (e) {
      console.error(e);
    }
  }

  function serializeFormSnapshot() {
    return JSON.stringify({ ...form, useLlmLink: useLlmLink.value });
  }

  function syncStoredFromPreset() {
    const p = activePreset.value;
    if (p) form.protocol = p.storedValue;
  }

  function applyPreset(keepApiKey = false) {
    const p = activePreset.value;
    if (!p) return;
    form.vendorPresetId = p.id;
    form.protocol = p.storedValue;
    if (!keepApiKey && (hub.editorMode.value === "create" || !form.name.trim())) {
      form.name = hub.t(p.labelKey);
    }
    if (hub.editorMode.value === "create") {
      form.config = CONFIG_HINTS[p.storedValue] ?? "{}";
      if (p.baseUrl) form.baseUrl = p.baseUrl;
    }
    if (p.storedValue === "openai") {
      if (!useLlmLink.value && (hub.editorMode.value === "create" || !form.modelCode.trim())) {
        form.modelCode = "gpt-4o-mini";
      }
    } else {
      useLlmLink.value = false;
      form.llmConfigId = "";
      if (hub.editorMode.value === "create" || !form.modelCode.trim() || form.modelCode === "gpt-4o-mini") {
        form.modelCode = "-";
      }
    }
  }

  function resetForm() {
    form.id = "";
    form.code = "";
    form.vendorPresetId = "openai";
    applyPreset();
    form.baseUrl = "";
    form.apiKey = "";
    form.apiSecret = "";
    form.llmConfigId = "";
    form.status = "inactive";
    form.sortOrder = 0;
    form.remark = "";
    useLlmLink.value = false;
    hub.snapshot.value = serializeFormSnapshot();
  }

  function loadFormFromRow(row: Record<string, unknown>) {
    const stored = String(row.protocol ?? "openai");
    form.id = String(row.id ?? "");
    form.code = String(row.code ?? "");
    form.name = String(row.name ?? "");
    form.vendorPresetId = resolvePresetId("translate", stored);
    syncStoredFromPreset();
    form.baseUrl = String(row.baseUrl ?? "");
    form.apiKey = String(row.apiKey ?? "");
    form.apiSecret = String(row.apiSecret ?? "");
    form.modelCode = String(row.modelCode ?? "");
    form.llmConfigId = String(row.llmConfigId ?? "");
    form.config = row.config != null ? String(row.config) : "{}";
    form.status = String(row.status ?? "inactive");
    form.sortOrder = Number(row.sortOrder ?? 0);
    form.remark = String(row.remark ?? "");
    useLlmLink.value = form.protocol === "openai" && form.llmConfigId.trim().length > 0;
    hub.snapshot.value = serializeFormSnapshot();
  }

  function llmLinkDisplay(id: unknown): string {
    const s = String(id ?? "").trim();
    if (!s) return "";
    const row = llmConfigs.value.find((r) => String(r.id) === s);
    return row ? String(row.name ?? row.code) : `LLM ${s.slice(0, 8)}…`;
  }

  const dirty = computed(() => serializeFormSnapshot() !== hub.snapshot.value);

  const tiles = computed(() =>
    hub.list.value.map((row) => {
      const pid = resolvePresetId("translate", String(row.protocol ?? ""));
      const modelOrLink = row.llmConfigId
        ? hub.t("pages.translateConfigs.linkedPrefix", { label: llmLinkDisplay(row.llmConfigId) })
        : String(row.modelCode ?? "");
      return {
        vendorId: pid,
        title: String(row.name ?? ""),
        subtitle: presetLabel("translate", pid, hub.t),
        model: modelOrLink,
        active: String(row.status) === "active",
        meta: serviceUsageTodayLine(row.usage as Record<string, unknown> | undefined),
        row,
      };
    }),
  );

  watch(
    () => form.vendorPresetId,
    () => applyPreset(true),
  );

  watch(useLlmLink, (linked) => {
    if (!linked) {
      form.llmConfigId = "";
      if (form.protocol === "openai" && !form.modelCode.trim()) form.modelCode = "gpt-4o-mini";
    } else {
      form.baseUrl = "";
      form.apiKey = "";
      if (!form.llmConfigId && llmConfigs.value.length === 1) {
        form.llmConfigId = String(llmConfigs.value[0]?.id ?? "");
      }
    }
  });

  onMounted(() => void loadLlmConfigs());

  watch(
    () => hub.loading.value,
    (v) => {
      if (!v) void loadLlmConfigs();
    },
  );

  function modelRequired(): boolean {
    return form.protocol === "openai" && !useLlmLink.value;
  }

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
    if (isOpenAi.value && useLlmLink.value) {
      if (!form.llmConfigId.trim()) {
        hub.toast.warning(hub.t("validation.selectLlmConfig"));
        return;
      }
    } else if (modelRequired() && !form.modelCode.trim()) {
      hub.toast.warning(hub.t("validation.fillModelOrLinkLlm"));
      return;
    }
    if (!modelRequired() && !form.modelCode.trim()) form.modelCode = "-";
    syncStoredFromPreset();
    let configStr = (form.config ?? "").trim();
    try {
      JSON.parse(configStr || "{}");
    } catch {
      hub.toast.error(hub.t("validation.invalidJson", { field: hub.t("fields.extJson") }));
      return;
    }
    const linked = isOpenAi.value && useLlmLink.value;
    const body = {
      code: hub.editorMode.value === "create" ? hub.autoCode(form.protocol) : form.code.trim(),
      name: form.name.trim(),
      protocol: form.protocol,
      baseUrl: linked ? null : form.baseUrl.trim() || null,
      apiKey: linked ? null : form.apiKey.trim() || null,
      apiSecret: needsApiSecret(form.protocol) ? form.apiSecret.trim() || null : null,
      modelCode: form.modelCode.trim(),
      llmConfigId: linked ? form.llmConfigId.trim() : null,
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
      message: hub.t("confirm.deleteTranslateConfig"),
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

  return {
    ...hub,
    form,
    dirty,
    tiles,
    vendorGridItems,
    activePreset,
    isOpenAi,
    useLlmLink,
    llmConfigOptions,
    llmConfigs,
    drawerTitle,
    drawerSubtitle,
    apiKeyLabel,
    apiSecretLabel,
    needsApiSecret,
    needsBaseUrl,
    modelRequired,
    openCreate,
    openEdit,
    goToConnect,
    submit,
    removeCurrent,
    configHints: CONFIG_HINTS,
  };
}
