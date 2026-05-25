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
    throw new Error("Azure TTS 缺少 API Key 或 region");
  }
  const endpoint = `https://${r}.tts.speech.microsoft.com/cognitiveservices/v1`;
  const fmt = outputFormat.trim() || "audio-16khz-128kbitrate-mono-mp3";
  const lang = localeFromAzureVoice(voice);
  const ssml = `<speak version="1.0" xmlns="http://www.w3.org/2001/10/synthesis" xml:lang="${lang}"><voice name="${voice}">${escapeXmlText(text)}</voice></speak>`;
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
    throw new Error(`Azure TTS ${res.status}: ${body.slice(0, 200)}`);
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
    throw new Error("OpenAI TTS 缺少 API Key");
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

/** 管理端试听合成：当前支持 azure_speech_rest、openai_rest（与默认种子一致）。 */
export async function synthesizePreviewAudio(
  tts: TtsRow,
  text: string,
): Promise<{ data: Buffer; mimeType: string; ext: string }> {
  const provider = (tts.provider ?? "").trim().toLowerCase();
  const ex = parseJsonConfig(tts.config);
  const voice = tts.voiceCode.trim();
  if (!voice) {
    throw new Error("音色代码为空");
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
    `管理端试听暂不支持 provider「${provider}」，请改用 azure_speech_rest 或 openai_rest，或在 Go 对话链路中验证该厂商`,
  );
}
