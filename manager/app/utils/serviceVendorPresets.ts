/** 各服务内置厂商/协议预设（与 serverServiceCatalog 对齐） */

export type ServicePreset = {
  id: string;
  /** 入库 protocol 或 provider 字段 */
  storedValue: string;
  labelKey: string;
  descriptionKey?: string;
  groupKey: string;
  baseUrl?: string;
  defaultModel?: string;
  region?: string;
};

const G = {
  global: "serverCatalog.group.global",
  china: "serverCatalog.group.china",
  native: "serverCatalog.group.native",
  custom: "serverCatalog.group.custom",
  local: "agentHub.groups.local",
};

export const STT_PRESETS: ServicePreset[] = [
  {
    id: "openai",
    storedValue: "openai",
    labelKey: "agentHub.stt.openai",
    descriptionKey: "pages.sttConfigs.protocolOpenai",
    groupKey: G.global,
    defaultModel: "whisper-1",
  },
  {
    id: "azure_speech_rest",
    storedValue: "azure_speech_rest",
    labelKey: "agentHub.stt.azure",
    descriptionKey: "pages.sttConfigs.protocolAzure",
    groupKey: G.global,
    defaultModel: "-",
  },
];

export const TTS_PRESETS: ServicePreset[] = [
  { id: "openai_rest", storedValue: "openai_rest", labelKey: "serverCatalog.tts.openai", groupKey: G.global, defaultModel: "tts-1" },
  { id: "azure_speech_rest", storedValue: "azure_speech_rest", labelKey: "serverCatalog.tts.azure", groupKey: G.global, defaultModel: "-", region: "eastus" },
  { id: "google_cloud_tts", storedValue: "google_cloud_tts", labelKey: "serverCatalog.tts.googleCloud", groupKey: G.global, defaultModel: "-" },
  { id: "gemini_tts", storedValue: "gemini_tts", labelKey: "serverCatalog.tts.gemini", groupKey: G.global, defaultModel: "gemini-2.5-flash-preview-tts" },
  { id: "elevenlabs", storedValue: "elevenlabs", labelKey: "serverCatalog.tts.elevenlabs", groupKey: G.global, defaultModel: "-" },
  { id: "deepgram", storedValue: "deepgram", labelKey: "serverCatalog.tts.deepgram", groupKey: G.global, defaultModel: "aura-asteria-en" },
  { id: "tencent_tts", storedValue: "tencent_tts", labelKey: "serverCatalog.tts.tencent", groupKey: G.china, defaultModel: "-", region: "ap-guangzhou" },
  { id: "aliyun_nls", storedValue: "aliyun_nls", labelKey: "serverCatalog.tts.aliyun", groupKey: G.china, defaultModel: "-", region: "cn-shanghai" },
  { id: "baidu_tts", storedValue: "baidu_tts", labelKey: "serverCatalog.tts.baidu", groupKey: G.china, defaultModel: "-" },
  { id: "xunfei", storedValue: "xunfei", labelKey: "serverCatalog.tts.xunfei", groupKey: G.china, defaultModel: "-" },
  { id: "minimax", storedValue: "minimax", labelKey: "serverCatalog.tts.minimax", groupKey: G.china, defaultModel: "speech-02-turbo" },
  { id: "volcengine", storedValue: "volcengine", labelKey: "serverCatalog.tts.volcengine", groupKey: G.china, defaultModel: "-" },
  { id: "aws_polly", storedValue: "aws_polly", labelKey: "serverCatalog.tts.awsPolly", groupKey: G.global, defaultModel: "-", region: "us-east-1" },
  { id: "ibm_watson", storedValue: "ibm_watson", labelKey: "serverCatalog.tts.ibmWatson", groupKey: G.global, defaultModel: "-" },
  { id: "playht", storedValue: "playht", labelKey: "serverCatalog.tts.playht", groupKey: G.global, defaultModel: "-" },
];

export const TRANSLATE_PRESETS: ServicePreset[] = [
  { id: "openai", storedValue: "openai", labelKey: "agentHub.translate.openai", descriptionKey: "agentHub.translate.openaiDesc", groupKey: G.global },
  { id: "azure_translator", storedValue: "azure_translator", labelKey: "agentHub.translate.azure", groupKey: G.global },
  { id: "deepl", storedValue: "deepl", labelKey: "agentHub.translate.deepl", groupKey: G.global },
  { id: "google_translate", storedValue: "google_translate", labelKey: "agentHub.translate.google", groupKey: G.global },
  { id: "aws_translate", storedValue: "aws_translate", labelKey: "agentHub.translate.aws", groupKey: G.global, region: "us-east-1" },
  { id: "ibm_watson_translate", storedValue: "ibm_watson_translate", labelKey: "agentHub.translate.ibm", groupKey: G.global, region: "us-south" },
  { id: "papago_translate", storedValue: "papago_translate", labelKey: "agentHub.translate.papago", groupKey: G.global },
  { id: "libretranslate", storedValue: "libretranslate", labelKey: "agentHub.translate.libre", groupKey: G.global, baseUrl: "http://localhost:5000" },
  { id: "baidu_translate", storedValue: "baidu_translate", labelKey: "agentHub.translate.baidu", groupKey: G.china },
  { id: "youdao_translate", storedValue: "youdao_translate", labelKey: "agentHub.translate.youdao", groupKey: G.china },
  { id: "tencent_translate", storedValue: "tencent_translate", labelKey: "agentHub.translate.tencent", groupKey: G.china, region: "ap-guangzhou" },
  { id: "aliyun_translate", storedValue: "aliyun_translate", labelKey: "agentHub.translate.aliyun", groupKey: G.china, region: "cn-hangzhou" },
  { id: "xunfei_translate", storedValue: "xunfei_translate", labelKey: "agentHub.translate.xunfei", groupKey: G.china },
  { id: "volcengine_translate", storedValue: "volcengine_translate", labelKey: "agentHub.translate.volcengine", groupKey: G.china, region: "cn-north-1" },
];

export const SMS_PRESETS: ServicePreset[] = [
  { id: "aliyun", storedValue: "aliyun", labelKey: "pages.smsService.providerAliyun", groupKey: G.china, region: "cn-hangzhou" },
  { id: "tencent", storedValue: "tencent", labelKey: "pages.smsService.providerTencent", groupKey: G.china, region: "ap-guangzhou" },
];

export const OBJECT_STORAGE_PRESETS: ServicePreset[] = [
  { id: "local", storedValue: "local", labelKey: "pages.objectStorage.providerLocal", groupKey: G.local },
  { id: "cloudflare_r2", storedValue: "cloudflare_r2", labelKey: "pages.objectStorage.providerR2", groupKey: G.global },
  { id: "qiniu", storedValue: "qiniu", labelKey: "pages.objectStorage.providerQiniu", groupKey: G.china },
  { id: "aliyun_oss", storedValue: "aliyun_oss", labelKey: "pages.objectStorage.providerAliyunOss", groupKey: G.china },
];

const maps = {
  stt: new Map(STT_PRESETS.map((p) => [p.id, p])),
  tts: new Map(TTS_PRESETS.map((p) => [p.id, p])),
  translate: new Map(TRANSLATE_PRESETS.map((p) => [p.id, p])),
  sms: new Map(SMS_PRESETS.map((p) => [p.id, p])),
  storage: new Map(OBJECT_STORAGE_PRESETS.map((p) => [p.id, p])),
};

export type ServicePresetKind = keyof typeof maps;

export function getServicePreset(kind: ServicePresetKind, id: string): ServicePreset | undefined {
  return maps[kind].get(id);
}

export function resolvePresetId(kind: ServicePresetKind, storedValue: string): string {
  const v = storedValue.trim().toLowerCase();
  for (const p of maps[kind].values()) {
    if (p.storedValue === v) return p.id;
  }
  const fallback = [...maps[kind].keys()][0];
  return v || fallback || "";
}

export function presetGridItems(
  kind: ServicePresetKind,
  t: (key: string) => string,
): { id: string; label: string; description?: string; group: string }[] {
  return [...maps[kind].values()].map((p) => ({
    id: p.id,
    label: t(p.labelKey),
    description: p.descriptionKey ? t(p.descriptionKey) : undefined,
    group: t(p.groupKey),
  }));
}
