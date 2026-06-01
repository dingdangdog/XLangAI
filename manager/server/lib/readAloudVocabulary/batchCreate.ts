import { createError } from "h3";
import prisma from "../prisma";

export type BatchCreateVocabularyItem = {
  word: string;
  exampleSentence: string;
};

export type BatchCreateVocabularyInput = {
  categoryId: string;
  languageId: string;
  voiceRoleId: string;
  items: BatchCreateVocabularyItem[];
  status?: string;
  remark?: string | null;
};

export type BatchCreateVocabularyResult = {
  created: number;
  ids: string[];
};

export async function batchCreateReadAloudVocabulary(
  input: BatchCreateVocabularyInput,
): Promise<BatchCreateVocabularyResult> {
  const categoryId = input.categoryId.trim();
  const languageId = input.languageId.trim();
  const voiceRoleId = input.voiceRoleId.trim();
  if (!categoryId || !languageId || !voiceRoleId) {
    throw createError({ statusCode: 400, message: "缺少场景、语言或语音角色" });
  }

  const items = (input.items ?? [])
    .map((it) => ({
      word: String(it.word ?? "").trim(),
      exampleSentence: String(it.exampleSentence ?? "").trim(),
    }))
    .filter((it) => it.word && it.exampleSentence);

  if (items.length === 0) {
    throw createError({ statusCode: 400, message: "没有可导入的词汇" });
  }
  if (items.length > 50) {
    throw createError({ statusCode: 400, message: "单次最多导入 50 条" });
  }

  const [category, language, voiceRole] = await Promise.all([
    prisma.readAloudCategory.findUnique({ where: { id: categoryId } }),
    prisma.language.findUnique({ where: { id: languageId } }),
    prisma.voiceRole.findUnique({ where: { id: voiceRoleId } }),
  ]);

  if (!category) throw createError({ statusCode: 400, message: "跟读场景不存在" });
  if (!language) throw createError({ statusCode: 400, message: "语言不存在" });
  if (!voiceRole) throw createError({ statusCode: 400, message: "语音角色不存在" });

  const synthesisType = (voiceRole.synthesisType ?? "tts").trim().toLowerCase();
  if (synthesisType !== "tts") {
    throw createError({ statusCode: 400, message: "批量导入须选择 TTS 类型语音角色" });
  }

  const maxSort = await prisma.readAloudVocabulary.aggregate({
    where: { categoryId, languageId },
    _max: { sortOrder: true },
  });
  let sortOrder = (maxSort._max.sortOrder ?? 0) + 1;

  const status = (input.status ?? "active").trim() || "active";
  const remark = (input.remark ?? "LLM batch import").trim() || null;

  const ids: string[] = [];
  await prisma.$transaction(async (tx) => {
    for (const it of items) {
      const row = await tx.readAloudVocabulary.create({
        data: {
          categoryId,
          languageId,
          word: it.word,
          exampleSentence: it.exampleSentence,
          voiceRoleId,
          sortOrder: sortOrder++,
          status,
          remark,
        },
      });
      ids.push(row.id);
    }
  });

  return { created: ids.length, ids };
}
