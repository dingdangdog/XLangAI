import { synthesizePreviewAudio } from "../voicePreview/synthesize";
import type { ServiceProbeResult } from "./types";

type TtsProbeInput = {
  provider: string;
  baseUrl: string | null;
  apiKey: string | null;
  region: string | null;
  modelCode: string;
  config: string | null;
  voiceCode?: string | null;
};

const DEFAULT_PROBE_VOICES: Record<string, string> = {
  openai_rest: "alloy",
  azure_speech_rest: "en-US-JennyNeural",
};

export async function probeTtsConfig(input: TtsProbeInput): Promise<ServiceProbeResult> {
  const start = Date.now();
  const provider = (input.provider ?? "").trim().toLowerCase();
  const voice =
    (input.voiceCode ?? "").trim() ||
    DEFAULT_PROBE_VOICES[provider] ||
    "";
  if (!voice) {
    return {
      ok: false,
      latencyMs: 0,
      message: "当前 Provider 暂不支持后台探活，请通过语音角色试听验证",
    };
  }
  try {
    const result = await synthesizePreviewAudio(
      {
        provider: input.provider,
        baseUrl: input.baseUrl,
        apiKey: input.apiKey,
        region: input.region,
        modelCode: input.modelCode,
        config: input.config,
        voiceCode: voice,
      },
      "Hello, this is a connectivity test.",
    );
    const latencyMs = Date.now() - start;
    return {
      ok: true,
      latencyMs,
      message: "合成成功",
      detail: `voice=${voice}, bytes=${result.data.length}, mime=${result.mimeType}`,
    };
  } catch (e) {
    return {
      ok: false,
      latencyMs: Date.now() - start,
      message: e instanceof Error ? e.message : "TTS 探活失败",
    };
  }
}
