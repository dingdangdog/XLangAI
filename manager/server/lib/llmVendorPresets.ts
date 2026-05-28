/**
 * LLM 内置服务商预设（OpenAI 兼容为主）。
 * 协议均落在 Go llmchat.IsOpenAICompatible 支持的 stored protocol 上。
 * 用户选预设后只需填 API Key；模型通过 /v1/models 拉取（不支持则手动输入）。
 */

export type LlmVendorPresetKind = "openai_compat" | "claude" | "gemini" | "custom_openai";

export type LlmVendorPreset = {
  id: string;
  /** 入库 protocol 字段（Go 路由用） */
  storedProtocol: string;
  kind: LlmVendorPresetKind;
  labelKey: string;
  descriptionKey?: string;
  /** 固定 Base URL；custom 为空由用户填写 */
  baseUrl: string;
  defaultModel: string;
  /** 是否支持 GET /v1/models */
  modelsApi: boolean;
  groupKey: string;
  /** 是否为「自定义 OpenAI 兼容」入口 */
  isCustom?: boolean;
};

export const LLM_VENDOR_PRESET_CUSTOM_OPENAI = "custom_openai";

export const LLM_VENDOR_PRESETS: LlmVendorPreset[] = [
  {
    id: "openai",
    storedProtocol: "openai",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.openai",
    baseUrl: "https://api.openai.com",
    defaultModel: "gpt-4o-mini",
    modelsApi: true,
    groupKey: "serverCatalog.group.global",
  },
  {
    id: "azure_openai",
    storedProtocol: "azure_openai",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.azureOpenai",
    descriptionKey: "serverCatalog.llm.vendors.azureOpenaiDesc",
    baseUrl: "",
    defaultModel: "gpt-4o-mini",
    modelsApi: true,
    groupKey: "serverCatalog.group.global",
  },
  {
    id: "deepseek",
    storedProtocol: "deepseek",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.deepseek",
    baseUrl: "https://api.deepseek.com",
    defaultModel: "deepseek-chat",
    modelsApi: true,
    groupKey: "serverCatalog.group.china",
  },
  {
    id: "zhipu",
    storedProtocol: "zhipu",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.zhipu",
    baseUrl: "https://openai.zhipuai.cn",
    defaultModel: "glm-4-flash",
    modelsApi: true,
    groupKey: "serverCatalog.group.china",
  },
  {
    id: "moonshot",
    storedProtocol: "moonshot",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.moonshot",
    baseUrl: "https://api.moonshot.cn",
    defaultModel: "moonshot-v1-8k",
    modelsApi: true,
    groupKey: "serverCatalog.group.china",
  },
  {
    id: "siliconflow",
    storedProtocol: "siliconflow",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.siliconflow",
    baseUrl: "https://api.siliconflow.cn",
    defaultModel: "deepseek-ai/DeepSeek-V3",
    modelsApi: true,
    groupKey: "serverCatalog.group.china",
  },
  {
    id: "openrouter",
    storedProtocol: "openrouter",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.openrouter",
    baseUrl: "https://openrouter.ai/api",
    defaultModel: "openai/gpt-4o-mini",
    modelsApi: true,
    groupKey: "serverCatalog.group.global",
  },
  {
    id: "groq",
    storedProtocol: "groq",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.groq",
    baseUrl: "https://api.groq.com/openai",
    defaultModel: "llama-3.3-70b-versatile",
    modelsApi: true,
    groupKey: "serverCatalog.group.global",
  },
  {
    id: "together",
    storedProtocol: "together",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.together",
    baseUrl: "https://api.together.xyz",
    defaultModel: "meta-llama/Llama-3.3-70B-Instruct-Turbo",
    modelsApi: true,
    groupKey: "serverCatalog.group.global",
  },
  {
    id: "ollama",
    storedProtocol: "ollama",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.ollama",
    descriptionKey: "serverCatalog.llm.vendors.ollamaDesc",
    baseUrl: "http://127.0.0.1:11434",
    defaultModel: "llama3",
    modelsApi: true,
    groupKey: "serverCatalog.group.global",
  },
  {
    id: "nvidia_nim",
    storedProtocol: "nvidia_nim",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.nvidiaNim",
    baseUrl: "https://integrate.api.nvidia.com",
    defaultModel: "meta/llama-3.1-8b-instruct",
    modelsApi: true,
    groupKey: "serverCatalog.group.global",
  },
  {
    id: "mistral",
    storedProtocol: "mistral",
    kind: "openai_compat",
    labelKey: "serverCatalog.llm.vendors.mistral",
    baseUrl: "https://api.mistral.ai",
    defaultModel: "mistral-small-latest",
    modelsApi: true,
    groupKey: "serverCatalog.group.global",
  },
  {
    id: "claude",
    storedProtocol: "claude",
    kind: "claude",
    labelKey: "serverCatalog.llm.vendors.claude",
    baseUrl: "https://api.anthropic.com",
    defaultModel: "claude-3-5-sonnet-20241022",
    modelsApi: false,
    groupKey: "serverCatalog.group.native",
  },
  {
    id: "gemini",
    storedProtocol: "gemini",
    kind: "gemini",
    labelKey: "serverCatalog.llm.vendors.gemini",
    baseUrl: "https://generativelanguage.googleapis.com",
    defaultModel: "gemini-1.5-flash",
    modelsApi: false,
    groupKey: "serverCatalog.group.native",
  },
  {
    id: LLM_VENDOR_PRESET_CUSTOM_OPENAI,
    storedProtocol: "openai",
    kind: "custom_openai",
    labelKey: "serverCatalog.llm.vendors.customOpenai",
    descriptionKey: "serverCatalog.llm.vendors.customOpenaiDesc",
    baseUrl: "",
    defaultModel: "",
    modelsApi: true,
    groupKey: "serverCatalog.group.custom",
    isCustom: true,
  },
];

const PRESET_BY_ID = new Map(LLM_VENDOR_PRESETS.map((p) => [p.id, p]));
const PRESET_BY_PROTOCOL = new Map(
  LLM_VENDOR_PRESETS.filter((p) => !p.isCustom).map((p) => [p.storedProtocol, p]),
);

export function getLlmVendorPreset(id: string): LlmVendorPreset | undefined {
  return PRESET_BY_ID.get(id);
}

/** 从已有配置反推预设；无法匹配时归为 custom_openai 或 claude/gemini */
export function resolveLlmVendorPresetId(args: {
  protocol: string;
  baseUrl: string | null;
}): string {
  const protocol = args.protocol.trim().toLowerCase();
  if (protocol === "claude" || protocol === "anthropic") return "claude";
  if (protocol === "gemini" || protocol === "google_gemini") return "gemini";

  const preset = PRESET_BY_PROTOCOL.get(protocol);
  if (!preset) return LLM_VENDOR_PRESET_CUSTOM_OPENAI;

  const storedBase = (preset.baseUrl ?? "").replace(/\/$/, "").toLowerCase();
  const rowBase = (args.baseUrl ?? "").trim().replace(/\/$/, "").toLowerCase();
  if (preset.id === "azure_openai") {
    return rowBase.includes("openai.azure.com") || protocol === "azure_openai" ? "azure_openai" : LLM_VENDOR_PRESET_CUSTOM_OPENAI;
  }
  if (storedBase && rowBase && storedBase !== rowBase) {
    return LLM_VENDOR_PRESET_CUSTOM_OPENAI;
  }
  return preset.id;
}

export function isOpenAiCompatPreset(preset: LlmVendorPreset): boolean {
  return preset.kind === "openai_compat" || preset.kind === "custom_openai";
}
