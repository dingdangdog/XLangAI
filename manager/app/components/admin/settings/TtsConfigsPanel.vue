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

const API = "/api/admin/tts-service-configs";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();
const { probe, probing, lastResult, clearProbe } = useServiceConfigProbe(API);

const PROVIDERS = [
  { value: "openai_rest", label: "OpenAI", group: "Global" },
  { value: "azure_speech_rest", label: "Microsoft Azure", group: "Global" },
  { value: "google_cloud_tts", label: "Google Cloud TTS", group: "Global" },
  { value: "gemini_tts", label: "Google Gemini TTS", group: "Global" },
  { value: "aws_polly", label: "Amazon Polly", group: "Global" },
  { value: "elevenlabs", label: "ElevenLabs", group: "Global" },
  { value: "deepgram", label: "Deepgram", group: "Global" },
  { value: "ibm_watson", label: "IBM Watson", group: "Global" },
  { value: "playht", label: "PlayHT", group: "Global" },
  { value: "tencent_tts", label: "腾讯云", group: "China" },
  { value: "aliyun_nls", label: "阿里云 NLS", group: "China" },
  { value: "baidu_tts", label: "百度", group: "China" },
  { value: "xunfei", label: "讯飞", group: "China" },
  { value: "minimax", label: "MiniMax", group: "China" },
  { value: "volcengine", label: "火山引擎", group: "China" },
] as const;

const DEFAULT_MODEL: Record<string, string> = {
  openai_rest: "tts-1",
  azure_speech_rest: "-",
  google_cloud_tts: "-",
  gemini_tts: "gemini-2.5-flash-preview-tts",
  aws_polly: "-",
  elevenlabs: "-",
  deepgram: "aura-asteria-en",
  ibm_watson: "-",
  tencent_tts: "-",
  aliyun_nls: "-",
  baidu_tts: "-",
  xunfei: "-",
  minimax: "speech-02-turbo",
  volcengine: "-",
  playht: "-",
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
const probeVoiceCode = ref("alloy");

const form = reactive({
  id: "",
  code: "",
  name: "",
  provider: "azure_speech_rest",
  baseUrl: "",
  apiKey: "",
  region: "",
  modelCode: "-",
  config: "{}",
  status: "active",
  sortOrder: 0,
  remark: "",
});

const schemaFields = ref<Record<string, string | number | boolean>>({});
const snapshot = ref("");

const providerOptions = computed(() => PROVIDERS.map((p) => ({ ...p })));
const configSchema = computed(() => getConfigSchema("tts", form.provider));

function autoTtsCode(provider: string) {
  const p = (provider || "openai_rest").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
}

function regionRequired(): boolean {
  return ["azure_speech_rest", "aliyun_nls", "tencent_tts", "aws_polly"].includes(form.provider);
}

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

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.provider = "azure_speech_rest";
  form.baseUrl = "";
  form.apiKey = "";
  form.region = "eastus";
  form.modelCode = DEFAULT_MODEL.azure_speech_rest ?? "-";
  form.config = "{}";
  form.status = "active";
  form.sortOrder = 0;
  form.remark = "";
  probeVoiceCode.value = form.provider === "azure_speech_rest" ? "en-US-JennyNeural" : "alloy";
  syncSchemaFromConfig();
  snapshot.value = serializeFormSnapshot();
  clearProbe();
}

function loadFormFromRow(row: Record<string, unknown>) {
  form.id = String(row.id ?? "");
  form.code = String(row.code ?? "");
  form.name = String(row.name ?? "");
  form.provider = String(row.provider ?? "azure_speech_rest");
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
    const def = DEFAULT_MODEL[v];
    if (def) form.modelCode = def;
    if (v === "azure_speech_rest" && !form.region.trim()) form.region = "eastus";
    if (v === "tencent_tts" && !form.region.trim()) form.region = "ap-guangzhou";
    probeVoiceCode.value = v === "azure_speech_rest" ? "en-US-JennyNeural" : "alloy";
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

async function submitForm() {
  if (!form.name.trim()) {
    toast.warning(t("validation.fillName"));
    return;
  }
  if (regionRequired() && !form.region.trim()) {
    toast.warning(t("validation.regionRequired"));
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
    code: editorMode.value === "create" ? autoTtsCode(form.provider) : form.code.trim(),
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
    message: t("confirm.deleteTtsConfig"),
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
      baseUrl: form.baseUrl,
      apiKey: form.apiKey,
      region: form.region,
      modelCode: form.modelCode,
      config: form.config,
      voiceCode: probeVoiceCode.value,
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
    <AdminAlert :title="$t('pages.ttsConfigs.providersCount')">
      <p class="text-sm">{{ $t("pages.ttsConfigs.providersAlert") }}</p>
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
      :get-meta="
        (row) =>
          `${String(row.modelCode ?? '')} · ${serviceUsageMonthLine(row.usage as Record<string, unknown> | undefined)}`
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
        :title="editorMode === 'create' ? $t('pages.ttsConfigs.createDialog') : form.name"
        :subtitle="editorMode === 'edit' ? `${form.provider} · ${form.region || t('common.emDash')}` : $t('serviceConfig.createHint')"
        v-model:editor-tab="editorTab"
        :saving="saving"
        :deleting="deleting"
        :activating="activatingId === form.id"
        :can-activate="editorMode === 'edit' && form.status !== 'active'"
        :can-delete="editorMode === 'edit'"
        :dirty="dirty"
        @save="submitForm"
        @cancel="cancelEdit"
        @delete="removeCurrent"
        @activate="selectedRow && activateRow(selectedRow)"
      >
        <template #connection>
          <div v-if="editorMode === 'create'" class="mb-4">
            <AdminFormField :label="$t('common.provider')" required>
              <AdminProviderPicker v-model="form.provider" :options="providerOptions" />
            </AdminFormField>
          </div>
          <AdminFormField v-else :label="$t('common.provider')" required>
            <AdminSelect v-model="form.provider" :options="providerOptions.map((p) => ({ value: p.value, label: p.label }))" />
          </AdminFormField>
          <AdminFormField :label="$t('common.name')" required>
            <AdminInput v-model="form.name" />
          </AdminFormField>
          <AdminFormField :label="$t('fields.baseUrl')" :hint="$t('pages.ttsConfigs.baseUrlHint')">
            <AdminInput v-model="form.baseUrl" />
          </AdminFormField>
          <AdminFormField :label="$t('fields.apiKeySecret')" :hint="$t('pages.ttsConfigs.defaultApiKeyHint')">
            <AdminInput v-model="form.apiKey" type="password" />
          </AdminFormField>
          <AdminFormField
            :label="$t('common.region')"
            :hint="regionRequired() ? $t('pages.ttsConfigs.regionRequired') : $t('pages.ttsConfigs.regionOptional')"
          >
            <AdminInput v-model="form.region" placeholder="eastus / ap-guangzhou / us-east-1" />
          </AdminFormField>
          <AdminFormField :label="$t('fields.modelCode')" :hint="$t('pages.ttsConfigs.modelHint')">
            <AdminInput v-model="form.modelCode" />
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
          <AdminFormField :label="$t('serviceConfig.probeVoiceCode')" :hint="$t('serviceConfig.probeVoiceHint')">
            <AdminInput v-model="probeVoiceCode" placeholder="alloy / en-US-JennyNeural" />
          </AdminFormField>
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
            :hint="$t('serviceConfig.ttsProbeHint')"
            @probe="runProbe"
          />
        </template>
      </AdminConfigEditorShell>
    </AdminServiceConfigLayout>
  </div>
</template>
