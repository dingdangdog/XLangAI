import { createError } from "h3";
import prisma from "../prisma";
import { llmChatComplete } from "../llmChat/complete";

export type TranslateCategoryTitlesInput = {
  sourceLanguageId: string;
  title: string;
  llmServiceConfigId?: string;
};

export type TranslateCategoryTitlesResult = {
  titlesByLanguageId: Record<string, string>;
  sourceLanguageId: string;
  llmServiceConfigId: string;
};

function extractJsonObject(raw: string): Record<string, unknown> | null {
  const text = raw.trim();
  if (!text) return null;

  const tryParse = (s: string): Record<string, unknown> | null => {
    try {
      const parsed = JSON.parse(s) as unknown;
      return parsed && typeof parsed === "object" && !Array.isArray(parsed)
        ? (parsed as Record<string, unknown>)
        : null;
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

  const start = text.indexOf("{");
  const end = text.lastIndexOf("}");
  if (start >= 0 && end > start) {
    return tryParse(text.slice(start, end + 1));
  }

  return null;
}

async function resolveLlm(llmId?: string) {
  if (llmId?.trim()) {
    const row = await prisma.sysLlmServiceConfig.findUnique({ where: { id: llmId.trim() } });
    if (row && row.status === "active") return row;
    throw createError({ statusCode: 400, message: "LLM 服务配置不存在或未启用" });
  }
  const row = await prisma.sysLlmServiceConfig.findFirst({
    where: { status: "active" },
    orderBy: [{ sortOrder: "asc" }, { code: "asc" }],
  });
  if (!row) {
    throw createError({ statusCode: 400, message: "请先在 AI 设置中启用至少一个 LLM 服务" });
  }
  return row;
}

export async function translateReadAloudCategoryTitles(
  input: TranslateCategoryTitlesInput,
): Promise<TranslateCategoryTitlesResult> {
  const sourceLanguageId = input.sourceLanguageId.trim();
  const title = input.title.trim();
  if (!sourceLanguageId || !title) {
    throw createError({ statusCode: 400, message: "缺少源语言或场景名称" });
  }

  const [llm, sourceLang, targets] = await Promise.all([
    resolveLlm(input.llmServiceConfigId),
    prisma.language.findUnique({ where: { id: sourceLanguageId } }),
    prisma.language.findMany({
      where: { status: "active" },
      orderBy: [{ sortOrder: "asc" }, { code: "asc" }],
    }),
  ]);

  if (!sourceLang || sourceLang.status !== "active") {
    throw createError({ statusCode: 400, message: "源语言不存在或未启用" });
  }

  const targetLangs = targets.filter((l) => l.id !== sourceLanguageId);
  if (!targetLangs.length) {
    throw createError({ statusCode: 400, message: "没有其他可翻译的目标语言" });
  }

  const targetLines = targetLangs
    .map((l) => `- ${l.code} (${l.nameNative?.trim() || l.name})`)
    .join("\n");

  const system = `You translate short UI scenario titles for a language-learning app.
Return ONLY a JSON object: keys are language codes (lowercase), values are concise natural titles in that language.
Do not add markdown or explanations.`;

  const user = `Source language: ${sourceLang.code} (${sourceLang.nameNative?.trim() || sourceLang.name})
Source title: "${title}"

Translate this scenario title into each target language below. Keep titles short (typically 2–6 words), natural for learners.

Target languages:
${targetLines}

JSON keys must be exactly these codes: ${targetLangs.map((l) => l.code).join(", ")}`;

  const rawText = await llmChatComplete(
    {
      protocol: llm.protocol,
      baseUrl: llm.baseUrl,
      apiKey: llm.apiKey,
      modelCode: llm.modelCode,
      config: llm.config,
    },
    system,
    user,
  );

  const parsed = extractJsonObject(rawText);
  if (!parsed) {
    throw createError({
      statusCode: 502,
      message: "LLM 未返回有效 JSON，请更换模型后重试",
    });
  }

  const codeToId = new Map(
    targetLangs.map((l) => [String(l.code).trim().toLowerCase(), l.id]),
  );
  const titlesByLanguageId: Record<string, string> = {
    [sourceLanguageId]: title,
  };

  for (const [code, value] of Object.entries(parsed)) {
    const langId = codeToId.get(String(code).trim().toLowerCase());
    const text = String(value ?? "").trim();
    if (langId && text) titlesByLanguageId[langId] = text;
  }

  const filled = Object.keys(titlesByLanguageId).length;
  if (filled <= 1) {
    throw createError({
      statusCode: 502,
      message: "未能解析出其他语言的译名，请重试或手动填写",
    });
  }

  return {
    titlesByLanguageId,
    sourceLanguageId,
    llmServiceConfigId: llm.id,
  };
}
