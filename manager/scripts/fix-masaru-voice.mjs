/**
 * 将不可用的 ja-JP-MasaruMultilingualNeural 迁移为 ja-JP-KeitaNeural。
 * 用法：cd servers/manager && node scripts/fix-masaru-voice.mjs
 */
import pg from "pg";
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const envText = readFileSync(join(__dirname, "..", ".env"), "utf8");
for (const line of envText.split(/\r?\n/)) {
  const m = line.match(/^DATABASE_URL=(.*)$/);
  if (!m) continue;
  let v = m[1].trim();
  if (
    (v.startsWith('"') && v.endsWith('"')) ||
    (v.startsWith("'") && v.endsWith("'"))
  ) {
    v = v.slice(1, -1);
  }
  process.env.DATABASE_URL = v;
}

const FROM = "ja-JP-MasaruMultilingualNeural";
const TO = "ja-JP-KeitaNeural";

const pool = new pg.Pool({ connectionString: process.env.DATABASE_URL });
const { rows } = await pool.query(
  `SELECT id, name, voice_code FROM sys_voice_roles WHERE voice_code = $1`,
  [FROM],
);
console.log("found", rows.length, "roles with", FROM);

for (const row of rows) {
  const newName = row.name === "Masaru" ? "Keita" : row.name;
  await pool.query(
    `UPDATE sys_voice_roles
     SET voice_code = $1,
         name = $2,
         preview_audio_url = NULL,
         preview_local_filename = NULL,
         preview_generated_at = NULL,
         updated_at = NOW()
     WHERE id = $3`,
    [TO, newName, row.id],
  );
  console.log("migrated", row.id, row.name, "→", newName, TO);
}

// Verify Keita synthesizes
const { rows: ttsRows } = await pool.query(
  `SELECT api_key, region, config FROM sys_tts_service_configs WHERE provider='azure_speech_rest' LIMIT 1`,
);
const key = ttsRows[0].api_key;
const cfg = ttsRows[0].config ? JSON.parse(ttsRows[0].config) : {};
const region = String(ttsRows[0].region || cfg.region || "")
  .trim()
  .toLowerCase();
const lang = "ja-JP";
const text = "こんにちは、Keitaです";
const ssml = `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="${lang}"><voice name="${TO}">${text}</voice></speak>`;
const res = await fetch(
  `https://${region}.tts.speech.microsoft.com/cognitiveservices/v1`,
  {
    method: "POST",
    headers: {
      "Ocp-Apim-Subscription-Key": key,
      "Content-Type": "application/ssml+xml",
      "X-Microsoft-OutputFormat": "audio-16khz-128kbitrate-mono-mp3",
      "User-Agent": "xlangai-fix",
    },
    body: ssml,
  },
);
console.log("Keita TTS verify status", res.status, "bytes", (await res.arrayBuffer()).byteLength);

await pool.end();
