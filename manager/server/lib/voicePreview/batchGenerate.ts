import { createError } from "h3";
import prisma from "../prisma";
import {
  generateVoiceRolePreview,
  type GeneratePreviewResult,
} from "./generate";

export const BATCH_GENERATE_PREVIEW_MAX = 30;

export type BatchGeneratePreviewInput = {
  /** 显式指定角色 id；未传则按 onlyMissing / 全量筛选 */
  ids?: string[];
  /** 仅处理尚未生成试听的角色（默认 true） */
  onlyMissing?: boolean;
};

export type BatchGeneratePreviewFailure = {
  id: string;
  name?: string;
  error: string;
};

export type BatchGeneratePreviewSkipped = {
  id: string;
  name?: string;
  reason: string;
};

export type BatchGeneratePreviewResult = {
  totalMatched: number;
  processed: number;
  ok: number;
  failed: BatchGeneratePreviewFailure[];
  skipped: BatchGeneratePreviewSkipped[];
  remaining: number;
  results: GeneratePreviewResult[];
};

function hasPreview(row: {
  previewAudioUrl: string | null;
  previewLocalFilename: string | null;
}): boolean {
  return (
    !!String(row.previewAudioUrl ?? "").trim() ||
    !!String(row.previewLocalFilename ?? "").trim()
  );
}

function errorMessage(e: unknown): string {
  if (e && typeof e === "object" && "statusMessage" in e) {
    const sm = String((e as { statusMessage?: unknown }).statusMessage ?? "").trim();
    if (sm) return sm;
  }
  if (e instanceof Error && e.message.trim()) return e.message;
  return "生成试听失败";
}

export async function batchGenerateVoiceRolePreviews(
  input: BatchGeneratePreviewInput,
): Promise<BatchGeneratePreviewResult> {
  const onlyMissing = input.onlyMissing !== false;
  const explicitIds = (input.ids ?? [])
    .map((id) => String(id ?? "").trim())
    .filter(Boolean);

  let candidates: Array<{
    id: string;
    name: string;
    synthesisType: string | null;
    previewAudioUrl: string | null;
    previewLocalFilename: string | null;
  }> = [];

  if (explicitIds.length > 0) {
    if (explicitIds.length > BATCH_GENERATE_PREVIEW_MAX) {
      throw createError({
        statusCode: 400,
        message: `单次最多处理 ${BATCH_GENERATE_PREVIEW_MAX} 条`,
      });
    }
    const rows = await prisma.voiceRole.findMany({
      where: { id: { in: explicitIds } },
      select: {
        id: true,
        name: true,
        synthesisType: true,
        previewAudioUrl: true,
        previewLocalFilename: true,
      },
    });
    const byId = new Map(rows.map((r) => [r.id, r]));
    candidates = explicitIds
      .map((id) => byId.get(id))
      .filter((r): r is NonNullable<typeof r> => !!r);
  } else {
    candidates = await prisma.voiceRole.findMany({
      orderBy: [{ sortOrder: "asc" }, { createdAt: "asc" }],
      select: {
        id: true,
        name: true,
        synthesisType: true,
        previewAudioUrl: true,
        previewLocalFilename: true,
      },
    });
  }

  const skipped: BatchGeneratePreviewSkipped[] = [];
  const planned: Array<{ id: string; name: string }> = [];

  for (const row of candidates) {
    const synthesisType = (row.synthesisType?.trim() || "tts").toLowerCase();
    if (synthesisType === "native_audio_in_text") {
      skipped.push({
        id: row.id,
        name: row.name,
        reason: "native_audio_in_text 无音频试听",
      });
      continue;
    }
    if (onlyMissing && hasPreview(row)) continue;
    planned.push({ id: row.id, name: row.name });
  }

  const totalMatched = planned.length;
  const batch = planned.slice(0, BATCH_GENERATE_PREVIEW_MAX);
  const remaining = Math.max(0, totalMatched - batch.length);

  const results: GeneratePreviewResult[] = [];
  const failed: BatchGeneratePreviewFailure[] = [];

  // 单条失败不影响后续：逐条 try/catch，串行避免 TTS 限流尖刺
  for (const item of batch) {
    try {
      const result = await generateVoiceRolePreview(item.id);
      results.push(result);
    } catch (e) {
      failed.push({ id: item.id, name: item.name, error: errorMessage(e) });
    }
  }

  return {
    totalMatched,
    processed: batch.length,
    ok: results.length,
    failed,
    skipped: explicitIds.length > 0 ? skipped : [],
    remaining,
    results,
  };
}
