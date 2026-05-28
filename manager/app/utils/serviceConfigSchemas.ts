export type ConfigFieldType = "string" | "number" | "boolean" | "select";

export type ConfigFieldSchema = {
  key: string;
  type: ConfigFieldType;
  labelKey: string;
  hintKey?: string;
  default?: string | number | boolean;
  required?: boolean;
  options?: { value: string; label: string }[];
  min?: number;
  max?: number;
  step?: number;
  placeholder?: string;
};

export function parseConfigObject(raw: string | null | undefined): Record<string, unknown> {
  if (!raw?.trim()) return {};
  try {
    const parsed = JSON.parse(raw) as unknown;
    return parsed && typeof parsed === "object" && !Array.isArray(parsed)
      ? (parsed as Record<string, unknown>)
      : {};
  } catch {
    return {};
  }
}

export function stringifyConfigObject(obj: Record<string, unknown>): string {
  return JSON.stringify(obj, null, 2);
}

export function mergeSchemaDefaults(
  schema: ConfigFieldSchema[],
  raw: Record<string, unknown>,
): Record<string, unknown> {
  const next = { ...raw };
  for (const field of schema) {
    if (next[field.key] === undefined && field.default !== undefined) {
      next[field.key] = field.default;
    }
  }
  return next;
}

export function configObjectFromSchemaFields(
  schema: ConfigFieldSchema[],
  fields: Record<string, string | number | boolean>,
): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const field of schema) {
    const v = fields[field.key];
    if (v === "" || v === undefined || v === null) continue;
    if (field.type === "number") {
      const n = Number(v);
      if (!Number.isNaN(n)) out[field.key] = n;
    } else if (field.type === "boolean") {
      out[field.key] = v === true || v === "true";
    } else {
      out[field.key] = String(v);
    }
  }
  return out;
}

export function schemaFieldsFromConfigObject(
  schema: ConfigFieldSchema[],
  raw: Record<string, unknown>,
): Record<string, string | number | boolean> {
  const merged = mergeSchemaDefaults(schema, raw);
  const out: Record<string, string | number | boolean> = {};
  for (const field of schema) {
    const v = merged[field.key];
    if (v === undefined || v === null) {
      if (field.default !== undefined) out[field.key] = field.default;
      else if (field.type === "boolean") out[field.key] = false;
      else out[field.key] = "";
      continue;
    }
    if (field.type === "boolean") out[field.key] = Boolean(v);
    else if (field.type === "number") out[field.key] = Number(v);
    else out[field.key] = String(v);
  }
  return out;
}

const LLM_CLAUDE: ConfigFieldSchema[] = [
  {
    key: "anthropic_version",
    type: "string",
    labelKey: "serviceConfig.schema.anthropicVersion",
    default: "2023-06-01",
  },
  {
    key: "max_tokens",
    type: "number",
    labelKey: "serviceConfig.schema.maxTokens",
    default: 4096,
    min: 1,
    max: 128000,
  },
  {
    key: "temperature",
    type: "number",
    labelKey: "serviceConfig.schema.temperature",
    default: 0.7,
    min: 0,
    max: 2,
    step: 0.1,
  },
];

const LLM_GEMINI: ConfigFieldSchema[] = [
  {
    key: "max_tokens",
    type: "number",
    labelKey: "serviceConfig.schema.maxTokens",
    default: 8192,
    min: 1,
    max: 128000,
  },
  {
    key: "temperature",
    type: "number",
    labelKey: "serviceConfig.schema.temperature",
    default: 0.7,
    min: 0,
    max: 2,
    step: 0.1,
  },
];

const LLM_AZURE: ConfigFieldSchema[] = [
  {
    key: "api_version",
    type: "string",
    labelKey: "serviceConfig.schema.apiVersion",
    default: "2024-02-15-preview",
  },
];

export const LLM_CONFIG_SCHEMAS: Record<string, ConfigFieldSchema[]> = {
  claude: LLM_CLAUDE,
  gemini: LLM_GEMINI,
  azure_openai: LLM_AZURE,
};

export const TTS_CONFIG_SCHEMAS: Record<string, ConfigFieldSchema[]> = {
  azure_speech_rest: [
    {
      key: "output_format",
      type: "string",
      labelKey: "serviceConfig.schema.outputFormat",
      default: "audio-16khz-128kbitrate-mono-mp3",
    },
    {
      key: "region",
      type: "string",
      labelKey: "serviceConfig.schema.regionInConfig",
      hintKey: "serviceConfig.schema.regionInConfigHint",
    },
  ],
  google_cloud_tts: [
    {
      key: "language_code",
      type: "string",
      labelKey: "serviceConfig.schema.languageCode",
      default: "en-US",
    },
    {
      key: "audio_encoding",
      type: "select",
      labelKey: "serviceConfig.schema.audioEncoding",
      default: "MP3",
      options: [
        { value: "MP3", label: "MP3" },
        { value: "OGG_OPUS", label: "OGG_OPUS" },
        { value: "LINEAR16", label: "LINEAR16" },
      ],
    },
  ],
  elevenlabs: [
    {
      key: "model_id",
      type: "string",
      labelKey: "serviceConfig.schema.modelId",
      default: "eleven_multilingual_v2",
    },
  ],
  deepgram: [
    {
      key: "model_id",
      type: "string",
      labelKey: "serviceConfig.schema.modelId",
      default: "aura-asteria-en",
    },
  ],
  tencent_tts: [
    {
      key: "secret_id",
      type: "string",
      labelKey: "serviceConfig.schema.secretIdInConfig",
      hintKey: "serviceConfig.schema.secretIdInConfigHint",
    },
    { key: "codec", type: "string", labelKey: "serviceConfig.schema.codec", default: "mp3" },
  ],
  aliyun_nls: [
    {
      key: "app_key",
      type: "string",
      labelKey: "serviceConfig.schema.appKey",
      required: true,
    },
    { key: "format", type: "string", labelKey: "serviceConfig.schema.format", default: "mp3" },
  ],
  baidu_tts: [
    { key: "cuid", type: "string", labelKey: "serviceConfig.schema.cuid", default: "xlangai" },
    { key: "spd", type: "number", labelKey: "serviceConfig.schema.speed", default: 5, min: 0, max: 15 },
    { key: "pit", type: "number", labelKey: "serviceConfig.schema.pitch", default: 5, min: 0, max: 15 },
    { key: "vol", type: "number", labelKey: "serviceConfig.schema.volume", default: 5, min: 0, max: 15 },
  ],
  xunfei: [
    { key: "app_id", type: "string", labelKey: "serviceConfig.schema.appId", required: true },
    { key: "api_secret", type: "string", labelKey: "serviceConfig.schema.apiSecret" },
  ],
  volcengine: [
    { key: "app_id", type: "string", labelKey: "serviceConfig.schema.appId", required: true },
    {
      key: "cluster",
      type: "string",
      labelKey: "serviceConfig.schema.cluster",
      default: "volcano_tts",
    },
  ],
  playht: [
    { key: "user_id", type: "string", labelKey: "serviceConfig.schema.playhtUserId" },
  ],
  aws_polly: [
    {
      key: "secret_id",
      type: "string",
      labelKey: "serviceConfig.schema.awsAccessKeyId",
    },
    {
      key: "api_secret",
      type: "string",
      labelKey: "serviceConfig.schema.awsSecretKey",
    },
    {
      key: "region",
      type: "string",
      labelKey: "serviceConfig.schema.regionInConfig",
      default: "us-east-1",
    },
  ],
};

export const SMS_CONFIG_SCHEMAS: Record<string, ConfigFieldSchema[]> = {
  aliyun: [
    {
      key: "endpoint",
      type: "string",
      labelKey: "serviceConfig.schema.smsEndpoint",
      default: "dysmsapi.aliyuncs.com",
    },
    {
      key: "template_param_key",
      type: "string",
      labelKey: "serviceConfig.schema.templateParamKey",
      default: "code",
    },
  ],
  tencent: [
    {
      key: "sdk_app_id",
      type: "string",
      labelKey: "serviceConfig.schema.sdkAppId",
      required: true,
    },
    {
      key: "endpoint",
      type: "string",
      labelKey: "serviceConfig.schema.smsEndpoint",
      default: "sms.tencentcloudapi.com",
    },
  ],
};

export function getConfigSchema(
  kind: "llm" | "tts" | "sms",
  key: string,
  options?: { llmOpenAiFlavor?: "generic" | "azure" },
): ConfigFieldSchema[] {
  if (kind === "llm") {
    if (key === "openai" && options?.llmOpenAiFlavor === "azure") {
      return LLM_CONFIG_SCHEMAS.azure_openai ?? [];
    }
    return LLM_CONFIG_SCHEMAS[key] ?? [];
  }
  if (kind === "tts") return TTS_CONFIG_SCHEMAS[key] ?? [];
  return SMS_CONFIG_SCHEMAS[key] ?? [];
}
