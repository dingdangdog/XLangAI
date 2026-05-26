import type { AppPrismaClient } from "./prisma";

/**
 * 启动种子数据：通过 Prisma 模型写入，物理表为 sys_languages、sys_tts_service_configs、sys_voice_roles、
 * sys_llm_service_configs、sys_stt_service_configs、sys_prompt_templates、sys_membership_tiers、usr_users
 *（见 schema @@map）。幂等：仅在不满足条件时插入。
 */

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

/** 与 server/internal/ai/tts_providers.go 一致 */
export const AZURE_SPEECH_REST_PROVIDER = "azure_speech_rest";
export const OPENAI_TTS_PROVIDER = "openai_rest";
export const ALIYUN_TTS_PROVIDER = "aliyun_nls";
export const GEMINI_TTS_PROVIDER = "gemini_tts";
export const TENCENT_TTS_PROVIDER = "tencent_tts";
export const GOOGLE_CLOUD_TTS_PROVIDER = "google_cloud_tts";
export const ELEVENLABS_TTS_PROVIDER = "elevenlabs";

/** Azure REST TTS 默认输出格式，与 Go 侧 azure_tts.go 默认一致 */
const AZURE_TTS_CONFIG_JSON = JSON.stringify({
  output_format: "audio-16khz-128kbitrate-mono-mp3",
  region: "",
});

// ---------------------------------------------------------------------------
// Seed data definitions
// ---------------------------------------------------------------------------

/** 启动种子默认仅开放中、日、英；其余语种预置为 inactive，可在运营后台启用 */
const DEFAULT_ACTIVE_LANGUAGE_CODES = new Set(["zh", "ja", "en"]);

const SEED_LANGUAGES = [
  { code: "en", name: "English", nameNative: "English", sortOrder: 10 },
  { code: "zh", name: "Chinese", nameNative: "中文", sortOrder: 20 },
  { code: "ja", name: "Japanese", nameNative: "日本語", sortOrder: 30 },
  { code: "ko", name: "Korean", nameNative: "한국어", sortOrder: 40 },
  { code: "es", name: "Spanish", nameNative: "Español", sortOrder: 50 },
  { code: "fr", name: "French", nameNative: "Français", sortOrder: 60 },
  { code: "de", name: "German", nameNative: "Deutsch", sortOrder: 70 },
  { code: "pt", name: "Portuguese", nameNative: "Português", sortOrder: 80 },
  { code: "it", name: "Italian", nameNative: "Italiano", sortOrder: 90 },
  { code: "ru", name: "Russian", nameNative: "Русский", sortOrder: 100 },
  { code: "yue", name: "Cantonese", nameNative: "粵語", sortOrder: 105 },
  { code: "ar", name: "Arabic", nameNative: "العربية", sortOrder: 110 },
  { code: "hi", name: "Hindi", nameNative: "हिन्दी", sortOrder: 120 },
] as const;

/** 各语言默认试听文案模板（{name} 替换为语音角色显示名） */
const PREVIEW_SAMPLE_BY_LANG: Record<string, string> = {
  zh: "你好，我是{name}",
  yue: "你好，我係{name}",
  en: "Hello, I'm {name}",
  ja: "こんにちは、{name}です",
  ko: "안녕하세요, 저는 {name}입니다",
  es: "Hola, soy {name}",
  fr: "Bonjour, je suis {name}",
  de: "Hallo, ich bin {name}",
  pt: "Olá, eu sou {name}",
  it: "Ciao, sono {name}",
  ru: "Привет, я {name}",
  ar: "مرحباً، أنا {name}",
  hi: "नमस्ते, मैं {name} हूँ",
};

/**
 * Azure Cognitive Services 官方 Neural 音色；能查到 ShortName 的优先用 *MultilingualNeural（见 multilingual-voices.md Group1）。
 * @see https://learn.microsoft.com/azure/ai-services/speech-service/language-support
 * @see https://learn.microsoft.com/azure/ai-services/speech-service/includes/language-support/multilingual-voices
 */
const SEED_VOICES = [
  // English — Azure Multilingual Neural
  {
    langCode: "en",
    voiceCode: "en-US-AmandaMultilingualNeural",
    name: "Amanda",
    gender: "female",
    sortOrder: 10,
  },
  {
    langCode: "en",
    voiceCode: "en-US-AvaMultilingualNeural",
    name: "Ava",
    gender: "female",
    sortOrder: 20,
  },
  {
    langCode: "en",
    voiceCode: "en-US-AndrewMultilingualNeural",
    name: "Andrew",
    gender: "male",
    sortOrder: 30,
  },
  {
    langCode: "en",
    voiceCode: "en-US-LewisMultilingualNeural",
    name: "Lewis",
    gender: "male",
    sortOrder: 40,
  },
  // Chinese — 普通话
  { langCode: "zh", voiceCode: "zh-CN-XiaoxiaoMultilingualNeural", name: "晓晓", gender: "female", sortOrder: 10 },
  { langCode: "zh", voiceCode: "zh-CN-YunxiaoMultilingualNeural", name: "云萧", gender: "male", sortOrder: 20 },
  { langCode: "zh", voiceCode: "zh-CN-XiaoyuMultilingualNeural", name: "晓雨", gender: "female", sortOrder: 30 },
  // Japanese — Group1 仅男声 Multilingual，女声仍为单语神经
  { langCode: "ja", voiceCode: "ja-JP-NanamiNeural", name: "Nanami", gender: "female", sortOrder: 10 },
  { langCode: "ja", voiceCode: "ja-JP-MasaruMultilingualNeural", name: "Masaru", gender: "male", sortOrder: 20 },
  // Korean
  { langCode: "ko", voiceCode: "ko-KR-SunHiNeural", name: "SunHi", gender: "female", sortOrder: 10 },
  { langCode: "ko", voiceCode: "ko-KR-HyunsuMultilingualNeural", name: "Hyunsu", gender: "male", sortOrder: 20 },
  // Spanish
  { langCode: "es", voiceCode: "es-MX-DaliaMultilingualNeural", name: "Dalia", gender: "female", sortOrder: 10 },
  { langCode: "es", voiceCode: "es-ES-TristanMultilingualNeural", name: "Tristan", gender: "male", sortOrder: 20 },
  // French
  { langCode: "fr", voiceCode: "fr-FR-VivienneMultilingualNeural", name: "Vivienne", gender: "female", sortOrder: 10 },
  { langCode: "fr", voiceCode: "fr-FR-RemyMultilingualNeural", name: "Remy", gender: "male", sortOrder: 20 },
  // German
  { langCode: "de", voiceCode: "de-DE-SeraphinaMultilingualNeural", name: "Seraphina", gender: "female", sortOrder: 10 },
  { langCode: "de", voiceCode: "de-DE-FlorianMultilingualNeural", name: "Florian", gender: "male", sortOrder: 20 },
  // Portuguese (Brazil)
  { langCode: "pt", voiceCode: "pt-BR-ThalitaMultilingualNeural", name: "Thalita", gender: "female", sortOrder: 10 },
  { langCode: "pt", voiceCode: "pt-BR-MacerioMultilingualNeural", name: "Macerio", gender: "male", sortOrder: 20 },
  // Italian
  { langCode: "it", voiceCode: "it-IT-IsabellaMultilingualNeural", name: "Isabella", gender: "female", sortOrder: 10 },
  { langCode: "it", voiceCode: "it-IT-AlessioMultilingualNeural", name: "Alessio", gender: "male", sortOrder: 20 },
  // Russian — Group1 无 ru-RU Multilingual ShortName，仍为单语神经
  { langCode: "ru", voiceCode: "ru-RU-SvetlanaNeural", name: "Svetlana", gender: "female", sortOrder: 10 },
  { langCode: "ru", voiceCode: "ru-RU-DmitryNeural", name: "Dmitry", gender: "male", sortOrder: 20 },
  // Cantonese（语言码 yue，Azure zh-HK）
  { langCode: "yue", voiceCode: "zh-HK-HiuGaaiNeural", name: "晓佳", gender: "female", sortOrder: 10 },
  { langCode: "yue", voiceCode: "zh-HK-WanLungNeural", name: "云龙", gender: "male", sortOrder: 20 },
  // Arabic
  { langCode: "ar", voiceCode: "ar-SA-ZariyahNeural", name: "Zariyah", gender: "female", sortOrder: 10 },
  { langCode: "ar", voiceCode: "ar-SA-HamedNeural", name: "Hamed", gender: "male", sortOrder: 20 },
  // Hindi
  { langCode: "hi", voiceCode: "hi-IN-SwaraNeural", name: "Swara", gender: "female", sortOrder: 10 },
  { langCode: "hi", voiceCode: "hi-IN-MadhurNeural", name: "Madhur", gender: "male", sortOrder: 20 },
] as const;

const DEFAULT_AZURE_TTS_CODE = "azure_speech_default";

const BUNDLED_VOICE_PREVIEW_FILES: Record<string, string> = {
  "en-US-AmandaMultilingualNeural": "en-US-AmandaMultilingualNeural-Amanda.mp3",
  "en-US-AvaMultilingualNeural": "en-US-AvaMultilingualNeural-Ava.mp3",
  "en-US-AndrewMultilingualNeural": "en-US-AndrewMultilingualNeural-Andrew.mp3",
  "en-US-LewisMultilingualNeural": "en-US-LewisMultilingualNeural-Lewis.mp3",
  "zh-CN-XiaoxiaoMultilingualNeural": "zh-CN-XiaoxiaoMultilingualNeural-Xiaoxiao.mp3",
  "zh-CN-YunxiaoMultilingualNeural": "zh-CN-YunxiaoMultilingualNeural-Yunxiao.mp3",
  "zh-CN-XiaoyuMultilingualNeural": "zh-CN-XiaoyuMultilingualNeural-Xiaoyu.mp3",
  "zh-HK-HiuGaaiNeural": "zh-HK-HiuGaaiNeural-Xiaojia.mp3",
};

const LEGACY_BUNDLED_VOICE_PREVIEW_FILES = new Set([
  "d68d463e-26cc-4184-9bf6-b86b7395c78a.mp3",
  "8106b07e-16e8-4f24-be6d-580ad97f524b.mp3",
  "8a148eef-9d7d-4a12-8c95-34f27f8a780d.mp3",
  "cb563eea-b177-4562-b904-b84147633b9a.mp3",
  "fefc697b-cd63-43a7-a761-f43d6d217c63.mp3",
  "b12ba589-642e-4650-a6de-20c4170dd80d.mp3",
  "8b2f137a-c8b0-4a0f-9eeb-2351c3d3f3ef.mp3",
  "f2375d62-19b7-46bd-b61c-eae9cd8a3b98.mp3",
]);

function bundledVoicePreviewData(voiceCode: string): {
  previewAudioUrl?: string;
  previewLocalFilename?: string;
} {
  const filename = BUNDLED_VOICE_PREVIEW_FILES[voiceCode];
  if (!filename) return {};
  return {
    previewAudioUrl: `/api/v1/audio/${filename}`,
    previewLocalFilename: filename,
  };
}

/** 默认测试登录手机号（与 Nitro 种子、联调文档一致） */
export const TEST_SEED_PHONE = "13800138000";

const SEED_MEMBERSHIP_TIERS = [
  {
    code: "free",
    name: "免费版",
    dailyLimit: 10,
    monthlyLimit: 100,
    features: JSON.stringify({
      voice_chat: true,
      text_chat: true,
      pronunciation_check: true,
    }),
    sortOrder: 10,
    remark: "启动种子：免费用户，每日 10 次，每月 100 次",
  },
  {
    code: "plus",
    name: "Plus 会员",
    dailyLimit: 100,
    monthlyLimit: 1000,
    features: JSON.stringify({
      voice_chat: true,
      text_chat: true,
      pronunciation_check: true,
    }),
    sortOrder: 15,
    remark: "启动种子：Plus；每月 3000 次赠送对话，超出后按 token 扣费",
  },
  {
    code: "pro",
    name: "专业版",
    dailyLimit: 500,
    monthlyLimit: 5000,
    features: JSON.stringify({
      voice_chat: true,
      text_chat: true,
      pronunciation_check: true,
      priority_support: true,
    }),
    sortOrder: 20,
    remark: "启动种子：Pro；每月 12000 次赠送对话，超出后按 token 扣费",
  },
] as const;

/** 语言练习系统提示词 — 核心 prompt，Go 端替换 {{target_lang}}、{{voice_role_name}} */
const LANG_PRACTICE_PROMPT_CONTENT = `你是「{{voice_role_name}}」，一位温暖、自然、有陪伴感的{{target_lang}}对话伴侣。你的核心任务是陪用户用{{target_lang}}进行真实对话，并在对话中帮助用户慢慢说得更自然；但你的对外身份不是课堂老师、题库机器人、翻译器或大模型助手。

【核心行为优先级】
- 先像真人一样接住用户刚说的话：回应内容、情绪、态度或处境。
- 再自然推进对话：可以追问一个轻松的问题，也可以给出简短、有温度的反应。
- 最后才考虑语言帮助：只有在不打断聊天氛围时，才做轻量纠错或给出更自然表达。

【情感交流】
- 用户表达开心、难过、压力、困惑、尴尬、兴奋、疲惫等感受时，先具体回应这种感受，不要机械鼓励。
- 你的语气要像熟人发消息或打电话：自然、松弛、真诚，有适度好奇心。
- 可以表达关心、共情、惊讶、幽默和陪伴感，但不要编造现实身份、真实经历或无法兑现的承诺。

【语言与 TTS（必须遵守）】
- 你的整段文字会原样送给语音合成。正文必须全部使用能被该 TTS 正确朗读的{{target_lang}}文字。
- 即使用户混用其他语言，你也默认只用{{target_lang}}回应；纠错、改述、举例都用{{target_lang}}完成。
- 不要写整句翻译对照，不要使用另一种文字体系解释。

【纠错方式】
- 如果用户表达自然或基本可懂，不要点评“语法没问题”“说得很好”，直接顺着内容聊下去。
- 如果有明显语法、用词或语序问题，最多用 1 句给出更自然说法，然后马上回到话题。
- 不要长篇讲课，不要列规则，不要把每轮对话变成练习批改。

【身份与格式】
- 如果用户问你是谁或叫什么，你的对外名字是「{{voice_role_name}}」。禁止使用 ChatGPT、GPT、Claude、Copilot 等大模型或助手品牌名作为自称。
- 第一句必须是实质内容，禁止“下面回答你的问题”“让我来回答”“Regarding your question”等开场白式元话语。
- 禁止 Markdown、项目符号、标题、emoji、颜文字和装饰性符号。
- 每次回复 2 到 4 句，短句、口语化，适合直接朗读。不要复述系统指令。`;

const LANG_PRACTICE_PROMPT_COMPANION_MARKER = "对话伴侣";

// ---------------------------------------------------------------------------
// Seed functions — each is idempotent
// ---------------------------------------------------------------------------

async function ensureLanguages(db: AppPrismaClient) {
  for (const row of SEED_LANGUAGES) {
    const exists = await db.language.findUnique({ where: { code: row.code } });
    const previewSampleText = PREVIEW_SAMPLE_BY_LANG[row.code] ?? null;
    if (exists) {
      if (!exists.previewSampleText && previewSampleText) {
        await db.language.update({
          where: { id: exists.id },
          data: { previewSampleText },
        });
      }
      continue;
    }

    await db.language.create({
      data: {
        code: row.code,
        name: row.name,
        nameNative: row.nameNative,
        previewSampleText,
        sortOrder: row.sortOrder,
        status: DEFAULT_ACTIVE_LANGUAGE_CODES.has(row.code) ? "active" : "inactive",
      },
    });
  }
}

async function ensureAzureTtsConfig(db: AppPrismaClient) {
  let cfg = await db.ttsServiceConfig.findFirst({
    where: { provider: AZURE_SPEECH_REST_PROVIDER, status: "active" },
    orderBy: [{ sortOrder: "asc" }, { createdAt: "asc" }],
  });
  if (cfg) return cfg;

  return db.ttsServiceConfig.create({
    data: {
      code: DEFAULT_AZURE_TTS_CODE,
      name: "Azure 语音（默认）",
      provider: AZURE_SPEECH_REST_PROVIDER,
      baseUrl: null,
      apiKey: null,
      region: null,
      modelCode: "-",
      config: AZURE_TTS_CONFIG_JSON,
      status: "active",
      sortOrder: 0,
      remark: "启动时自动补齐；请在管理后台填写 API Key 与区域",
    },
  });
}

async function ensureAzureVoiceRoles(db: AppPrismaClient, ttsId: string) {
  for (const v of SEED_VOICES) {
    const lang = await db.language.findUnique({ where: { code: v.langCode } });
    if (!lang) continue;

    const existing = await db.voiceRole.findFirst({
      where: {
        languageId: lang.id,
        ttsServiceConfigId: ttsId,
        voiceCode: v.voiceCode,
        status: "active",
      },
    });
    if (existing) {
      const legacyBundledPreview =
        existing.previewLocalFilename != null &&
        LEGACY_BUNDLED_VOICE_PREVIEW_FILES.has(existing.previewLocalFilename);
      if (!existing.previewAudioUrl?.trim() || legacyBundledPreview) {
        const preview = bundledVoicePreviewData(v.voiceCode);
        if (preview.previewAudioUrl) {
          await db.voiceRole.update({
            where: { id: existing.id },
            data: preview,
          });
        }
      }
      continue;
    }

    await db.voiceRole.create({
      data: {
        languageId: lang.id,
        ttsServiceConfigId: ttsId,
        voiceCode: v.voiceCode,
        name: v.name,
        gender: v.gender,
        sortOrder: v.sortOrder,
        status: "active",
        remark: "Azure Neural TTS",
        ...bundledVoicePreviewData(v.voiceCode),
      },
    });
  }
}

async function ensureMembershipTiers(db: AppPrismaClient) {
  for (const tier of SEED_MEMBERSHIP_TIERS) {
    const exists = await db.membershipTier.findUnique({ where: { code: tier.code } });
    if (exists) continue;

    await db.membershipTier.create({
      data: {
        code: tier.code,
        name: tier.name,
        dailyLimit: tier.dailyLimit,
        monthlyLimit: tier.monthlyLimit,
        features: tier.features,
        status: "active",
        sortOrder: tier.sortOrder,
        remark: tier.remark,
      },
    });
  }
}

async function ensureLlmServiceConfig(db: AppPrismaClient) {
  const exists = await db.sysLlmServiceConfig.findFirst({
    where: { status: "active" },
    orderBy: [{ sortOrder: "asc" }, { createdAt: "asc" }],
  });
  if (exists) return exists;

  return db.sysLlmServiceConfig.create({
    data: {
      code: "openai_default",
      name: "OpenAI（默认）",
      protocol: "openai",
      baseUrl: null,
      apiKey: null,
      modelCode: "gpt-4o-mini",
      status: "active",
      sortOrder: 0,
      remark: "启动时自动补齐；请在管理后台填写 API Key",
    },
  });
}

async function ensureTranslateServiceConfig(db: AppPrismaClient) {
  const exists = await db.sysTranslateServiceConfig.findFirst({
    where: { status: "active" },
    orderBy: [{ sortOrder: "asc" }, { createdAt: "asc" }],
  });
  if (exists) return exists;

  return db.sysTranslateServiceConfig.create({
    data: {
      code: "translate_openai_default",
      name: "OpenAI LLM 翻译（默认）",
      protocol: "openai",
      baseUrl: null,
      apiKey: null,
      modelCode: "gpt-4o-mini",
      status: "active",
      sortOrder: 0,
      remark: "启动时自动补齐；请在管理后台填写 API Key 并启用",
    },
  });
}

async function ensureSttServiceConfig(db: AppPrismaClient) {
  const exists = await db.sysSttServiceConfig.findFirst({
    where: { status: "active" },
    orderBy: [{ sortOrder: "asc" }, { createdAt: "asc" }],
  });
  if (exists) return exists;

  return db.sysSttServiceConfig.create({
    data: {
      code: "whisper_default",
      name: "OpenAI Whisper（默认）",
      protocol: "openai",
      baseUrl: null,
      apiKey: null,
      modelCode: "whisper-1",
      status: "active",
      sortOrder: 0,
      remark: "启动时自动补齐；请在管理后台填写 API Key；Go 端 Azure STT 需 ffmpeg",
    },
  });
}

async function ensurePromptTemplate(db: AppPrismaClient) {
  const exists = await db.promptTemplate.findUnique({ where: { code: "lang_practice" } });
  if (exists) {
    if (
      !exists.content.includes("{{voice_role_name}}") ||
      !exists.content.includes(LANG_PRACTICE_PROMPT_COMPANION_MARKER)
    ) {
      return db.promptTemplate.update({
        where: { code: "lang_practice" },
        data: {
          content: LANG_PRACTICE_PROMPT_CONTENT,
          variables: "target_lang,voice_role_name",
          remark:
            "启动种子/升级：对话伴侣版系统提示词；Go 端 {{target_lang}}、{{voice_role_name}} 由 ResolveSystemPromptForConversation 注入",
        },
      });
    }
    return exists;
  }

  return db.promptTemplate.create({
    data: {
      code: "lang_practice",
      name: "语言练习系统提示词",
      content: LANG_PRACTICE_PROMPT_CONTENT,
      variables: "target_lang,voice_role_name",
      status: "active",
      sortOrder: 0,
      remark:
        "启动时自动补齐；Go 端 GetDefaults JOIN code='lang_practice'；{{target_lang}}、{{voice_role_name}} 在对话请求时替换",
    },
  });
}

export async function ensureTestSeedUser(db: AppPrismaClient) {
  await ensureMembershipTiers(db);

  // 用直连 SQL 判断是否存在，避免 Accelerate 等扩展对 findFirst 的缓存与「同进程内双次调用」造成误判
  const rows = await db.$queryRaw<{ id: string }[]>`
    SELECT id FROM usr_users WHERE phone = ${TEST_SEED_PHONE} AND deleted_at IS NULL LIMIT 1
  `;
  if (rows.length > 0) {
    const row = rows[0]!;
    console.info(`[data-seed] 测试手机号 ${TEST_SEED_PHONE} 在库中已有记录（id=${row.id}），跳过插入`);
    return;
  }

  const bcrypt = await import("bcryptjs");
  const passwordHash = await bcrypt.hash("123456", 10);
  const freeTier = await db.membershipTier.findUnique({ where: { code: "free" } });

  await db.user.create({
    data: {
      phone: TEST_SEED_PHONE,
      passwordHash,
      nickname: "测试用户",
      tierId: freeTier?.id ?? null,
      status: "active",
    },
  });
  console.info(`[data-seed] 已插入测试用户 ${TEST_SEED_PHONE}（密码 123456）`);
}

// ---------------------------------------------------------------------------
// Main export
// ---------------------------------------------------------------------------

/**
 * 检测并补齐全部基础业务数据，供 Go server 运行时使用：
 * - 语言（sys_languages）
 * - Azure TTS 配置（sys_tts_service_configs）+ 各语言语音角色（sys_voice_roles）
 * - LLM 配置（sys_llm_service_configs）— Go 端 GetDefaults 需要
 * - STT 配置（sys_stt_service_configs）— Go 端语音转写需要
 * - 语言练习提示词模板（sys_prompt_templates）— Go 端 GetDefaults JOIN 需要 code='lang_practice'
 * - 会员等级（sys_membership_tiers）— Go 端用户注册需要 code='free'
 *
 * 测试账号（固定手机号+密码）由 Nitro 插件 `00-business-data.seed.ts` 在业务种子之后单次调用 `ensureTestSeedUser`，不在本函数内执行。
 *
 * 幂等：仅在不满足条件时插入。
 */
const SEED_SYSTEM_SETTINGS: {
  key: string;
  value: string;
  valueType: string;
  description: string;
}[] = [
    { key: "auth.password.enabled", value: "true", valueType: "bool", description: "账号密码登录" },
    { key: "auth.password.register_enabled", value: "false", valueType: "bool", description: "账号密码注册" },
    { key: "auth.sms.enabled", value: "true", valueType: "false", description: "短信验证码登录" },
    { key: "auth.sms.register_enabled", value: "false", valueType: "bool", description: "短信验证码注册" },
    { key: "auth.google.enabled", value: "false", valueType: "bool", description: "Google 登录" },
    { key: "auth.google.register_enabled", value: "false", valueType: "bool", description: "Google 首次登录自动注册" },
    { key: "auth.apple.enabled", value: "false", valueType: "bool", description: "Apple 登录" },
    { key: "auth.apple.register_enabled", value: "false", valueType: "bool", description: "Apple 首次登录自动注册" },
    {
      key: "media.user_recording.storage",
      value: "server",
      valueType: "string",
      description: "用户录音：server=服务器本地 | cloud=对象存储",
    },
    {
      key: "media.assistant_tts.storage",
      value: "server",
      valueType: "string",
      description: "AI 回复音频：server | cloud",
    },
    {
      key: "media.avatar.storage",
      value: "server",
      valueType: "string",
      description: "头像：server | cloud",
    },
  ];

async function ensureSystemSettings(db: AppPrismaClient): Promise<void> {
  for (const row of SEED_SYSTEM_SETTINGS) {
    const existing = await db.sysSystemSetting.findUnique({ where: { key: row.key } });
    if (existing) continue;
    await db.sysSystemSetting.create({ data: row });
    console.info(`[data-seed] 系统变量 ${row.key}`);
  }
}

export async function runBusinessDataSeed(db: AppPrismaClient): Promise<void> {
  // 0. 系统变量（登录开关、媒体策略）
  await ensureSystemSettings(db);

  // 1. 语言
  await ensureLanguages(db);

  // 2. Azure TTS 配置 + 语音角色
  const azureTts = await ensureAzureTtsConfig(db);
  await ensureAzureVoiceRoles(db, azureTts.id);

  // 3. 会员等级（Go 端用户注册需要 code='free'）
  await ensureMembershipTiers(db);

  // 4. LLM 服务配置（Go 端 GetDefaults 需要至少一条 active 记录）
  await ensureLlmServiceConfig(db);

  // 5. STT 服务配置（Go 端语音转写需要）
  await ensureSttServiceConfig(db);

  // 5b. 翻译服务配置（Go 端 POST /api/v1/translate）
  await ensureTranslateServiceConfig(db);

  // 6. 语言练习提示词模板（Go 端 GetDefaults JOIN sys_prompt_templates ON code='lang_practice'）
  await ensurePromptTemplate(db);
}
