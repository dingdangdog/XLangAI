import { createError } from "h3";
import prisma from "../prisma";
import { geminiNativeAudioPreview } from "./nativeGemini";
import { resolvePreviewSampleText } from "./sampleText";
import { synthesizePreviewAudio } from "./synthesize";
import { saveAssistantPreviewAudio } from "./storage";
import { applyAssistantTtsLoudness } from "./ttsLoudness";

export type GeneratePreviewResult = {
  voiceRoleId: string;
  previewAudioUrl: string;
  previewLocalFilename: string;
  sampleText: string;
  generatedAt: string;
};

function previewAudioBaseName(args: {
  langCode: string;
  voiceCode: string;
  roleName: string;
}): string {
  return [args.langCode, args.voiceCode, args.roleName]
    .map((part) => part.trim())
    .filter(Boolean)
    .join("-");
}

export async function generateVoiceRolePreview(
  voiceRoleId: string,
): Promise<GeneratePreviewResult> {
  const role = await prisma.voiceRole.findUnique({ where: { id: voiceRoleId } });
  if (!role) {
    throw new Error("Voice role not found");
  }
  const synthesisType = (role.synthesisType?.trim() || "tts") as string;

  if (synthesisType === "native_audio_in_text") {
    throw createError({
      statusCode: 400,
      message: "native_audio_in_text roles have no audio preview; use text chat to verify",
    });
  }

  let langCode = "";
  let langTemplate: string | null = null;
  if (role.languageId) {
    const lang = await prisma.language.findUnique({ where: { id: role.languageId } });
    if (lang) {
      langCode = lang.code;
      langTemplate = lang.previewSampleText;
    }
  }

  const sampleText = resolvePreviewSampleText(langCode, langTemplate, role.name);

  let data: Buffer;
  let mimeType: string;
  let ext: string;

  if (synthesisType === "native_audio_io") {
    if (!role.llmServiceConfigId?.trim()) {
      throw new Error("Voice role has no LLM config for native audio preview");
    }
    const llm = await prisma.sysLlmServiceConfig.findUnique({
      where: { id: role.llmServiceConfigId },
    });
    if (!llm) {
      throw new Error("LLM service config not found");
    }
    const protocol = (llm.protocol ?? "").trim().toLowerCase();
    if (protocol !== "gemini" && protocol !== "google_gemini") {
      throw createError({
        statusCode: 400,
        message: "native_audio_io preview requires Gemini LLM protocol",
      });
    }
    const raw = await geminiNativeAudioPreview(llm, role.voiceCode, sampleText);
    const normalized = await applyAssistantTtsLoudness(raw.data, raw.mimeType);
    data = normalized.data;
    mimeType = normalized.mimeType;
    ext = normalized.ext;
  } else {
    if (!role.ttsServiceConfigId?.trim() || !role.voiceCode.trim()) {
      throw new Error("Voice role has no TTS config or voice code");
    }
    const tts = await prisma.ttsServiceConfig.findUnique({
      where: { id: role.ttsServiceConfigId },
    });
    if (!tts) {
      throw new Error("TTS service config not found");
    }
    const synthesized = await synthesizePreviewAudio(
      {
        provider: tts.provider,
        baseUrl: tts.baseUrl,
        apiKey: tts.apiKey,
        region: tts.region,
        modelCode: tts.modelCode,
        config: tts.config,
        voiceCode: role.voiceCode,
      },
      sampleText,
    );
    const normalized = await applyAssistantTtsLoudness(
      synthesized.data,
      synthesized.mimeType,
    );
    data = normalized.data;
    mimeType = normalized.mimeType;
    ext = normalized.ext;
  }

  const saved = await saveAssistantPreviewAudio(
    data,
    ext,
    mimeType,
    previewAudioBaseName({
      langCode,
      voiceCode: role.voiceCode,
      roleName: role.name,
    }),
  );
  const generatedAt = new Date();

  await prisma.voiceRole.update({
    where: { id: voiceRoleId },
    data: {
      previewAudioUrl: saved.previewAudioUrl,
      previewLocalFilename: saved.previewLocalFilename,
      previewGeneratedAt: generatedAt,
    },
  });

  return {
    voiceRoleId,
    previewAudioUrl: saved.previewAudioUrl,
    previewLocalFilename: saved.previewLocalFilename,
    sampleText,
    generatedAt: generatedAt.toISOString(),
  };
}
