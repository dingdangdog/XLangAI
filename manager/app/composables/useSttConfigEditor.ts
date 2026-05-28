import {
  configObjectFromSchemaFields,
  getConfigSchema,
  parseConfigObject,
  schemaFieldsFromConfigObject,
  stringifyConfigObject,
} from "~/utils/serviceConfigSchemas";
import { getServicePreset, resolvePresetId } from "~/utils/serviceVendorPresets";

const NVIDIA_INTEGRATE_HOST = "integrate.api.nvidia.com";

export function useSttConfigEditor() {
  const hub = useAgentHubList("/api/admin/stt-service-configs");
  const { confirm } = useConfirm();
  const { serviceUsageTodayLine } = useUsageDisplay();

  const form = reactive({
    id: "",
    code: "",
    name: "",
    vendorPresetId: "openai",
    protocol: "openai",
    baseUrl: "",
    apiKey: "",
    modelCode: "whisper-1",
    config: "{}",
    status: "active",
    sortOrder: 0,
    remark: "",
  });

  const vendorGridItems = vendorGridFor("stt", hub.t);
  const activePreset = computed(() => getServicePreset("stt", form.vendorPresetId));
  const isOpenAi = computed(() => form.protocol === "openai");
  const isAzure = computed(() => form.protocol === "azure_speech_rest");

  const drawerTitle = computed(() =>
    hub.editorMode.value === "create"
      ? hub.setupPhase.value === "vendor"
        ? hub.t("agentHub.addStt")
        : hub.t("agentHub.connectStt")
      : form.name || hub.t("agentHub.editStt"),
  );
  const drawerSubtitle = computed(() => {
    if (hub.setupPhase.value === "vendor") return hub.t("agentHub.pickVendorSubtitle");
    const p = activePreset.value;
    return p ? hub.t(p.labelKey) : "";
  });

  function serializeFormSnapshot() {
    return JSON.stringify({ ...form });
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
    if (hub.editorMode.value === "create" || !form.modelCode.trim()) {
      form.modelCode = p.defaultModel ?? "whisper-1";
    }
  }

  function resetForm() {
    form.id = "";
    form.code = "";
    form.vendorPresetId = "openai";
    applyPreset();
    form.baseUrl = "";
    form.apiKey = "";
    form.config = "{}";
    form.status = "active";
    form.sortOrder = 0;
    form.remark = "";
    hub.snapshot.value = serializeFormSnapshot();
  }

  function loadFormFromRow(row: Record<string, unknown>) {
    const stored = String(row.protocol ?? "openai");
    form.id = String(row.id ?? "");
    form.code = String(row.code ?? "");
    form.name = String(row.name ?? "");
    form.vendorPresetId = resolvePresetId("stt", stored);
    syncStoredFromPreset();
    form.baseUrl = String(row.baseUrl ?? "");
    form.apiKey = String(row.apiKey ?? "");
    form.modelCode = String(row.modelCode ?? "");
    form.config = row.config != null ? String(row.config) : "{}";
    form.status = String(row.status ?? "active");
    form.sortOrder = Number(row.sortOrder ?? 0);
    form.remark = String(row.remark ?? "");
    hub.snapshot.value = serializeFormSnapshot();
  }

  const dirty = computed(() => serializeFormSnapshot() !== hub.snapshot.value);

  const tiles = computed(() =>
    hub.list.value.map((row) => ({
      vendorId: resolvePresetId("stt", String(row.protocol ?? "")),
      title: String(row.name ?? ""),
      subtitle: presetLabel("stt", resolvePresetId("stt", String(row.protocol ?? "")), hub.t),
      model: String(row.modelCode ?? ""),
      active: String(row.status) === "active",
      meta: serviceUsageTodayLine(row.usage as Record<string, unknown> | undefined),
      row,
    })),
  );

  watch(
    () => form.vendorPresetId,
    () => applyPreset(true),
  );

  watch(
    () => form.protocol,
    (v) => {
      if (v === "azure_speech_rest" && (!form.modelCode.trim() || form.modelCode === "whisper-1")) {
        form.modelCode = "-";
      }
      if (v === "openai" && form.modelCode === "-") form.modelCode = "whisper-1";
    },
  );

  function vendorIdForRow(row: Record<string, unknown>) {
    return resolvePresetId("stt", String(row.protocol ?? ""));
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
    if (isOpenAi.value && !form.modelCode.trim()) {
      hub.toast.warning(hub.t("validation.sttModelRequired"));
      return;
    }
    if (isAzure.value && !form.modelCode.trim()) form.modelCode = "-";
    const bu = form.baseUrl.trim().toLowerCase();
    if (isOpenAi.value && bu.includes(NVIDIA_INTEGRATE_HOST)) {
      hub.toast.error(hub.t("pages.sttConfigs.nvidiaWarning"));
      return;
    }
    syncStoredFromPreset();
    let configStr = (form.config ?? "").trim();
    try {
      JSON.parse(configStr || "{}");
    } catch {
      hub.toast.error(hub.t("validation.invalidJson", { field: hub.t("fields.extJson") }));
      return;
    }
    const body = {
      code: hub.editorMode.value === "create" ? hub.autoCode(form.protocol) : form.code.trim(),
      name: form.name.trim(),
      protocol: form.protocol,
      baseUrl: form.baseUrl.trim() || null,
      apiKey: form.apiKey.trim() || null,
      modelCode: form.modelCode.trim(),
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
      message: hub.t("confirm.deleteSttConfig"),
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
    isAzure,
    drawerTitle,
    drawerSubtitle,
    vendorIdForRow,
    openCreate,
    openEdit,
    goToConnect,
    submit,
    removeCurrent,
  };
}
