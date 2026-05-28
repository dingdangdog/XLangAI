<script setup lang="ts">
import {
  configObjectFromSchemaFields,
  getConfigSchema,
  parseConfigObject,
  schemaFieldsFromConfigObject,
  stringifyConfigObject,
} from "~/utils/serviceConfigSchemas";

const { t } = useI18n();
const { serviceUsageMonthLine, serviceUsageTodayLine } = useUsageDisplay();

const API = "/api/admin/llm-service-configs";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();
const { probe, probing, lastResult, clearProbe } = useServiceConfigProbe(API);

const PROTOCOLS = [
  { value: "openai", label: "OpenAI 官方 / 兼容网关", group: "OpenAI 兼容" },
  { value: "azure_openai", label: "Azure OpenAI", group: "OpenAI 兼容" },
  { value: "ollama", label: "Ollama", group: "OpenAI 兼容" },
  { value: "deepseek", label: "DeepSeek", group: "OpenAI 兼容" },
  { value: "openrouter", label: "OpenRouter", group: "OpenAI 兼容" },
  { value: "groq", label: "Groq", group: "OpenAI 兼容" },
  { value: "together", label: "Together", group: "OpenAI 兼容" },
  { value: "zhipu", label: "智谱", group: "OpenAI 兼容" },
  { value: "moonshot", label: "Moonshot", group: "OpenAI 兼容" },
  { value: "siliconflow", label: "SiliconFlow", group: "OpenAI 兼容" },
  { value: "nvidia_nim", label: "NVIDIA NIM", group: "OpenAI 兼容" },
  { value: "claude", label: "Anthropic Claude", group: "原生协议" },
  { value: "gemini", label: "Google Gemini", group: "原生协议" },
] as const;

const DEFAULT_MODEL: Record<string, string> = {
  openai: "gpt-4o-mini",
  claude: "claude-3-5-sonnet-20241022",
  gemini: "gemini-1.5-flash",
  deepseek: "deepseek-chat",
  ollama: "llama3",
};

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
  protocol: "openai",
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

const protocolOptions = computed(() =>
  PROTOCOLS.map((p) => ({ value: p.value, label: p.label, group: p.group })),
);

const configSchema = computed(() => getConfigSchema("llm", form.protocol));

function autoLlmCode(protocol: string) {
  const p = (protocol || "openai").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
}

function isOpenAICompatible(protocol: string): boolean {
  const p = protocol.trim().toLowerCase();
  return p === "" || p === "openai" || !["claude", "gemini"].includes(p);
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
  form.protocol = "openai";
  form.baseUrl = "";
  form.apiKey = "";
  form.modelCode = DEFAULT_MODEL.openai ?? "gpt-4o-mini";
  form.config = "{}";
  form.status = "active";
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
  form.protocol = String(row.protocol ?? "openai");
  form.baseUrl = String(row.baseUrl ?? "");
  form.apiKey = String(row.apiKey ?? "");
  form.modelCode = String(row.modelCode ?? "");
  form.config = row.config != null ? String(row.config) : "{}";
  form.status = String(row.status ?? "active");
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
  () => form.protocol,
  (v) => {
    const def = DEFAULT_MODEL[v];
    if (def && (editorMode.value === "create" || !form.modelCode.trim())) {
      form.modelCode = def;
    }
    syncSchemaFromConfig();
  },
);

const dirty = computed(() => serializeFormSnapshot() !== snapshot.value);

const selectedRow = computed(() =>
  list.value.find((r) => String(r.id) === selectedId.value) ?? null,
);

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
  if (!form.name.trim() || !form.modelCode.trim()) {
    toast.warning(t("validation.fillNameAndModel"));
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
    code: editorMode.value === "create" ? autoLlmCode(form.protocol) : form.code.trim(),
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
    message: t("confirm.deleteLlmConfig"),
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
  if (!form.id && editorMode.value !== "edit") {
    toast.warning(t("serviceConfig.probeNeedSave"));
    return;
  }
  syncConfigFromSchema();
  try {
    const result = await probe(form.id, {
      protocol: form.protocol,
      baseUrl: form.baseUrl,
      apiKey: form.apiKey,
      modelCode: form.modelCode,
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
  exclusivity: "multi-active",
  reload: load,
});

const statusOptions = [
  { value: "active", label: "active" },
  { value: "inactive", label: "inactive" },
];
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col gap-4">
    <AdminAlert :title="$t('pages.llmConfigs.protocolAlertTitle')">
      <ul class="list-disc space-y-1 pl-5 text-sm">
        <li>{{ $t("pages.llmConfigs.protocolAlertOpenai") }}</li>
        <li>{{ $t("pages.llmConfigs.protocolAlertClaude") }}</li>
        <li>{{ $t("pages.llmConfigs.protocolAlertGemini") }}</li>
      </ul>
    </AdminAlert>

    <AdminServiceConfigLayout
      :items="list"
      :selected-id="selectedId"
      :loading="loading"
      :total="total"
      v-model:page="page"
      v-model:page-size="pageSize"
      :get-title="(row) => String(row.name ?? row.code ?? '')"
      :get-subtitle="(row) => String(row.protocol ?? '')"
      :get-meta="
        (row) =>
          `${String(row.modelCode ?? '')} · ${serviceUsageTodayLine(row.usage as Record<string, unknown> | undefined)}`
      "
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
        :title="editorMode === 'create' ? $t('pages.llmConfigs.createDialog') : form.name"
        :subtitle="editorMode === 'edit' ? `${form.protocol} · ${form.modelCode}` : $t('serviceConfig.createHint')"
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
            <AdminFormField :label="$t('fields.protocol')">
              <AdminProviderPicker v-model="form.protocol" :options="protocolOptions" />
            </AdminFormField>
          </div>
          <AdminFormField v-else :label="$t('fields.protocol')">
            <AdminSelect v-model="form.protocol" :options="protocolOptions.map((p) => ({ value: p.value, label: p.label }))" />
          </AdminFormField>
          <AdminFormField :label="$t('common.name')" required>
            <AdminInput v-model="form.name" />
          </AdminFormField>
          <AdminFormField
            :label="$t('fields.baseUrl')"
            :hint="
              isOpenAICompatible(form.protocol)
                ? $t('pages.llmConfigs.baseUrlHintOpenai')
                : form.protocol === 'claude'
                  ? $t('pages.llmConfigs.baseUrlHintClaude')
                  : $t('pages.llmConfigs.baseUrlHintGemini')
            "
          >
            <AdminInput v-model="form.baseUrl" :placeholder="$t('pages.llmConfigs.optionalPlaceholder')" />
          </AdminFormField>
          <AdminFormField :label="$t('fields.apiKey')" :hint="$t('pages.translateConfigs.apiKeyRequired')">
            <AdminInput v-model="form.apiKey" type="password" />
          </AdminFormField>
          <AdminFormField :label="$t('fields.modelCode')" required>
            <AdminInput v-model="form.modelCode" :placeholder="$t('pages.llmConfigs.modelPlaceholder')" />
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
        </template>

        <template #raw>
          <AdminFormField :label="$t('fields.extJson')">
            <AdminInput v-model="form.config" type="textarea" :rows="12" class="font-mono text-sm" />
          </AdminFormField>
        </template>

        <template #probe>
          <AdminConfigProbePanel
            :loading="probing"
            :result="lastResult"
            :disabled="!form.id"
            :hint="$t('serviceConfig.llmProbeHint')"
            @probe="runProbe"
          />
        </template>
      </AdminConfigEditorShell>
    </AdminServiceConfigLayout>
  </div>
</template>
