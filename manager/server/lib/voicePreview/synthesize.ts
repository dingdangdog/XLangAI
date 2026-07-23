type TtsRow = {
  provider: string;
  baseUrl: string | null;
  apiKey: string | null;
  region: string | null;
  modelCode: string;
  config: string | null;
  voiceCode: string;
};

function parseJsonConfig(raw: string | null): Record<string, unknown> {
  if (!raw?.trim()) return {};
  try {
    return JSON.parse(raw) as Record<string, unknown>;
  } catch {
    return {};
  }
}

function escapeXmlText(s: string): string {
  return s
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&apos;");
}

function localeFromAzureVoice(voice: string): string {
  const parts = voice.split("-");
  if (parts.length >= 2) {
    return `${parts[0]}-${parts[1]}`;
  }
  return "en-US";
}

async function azureTts(
  key: string,
  region: string,
  outputFormat: string,
  voice: string,
  text: string,
): Promise<Buffer> {
  const r = region.trim().toLowerCase();
  if (!key || !r) {
    throw new Error("Azure TTS missing API key or region");
  }
  const endpoint = `https://${r}.tts.speech.microsoft.com/cognitiveservices/v1`;
  const fmt = outputFormat.trim() || "audio-16khz-128kbitrate-mono-mp3";
  const lang = localeFromAzureVoice(voice);
  // Multilingual voices: wrap <lang> so speaking locale is explicit (Azure recommendation).
  const isMultilingual = /MultilingualNeural$/i.test(voice);
  const inner = isMultilingual
    ? `<lang xml:lang="${lang}">${escapeXmlText(text)}</lang>`
    : escapeXmlText(text);
  const ssml = `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xmlns:mstts="https://www.w3.org/2001/mstts" xml:lang="${lang}"><voice name="${voice}">${inner}</voice></speak>`;
  const res = await fetch(endpoint, {
    method: "POST",
    headers: {
      "Ocp-Apim-Subscription-Key": key,
      "Content-Type": "application/ssml+xml",
      "X-Microsoft-OutputFormat": fmt,
      "User-Agent": "xlangai-manager",
    },
    body: ssml,
  });
  if (!res.ok) {
    const body = await res.text();
    const detail = body.trim() || "(empty body)";
    throw new Error(
      `Azure TTS ${res.status} voice=${voice} region=${r}: ${detail.slice(0, 200)}` +
        (res.status === 400
          ? " — 该音色可能在当前区域不可用（preview 音色常见），请换 GA Neural 或支持 preview 的区域"
          : ""),
    );
  }
  return Buffer.from(await res.arrayBuffer());
}

async function openaiTts(
  baseUrl: string,
  apiKey: string,
  model: string,
  voice: string,
  text: string,
): Promise<Buffer> {
  if (!apiKey) {
    throw new Error("OpenAI TTS missing API key");
  }
  const root = (baseUrl.trim() || "https://api.openai.com").replace(/\/$/, "");
  const res = await fetch(`${root}/v1/audio/speech`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${apiKey}`,
    },
    body: JSON.stringify({
      model: model.trim() || "tts-1",
      input: text,
      voice,
    }),
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`OpenAI TTS ${res.status}: ${body.slice(0, 200)}`);
  }
  return Buffer.from(await res.arrayBuffer());
}

/** Manager preview synthesis: azure_speech_rest, openai_rest (matches default seed). */
export async function synthesizePreviewAudio(
  tts: TtsRow,
  text: string,
): Promise<{ data: Buffer; mimeType: string; ext: string }> {
  const provider = (tts.provider ?? "").trim().toLowerCase();
  const ex = parseJsonConfig(tts.config);
  const voice = tts.voiceCode.trim();
  if (!voice) {
    throw new Error("Voice code is empty");
  }

  if (provider === "azure_speech_rest") {
    const region =
      (tts.region ?? "").trim() ||
      String(ex.region ?? "").trim() ||
      String(ex.Region ?? "").trim();
    const outputFormat = String(ex.output_format ?? ex.OutputFormat ?? "").trim();
    const key = (tts.apiKey ?? "").trim();
    const data = await azureTts(key, region, outputFormat, voice, text);
    return { data, mimeType: "audio/mpeg", ext: ".mp3" };
  }

  if (provider === "openai_rest" || provider === "") {
    const model =
      tts.modelCode.trim() && tts.modelCode !== "-" ? tts.modelCode : "tts-1";
    const data = await openaiTts(tts.baseUrl ?? "", tts.apiKey ?? "", model, voice, text);
    return { data, mimeType: "audio/mpeg", ext: ".mp3" };
  }

  throw new Error(
    `Manager preview does not support provider "${provider}"; use azure_speech_rest or openai_rest, or test via Go conversation pipeline`,
  );
}
