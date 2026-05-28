import {
  getServicePreset,
  presetGridItems,
  resolvePresetId,
  type ServicePresetKind,
} from "~/utils/serviceVendorPresets";

/** 各 Agent Hub 面板共用的列表 / 抽屉状态 */
export function useAgentHubList(apiPath: string, pageSizeDefault = 24) {
  const { t } = useI18n();
  const api = useAdminResourceApi(apiPath);
  const toast = useToast();

  const page = ref(1);
  const pageSize = ref(pageSizeDefault);
  const total = ref(0);
  const list = ref<Record<string, unknown>[]>([]);
  const loading = ref(false);

  const drawerOpen = ref(false);
  const setupPhase = ref<"vendor" | "connect">("vendor");
  const editorMode = ref<"create" | "edit">("create");
  const saving = ref(false);
  const deleting = ref(false);
  const snapshot = ref("");

  const setupStepNumber = computed(() => (setupPhase.value === "vendor" ? 1 : 2));
  const activeCount = computed(() => list.value.filter((r) => String(r.status) === "active").length);

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

  function openCreate() {
    editorMode.value = "create";
    setupPhase.value = "vendor";
    drawerOpen.value = true;
  }

  function openEdit(row: Record<string, unknown>) {
    editorMode.value = "edit";
    setupPhase.value = "connect";
    drawerOpen.value = true;
  }

  function closeDrawer() {
    drawerOpen.value = false;
  }

  function goToConnect() {
    setupPhase.value = "connect";
  }

  function backToVendor() {
    if (editorMode.value === "create") setupPhase.value = "vendor";
  }

  function autoCode(prefix: string) {
    return `${prefix.trim().replace(/[^a-zA-Z0-9_-]/g, "_")}-${Date.now()}`;
  }

  return {
    t,
    api,
    toast,
    page,
    pageSize,
    total,
    list,
    loading,
    drawerOpen,
    setupPhase,
    setupStepNumber,
    editorMode,
    saving,
    deleting,
    snapshot,
    activeCount,
    load,
    openCreate,
    openEdit,
    closeDrawer,
    goToConnect,
    backToVendor,
    autoCode,
  };
}

export function vendorGridFor(kind: ServicePresetKind, t: (key: string) => string) {
  return computed(() => presetGridItems(kind, t));
}

export function presetLabel(
  kind: ServicePresetKind,
  presetId: string,
  t: (key: string) => string,
): string {
  const p = getServicePreset(kind, presetId);
  return p ? t(p.labelKey) : presetId;
}

export function resolveVendorId(kind: ServicePresetKind, storedValue: string): string {
  return resolvePresetId(kind, storedValue);
}
