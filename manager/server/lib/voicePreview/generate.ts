import prisma from "../prisma";
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
  if (!role.ttsServiceConfigId?.trim() || !role.voiceCode.trim()) {
    throw new Error("Voice role has no TTS config or voice code");
  }

  const tts = await prisma.ttsServiceConfig.findUnique({
    where: { id: role.ttsServiceConfigId },
  });
  if (!tts) {
    throw new Error("TTS service config not found");
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
  const data = normalized.data;
  const mimeType = normalized.mimeType;
  const ext = normalized.ext;

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
