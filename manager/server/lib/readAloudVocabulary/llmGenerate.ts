import { createError } from "h3";
import prisma from "../prisma";
import { llmChatComplete } from "../llmChat/complete";

export type LlmVocabItem = {
  word: string;
  exampleSentence: string;
};

export type LlmGenerateVocabularyInput = {
  llmServiceConfigId: string;
  categoryId: string;
  languageId: string;
  count?: number;
  extraInstructions?: string | null;
};

export type LlmGenerateVocabularyResult = {
  items: LlmVocabItem[];
  modelCode: string;
  llmServiceConfigId: string;
  existingWordCount: number;
};

function clampCount(n: unknown): number {
  const v = typeof n === "number" ? n : Number(n);
  if (!Number.isFinite(v)) return 10;
  return Math.min(30, Math.max(1, Math.round(v)));
}

function extractJsonArray(raw: string): unknown[] {
  const text = raw.trim();
  if (!text) return [];

  const tryParse = (s: string): unknown[] | null => {
    try {
      const parsed = JSON.parse(s) as unknown;
      return Array.isArray(parsed) ? parsed : null;
    } catch {
      return null;
    }
  };

  const direct = tryParse(text);
  if (direct) return direct;

  const fenced = text.match(/```(?:json)?\s*([\s\S]*?)```/i);
  if (fenced?.[1]) {
    const inner = tryParse(fenced[1].trim());
    if (inner) return inner;
  }

  const start = text.indexOf("[");
  const end = text.lastIndexOf("]");
  if (start >= 0 && end > start) {
    const slice = tryParse(text.slice(start, end + 1));
    if (slice) return slice;
  }

  return [];
}

function normalizeItems(raw: unknown[]): LlmVocabItem[] {
  const out: LlmVocabItem[] = [];
  const seen = new Set<string>();
  for (const row of raw) {
    if (!row || typeof row !== "object") continue;
    const obj = row as Record<string, unknown>;
    const word = String(obj.word ?? obj.term ?? obj.vocabulary ?? "").trim();
    const exampleSentence = String(
      obj.exampleSentence ?? obj.example_sentence ?? obj.sentence ?? obj.example ?? "",
    ).trim();
    if (!word || !exampleSentence) continue;
    const key = word.toLowerCase();
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ word, exampleSentence });
  }
  return out;
}

function buildSystemPrompt(): string {
  return `You are a language-learning content assistant. Generate vocabulary entries for read-aloud practice.
Output ONLY valid JSON: an array of objects with keys "word" and "exampleSentence".
No markdown fences, no commentary, no extra keys.`;
}

function buildUserPrompt(args: {
  categoryName: string;
  categoryDescription: string;
  languageName: string;
  languageCode: string;
  count: number;
  existingWords: string[];
  extraInstructions?: string | null;
}): string {
  const existing =
    args.existingWords.length > 0
      ? args.existingWords.slice(0, 80).join(", ")
      : "(none)";
  const extra = (args.extraInstructions ?? "").trim();
  return `Scenario: ${args.categoryName}
Scenario description: ${args.categoryDescription || "N/A"}
Target language: ${args.languageName} (code: ${args.languageCode})
Generate exactly ${args.count} new vocabulary items for this scenario.

Existing words/phrases to avoid duplicating: ${existing}

Requirements:
- "word": one useful word or short phrase for this scenario
- "exampleSentence": one natural spoken sentence using the word, entirely in ${args.languageName}
- Practical, conversational, suitable for TTS read-aloud
- Do not repeat existing words${extra ? `\n\nAdditional instructions:\n${extra}` : ""}

Return JSON array only, length ${args.count}.`;
}

export async function generateReadAloudVocabularyWithLlm(
  input: LlmGenerateVocabularyInput,
): Promise<LlmGenerateVocabularyResult> {
  const llmId = input.llmServiceConfigId.trim();
  const categoryId = input.categoryId.trim();
  const languageId = input.languageId.trim();
  if (!llmId || !categoryId || !languageId) {
    throw createError({ statusCode: 400, message: "缺少 LLM 配置、场景或语言" });
  }

  const count = clampCount(input.count);

  const [llm, category, language, existing] = await Promise.all([
    prisma.sysLlmServiceConfig.findUnique({ where: { id: llmId } }),
    prisma.readAloudCategory.findUnique({ where: { id: categoryId } }),
    prisma.language.findUnique({ where: { id: languageId } }),
    prisma.readAloudVocabulary.findMany({
      where: { categoryId, languageId, status: "active" },
      select: { word: true },
      take: 200,
      orderBy: { sortOrder: "asc" },
    }),
  ]);

  if (!llm || llm.status !== "active") {
    throw createError({ statusCode: 400, message: "LLM 服务配置不存在或未启用" });
  }
  if (!category) {
    throw createError({ statusCode: 400, message: "跟读场景不存在" });
  }
  if (!language) {
    throw createError({ statusCode: 400, message: "语言不存在" });
  }

  const existingWords = existing.map((r) => r.word.trim()).filter(Boolean);
  const rawText = await llmChatComplete(
    {
      protocol: llm.protocol,
      baseUrl: llm.baseUrl,
      apiKey: llm.apiKey,
      modelCode: llm.modelCode,
      config: llm.config,
    },
    buildSystemPrompt(),
    buildUserPrompt({
      categoryName: category.name,
      categoryDescription: category.description ?? "",
      languageName: language.nameNative?.trim() || language.name,
      languageCode: language.code,
      count,
      existingWords,
      extraInstructions: input.extraInstructions,
    }),
  );

  const parsed = normalizeItems(extractJsonArray(rawText));
  if (parsed.length === 0) {
    throw createError({
      statusCode: 502,
      message: "LLM 未返回有效词汇 JSON，请调整提示或更换模型后重试",
    });
  }

  return {
    items: parsed.slice(0, count),
    modelCode: llm.modelCode,
    llmServiceConfigId: llm.id,
    existingWordCount: existingWords.length,
  };
}
