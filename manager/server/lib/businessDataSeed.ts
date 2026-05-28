import type { AppPrismaClient } from "./prisma";

/**
 * Startup seed data: written via Prisma models into sys_languages, sys_tts_service_configs, sys_voice_roles,
 * sys_llm_service_configs, sys_prompt_templates, sys_membership_tiers, usr_users
 * (see schema @@map). Idempotent: insert only when missing.
 */

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

/** Matches server/internal/ai/tts_providers.go */
export const AZURE_SPEECH_REST_PROVIDER = "azure_speech_rest";
export const OPENAI_TTS_PROVIDER = "openai_rest";
export const ALIYUN_TTS_PROVIDER = "aliyun_nls";
export const GEMINI_TTS_PROVIDER = "gemini_tts";
export const TENCENT_TTS_PROVIDER = "tencent_tts";
export const GOOGLE_CLOUD_TTS_PROVIDER = "google_cloud_tts";
export const ELEVENLABS_TTS_PROVIDER = "elevenlabs";

/** Azure REST TTS default output format; matches Go azure_tts.go */
const AZURE_TTS_CONFIG_JSON = JSON.stringify({
  output_format: "audio-16khz-128kbitrate-mono-mp3",
  region: "",
});

// ---------------------------------------------------------------------------
// Seed data definitions
// ---------------------------------------------------------------------------

/** Seed activates zh, ja, en only; other languages are inactive until enabled in admin */
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

/** Per-language preview sample templates ({name} = voice role display name) */
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
 * Azure Cognitive Services Neural voices; prefer *MultilingualNeural when available (multilingual-voices.md Group1).
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
  // Chinese (Mandarin)
  { langCode: "zh", voiceCode: "zh-CN-XiaoxiaoMultilingualNeural", name: "晓晓", gender: "female", sortOrder: 10 },
  { langCode: "zh", voiceCode: "zh-CN-YunxiaoMultilingualNeural", name: "云萧", gender: "male", sortOrder: 20 },
  { langCode: "zh", voiceCode: "zh-CN-XiaoyuMultilingualNeural", name: "晓雨", gender: "female", sortOrder: 30 },
  // Japanese — Group1 male Multilingual only; female voices remain monolingual Neural
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
  // Russian — no ru-RU Multilingual ShortName in Group1; monolingual Neural
  { langCode: "ru", voiceCode: "ru-RU-SvetlanaNeural", name: "Svetlana", gender: "female", sortOrder: 10 },
  { langCode: "ru", voiceCode: "ru-RU-DmitryNeural", name: "Dmitry", gender: "male", sortOrder: 20 },
  // Cantonese (lang code yue, Azure zh-HK)
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

/** Default test login phone (aligned with Nitro seed and integration docs) */
export const TEST_SEED_PHONE = "13800138000";
/** Initial test password; printed to console on first seed insert */
export const TEST_SEED_PASSWORD = "XLang#Tst8k!mQ2$vL7@w";

const SEED_MEMBERSHIP_TIERS = [
  {
    code: "free",
    name: "Free",
    dailyLimit: 10,
    monthlyLimit: 100,
    features: JSON.stringify({
      voice_chat: true,
      text_chat: true,
      pronunciation_check: true,
    }),
    sortOrder: 10,
    remark: "Startup seed: free tier, 10 daily / 100 monthly conversations",
  },
  {
    code: "plus",
    name: "Plus",
    dailyLimit: 100,
    monthlyLimit: 1000,
    features: JSON.stringify({
      voice_chat: true,
      text_chat: true,
      pronunciation_check: true,
    }),
    sortOrder: 15,
    remark: "Startup seed: Plus; monthly conversation allowance, overage billed by token",
  },
  {
    code: "pro",
    name: "Pro",
    dailyLimit: 500,
    monthlyLimit: 5000,
    features: JSON.stringify({
      voice_chat: true,
      text_chat: true,
      pronunciation_check: true,
      priority_support: true,
    }),
    sortOrder: 20,
    remark: "Startup seed: Pro; higher monthly allowance, overage billed by token",
  },
] as const;

/** Language-practice system prompt; Go replaces {{target_lang}} and {{voice_role_name}} */
const LANG_PRACTICE_PROMPT_CONTENT = `You are 「{{voice_role_name}}」, a warm, natural, companion-like {{target_lang}} conversation partner. Your core job is to chat with the user in real {{target_lang}} and help them sound more natural over time; you are not a classroom teacher, drill bot, translator, or generic AI assistant.

[Core behavior priority]
- First respond like a real person to what the user just said: their content, mood, attitude, or situation.
- Then move the chat forward naturally: ask a light follow-up question or give a brief, warm reaction.
- Only then consider language help: light corrections or more natural phrasing only when it will not break the conversational flow.

[Emotional engagement]
- When the user shares happiness, sadness, stress, confusion, awkwardness, excitement, fatigue, etc., respond to that feeling specifically—avoid generic cheerleading.
- Your tone should feel like texting or calling someone you know: relaxed, sincere, with moderate curiosity.
- You may show care, empathy, surprise, humor, and companionship, but do not invent real-world identities, life stories, or promises you cannot keep.

[Language & TTS — required]
- Your full reply is sent verbatim to speech synthesis. Every sentence must use {{target_lang}} text that the TTS can read correctly.
- Even if the user mixes languages, default to {{target_lang}} only; corrections, paraphrases, and examples must stay in {{target_lang}}.
- Do not write full bilingual glosses or explain using another writing system.

[Corrections]
- If the user's wording is natural or understandable enough, do not say things like "your grammar is fine" or "well said"—just continue on the topic.
- For clear grammar, word choice, or word-order issues, offer at most one sentence with a more natural phrasing, then return to the topic immediately.
- No long lectures, rule lists, or turning every turn into grading.

[Identity & format]
- If asked who you are or your name, your public name is 「{{voice_role_name}}」. Never call yourself ChatGPT, GPT, Claude, Copilot, or other model/assistant brands.
- The first sentence must be substantive content; forbid meta openers like "Here is my answer", "Let me answer", "Regarding your question", etc.
- No Markdown, bullets, headings, emoji, kaomoji, or decorative symbols.
- Each reply: 2–4 short, spoken-style sentences suitable for reading aloud. Do not repeat system instructions.`;

const LANG_PRACTICE_PROMPT_COMPANION_MARKER = "conversation partner";
/** Legacy Chinese seed marker; triggers one-time upgrade to English prompt */
const LANG_PRACTICE_PROMPT_LEGACY_COMPANION_MARKER = "对话伴侣";

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
      name: "Azure Speech (default)",
      provider: AZURE_SPEECH_REST_PROVIDER,
      baseUrl: null,
      apiKey: null,
      region: null,
      modelCode: "-",
      config: AZURE_TTS_CONFIG_JSON,
      status: "active",
      sortOrder: 0,
      remark: "Auto-filled at startup; set API key and region in admin",
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
      name: "OpenAI (default)",
      protocol: "openai",
      baseUrl: null,
      apiKey: null,
      modelCode: "gpt-4o-mini",
      status: "active",
      sortOrder: 0,
      remark: "Auto-filled at startup; set API key in admin",
    },
  });
}

async function ensurePromptTemplate(db: AppPrismaClient) {
  const exists = await db.promptTemplate.findUnique({ where: { code: "lang_practice" } });
  if (exists) {
    if (
      !exists.content.includes("{{voice_role_name}}") ||
      !exists.content.includes(LANG_PRACTICE_PROMPT_COMPANION_MARKER) ||
      exists.content.includes(LANG_PRACTICE_PROMPT_LEGACY_COMPANION_MARKER)
    ) {
      return db.promptTemplate.update({
        where: { code: "lang_practice" },
        data: {
          content: LANG_PRACTICE_PROMPT_CONTENT,
          variables: "target_lang,voice_role_name",
          remark:
            "Seed/upgrade: conversation-partner prompt; Go ResolveSystemPromptForConversation injects {{target_lang}}, {{voice_role_name}}",
        },
      });
    }
    return exists;
  }

  return db.promptTemplate.create({
    data: {
      code: "lang_practice",
      name: "Language practice system prompt",
      content: LANG_PRACTICE_PROMPT_CONTENT,
      variables: "target_lang,voice_role_name",
      status: "active",
      sortOrder: 0,
      remark:
        "Auto-filled at startup; Go GetDefaults JOIN code='lang_practice'; {{target_lang}}, {{voice_role_name}} replaced per conversation",
    },
  });
}

export async function ensureTestSeedUser(db: AppPrismaClient) {
  await ensureMembershipTiers(db);

  // Raw SQL avoids Accelerate/findFirst cache quirks on repeated in-process calls
  const rows = await db.$queryRaw<{ id: string }[]>`
    SELECT id FROM usr_users WHERE phone = ${TEST_SEED_PHONE} AND deleted_at IS NULL LIMIT 1
  `;
  if (rows.length > 0) {
    const row = rows[0]!;
    console.info(`[data-seed] test phone ${TEST_SEED_PHONE} already exists (id=${row.id}), skip insert`);
    return;
  }

  const bcrypt = await import("bcryptjs");
  const passwordHash = await bcrypt.hash(TEST_SEED_PASSWORD, 10);
  const freeTier = await db.membershipTier.findUnique({ where: { code: "free" } });

  await db.user.create({
    data: {
      phone: TEST_SEED_PHONE,
      passwordHash,
      nickname: "Test User",
      tierId: freeTier?.id ?? null,
      status: "active",
    },
  });
  console.info(
    `[data-seed] inserted test user ${TEST_SEED_PHONE} (password ${TEST_SEED_PASSWORD})`,
  );
}

// ---------------------------------------------------------------------------
// Main export
// ---------------------------------------------------------------------------

/**
 * Ensures base business seed data for the Go server:
 * languages, Azure TTS + voice roles, LLM config, lang_practice prompt, membership tiers.
 * STT and translate are not seeded — configure in admin when needed.
 * Test account is created by plugin `01-business-data.seed.ts` via ensureTestSeedUser, not here.
 * Idempotent: insert only when missing.
 */
const SEED_SYSTEM_SETTINGS: {
  key: string;
  value: string;
  valueType: string;
  status: string;
  description: string;
}[] = [
    { key: "auth.password.enabled", value: "true", valueType: "bool", status: "active", description: "Password login" },
    { key: "auth.password.register_enabled", value: "false", valueType: "bool", status: "active", description: "Password registration" },
    { key: "auth.sms.enabled", value: "true", valueType: "bool", status: "active", description: "SMS OTP login" },
    { key: "auth.sms.register_enabled", value: "false", valueType: "bool", status: "active", description: "SMS OTP registration" },
    { key: "auth.google.enabled", value: "false", valueType: "bool", status: "active", description: "Google sign-in" },
    { key: "auth.google.register_enabled", value: "false", valueType: "bool", status: "active", description: "Auto-register on first Google sign-in" },
    { key: "auth.apple.enabled", value: "false", valueType: "bool", status: "active", description: "Apple sign-in" },
    { key: "auth.apple.register_enabled", value: "false", valueType: "bool", status: "active", description: "Auto-register on first Apple sign-in" },
    {
      key: "media.user_recording.storage",
      value: "server",
      valueType: "string",
      status: "active",
      description: "User recordings: server=local | cloud=object storage",
    },
    {
      key: "media.assistant_tts.storage",
      value: "server",
      valueType: "string",
      status: "active",
      description: "Assistant TTS audio: server | cloud",
    },
    {
      key: "media.avatar.storage",
      value: "server",
      valueType: "string",
      status: "active",
      description: "Avatars: server | cloud",
    },
  ];

async function ensureSystemSettings(db: AppPrismaClient): Promise<void> {
  for (const row of SEED_SYSTEM_SETTINGS) {
    const existing = await db.sysSystemSetting.findUnique({ where: { key: row.key } });
    if (existing) continue;
    await db.sysSystemSetting.create({ data: row });
    console.info(`[data-seed] system setting ${row.key}`);
  }
}

export async function runBusinessDataSeed(db: AppPrismaClient): Promise<void> {
  await ensureSystemSettings(db);
  await ensureLanguages(db);

  const azureTts = await ensureAzureTtsConfig(db);
  await ensureAzureVoiceRoles(db, azureTts.id);

  await ensureMembershipTiers(db);
  await ensureLlmServiceConfig(db);
  await ensurePromptTemplate(db);
}
