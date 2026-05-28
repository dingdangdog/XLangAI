import { getServicePreset, resolvePresetId } from "~/utils/serviceVendorPresets";

const CONFIG_HINTS: Record<string, string> = {
  local: "{}",
  cloudflare_r2: '{"path_style":true}',
  qiniu: '{"zone":"z0"}',
  aliyun_oss: "{}",
};

export function useObjectStorageConfigEditor() {
  const hub = useAgentHubList("/api/admin/object-storage-configs");
  const { confirm } = useConfirm();

  const form = reactive({
    id: "",
    code: "",
    name: "",
    vendorPresetId: "local",
    provider: "local",
    baseUrl: "",
    publicBaseUrl: "",
    apiKey: "",
    secretKey: "",
    bucket: "",
    region: "",
    config: "{}",
    status: "inactive",
    sortOrder: 0,
    remark: "",
  });

  const vendorGridItems = vendorGridFor("storage", hub.t);
  const activePreset = computed(() => getServicePreset("storage", form.vendorPresetId));
  const isR2 = computed(() => form.provider === "cloudflare_r2");
  const isLocal = computed(() => form.provider === "local");

  const drawerTitle = computed(() =>
    hub.editorMode.value === "create"
      ? hub.setupPhase.value === "vendor"
        ? hub.t("agentHub.addStorage")
        : hub.t("agentHub.connectStorage")
      : form.name || hub.t("agentHub.editStorage"),
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
    if (hub.editorMode.value === "create") {
      form.config = CONFIG_HINTS[p.storedValue] ?? "{}";
    }
  }

  function resetForm() {
    form.id = "";
    form.code = "";
    form.vendorPresetId = "local";
    applyPreset();
    form.baseUrl = "";
    form.publicBaseUrl = "";
    form.apiKey = "";
    form.secretKey = "";
    form.bucket = "";
    form.region = "";
    form.status = "inactive";
    form.sortOrder = 0;
    form.remark = "";
    hub.snapshot.value = serializeFormSnapshot();
  }

  function loadFormFromRow(row: Record<string, unknown>) {
    const stored = String(row.provider ?? "local");
    form.id = String(row.id ?? "");
    form.code = String(row.code ?? "");
    form.name = String(row.name ?? "");
    form.vendorPresetId = resolvePresetId("storage", stored);
    syncStoredFromPreset();
    form.baseUrl = String(row.baseUrl ?? "");
    form.publicBaseUrl = String(row.publicBaseUrl ?? "");
    form.apiKey = String(row.apiKey ?? "");
    form.secretKey = String(row.secretKey ?? "");
    form.bucket = String(row.bucket ?? "");
    form.region = String(row.region ?? "");
    form.config = row.config != null ? String(row.config) : "{}";
    form.status = String(row.status ?? "inactive");
    form.sortOrder = Number(row.sortOrder ?? 0);
    form.remark = String(row.remark ?? "");
    hub.snapshot.value = serializeFormSnapshot();
  }

  const dirty = computed(() => serializeFormSnapshot() !== hub.snapshot.value);

  const tiles = computed(() =>
    hub.list.value.map((row) => ({
      vendorId: resolvePresetId("storage", String(row.provider ?? "")),
      title: String(row.name ?? ""),
      subtitle: presetLabel("storage", resolvePresetId("storage", String(row.provider ?? "")), hub.t),
      model: String(row.bucket ?? ""),
      active: String(row.status) === "active",
      meta: String(row.publicBaseUrl ?? ""),
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
    syncStoredFromPreset();
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
      publicBaseUrl: form.publicBaseUrl.trim() || null,
      apiKey: form.apiKey.trim() || null,
      secretKey: form.secretKey.trim() || null,
      bucket: form.bucket.trim() || null,
      region: form.region.trim() || null,
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
      message: hub.t("confirm.deleteObjectStorage"),
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
    isR2,
    isLocal,
    drawerTitle,
    drawerSubtitle,
    openCreate,
    openEdit,
    goToConnect,
    submit,
    removeCurrent,
    configHints: CONFIG_HINTS,
  };
}
