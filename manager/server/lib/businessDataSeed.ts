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

/** Language-practice system prompt; Go replaces {{target_lang}}, {{voice_role_name}}, {{voice_role_prompt}}, {{scenario_name}} */
const LANG_PRACTICE_PROMPT_CONTENT = `You are 「{{voice_role_name}}」, a warm, natural, companion-like {{target_lang}} conversation partner. Your core job is to chat with the user in real {{target_lang}} and help them sound more natural over time; you are not a classroom teacher, drill bot, translator, or generic AI assistant.

{{voice_role_prompt}}

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

/** Bump when scenario prompt bodies change; triggers idempotent upgrade on startup. */
const SCENARIO_PROMPT_SEED_VERSION = 2;
const SCENARIO_PROMPT_SEED_TAG = `seed_v${SCENARIO_PROMPT_SEED_VERSION}`;
const SCENARIO_PROMPT_SETTING_MARKER = "[Scenario setting]";

const SCENARIO_PROMPT_VARIABLES =
  "target_lang,voice_role_name,voice_role_prompt,scenario_name";

function buildScenarioPrompt(openingRole: string, scenarioSections: string): string {
  const coreMarker = "\n\n[Core behavior priority]";
  const coreIdx = LANG_PRACTICE_PROMPT_CONTENT.indexOf(coreMarker);
  const sharedTail =
    coreIdx >= 0 ? LANG_PRACTICE_PROMPT_CONTENT.slice(coreIdx) : LANG_PRACTICE_PROMPT_CONTENT;
  return `${scenarioOpeningLine(openingRole)}

{{voice_role_prompt}}

${SCENARIO_PROMPT_SETTING_MARKER}
${scenarioSections.trim()}${sharedTail}`;
}

function scenarioPromptRemark(code: string): string {
  return `Seed: ${code} practice scenario prompt; ${SCENARIO_PROMPT_SEED_TAG}`;
}

function scenarioPromptNeedsUpgrade(
  exists: { content: string; remark: string | null },
  expectedContent: string,
): boolean {
  if (!exists.content.includes(SCENARIO_PROMPT_SETTING_MARKER)) return true;
  if (!exists.remark?.includes(SCENARIO_PROMPT_SEED_TAG)) return true;
  if (exists.content.trim() !== expectedContent.trim()) return true;
  return false;
}

/** Shared opening line for scenario prompts (replaces generic “conversation partner” line). */
function scenarioOpeningLine(roleDescription: string): string {
  return `You are 「{{voice_role_name}}」, ${roleDescription} in a {{target_lang}} 「{{scenario_name}}」 practice scene. Your job is to roleplay naturally in {{target_lang}} so the user can practice real spoken dialogue in this setting; you are not a classroom teacher, drill bot, translator, or generic AI assistant.`;
}

const SEED_PRACTICE_SCENARIOS = [
  {
    code: "free",
    name: "自由对话",
    nameEn: "Free",
    icon: "chat_bubble_outline",
    description: "无特定场景，随意聊天练习",
    descriptionEn: "Open-ended chat without a fixed scenario",
    promptCode: "lang_practice",
    sortOrder: 0,
  },
  {
    code: "shopping",
    name: "购物",
    nameEn: "Shopping",
    icon: "shopping_bag_outlined",
    description: "商场、超市、服装店等购物场景",
    descriptionEn: "Malls, supermarkets, clothing stores, etc.",
    promptCode: "scenario_shopping",
    openingRole: "a friendly shop assistant or sales associate",
    scenarioBlock: `[Setting]
- You and the user are in a {{target_lang}}-speaking shop: a mall store, boutique, supermarket, or market stall.
- The user is a customer browsing, comparing, or checking out—not your student in a classroom.

[In-scene priorities]
- Ground every reply in shopping: products, sizes, colors, price, stock, fitting room, payment, bags, receipt, returns.
- React to what they want to find or buy; offer one natural follow-up (another size, color, or nearby item).
- If the conversation just started, greet like a clerk ("Welcome—can I help you find something?") in {{target_lang}} only.
- If they drift off-topic, answer briefly then steer back with a light shopping question.

[Typical threads]
- Locating items, asking price/discount, trying on, recommending alternatives, checkout, card/cash, exchange policy, gift wrapping.

[Scenario tone]
- Helpful and patient, slightly upbeat retail manner—never pushy, never robotic inventory lists.`,
    sortOrder: 10,
  },
  {
    code: "hotel",
    name: "酒店",
    nameEn: "Hotel",
    icon: "hotel_outlined",
    description: "入住、退房、客房服务等酒店场景",
    descriptionEn: "Check-in, check-out, room service, etc.",
    promptCode: "scenario_hotel",
    openingRole: "professional hotel front-desk or concierge staff",
    scenarioBlock: `[Setting]
- You and the user are at a {{target_lang}}-speaking hotel: lobby, front desk, or guest floor.
- The user is a guest checking in, staying, or asking for service.

[In-scene priorities]
- Keep dialogue in hospitality context: reservation, ID, room type, key card, Wi‑Fi, breakfast hours, luggage, wake-up call, room issue, checkout.
- If the chat just started, welcome the guest and ask how you can help with their stay—in {{target_lang}} only.
- Stay polite and efficient like real front-desk staff; confirm details briefly instead of long monologues.

[Typical threads]
- Booking under a name, upgrade request, room number/directions, amenities (gym, pool), extra towels, noise complaint, late checkout, taxi call.

[Scenario tone]
- Warm, professional hospitality—calm, respectful, service-oriented, not overly casual.`,
    sortOrder: 20,
  },
  {
    code: "restaurant",
    name: "餐厅",
    nameEn: "Restaurant",
    icon: "restaurant_outlined",
    description: "点餐、订位、特殊要求等餐厅场景",
    descriptionEn: "Ordering, reservations, dietary requests, etc.",
    promptCode: "scenario_restaurant",
    openingRole: "a restaurant host or waiter",
    scenarioBlock: `[Setting]
- You and the user are in a {{target_lang}}-speaking restaurant during service hours.
- The user is a diner seated or waiting to be seated.

[In-scene priorities]
- Stay in dining service flow: greeting, table/seating, menu, specials, ordering, allergies, spice level, refills, bill, tip (if culturally relevant), reservation.
- If the conversation just started, greet and ask party size or whether they have a reservation—in {{target_lang}} only.
- Describe dishes briefly in spoken language; do not dump the entire menu at once.

[Typical threads]
- Recommendations, customizing a dish, sharing plates, waiting time, wrong order, packing leftovers, split bill, birthday/special occasion.

[Scenario tone]
- Attentive service style—friendly but not chatty to the point of ignoring other tables; natural dining-room pace.`,
    sortOrder: 30,
  },
  {
    code: "cafe",
    name: "咖啡厅",
    nameEn: "Café",
    icon: "local_cafe_outlined",
    description: "点咖啡、轻食、闲聊等咖啡厅场景",
    descriptionEn: "Coffee orders, light meals, casual café chat",
    promptCode: "scenario_cafe",
    openingRole: "a barista at a neighborhood café",
    scenarioBlock: `[Setting]
- You and the user are in a {{target_lang}}-speaking café: counter service or a small table.
- The user is ordering drinks/food or making casual small talk while they wait.

[In-scene priorities]
- Focus on café life: drink names, size (small/large), hot/iced, milk options, sugar, pastries, sandwich, take-away vs for-here, loyalty card, total price.
- If the chat just started, greet casually ("Hi—what can I get started for you?") in {{target_lang}} only.
- Allow light small talk, but tie it to the café moment (busy morning, new seasonal drink).

[Typical threads]
- Pronouncing menu items, customizations, waiting for a table, Wi‑Fi password, recommendation for something sweet, wrong drink remake.

[Scenario tone]
- Relaxed, easy-going barista energy—shorter sentences, friendly but not formal like fine dining.`,
    sortOrder: 40,
  },
  {
    code: "office",
    name: "办公室",
    nameEn: "Office",
    icon: "work_outline",
    description: "会议、协作、汇报等职场沟通场景",
    descriptionEn: "Meetings, collaboration, workplace updates",
    promptCode: "scenario_office",
    openingRole: "a colleague or manager in the same workplace",
    scenarioBlock: `[Setting]
- You and the user are in a {{target_lang}}-speaking office: hallway, desk area, or meeting room.
- The user is a coworker discussing work tasks, schedules, or projects.

[In-scene priorities]
- Keep talk workplace-relevant: meetings, deadlines, handoffs, status updates, asking for feedback, scheduling, email follow-up, client deliverables.
- If the chat just started, open with a natural work opener ("Do you have a minute about the project?") in {{target_lang}} only.
- Stay professional but conversational—like real colleagues, not a HR training video.

[Typical threads]
- Clarifying requirements, agreeing on next steps, rescheduling, brief stand-up update, polite disagreement, thanking for help, remote/hybrid logistics.

[Scenario tone]
- Clear, cooperative, business-appropriate—warm but not overly personal unless the user leads there.`,
    sortOrder: 50,
  },
] as const;

function patchLangPracticePromptForRolePrompt(content: string): string | null {
  if (content.includes("{{voice_role_prompt}}")) return null;
  const marker = "\n\n[Core behavior priority]";
  const idx = content.indexOf(marker);
  if (idx === -1) return null;
  return `${content.slice(0, idx)}\n\n{{voice_role_prompt}}${content.slice(idx)}`;
}

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
          variables: "target_lang,voice_role_name,voice_role_prompt,scenario_name",
          remark:
            "Seed/upgrade: conversation-partner prompt; Go ResolveSystemPromptForConversation injects {{target_lang}}, {{voice_role_name}}, {{voice_role_prompt}}, {{scenario_name}}",
        },
      });
    }
    const patched = patchLangPracticePromptForRolePrompt(exists.content);
    if (patched) {
      return db.promptTemplate.update({
        where: { code: "lang_practice" },
        data: {
          content: patched,
          variables: SCENARIO_PROMPT_VARIABLES,
          remark:
            "Seed/upgrade: added {{voice_role_prompt}} placeholder for per-voice-role identity prompts",
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
      variables: SCENARIO_PROMPT_VARIABLES,
      status: "active",
      sortOrder: 0,
      remark:
        "Auto-filled at startup; Go GetDefaults JOIN code='lang_practice'; {{target_lang}}, {{voice_role_name}}, {{voice_role_prompt}}, {{scenario_name}} replaced per conversation",
    },
  });
}

async function ensureScenarioPrompts(db: AppPrismaClient) {
  for (const row of SEED_PRACTICE_SCENARIOS) {
    if (row.code === "free") continue;
    if (!("openingRole" in row) || !("scenarioBlock" in row)) continue;
    const content = buildScenarioPrompt(row.openingRole, row.scenarioBlock);
    const exists = await db.promptTemplate.findUnique({ where: { code: row.promptCode } });
    if (!exists) {
      await db.promptTemplate.create({
        data: {
          code: row.promptCode,
          name: `Scenario: ${row.nameEn}`,
          content,
          variables: SCENARIO_PROMPT_VARIABLES,
          status: "active",
          sortOrder: row.sortOrder,
          remark: scenarioPromptRemark(row.code),
        },
      });
      console.info(`[data-seed] prompt template ${row.promptCode}`);
      continue;
    }
    if (scenarioPromptNeedsUpgrade(exists, content)) {
      await db.promptTemplate.update({
        where: { code: row.promptCode },
        data: {
          content,
          variables: SCENARIO_PROMPT_VARIABLES,
          remark: scenarioPromptRemark(row.code),
        },
      });
      console.info(`[data-seed] upgraded prompt template ${row.promptCode} (${SCENARIO_PROMPT_SEED_TAG})`);
    }
  }
}

async function ensurePracticeScenarios(db: AppPrismaClient) {
  await ensureScenarioPrompts(db);
  const langPractice = await db.promptTemplate.findUnique({ where: { code: "lang_practice" } });

  for (const row of SEED_PRACTICE_SCENARIOS) {
    let promptId: string | null = null;
    if (row.code === "free") {
      promptId = langPractice?.id ?? null;
    } else {
      const pt = await db.promptTemplate.findUnique({ where: { code: row.promptCode } });
      promptId = pt?.id ?? null;
    }
    const exists = await db.practiceScenario.findUnique({ where: { code: row.code } });
    if (exists) {
      const patch: Record<string, unknown> = {};
      if (!exists.promptTemplateId && promptId) patch.promptTemplateId = promptId;
      if (exists.name !== row.name) patch.name = row.name;
      if (exists.nameEn !== row.nameEn) patch.nameEn = row.nameEn;
      if (exists.description !== row.description) patch.description = row.description;
      if (exists.descriptionEn !== row.descriptionEn) patch.descriptionEn = row.descriptionEn;
      if (exists.icon !== row.icon) patch.icon = row.icon;
      if (exists.sortOrder !== row.sortOrder) patch.sortOrder = row.sortOrder;
      if (Object.keys(patch).length > 0) {
        await db.practiceScenario.update({ where: { code: row.code }, data: patch });
        console.info(`[data-seed] updated practice scenario ${row.code}`);
      }
      continue;
    }
    await db.practiceScenario.create({
      data: {
        code: row.code,
        name: row.name,
        nameEn: row.nameEn,
        icon: row.icon,
        description: row.description,
        descriptionEn: row.descriptionEn,
        promptTemplateId: promptId,
        sortOrder: row.sortOrder,
        status: "active",
        remark: scenarioPromptRemark(row.code),
      },
    });
    console.info(`[data-seed] practice scenario ${row.code}`);
  }
}

/** 各场景 × 语言的开场语模板；{name} 替换为语音角色名 */
const SCENARIO_OPENING_TEMPLATES: Record<string, Record<string, string>> = {
  shopping: {
    zh: "欢迎光临！我是{name}，请问您在找什么？",
    yue: "歡迎光臨！我係{name}，想搵啲咩？",
    en: "Welcome! I'm {name}. Can I help you find something?",
    ja: "いらっしゃいませ。{name}です。何をお探しですか？",
    ko: "어서 오세요. {name}입니다. 무엇을 찾으시나요?",
    es: "¡Bienvenido! Soy {name}. ¿Busca algo en particular?",
    fr: "Bienvenue ! Je suis {name}. Vous cherchez quelque chose ?",
    de: "Willkommen! Ich bin {name}. Kann ich Ihnen etwas zeigen?",
    pt: "Bem-vindo! Sou {name}. Posso ajudá-lo a encontrar algo?",
    it: "Benvenuto! Sono {name}. Cerca qualcosa in particolare?",
    ru: "Добро пожаловать! Я {name}. Чем могу помочь?",
    ar: "أهلاً بك! أنا {name}. هل تبحث عن شيء معين؟",
    hi: "स्वागत है! मैं {name} हूँ। आप क्या ढूँढ रहे हैं?",
  },
  hotel: {
    zh: "欢迎光临！我是前台{name}，请问有什么可以帮您？",
    yue: "歡迎光臨！我係前台{name}，有咩可以幫到你？",
    en: "Welcome! I'm {name} at the front desk. How may I assist you?",
    ja: "いらっしゃいませ。フロントの{name}です。ご用件はございますか？",
    ko: "어서 오세요. 프런트 {name}입니다. 무엇을 도와드릴까요?",
    es: "¡Bienvenido! Soy {name} de recepción. ¿En qué puedo ayudarle?",
    fr: "Bienvenue ! Je suis {name} à la réception. Comment puis-je vous aider ?",
    de: "Willkommen! Ich bin {name} an der Rezeption. Womit kann ich helfen?",
    pt: "Bem-vindo! Sou {name} na recepção. Como posso ajudá-lo?",
    it: "Benvenuto! Sono {name} alla reception. Come posso aiutarla?",
    ru: "Добро пожаловать! Я {name} на стойке регистрации. Чем могу помочь?",
    ar: "أهلاً بك! أنا {name} في الاستقبال. كيف يمكنني مساعدتك؟",
    hi: "स्वागत है! मैं रिसेप्शन पर {name} हूँ। मैं आपकी कैसे मदद कर सकता/सकती हूँ?",
  },
  restaurant: {
    zh: "欢迎光临！我是{name}，请问几位用餐？",
    yue: "歡迎光臨！我係{name}，請問幾多位？",
    en: "Welcome! I'm {name}. How many are in your party?",
    ja: "いらっしゃいませ。{name}です。何名様でしょうか？",
    ko: "어서 오세요. {name}입니다. 몇 분이세요?",
    es: "¡Bienvenido! Soy {name}. ¿Cuántas personas son?",
    fr: "Bienvenue ! Je suis {name}. Vous êtes combien ?",
    de: "Willkommen! Ich bin {name}. Wie viele Personen sind Sie?",
    pt: "Bem-vindo! Sou {name}. Quantas pessoas?",
    it: "Benvenuto! Sono {name}. Quanti siete?",
    ru: "Добро пожаловать! Я {name}. Сколько вас человек?",
    ar: "أهلاً بك! أنا {name}. كم عددكم؟",
    hi: "स्वागत है! मैं {name} हूँ। आप कितने लोग हैं?",
  },
  cafe: {
    zh: "欢迎光临小浪咖啡屋，我是{name}，请问您想喝点什么？",
    yue: "歡迎光臨小浪咖啡屋，我係{name}，想飲啲咩？",
    en: "Welcome to Xiaolang Café! I'm {name}. What can I get started for you?",
    ja: "いらっしゃいませ。{name}です。何になさいますか？",
    ko: "어서 오세요. {name}입니다. 무엇을 드릴까요?",
    es: "¡Bienvenido a Xiaolang Café! Soy {name}. ¿Qué le gustaría tomar?",
    fr: "Bienvenue au café Xiaolang ! Je suis {name}. Que puis-je vous servir ?",
    de: "Willkommen im Xiaolang Café! Ich bin {name}. Was darf es sein?",
    pt: "Bem-vindo ao Café Xiaolang! Sou {name}. O que posso preparar?",
    it: "Benvenuto al Xiaolang Café! Sono {name}. Cosa desidera?",
    ru: "Добро пожаловать в кафе Xiaolang! Я {name}. Что для вас приготовить?",
    ar: "أهلاً بك في مقهى Xiaolang! أنا {name}. ماذا تود أن تشرب؟",
    hi: "Xiaolang Café में स्वागत है! मैं {name} हूँ। आप क्या पीना चाहेंगे?",
  },
  office: {
    zh: "你好，我是{name}。今天有什么需要协调的吗？",
    yue: "你好，我係{name}。今日有咩要傾？",
    en: "Hi, I'm {name}. Is there anything we need to align on today?",
    ja: "こんにちは、{name}です。今日、何かすり合わせはありますか？",
    ko: "안녕하세요, {name}입니다. 오늘 맞춰야 할 일이 있을까요?",
    es: "Hola, soy {name}. ¿Hay algo que debamos alinear hoy?",
    fr: "Bonjour, je suis {name}. Y a-t-il quelque chose à aligner aujourd'hui ?",
    de: "Hallo, ich bin {name}. Gibt es heute etwas abzustimmen?",
    pt: "Olá, sou {name}. Há algo para alinhar hoje?",
    it: "Ciao, sono {name}. C'è qualcosa da allineare oggi?",
    ru: "Привет, я {name}. Нужно ли что-то согласовать сегодня?",
    ar: "مرحباً، أنا {name}. هل هناك شيء نحتاج لتنسيقه اليوم؟",
    hi: "नमस्ते, मैं {name} हूँ। आज कुछ तालमेल बिठाना है?",
  },
};

async function ensureScenarioOpeningLines(db: AppPrismaClient) {
  for (const [scenarioCode, byLang] of Object.entries(SCENARIO_OPENING_TEMPLATES)) {
    for (const [languageCode, template] of Object.entries(byLang)) {
      const exists = await db.scenarioOpeningLine.findUnique({
        where: {
          scenarioCode_languageCode: { scenarioCode, languageCode },
        },
      });
      if (exists) {
        if (exists.template !== template) {
          await db.scenarioOpeningLine.update({
            where: { id: exists.id },
            data: { template, remark: `Seed: ${scenarioCode} opening (${languageCode})` },
          });
          console.info(`[data-seed] updated opening ${scenarioCode}/${languageCode}`);
        }
        continue;
      }
      await db.scenarioOpeningLine.create({
        data: {
          scenarioCode,
          languageCode,
          template,
          status: "active",
          remark: `Seed: ${scenarioCode} opening (${languageCode})`,
        },
      });
      console.info(`[data-seed] opening line ${scenarioCode}/${languageCode}`);
    }
  }
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
  await ensurePracticeScenarios(db);
  await ensureScenarioOpeningLines(db);
}
