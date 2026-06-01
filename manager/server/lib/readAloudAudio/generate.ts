import { createError } from "h3";
import prisma from "../prisma";
import { synthesizePreviewAudio } from "../voicePreview/synthesize";
import { saveAssistantPreviewAudio } from "../voicePreview/storage";
import { applyAssistantTtsLoudness } from "../voicePreview/ttsLoudness";

export type ReadAloudAudioPart = "word" | "sentence" | "both";

export type GenerateReadAloudAudioResult = {
  vocabularyId: string;
  part: ReadAloudAudioPart;
  wordAudioUrl?: string | null;
  sentenceAudioUrl?: string | null;
  generatedAt: string;
};

function audioBaseName(args: {
  categoryCode: string;
  langCode: string;
  word: string;
  part: "word" | "sentence";
}): string {
  return ["readaloud", args.categoryCode, args.langCode, args.part, args.word]
    .map((p) => p.trim())
    .filter(Boolean)
    .join("-");
}

async function synthesizeText(voiceRoleId: string, text: string): Promise<{
  data: Buffer;
  mimeType: string;
  ext: string;
}> {
  const role = await prisma.voiceRole.findUnique({ where: { id: voiceRoleId } });
  if (!role) {
    throw new Error("Voice role not found");
  }
  const synthesisType = (role.synthesisType?.trim() || "tts").toLowerCase();
  if (synthesisType !== "tts") {
    throw createError({
      statusCode: 400,
      message: "跟读参考音频仅支持 TTS 类型语音角色",
    });
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
    text,
  );
  const normalized = await applyAssistantTtsLoudness(
    synthesized.data,
    synthesized.mimeType,
  );
  return normalized;
}

export async function generateReadAloudVocabularyAudio(
  vocabularyId: string,
  part: ReadAloudAudioPart = "both",
): Promise<GenerateReadAloudAudioResult> {
  const vocab = await prisma.readAloudVocabulary.findUnique({
    where: { id: vocabularyId },
  });
  if (!vocab) {
    throw new Error("Vocabulary not found");
  }

  const category = await prisma.readAloudCategory.findUnique({
    where: { id: vocab.categoryId },
  });
  if (!category) {
    throw new Error("Category not found");
  }

  let langCode = "";
  if (vocab.languageId) {
    const lang = await prisma.language.findUnique({ where: { id: vocab.languageId } });
    if (lang) langCode = lang.code;
  }

  const generatedAt = new Date();
  const updateData: Record<string, unknown> = { updatedAt: generatedAt };
  let wordAudioUrl = vocab.wordAudioUrl;
  let sentenceAudioUrl = vocab.sentenceAudioUrl;

  const parts: Array<"word" | "sentence"> =
    part === "both" ? ["word", "sentence"] : [part];

  for (const p of parts) {
    const text = p === "word" ? vocab.word.trim() : vocab.exampleSentence.trim();
    if (!text) {
      throw new Error(p === "word" ? "词汇为空" : "例句为空");
    }
    const audio = await synthesizeText(vocab.voiceRoleId, text);
    const saved = await saveAssistantPreviewAudio(
      audio.data,
      audio.ext,
      audio.mimeType,
      audioBaseName({
        categoryCode: category.code,
        langCode,
        word: vocab.word,
        part: p,
      }),
    );
    if (p === "word") {
      wordAudioUrl = saved.previewAudioUrl;
      updateData.wordAudioUrl = saved.previewAudioUrl;
      updateData.wordAudioLocalFilename = saved.previewLocalFilename;
      updateData.wordAudioGeneratedAt = generatedAt;
    } else {
      sentenceAudioUrl = saved.previewAudioUrl;
      updateData.sentenceAudioUrl = saved.previewAudioUrl;
      updateData.sentenceAudioLocalFilename = saved.previewLocalFilename;
      updateData.sentenceAudioGeneratedAt = generatedAt;
    }
  }

  await prisma.readAloudVocabulary.update({
    where: { id: vocabularyId },
    data: updateData,
  });

  return {
    vocabularyId,
    part,
    wordAudioUrl,
    sentenceAudioUrl,
    generatedAt: generatedAt.toISOString(),
  };
}
