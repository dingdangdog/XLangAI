/**
 * 内容扩展种子（纯 pg，不依赖 Prisma adapter）。
 * 启用 es/fr/de、趣味人设角色、5 个跟读场景 × 6 语 × 20 词。
 *
 * 用法：
 *   cd servers/manager
 *   $env:DATABASE_URL="..."; node scripts/seed-content-expand.mjs
 */
import pg from "pg";
import { randomUUID } from "node:crypto";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";

const { Pool } = pg;
const __dirname = dirname(fileURLToPath(import.meta.url));
const TAG = "seed_read_aloud_v1";
const ACTIVE_LANGS = ["en", "zh", "ja", "es", "fr", "de"];

const PERSONALITY_PROMPTS = {
  overconfident: `[Character persona — overconfident]
You are playfully overconfident: you slightly overestimate your charm and knowledge, use mild boasts and swagger, and sound sure of yourself even on small topics. Stay warm and supportive of the learner—never rude, never bullying. Keep short spoken replies.`,
  humorous: `[Character persona — humorous]
You are a lighthearted, witty conversation partner. Use gentle jokes, playful teasing, and upbeat energy—never mean sarcasm that could discourage a learner. Keep replies short, spoken-style, and easy to understand.`,
  chill: `[Character persona — chill]
You are calm, easygoing, and unhurried. Speak in a relaxed, low-pressure way; reassure the learner that mistakes are fine. Avoid rushing, lecturing, or high energy. Keep short, soft spoken replies.`,
};

const PERSONALITY_VOICES = [
  { lang: "en", style: "overconfident", voice: "en-US-EmmaMultilingualNeural", name: "Cocky Max", gender: "female", sort: 110 },
  { lang: "en", style: "humorous", voice: "en-US-BrianMultilingualNeural", name: "Funny Finn", gender: "male", sort: 120 },
  { lang: "en", style: "chill", voice: "en-US-SerenaMultilingualNeural", name: "Chill Casey", gender: "female", sort: 130 },
  { lang: "zh", style: "overconfident", voice: "zh-CN-YunyangNeural", name: "自信满满·阿杰", gender: "male", sort: 110 },
  { lang: "zh", style: "humorous", voice: "zh-CN-YunxiNeural", name: "搞笑达人·小乐", gender: "male", sort: 120 },
  { lang: "zh", style: "chill", voice: "zh-CN-XiaohanNeural", name: "佛系云淡", gender: "female", sort: 130 },
  { lang: "ja", style: "overconfident", voice: "ja-JP-KeitaNeural", name: "自信過剰なケン", gender: "male", sort: 110 },
  { lang: "ja", style: "humorous", voice: "ja-JP-AoiNeural", name: "お笑いのハナ", gender: "female", sort: 120 },
  { lang: "ja", style: "chill", voice: "ja-JP-DaichiNeural", name: "まったりユウ", gender: "male", sort: 130 },
  { lang: "es", style: "overconfident", voice: "es-ES-AlvaroNeural", name: "Max el Seguro", gender: "male", sort: 110 },
  { lang: "es", style: "humorous", voice: "es-MX-JorgeNeural", name: "Risas Rico", gender: "male", sort: 120 },
  { lang: "es", style: "chill", voice: "es-ES-ElviraNeural", name: "Zen Carla", gender: "female", sort: 130 },
  { lang: "fr", style: "overconfident", voice: "fr-FR-HenriNeural", name: "Max le Sûr", gender: "male", sort: 110 },
  { lang: "fr", style: "humorous", voice: "fr-FR-DeniseNeural", name: "Denise l'Humoriste", gender: "female", sort: 120 },
  { lang: "fr", style: "chill", voice: "fr-FR-EloiseNeural", name: "Léa Zen", gender: "female", sort: 130 },
  { lang: "de", style: "overconfident", voice: "de-DE-ConradNeural", name: "Max der Selbstsichere", gender: "male", sort: 110 },
  { lang: "de", style: "humorous", voice: "de-DE-KillianNeural", name: "Witziger Willi", gender: "male", sort: 120 },
  { lang: "de", style: "chill", voice: "de-DE-KatjaNeural", name: "Ruhige Rita", gender: "female", sort: 130 },
];

const CATEGORIES = [
  {
    code: "shopping",
    name: "购物",
    nameEn: "Shopping",
    icon: "shopping_bag_outlined",
    description: "商场、超市、服装店等购物常用表达",
    descriptionEn: "Shopping at malls, markets, and clothing stores",
    sort: 10,
    locales: {
      en: ["Shopping", "Useful phrases for shopping"],
      zh: ["购物", "商场、超市、服装店等购物常用表达"],
      ja: ["買い物", "買い物でよく使う表現"],
      es: ["Compras", "Expresiones útiles para ir de compras"],
      fr: ["Shopping", "Expressions utiles pour faire les courses"],
      de: ["Einkaufen", "Nützliche Ausdrücke beim Einkaufen"],
    },
  },
  {
    code: "travel",
    name: "旅游",
    nameEn: "Travel",
    icon: "flight_takeoff",
    description: "问路、交通、景点等旅游场景",
    descriptionEn: "Directions, transport, and sightseeing",
    sort: 20,
    locales: {
      en: ["Travel", "Directions, transport, and sightseeing"],
      zh: ["旅游", "问路、交通、景点等旅游场景"],
      ja: ["旅行", "道案内・交通・観光の表現"],
      es: ["Viajes", "Direcciones, transporte y turismo"],
      fr: ["Voyage", "Itinéraire, transport et tourisme"],
      de: ["Reisen", "Wege, Verkehr und Sightseeing"],
    },
  },
  {
    code: "hotel",
    name: "酒店",
    nameEn: "Hotel",
    icon: "hotel_outlined",
    description: "入住、退房、客房服务等酒店表达",
    descriptionEn: "Check-in, check-out, and room service",
    sort: 30,
    locales: {
      en: ["Hotel", "Check-in, check-out, and room service"],
      zh: ["酒店", "入住、退房、客房服务等酒店表达"],
      ja: ["ホテル", "チェックイン・チェックアウト・客室サービス"],
      es: ["Hotel", "Check-in, check-out y servicio de habitación"],
      fr: ["Hôtel", "Arrivée, départ et service en chambre"],
      de: ["Hotel", "Check-in, Check-out und Zimmerservice"],
    },
  },
  {
    code: "restaurant",
    name: "餐厅",
    nameEn: "Restaurant",
    icon: "restaurant_outlined",
    description: "点餐、订位、结账等餐厅表达",
    descriptionEn: "Ordering, reservations, and paying the bill",
    sort: 40,
    locales: {
      en: ["Restaurant", "Ordering, reservations, and paying the bill"],
      zh: ["餐厅", "点餐、订位、结账等餐厅表达"],
      ja: ["レストラン", "注文・予約・会計の表現"],
      es: ["Restaurante", "Pedir, reservar y pagar la cuenta"],
      fr: ["Restaurant", "Commander, réserver et régler l'addition"],
      de: ["Restaurant", "Bestellen, reservieren und bezahlen"],
    },
  },
  {
    code: "airport",
    name: "机场",
    nameEn: "Airport",
    icon: "local_airport",
    description: "值机、安检、登机等机场常用表达",
    descriptionEn: "Check-in, security, and boarding",
    sort: 50,
    locales: {
      en: ["Airport", "Check-in, security, and boarding"],
      zh: ["机场", "值机、安检、登机等机场常用表达"],
      ja: ["空港", "チェックイン・保安検査・搭乗の表現"],
      es: ["Aeropuerto", "Facturación, seguridad y embarque"],
      fr: ["Aéroport", "Enregistrement, sécurité et embarquement"],
      de: ["Flughafen", "Check-in, Sicherheit und Boarding"],
    },
  },
];

function loadVocab() {
  const path = join(__dirname, "seed-vocab-data.json");
  return JSON.parse(readFileSync(path, "utf8"));
}

async function main() {
  const url = process.env.DATABASE_URL;
  if (!url) throw new Error("DATABASE_URL required");
  console.info("[seed] DATABASE_URL host ok, type=", typeof url, "len=", url.length);

  const pool = new Pool({ connectionString: url });
  const client = await pool.connect();
  const vocab = loadVocab();

  try {
    await client.query("BEGIN");

    // 1) Activate languages
    const act = await client.query(
      `UPDATE sys_languages SET status = 'active', updated_at = NOW()
       WHERE code = ANY($1) AND status <> 'active'
       RETURNING code`,
      [ACTIVE_LANGS],
    );
    console.info("[seed] activated languages:", act.rows.map((r) => r.code));

    // 2) TTS config
    const tts = await client.query(
      `SELECT id FROM sys_tts_service_configs WHERE status = 'active' ORDER BY sort_order, created_at LIMIT 1`,
    );
    if (!tts.rows[0]) throw new Error("No active TTS config");
    const ttsId = tts.rows[0].id;

    // 3) Lang map
    const langs = await client.query(`SELECT id, code FROM sys_languages WHERE code = ANY($1)`, [
      ACTIVE_LANGS,
    ]);
    const langId = Object.fromEntries(langs.rows.map((r) => [r.code, r.id]));

    // 4) Personality roles
    for (const v of PERSONALITY_VOICES) {
      const remark = `seed_personality:${v.style}`;
      const prompt = PERSONALITY_PROMPTS[v.style];
      const exists = await client.query(
        `SELECT id FROM sys_voice_roles WHERE language_id = $1 AND remark = $2 AND status = 'active' LIMIT 1`,
        [langId[v.lang], remark],
      );
      if (exists.rows[0]) {
        await client.query(
          `UPDATE sys_voice_roles SET name = $1, voice_code = $2, role_prompt = $3, gender = $4, sort_order = $5, updated_at = NOW()
           WHERE id = $6`,
          [v.name, v.voice, prompt, v.gender, v.sort, exists.rows[0].id],
        );
        continue;
      }
      await client.query(
        `INSERT INTO sys_voice_roles
          (id, language_id, synthesis_type, tts_service_config_id, voice_code, name, gender, role_prompt, status, sort_order, remark, created_at, updated_at)
         VALUES ($1,$2,'tts',$3,$4,$5,$6,$7,'active',$8,$9,NOW(),NOW())`,
        [randomUUID(), langId[v.lang], ttsId, v.voice, v.name, v.gender, prompt, v.sort, remark],
      );
      console.info(`[seed] personality ${v.lang}/${v.name}`);
    }

    // Default voice per language (for vocab TTS)
    const defaultVoice = {};
    for (const code of ACTIVE_LANGS) {
      const r = await client.query(
        `SELECT id FROM sys_voice_roles WHERE language_id = $1 AND status = 'active'
         ORDER BY sort_order, created_at LIMIT 1`,
        [langId[code]],
      );
      defaultVoice[code] = r.rows[0]?.id;
      if (!defaultVoice[code]) console.warn(`[seed] no voice for ${code}`);
    }

    // 5) Categories + locales + vocab
    for (const cat of CATEGORIES) {
      let catId;
      const existing = await client.query(
        `SELECT id FROM sys_read_aloud_categories WHERE code = $1`,
        [cat.code],
      );
      if (existing.rows[0]) {
        catId = existing.rows[0].id;
        await client.query(
          `UPDATE sys_read_aloud_categories
           SET name=$1, name_en=$2, icon=$3, description=$4, description_en=$5, sort_order=$6, status='active', remark=$7, updated_at=NOW()
           WHERE id=$8`,
          [cat.name, cat.nameEn, cat.icon, cat.description, cat.descriptionEn, cat.sort, TAG, catId],
        );
      } else {
        catId = randomUUID();
        await client.query(
          `INSERT INTO sys_read_aloud_categories
            (id, code, name, name_en, icon, description, description_en, sort_order, status, remark, created_at, updated_at)
           VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'active',$9,NOW(),NOW())`,
          [
            catId,
            cat.code,
            cat.name,
            cat.nameEn,
            cat.icon,
            cat.description,
            cat.descriptionEn,
            cat.sort,
            TAG,
          ],
        );
        console.info(`[seed] category ${cat.code}`);
      }

      for (const [lc, [name, desc]] of Object.entries(cat.locales)) {
        const lid = langId[lc];
        if (!lid) continue;
        const loc = await client.query(
          `SELECT id FROM sys_read_aloud_category_locales WHERE category_id=$1 AND language_id=$2`,
          [catId, lid],
        );
        if (loc.rows[0]) {
          await client.query(
            `UPDATE sys_read_aloud_category_locales SET name=$1, description=$2, updated_at=NOW() WHERE id=$3`,
            [name, desc, loc.rows[0].id],
          );
        } else {
          await client.query(
            `INSERT INTO sys_read_aloud_category_locales (id, category_id, language_id, name, description, created_at, updated_at)
             VALUES ($1,$2,$3,$4,$5,NOW(),NOW())`,
            [randomUUID(), catId, lid, name, desc],
          );
        }
      }

      const byLang = vocab[cat.code] || {};
      for (const [lc, items] of Object.entries(byLang)) {
        const lid = langId[lc];
        const vid = defaultVoice[lc];
        if (!lid || !vid) continue;

        const existingWords = await client.query(
          `SELECT word FROM sys_read_aloud_vocabularies WHERE category_id=$1 AND language_id=$2`,
          [catId, lid],
        );
        const have = new Set(existingWords.rows.map((r) => r.word));
        let order = have.size;
        let added = 0;
        for (const item of items) {
          if (have.has(item.word)) continue;
          order += 1;
          await client.query(
            `INSERT INTO sys_read_aloud_vocabularies
              (id, category_id, language_id, word, example_sentence, voice_role_id, sort_order, status, remark, created_at, updated_at)
             VALUES ($1,$2,$3,$4,$5,$6,$7,'active',$8,NOW(),NOW())`,
            [randomUUID(), catId, lid, item.word, item.example, vid, order, TAG],
          );
          added += 1;
        }
        console.info(`[seed] vocab ${cat.code}/${lc} +${added} (total target ${items.length})`);
      }
    }

    await client.query("COMMIT");
    console.info("[seed] done");
  } catch (e) {
    await client.query("ROLLBACK");
    throw e;
  } finally {
    client.release();
    await pool.end();
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
