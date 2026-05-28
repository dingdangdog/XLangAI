import {
  configObjectFromSchemaFields,
  getConfigSchema,
  parseConfigObject,
  schemaFieldsFromConfigObject,
  stringifyConfigObject,
} from "~/utils/serviceConfigSchemas";
import { getServicePreset, resolvePresetId } from "~/utils/serviceVendorPresets";

export function useSmsConfigEditor() {
  const hub = useAgentHubList("/api/admin/sms-service-configs");
  const { confirm } = useConfirm();
  const { probe, probing, lastResult, clearProbe } = useServiceConfigProbe("/api/admin/sms-service-configs");

  const form = reactive({
    id: "",
    code: "",
    name: "",
    vendorPresetId: "aliyun",
    provider: "aliyun",
    apiKey: "",
    secretKey: "",
    region: "cn-hangzhou",
    signName: "",
    templateCode: "",
    config: "{}",
    status: "inactive",
    sortOrder: 0,
    remark: "",
  });

  const schemaFields = ref<Record<string, string | number | boolean>>({});

  const vendorGridItems = vendorGridFor("sms", hub.t);
  const activePreset = computed(() => getServicePreset("sms", form.vendorPresetId));
  const configSchema = computed(() => getConfigSchema("sms", form.provider));
  const isTencent = computed(() => form.provider === "tencent");

  const drawerTitle = computed(() =>
    hub.editorMode.value === "create"
      ? hub.setupPhase.value === "vendor"
        ? hub.t("agentHub.addSms")
        : hub.t("agentHub.connectSms")
      : form.name || hub.t("agentHub.editSms"),
  );
  const drawerSubtitle = computed(() => {
    if (hub.setupPhase.value === "vendor") return hub.t("agentHub.pickVendorSubtitle");
    const p = activePreset.value;
    return p ? hub.t(p.labelKey) : "";
  });

  function serializeFormSnapshot() {
    return JSON.stringify({ ...form, schemaFields: schemaFields.value });
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
    if (p.region && (hub.editorMode.value === "create" || !form.region.trim())) {
      form.region = p.region;
    }
    syncSchemaFromConfig();
  }

  function resetForm() {
    form.id = "";
    form.code = "";
    form.vendorPresetId = "aliyun";
    applyPreset();
    form.apiKey = "";
    form.secretKey = "";
    form.signName = "";
    form.templateCode = "";
    form.config = "{}";
    form.status = "inactive";
    form.sortOrder = 0;
    form.remark = "";
    hub.snapshot.value = serializeFormSnapshot();
    clearProbe();
  }

  function loadFormFromRow(row: Record<string, unknown>) {
    const stored = String(row.provider ?? "aliyun");
    form.id = String(row.id ?? "");
    form.code = String(row.code ?? "");
    form.name = String(row.name ?? "");
    form.vendorPresetId = resolvePresetId("sms", stored);
    syncStoredFromPreset();
    form.apiKey = String(row.apiKey ?? "");
    form.secretKey = String(row.secretKey ?? "");
    form.region = String(row.region ?? "");
    form.signName = String(row.signName ?? "");
    form.templateCode = String(row.templateCode ?? "");
    form.config = row.config != null ? String(row.config) : "{}";
    form.status = String(row.status ?? "inactive");
    form.sortOrder = Number(row.sortOrder ?? 0);
    form.remark = String(row.remark ?? "");
    syncSchemaFromConfig();
    hub.snapshot.value = serializeFormSnapshot();
    clearProbe();
  }

  const dirty = computed(() => serializeFormSnapshot() !== hub.snapshot.value);

  const tiles = computed(() =>
    hub.list.value.map((row) => ({
      vendorId: resolvePresetId("sms", String(row.provider ?? "")),
      title: String(row.name ?? ""),
      subtitle: presetLabel("sms", resolvePresetId("sms", String(row.provider ?? "")), hub.t),
      model: `${String(row.signName ?? "")} · ${String(row.templateCode ?? "")}`,
      active: String(row.status) === "active",
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
    if (!form.signName.trim() || !form.templateCode.trim()) {
      hub.toast.warning(hub.t("pages.smsService.fillSignAndTemplate"));
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
      apiKey: form.apiKey.trim() || null,
      secretKey: form.secretKey.trim() || null,
      region: form.region.trim() || null,
      signName: form.signName.trim() || null,
      templateCode: form.templateCode.trim() || null,
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
      message: hub.t("confirm.deleteSmsConfig"),
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
        apiKey: form.apiKey,
        secretKey: form.secretKey,
        region: form.region,
        signName: form.signName,
        templateCode: form.templateCode,
        config: form.config,
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
    isTencent,
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
