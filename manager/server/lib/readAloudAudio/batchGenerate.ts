import { createError } from "h3";
import prisma from "../prisma";
import {
  generateReadAloudVocabularyAudio,
  type GenerateReadAloudAudioResult,
  type ReadAloudAudioPart,
} from "./generate";

export const BATCH_GENERATE_AUDIO_MAX = 30;

export type BatchGenerateAudioInput = {
  /** 显式指定词汇 id；与 categoryId 二选一（优先 ids） */
  ids?: string[];
  categoryId?: string;
  languageId?: string;
  /** 仅处理缺失词/句音频的条目（默认 true） */
  onlyMissing?: boolean;
  part?: ReadAloudAudioPart;
};

export type BatchGenerateAudioFailure = {
  id: string;
  word?: string;
  error: string;
};

export type BatchGenerateAudioResult = {
  totalMatched: number;
  processed: number;
  ok: number;
  failed: BatchGenerateAudioFailure[];
  remaining: number;
  results: GenerateReadAloudAudioResult[];
};

function resolvePartForRow(
  row: { wordAudioUrl: string | null; sentenceAudioUrl: string | null },
  requested: ReadAloudAudioPart,
  onlyMissing: boolean,
): ReadAloudAudioPart | null {
  if (!onlyMissing) return requested;

  const needWord =
    !String(row.wordAudioUrl ?? "").trim() &&
    (requested === "word" || requested === "both");
  const needSentence =
    !String(row.sentenceAudioUrl ?? "").trim() &&
    (requested === "sentence" || requested === "both");

  if (!needWord && !needSentence) return null;
  if (needWord && needSentence) return "both";
  if (needWord) return "word";
  return "sentence";
}

export async function batchGenerateReadAloudVocabularyAudio(
  input: BatchGenerateAudioInput,
): Promise<BatchGenerateAudioResult> {
  const part: ReadAloudAudioPart =
    input.part === "word" || input.part === "sentence" || input.part === "both"
      ? input.part
      : "both";
  const onlyMissing = input.onlyMissing !== false;

  const explicitIds = (input.ids ?? [])
    .map((id) => String(id ?? "").trim())
    .filter(Boolean);

  let candidates: Array<{
    id: string;
    word: string;
    wordAudioUrl: string | null;
    sentenceAudioUrl: string | null;
  }> = [];

  if (explicitIds.length > 0) {
    if (explicitIds.length > BATCH_GENERATE_AUDIO_MAX) {
      throw createError({
        statusCode: 400,
        message: `单次最多处理 ${BATCH_GENERATE_AUDIO_MAX} 条`,
      });
    }
    const rows = await prisma.readAloudVocabulary.findMany({
      where: { id: { in: explicitIds } },
      select: {
        id: true,
        word: true,
        wordAudioUrl: true,
        sentenceAudioUrl: true,
      },
    });
    const byId = new Map(rows.map((r) => [r.id, r]));
    candidates = explicitIds
      .map((id) => byId.get(id))
      .filter((r): r is NonNullable<typeof r> => !!r);
  } else {
    const categoryId = String(input.categoryId ?? "").trim();
    if (!categoryId) {
      throw createError({ statusCode: 400, message: "请指定词汇 ids 或场景 categoryId" });
    }
    const languageId = String(input.languageId ?? "").trim();
    candidates = await prisma.readAloudVocabulary.findMany({
      where: {
        categoryId,
        ...(languageId ? { languageId } : {}),
      },
      orderBy: [{ sortOrder: "asc" }, { createdAt: "asc" }],
      select: {
        id: true,
        word: true,
        wordAudioUrl: true,
        sentenceAudioUrl: true,
      },
    });
  }

  const planned: Array<{ id: string; word: string; part: ReadAloudAudioPart }> = [];
  for (const row of candidates) {
    const resolved = resolvePartForRow(row, part, onlyMissing);
    if (!resolved) continue;
    planned.push({ id: row.id, word: row.word, part: resolved });
  }

  const totalMatched = planned.length;
  const batch = planned.slice(0, BATCH_GENERATE_AUDIO_MAX);
  const remaining = Math.max(0, totalMatched - batch.length);

  const results: GenerateReadAloudAudioResult[] = [];
  const failed: BatchGenerateAudioFailure[] = [];

  for (const item of batch) {
    try {
      const result = await generateReadAloudVocabularyAudio(item.id, item.part);
      results.push(result);
    } catch (e) {
      const msg = e instanceof Error ? e.message : "生成音频失败";
      failed.push({ id: item.id, word: item.word, error: msg });
    }
  }

  return {
    totalMatched,
    processed: batch.length,
    ok: results.length,
    failed,
    remaining,
    results,
  };
}
