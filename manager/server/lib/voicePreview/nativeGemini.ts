type LlmRow = {
  baseUrl: string | null;
  apiKey: string | null;
  modelCode: string;
  config: string | null;
};

/** Gemini 多模态：文本入 → 音频+文本出（用于 native_audio_io 试听） */
export async function geminiNativeAudioPreview(
  llm: LlmRow,
  voiceCode: string,
  sampleText: string,
): Promise<{ data: Buffer; mimeType: string }> {
  const key = llm.apiKey?.trim();
  if (!key) {
    throw new Error("LLM API key missing for native audio preview");
  }
  const base = (llm.baseUrl?.trim() || "https://generativelanguage.googleapis.com").replace(
    /\/$/,
    "",
  );
  let model = llm.modelCode?.trim() || "gemini-2.5-flash-preview-tts";
  model = model.replace(/^models\//, "");
  const voice = voiceCode.trim() || "Kore";

  const reqBody = {
    contents: [{ parts: [{ text: sampleText }] }],
    generationConfig: {
      responseModalities: ["AUDIO", "TEXT"],
      speechConfig: {
        voiceConfig: {
          prebuiltVoiceConfig: { voiceName: voice },
        },
      },
    },
  };

  const res = await fetch(`${base}/v1beta/models/${model}:generateContent`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-goog-api-key": key,
    },
    body: JSON.stringify(reqBody),
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`Gemini native preview ${res.status}: ${body.slice(0, 300)}`);
  }
  const parsed = (await res.json()) as {
    candidates?: Array<{
      content?: { parts?: Array<{ inlineData?: { mimeType?: string; data?: string } }> };
    }>;
  };
  for (const c of parsed.candidates ?? []) {
    for (const p of c.content?.parts ?? []) {
      if (p.inlineData?.data) {
        const data = Buffer.from(p.inlineData.data, "base64");
        const mimeType = p.inlineData.mimeType?.trim() || "audio/mpeg";
        return { data, mimeType };
      }
    }
  }
  throw new Error("Gemini native preview: no audio in response");
}
