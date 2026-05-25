<script setup lang="ts">
import {
  serviceUsageMonthLine,
  serviceUsageTodayLine,
} from "~/utils/usageDisplay";

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
    toast.error("加载失败");
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
  const m: Record<string, string> = {
    openai_rest: "OpenAI API Key（必填）",
    azure_speech_rest: "Azure 订阅密钥（必填）",
    google_cloud_tts: "Google API Key（必填）",
    gemini_tts: "Google API Key（必填）",
    aws_polly: "SecretAccessKey（必填）；AccessKeyId 放 config.secret_id",
    elevenlabs: "xi-api-key（必填）",
    deepgram: "Deepgram API Key（必填）",
    ibm_watson: "IBM apikey（必填）",
    tencent_tts: "SecretKey（必填）；SecretId 放 config.secret_id",
    aliyun_nls: "NLS Token（必填）；app_key 放 config",
    baidu_tts: "access_token（必填）",
    xunfei: "api_key（必填）；app_id、api_secret 放 config",
    minimax: "API Key（必填）",
    volcengine: "access_token（必填）；app_id 放 config",
    playht: "API Key（必填）；user_id 放 config",
  };
  return m[form.provider] ?? "厂商 API 密钥（必填）";
}

function voiceHint(): string {
  const m: Record<string, string> = {
    azure_speech_rest: "Azure 音色 ShortName（语音角色 voice_code）",
    tencent_tts: "VoiceType 数字，如 1001",
    openai_rest: "alloy / nova 等",
    elevenlabs: "ElevenLabs voice_id",
    aws_polly: "Joanna 等 VoiceId",
    google_cloud_tts: "en-US-Neural2-A",
    gemini_tts: "Kore 等",
    xunfei: "xiaoyan 等发音人",
    volcengine: "BV001_streaming 等",
  };
  return m[form.provider] ?? "在「语音角色」中配置 voice_code";
}

async function submitForm() {
  if (!form.code.trim() || !form.name.trim()) {
    toast.warning("请填写编码与名称");
    return;
  }
  if (regionRequired() && !form.region.trim()) {
    toast.warning("该 Provider 须填写区域");
    return;
  }
  let configStr = (form.config ?? "").trim();
  if (configStr) {
    try {
      JSON.parse(configStr);
    } catch {
      toast.error("扩展配置须为合法 JSON");
      return;
    }
  } else {
    configStr = "{}";
  }
  const body = {
    code: form.code.trim(),
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
      toast.success("已创建");
    } else {
      await api.update(form.id, { id: form.id, ...body });
      toast.success("已保存");
    }
    dialogVisible.value = false;
    await load();
  } catch (e) {
    toast.error("保存失败");
    console.error(e);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: "确认删除该 TTS 配置？语音角色若仍引用该 ID 可能导致合成失败。",
    danger: true,
    confirmLabel: "删除",
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success("已删除");
    await load();
  } catch (e) {
    toast.error("删除失败");
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
        title="TTS 服务配置"
        description="支持全球主流语音合成厂商。列表「今日/本月用量」为成功合成次数与输出文本字符数（UTC 自然日）。"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">新建</AdminButton>
        </template>
      </AdminPageHeader>

      <AdminAlert title="已接入 Provider（15）">
        <p class="text-sm">
          北美/全球：OpenAI、Azure、Google Cloud、Gemini、AWS Polly、ElevenLabs、Deepgram、IBM Watson、PlayHT。
          中国：腾讯云、阿里云、百度、讯飞、MiniMax、火山引擎。新增厂商只需在后台新建配置并绑定语音角色。
        </p>
      </AdminAlert>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>编码</AdminTh>
          <AdminTh>名称</AdminTh>
          <AdminTh width="160px">Provider</AdminTh>
          <AdminTh width="96px">区域</AdminTh>
          <AdminTh>模型</AdminTh>
          <AdminTh>API Key</AdminTh>
          <AdminTh width="80px">状态</AdminTh>
          <AdminTh width="120px">今日用量</AdminTh>
          <AdminTh width="120px">本月用量</AdminTh>
          <AdminTh width="64px">排序</AdminTh>
          <AdminTh>更新时间</AdminTh>
          <AdminTh width="200px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap>{{ row.code }}</AdminTd>
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.provider }}</AdminTd>
          <AdminTd>{{ row.region ?? "—" }}</AdminTd>
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
              启用
            </AdminButton>
            <AdminButton variant="link" @click="openEdit(row)">编辑</AdminButton>
            <AdminButton variant="link" class="!text-danger-600" @click="removeRow(row)">
              删除
            </AdminButton>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建 TTS 配置' : '编辑 TTS 配置'"
      width="lg"
    >
      <AdminFormField v-if="dialogMode === 'edit'" label="ID">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField label="编码" required>
        <AdminInput
          v-model="form.code"
          :disabled="dialogMode === 'edit'"
          placeholder="唯一，如 tencent_zh_female"
        />
      </AdminFormField>
      <AdminFormField label="名称" required>
        <AdminInput v-model="form.name" />
      </AdminFormField>
      <AdminFormField label="Provider" required>
        <AdminSelect v-model="form.provider" :options="providerOptions" />
      </AdminFormField>
      <AdminFormField label="Base URL" hint="可选；IBM / MiniMax / ElevenLabs 自定义端点">
        <AdminInput v-model="form.baseUrl" />
      </AdminFormField>
      <AdminFormField label="API Key / 密钥" :hint="apiKeyHint()">
        <AdminInput v-model="form.apiKey" type="password" />
      </AdminFormField>
      <AdminFormField label="区域" :hint="regionRequired() ? '必填' : '可选'">
        <AdminInput v-model="form.region" placeholder="eastus / ap-guangzhou / us-east-1" />
      </AdminFormField>
      <AdminFormField label="模型 code" hint="部分厂商必填，见协议说明">
        <AdminInput v-model="form.modelCode" />
      </AdminFormField>
      <AdminFormField label="扩展 JSON" :hint="CONFIG_HINTS[form.provider]">
        <AdminInput v-model="form.config" type="textarea" :rows="4" class="font-mono text-sm" />
      </AdminFormField>
      <AdminFormField label="音色说明">
        <p class="text-sm text-gray-500">{{ voiceHint() }}</p>
      </AdminFormField>
      <AdminFormField label="状态">
        <AdminSelect v-model="form.status" :options="statusOptions" />
      </AdminFormField>
      <AdminFormField label="排序">
        <AdminInput v-model="form.sortOrder" type="number" />
      </AdminFormField>
      <AdminFormField label="备注">
        <AdminInput v-model="form.remark" type="textarea" :rows="2" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">取消</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submitForm">保存</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
