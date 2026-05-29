/**
 * 与 Go Server 运行时能力对齐的服务目录（单一事实来源）。
 * 对应：
 * - LLM: server/internal/llmchat/config.go + service.go
 * - TTS: server/internal/ai/tts_providers.go + synthesize.go
 * - STT: server/internal/handler/ai_handler.go + repository/stt_config_repo.go
 * - SMS: server/internal/sms/service.go
 * - Translate: server/internal/translate/service.go
 */

export type LlmProtocolFamily = "openai" | "claude" | "gemini";
export type LlmOpenAiFlavor = "generic" | "azure";

/** Go IsOpenAICompatible 完整列表 */
export const LLM_OPENAI_COMPAT_STORED_PROTOCOLS = [
  "openai",
  "azure_openai",
  "ollama",
  "deepseek",
  "openrouter",
  "groq",
  "together",
  "zhipu",
  "moonshot",
  "siliconflow",
  "nvidia_nim",
  "mistral",
] as const;

export type LlmOpenAiCompatStored = (typeof LLM_OPENAI_COMPAT_STORED_PROTOCOLS)[number];

/** 后台 UI 仅展示 Server 实际路由的协议族 */
export const LLM_PROTOCOL_FAMILIES = [
  {
    value: "openai" as const,
    labelKey: "serverCatalog.llm.familyOpenai",
    descriptionKey: "serverCatalog.llm.familyOpenaiDesc",
  },
  {
    value: "claude" as const,
    labelKey: "serverCatalog.llm.familyClaude",
    descriptionKey: "serverCatalog.llm.familyClaudeDesc",
  },
  {
    value: "gemini" as const,
    labelKey: "serverCatalog.llm.familyGemini",
    descriptionKey: "serverCatalog.llm.familyGeminiDesc",
  },
];

/** OpenAI 兼容族内的存储变体（均为同一 Chat Completions 路由） */
export const LLM_OPENAI_FLAVORS = [
  {
    value: "generic" as const,
    storedProtocol: "openai" as const,
    labelKey: "serverCatalog.llm.flavorGeneric",
    descriptionKey: "serverCatalog.llm.flavorGenericDesc",
  },
  {
    value: "azure" as const,
    storedProtocol: "azure_openai" as const,
    labelKey: "serverCatalog.llm.flavorAzure",
    descriptionKey: "serverCatalog.llm.flavorAzureDesc",
  },
];

const LLM_OPENAI_COMPAT_SET = new Set<string>(LLM_OPENAI_COMPAT_STORED_PROTOCOLS);

const LLM_NATIVE_PROTOCOLS = new Set(["claude", "anthropic", "gemini", "google_gemini"]);

export function isLlmOpenAiCompatibleProtocol(protocol: string): boolean {
  return LLM_OPENAI_COMPAT_SET.has(protocol.trim().toLowerCase());
}

export function isSupportedLlmProtocol(protocol: string): boolean {
  const p = protocol.trim().toLowerCase();
  if (!p) return true;
  return isLlmOpenAiCompatibleProtocol(p) || LLM_NATIVE_PROTOCOLS.has(p);
}

/** 将库中 protocol 映射为 UI 协议族 */
export function llmProtocolToFamily(protocol: string): LlmProtocolFamily {
  const p = protocol.trim().toLowerCase();
  if (p === "claude" || p === "anthropic") return "claude";
  if (p === "gemini" || p === "google_gemini") return "gemini";
  if (isLlmOpenAiCompatibleProtocol(p) || p === "") return "openai";
  return "openai";
}

/** 从库 protocol 推断 OpenAI 兼容变体 */
export function llmOpenAiFlavorFromProtocol(protocol: string): LlmOpenAiFlavor {
  return protocol.trim().toLowerCase() === "azure_openai" ? "azure" : "generic";
}

/** 保存时：协议族 + OpenAI 变体 → 入库 protocol */
export function llmStoredProtocol(family: LlmProtocolFamily, openAiFlavor: LlmOpenAiFlavor): string {
  if (family === "claude") return "claude";
  if (family === "gemini") return "gemini";
  const flavor = LLM_OPENAI_FLAVORS.find((f) => f.value === openAiFlavor);
  return flavor?.storedProtocol ?? "openai";
}

export function llmDefaultModel(family: LlmProtocolFamily): string {
  switch (family) {
    case "claude":
      return "claude-3-5-sonnet-20241022";
    case "gemini":
      return "gemini-1.5-flash";
    default:
      return "gpt-4o-mini";
  }
}

/** Go synthesize.go switch 支持的 provider（规范化后） */
export const TTS_SUPPORTED_PROVIDERS = [
  { value: "openai_rest", labelKey: "serverCatalog.tts.openai", groupKey: "serverCatalog.group.global" },
  { value: "azure_speech_rest", labelKey: "serverCatalog.tts.azure", groupKey: "serverCatalog.group.global" },
  { value: "google_cloud_tts", labelKey: "serverCatalog.tts.googleCloud", groupKey: "serverCatalog.group.global" },
  { value: "gemini_tts", labelKey: "serverCatalog.tts.gemini", groupKey: "serverCatalog.group.global" },
  { value: "aws_polly", labelKey: "serverCatalog.tts.awsPolly", groupKey: "serverCatalog.group.global" },
  { value: "elevenlabs", labelKey: "serverCatalog.tts.elevenlabs", groupKey: "serverCatalog.group.global" },
  { value: "deepgram", labelKey: "serverCatalog.tts.deepgram", groupKey: "serverCatalog.group.global" },
  { value: "ibm_watson", labelKey: "serverCatalog.tts.ibmWatson", groupKey: "serverCatalog.group.global" },
  { value: "playht", labelKey: "serverCatalog.tts.playht", groupKey: "serverCatalog.group.global" },
  { value: "tencent_tts", labelKey: "serverCatalog.tts.tencent", groupKey: "serverCatalog.group.china" },
  { value: "aliyun_nls", labelKey: "serverCatalog.tts.aliyun", groupKey: "serverCatalog.group.china" },
  { value: "baidu_tts", labelKey: "serverCatalog.tts.baidu", groupKey: "serverCatalog.group.china" },
  { value: "xunfei", labelKey: "serverCatalog.tts.xunfei", groupKey: "serverCatalog.group.china" },
  { value: "minimax", labelKey: "serverCatalog.tts.minimax", groupKey: "serverCatalog.group.china" },
  { value: "volcengine", labelKey: "serverCatalog.tts.volcengine", groupKey: "serverCatalog.group.china" },
] as const;

const TTS_PROVIDER_SET = new Set(TTS_SUPPORTED_PROVIDERS.map((p) => p.value));

export function isSupportedTtsProvider(provider: string): boolean {
  return TTS_PROVIDER_SET.has(provider.trim().toLowerCase() as (typeof TTS_SUPPORTED_PROVIDERS)[number]["value"]);
}

export const TTS_DEFAULT_MODEL: Record<string, string> = {
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

export const SMS_SUPPORTED_PROVIDERS = [
  { value: "aliyun", labelKey: "pages.smsService.providerAliyun" },
  { value: "tencent", labelKey: "pages.smsService.providerTencent" },
] as const;

const SMS_PROVIDER_SET = new Set(SMS_SUPPORTED_PROVIDERS.map((p) => p.value));

export function isSupportedSmsProvider(provider: string): boolean {
  const p = provider.trim().toLowerCase();
  return p === "aliyun" || p === "alibaba" || p === "aliyun_sms" || p === "tencent" || p === "tencentcloud" || p === "qcloud";
}

export function normalizeSmsProvider(provider: string): string {
  const p = provider.trim().toLowerCase();
  if (p === "tencent" || p === "tencentcloud" || p === "qcloud") return "tencent";
  return "aliyun";
}

export const STT_SUPPORTED_PROTOCOLS = [
  { value: "openai", labelKey: "pages.sttConfigs.protocolOpenai" },
  { value: "azure_speech_rest", labelKey: "pages.sttConfigs.protocolAzure" },
] as const;

export const TRANSLATE_SUPPORTED_PROTOCOLS = [
  "openai",
  "azure_translator",
  "deepl",
  "google_translate",
  "baidu_translate",
  "tencent_translate",
  "aliyun_translate",
  "aws_translate",
  "youdao_translate",
  "papago_translate",
  "ibm_watson_translate",
  "libretranslate",
  "xunfei_translate",
  "volcengine_translate",
] as const;

export function isSupportedTranslateProtocol(protocol: string): boolean {
  return (TRANSLATE_SUPPORTED_PROTOCOLS as readonly string[]).includes(protocol.trim().toLowerCase());
}

export function isSupportedSttProtocol(protocol: string): boolean {
  const p = protocol.trim().toLowerCase();
  return p === "openai" || p === "azure_speech_rest";
}
