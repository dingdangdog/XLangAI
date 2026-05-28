<script setup lang="ts">
import {
  configObjectFromSchemaFields,
  getConfigSchema,
  parseConfigObject,
  schemaFieldsFromConfigObject,
  stringifyConfigObject,
} from "~/utils/serviceConfigSchemas";

const { t } = useI18n();

const API = "/api/admin/sms-service-configs";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();
const { probe, probing, lastResult, clearProbe } = useServiceConfigProbe(API);

const PROVIDERS = [
  { value: "aliyun", labelKey: "pages.smsService.providerAliyun" },
  { value: "tencent", labelKey: "pages.smsService.providerTencent" },
] as const;

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const selectedId = ref<string | null>(null);
const editorMode = ref<"idle" | "create" | "edit">("idle");
const editorTab = ref<"connection" | "advanced" | "raw">("connection");
const saving = ref(false);
const deleting = ref(false);

const form = reactive({
  id: "",
  code: "",
  name: "",
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
const snapshot = ref("");

const providerOptions = computed(() =>
  PROVIDERS.map((p) => ({ value: p.value, label: t(p.labelKey) })),
);

const configSchema = computed(() => getConfigSchema("sms", form.provider));
const isTencent = computed(() => form.provider === "tencent");

const activeCount = computed(
  () => list.value.filter((r) => String(r.status) === "active").length,
);

function autoSmsCode(provider: string) {
  const p = (provider || "aliyun").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
}

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

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.provider = "aliyun";
  form.apiKey = "";
  form.secretKey = "";
  form.region = "cn-hangzhou";
  form.signName = "";
  form.templateCode = "";
  form.config = "{}";
  form.status = "inactive";
  form.sortOrder = 0;
  form.remark = "";
  syncSchemaFromConfig();
  snapshot.value = serializeFormSnapshot();
  clearProbe();
}

function loadFormFromRow(row: Record<string, unknown>) {
  form.id = String(row.id ?? "");
  form.code = String(row.code ?? "");
  form.name = String(row.name ?? "");
  form.provider = String(row.provider ?? "aliyun");
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
  snapshot.value = serializeFormSnapshot();
  clearProbe();
}

async function load() {
  loading.value = true;
  try {
    const res = await api.list({ page: page.value, pageSize: pageSize.value });
    list.value = res.items;
    total.value = res.total;
    if (selectedId.value && !res.items.some((r) => String(r.id) === selectedId.value)) {
      selectedId.value = null;
      editorMode.value = "idle";
    }
  } catch (e) {
    toast.error(t("toast.loadFailed"));
    console.error(e);
  } finally {
    loading.value = false;
  }
}

watch([page, pageSize], () => void load(), { immediate: true });

watch(
  () => form.provider,
  (v) => {
    if (v === "tencent" && (!form.region || form.region === "cn-hangzhou")) {
      form.region = "ap-guangzhou";
    }
    if (v === "aliyun" && form.region === "ap-guangzhou") {
      form.region = "cn-hangzhou";
    }
    syncSchemaFromConfig();
  },
);

const dirty = computed(() => serializeFormSnapshot() !== snapshot.value);
const selectedRow = computed(() => list.value.find((r) => String(r.id) === selectedId.value) ?? null);

function openCreate() {
  editorMode.value = "create";
  selectedId.value = null;
  editorTab.value = "connection";
  resetForm();
}

function selectRow(row: Record<string, unknown>) {
  selectedId.value = String(row.id);
  editorMode.value = "edit";
  editorTab.value = "connection";
  loadFormFromRow(row);
}

function cancelEdit() {
  if (editorMode.value === "create") {
    editorMode.value = "idle";
    selectedId.value = null;
    resetForm();
    return;
  }
  if (selectedRow.value) loadFormFromRow(selectedRow.value);
}

async function submit() {
  if (!form.name.trim()) {
    toast.warning(t("validation.fillName"));
    return;
  }
  if (!form.signName.trim() || !form.templateCode.trim()) {
    toast.warning(t("pages.smsService.fillSignAndTemplate"));
    return;
  }
  syncConfigFromSchema();
  let configStr = (form.config ?? "").trim();
  try {
    JSON.parse(configStr);
  } catch {
    toast.error(t("validation.invalidJson", { field: t("fields.extJson") }));
    editorTab.value = "raw";
    return;
  }
  const body = {
    code: editorMode.value === "create" ? autoSmsCode(form.provider) : form.code.trim(),
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
  saving.value = true;
  try {
    if (editorMode.value === "create") {
      const created = (await api.create(body)) as Record<string, unknown>;
      toast.success(t("toast.created"));
      await load();
      if (created?.id) {
        selectedId.value = String(created.id);
        editorMode.value = "edit";
      }
    } else {
      await api.update(form.id, { id: form.id, ...body });
      toast.success(t("toast.saved"));
      await load();
      const row = list.value.find((r) => String(r.id) === form.id);
      if (row) loadFormFromRow(row);
    }
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
    message: t("confirm.deleteSmsConfig"),
    danger: true,
    confirmLabel: t("common.delete"),
  });
  if (!ok) return;
  deleting.value = true;
  try {
    await api.remove(form.id);
    toast.success(t("toast.deleted"));
    selectedId.value = null;
    editorMode.value = "idle";
    await load();
  } catch (e) {
    toast.error(t("toast.deleteFailed"));
    console.error(e);
  } finally {
    deleting.value = false;
  }
}

async function runProbe() {
  if (!form.id) {
    toast.warning(t("serviceConfig.probeNeedSave"));
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
    if (result.ok) toast.success(t("serviceConfig.probeOk"));
    else toast.error(result.message);
  } catch (e) {
    toast.error(t("serviceConfig.probeFailed"));
    console.error(e);
  }
}

const { activateRow, activatingId } = useActivateConfigRow({
  api,
  getList: () => list.value,
  exclusivity: "single-active",
  reload: load,
});

const statusOptions = [
  { value: "active", label: "active" },
  { value: "inactive", label: "inactive" },
];
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col gap-4">
    <AdminAlert v-if="activeCount > 1" :title="$t('pages.smsService.configAnomalyTitle')" variant="warning">
      {{ $t("pages.smsService.configAnomaly", { count: activeCount }) }}
    </AdminAlert>

    <AdminAlert :title="$t('pages.smsService.hintTitle')" variant="info">
      {{ $t("pages.smsService.hintBody") }}
    </AdminAlert>

    <AdminServiceConfigLayout
      :items="list"
      :selected-id="selectedId"
      :loading="loading"
      :total="total"
      v-model:page="page"
      v-model:page-size="pageSize"
      :get-title="(row) => String(row.name ?? row.code ?? '')"
      :get-subtitle="(row) => String(row.provider ?? '')"
      :get-meta="(row) => `${String(row.signName ?? '')} · ${String(row.templateCode ?? '')}`"
      :get-badge="
        (row) =>
          String(row.status) === 'active'
            ? { text: 'active', variant: 'success' }
            : { text: 'inactive', variant: 'muted' }
      "
      @create="openCreate"
      @select="selectRow"
    >
      <AdminPanel v-if="editorMode === 'idle'" :fill="false" class="flex h-full min-h-[520px] items-center justify-center">
        <div class="max-w-sm px-6 text-center">
          <p class="text-sm text-muted">{{ $t("serviceConfig.selectOrCreate") }}</p>
          <AdminButton class="mt-4" variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
        </div>
      </AdminPanel>

      <AdminConfigEditorShell
        v-else
        :title="editorMode === 'create' ? $t('pages.smsService.createDialog') : form.name"
        :subtitle="editorMode === 'edit' ? `${form.provider} · ${form.signName}` : $t('serviceConfig.smsCreateHint')"
        v-model:editor-tab="editorTab"
        :saving="saving"
        :deleting="deleting"
        :activating="activatingId === form.id"
        :can-activate="editorMode === 'edit' && form.status !== 'active'"
        :can-delete="editorMode === 'edit'"
        :dirty="dirty"
        @save="submit"
        @cancel="cancelEdit"
        @delete="removeCurrent"
        @activate="selectedRow && activateRow(selectedRow)"
      >
        <template #connection>
          <div v-if="editorMode === 'create'" class="mb-4">
            <AdminFormField :label="$t('common.provider')">
              <AdminProviderPicker v-model="form.provider" :options="providerOptions" />
            </AdminFormField>
          </div>
          <AdminFormField v-else :label="$t('common.provider')">
            <AdminSelect v-model="form.provider" :options="providerOptions" />
          </AdminFormField>
          <AdminFormField :label="$t('common.name')" required>
            <AdminInput v-model="form.name" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.smsService.accessKey')" required>
            <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.smsService.secretKey')" required>
            <AdminInput v-model="form.secretKey" type="password" autocomplete="off" />
          </AdminFormField>
          <AdminFormField :label="$t('common.region')">
            <AdminInput v-model="form.region" :placeholder="isTencent ? 'ap-guangzhou' : 'cn-hangzhou'" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.smsService.signName')" required>
            <AdminInput v-model="form.signName" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.smsService.templateCode')" required>
            <AdminInput v-model="form.templateCode" class="font-mono text-sm" />
          </AdminFormField>
          <div class="grid gap-1 sm:grid-cols-2">
            <AdminFormField :label="$t('common.status')">
              <AdminSelect v-model="form.status" :options="statusOptions" />
            </AdminFormField>
            <AdminFormField :label="$t('common.sort')">
              <AdminInput v-model="form.sortOrder" type="number" />
            </AdminFormField>
          </div>
          <AdminFormField :label="$t('common.remark')">
            <AdminInput v-model="form.remark" type="textarea" :rows="2" />
          </AdminFormField>
        </template>

        <template #advanced>
          <AdminConfigSchemaFields v-model="schemaFields" :schema="configSchema" />
          <p v-if="isTencent" class="text-xs text-muted">{{ $t("pages.smsService.tencentConfigHint") }}</p>
        </template>

        <template #raw>
          <AdminFormField :label="$t('fields.extJson')">
            <AdminInput v-model="form.config" type="textarea" :rows="10" class="font-mono text-sm" />
          </AdminFormField>
        </template>

        <template #probe>
          <AdminConfigProbePanel
            :loading="probing"
            :result="lastResult"
            :disabled="!form.id"
            :hint="$t('serviceConfig.smsProbeHint')"
            @probe="runProbe"
          />
        </template>
      </AdminConfigEditorShell>
    </AdminServiceConfigLayout>
  </div>
</template>
