<script setup lang="ts">
const { t } = useI18n();

const { serviceUsageMonthLine, serviceUsageTodayLine } = useUsageDisplay();

const API = "/api/admin/tts-service-configs";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

/** 全球主流 TTS；与 server/internal/ai/tts_providers.go NormalizeTTSProvider 对齐 */
const PROVIDERS = [
  { value: "openai_rest", label: "OpenAI（/v1/audio/speech）" },
  { value: "azure_speech_rest", label: "Microsoft Azure 语音" },
  { value: "google_cloud_tts", label: "Google Cloud Text-to-Speech" },
  { value: "gemini_tts", label: "Google Gemini TTS（AUDIO）" },
  { value: "aws_polly", label: "Amazon Polly" },
  { value: "elevenlabs", label: "ElevenLabs" },
  { value: "deepgram", label: "Deepgram Speak" },
  { value: "ibm_watson", label: "IBM Watson TTS" },
  { value: "tencent_tts", label: "腾讯云语音合成" },
  { value: "aliyun_nls", label: "阿里云 NLS TTS" },
  { value: "baidu_tts", label: "百度语音合成" },
  { value: "xunfei", label: "讯飞开放平台（iFlytek）" },
  { value: "minimax", label: "MiniMax 语音" },
  { value: "volcengine", label: "火山引擎 / 豆包 OpenSpeech" },
  { value: "playht", label: "PlayHT" },
] as const;

const CONFIG_HINTS: Record<string, string> = {
  openai_rest: "{}",
  azure_speech_rest: '{"output_format":"audio-16khz-128kbitrate-mono-mp3","region":"eastus"}',
  google_cloud_tts: '{"language_code":"en-US","audio_encoding":"MP3"}',
  gemini_tts: "{}",
  aws_polly: '{"secret_id":"AWS_ACCESS_KEY_ID","api_secret":"AWS_SECRET","region":"us-east-1"}',
  elevenlabs: '{"model_id":"eleven_multilingual_v2"}',
  deepgram: '{"model_id":"aura-asteria-en"}',
  ibm_watson: "{}",
  tencent_tts: '{"secret_id":"SecretId","codec":"mp3","region":"ap-guangzhou"}',
  aliyun_nls: '{"app_key":"AppKey","format":"mp3","region":"cn-shanghai"}',
  baidu_tts: '{"cuid":"xlangai","spd":5,"pit":5,"vol":5}',
  xunfei: '{"app_id":"APPID","api_secret":"APISecret"}',
  minimax: "{}",
  volcengine: '{"app_id":"AppId","cluster":"volcano_tts"}',
  playht: '{"user_id":"PlayHT用户ID"}',
};

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

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

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

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.provider = "azure_speech_rest";
  form.baseUrl = "";
  form.apiKey = "";
  form.region = "";
  form.modelCode = "-";
  form.config = CONFIG_HINTS.azure_speech_rest ?? "{}";
  form.status = "active";
  form.sortOrder = 0;
  form.remark = "";
}

function onProviderChange(v: string) {
  form.config = CONFIG_HINTS[v] ?? "{}";
  const def = DEFAULT_MODEL[v];
  if (def) form.modelCode = def;
  if (v === "azure_speech_rest" && !form.region.trim()) form.region = "eastus";
  if (v === "tencent_tts" && !form.region.trim()) form.region = "ap-guangzhou";
}

watch(() => form.provider, onProviderChange);

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
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
  dialogVisible.value = true;
}

function regionRequired(): boolean {
  return ["azure_speech_rest", "aliyun_nls", "tencent_tts", "aws_polly"].includes(form.provider);
}

function apiKeyHint(): string {
  return t("pages.ttsConfigs.defaultApiKeyHint");
}

function voiceHint(): string {
  return t("pages.ttsConfigs.defaultVoiceHint");
}

/** 库表 code 唯一；Go 按 id 查 TTS 配置，不按 code 查。新建时自动生成。 */
function autoTtsCode(provider: string) {
  const p = (provider || "openai_rest").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
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
  let configStr = (form.config ?? "").trim();
  if (configStr) {
    try {
      JSON.parse(configStr);
    } catch {
      toast.error(t("validation.invalidJson", { field: t("fields.extJson") }));
      return;
    }
  } else {
    configStr = "{}";
  }
  const body = {
    code:
      dialogMode.value === "create"
        ? autoTtsCode(form.provider)
        : form.code.trim(),
    name: form.name.trim(),
    provider: form.provider,
    baseUrl: form.baseUrl.trim() || null,
    apiKey: form.apiKey.trim() || null,
    region: form.region.trim() || null,
    modelCode: form.modelCode.trim() || "-",
    config: configStr,
    status: form.status,
    sortOrder: form.sortOrder,
    remark: form.remark.trim() || null,
  };
  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await api.create(body);
      toast.success(t("toast.created"));
    } else {
      await api.update(form.id, { id: form.id, ...body });
      toast.success(t("toast.saved"));
    }
    dialogVisible.value = false;
    await load();
  } catch (e) {
    toast.error(t("toast.saveFailed"));
    console.error(e);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: t("confirm.deleteTtsConfig"),
    danger: true,
    confirmLabel: t("common.delete"),
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success(t("toast.deleted"));
    await load();
  } catch (e) {
    toast.error(t("toast.deleteFailed"));
    console.error(e);
  }
}

const statusOptions = [
  { value: "active", label: "active" },
  { value: "inactive", label: "inactive" },
];

const providerOptions = PROVIDERS.map((p) => ({ value: p.value, label: p.label }));

const { activateRow, activatingId } = useActivateConfigRow({
  api,
  getList: () => list.value,
  exclusivity: "multi-active",
  reload: load,
});
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.ttsConfigs.title')"
        :description="$t('pages.ttsConfigs.description')"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
        </template>
      </AdminPageHeader>

      <AdminAlert :title="$t('pages.ttsConfigs.providersCount')">
        <p class="text-sm">{{ $t("pages.ttsConfigs.providersAlert") }}</p>
      </AdminAlert>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("common.code") }}</AdminTh>
          <AdminTh>{{ $t("common.name") }}</AdminTh>
          <AdminTh width="160px">{{ $t("common.provider") }}</AdminTh>
          <AdminTh width="96px">{{ $t("common.region") }}</AdminTh>
          <AdminTh>{{ $t("common.model") }}</AdminTh>
          <AdminTh>{{ $t("fields.apiKey") }}</AdminTh>
          <AdminTh width="80px">{{ $t("common.status") }}</AdminTh>
          <AdminTh width="120px">{{ $t("common.todayUsage") }}</AdminTh>
          <AdminTh width="120px">{{ $t("common.monthUsage") }}</AdminTh>
          <AdminTh width="64px">{{ $t("common.sort") }}</AdminTh>
          <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>
          <AdminTh width="200px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap>{{ row.code }}</AdminTd>
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.provider }}</AdminTd>
          <AdminTd>{{ row.region ?? t("common.emDash") }}</AdminTd>
          <AdminTd>{{ row.modelCode }}</AdminTd>
          <AdminTd><AdminMaskedKey :value="row.apiKey as string | null" /></AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
          <AdminTd class="text-sm tabular-nums">
            <div>{{ serviceUsageTodayLine(row.usage as Record<string, unknown> | undefined) }}</div>
          </AdminTd>
          <AdminTd class="text-sm tabular-nums">
            <div>{{ serviceUsageMonthLine(row.usage as Record<string, unknown> | undefined) }}</div>
          </AdminTd>
          <AdminTd>{{ row.sortOrder }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
          <AdminTd align="right" class="whitespace-nowrap">
            <AdminButton
              v-if="String(row.status) !== 'active'"
              variant="link"
              :loading="activatingId === String(row.id)"
              @click="activateRow(row)"
            >
              {{ $t("common.enable") }}
            </AdminButton>
            <AdminButton variant="link" @click="openEdit(row)">{{ $t("common.edit") }}</AdminButton>
            <AdminButton variant="link" class="!text-danger-600" @click="removeRow(row)">
              {{ $t("common.delete") }}
            </AdminButton>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? $t('pages.ttsConfigs.createDialog') : $t('pages.ttsConfigs.editDialog')"
      width="lg"
    >
      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField :label="$t('common.name')" required>
        <AdminInput v-model="form.name" />
      </AdminFormField>
      <AdminFormField :label="$t('common.provider')" required>
        <AdminSelect v-model="form.provider" :options="providerOptions" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.baseUrl')" :hint="$t('pages.ttsConfigs.baseUrlHint')">
        <AdminInput v-model="form.baseUrl" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.apiKeySecret')" :hint="apiKeyHint()">
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
      <AdminFormField :label="$t('fields.extJson')" :hint="CONFIG_HINTS[form.provider]">
        <AdminInput v-model="form.config" type="textarea" :rows="4" class="font-mono text-sm" />
      </AdminFormField>
      <AdminFormField :label="$t('pages.ttsConfigs.voiceDescription')">
        <p class="text-sm text-gray-500">{{ voiceHint() }}</p>
      </AdminFormField>
      <AdminFormField :label="$t('common.status')">
        <AdminSelect v-model="form.status" :options="statusOptions" />
      </AdminFormField>
      <AdminFormField :label="$t('common.sort')">
        <AdminInput v-model="form.sortOrder" type="number" />
      </AdminFormField>
      <AdminFormField :label="$t('common.remark')">
        <AdminInput v-model="form.remark" type="textarea" :rows="2" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submitForm">{{ $t("common.save") }}</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
