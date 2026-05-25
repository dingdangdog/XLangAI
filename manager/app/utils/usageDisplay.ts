/** 后台用量展示辅助（用户列表、服务配置列表共用） */

export function formatCompactNumber(n: number | string | bigint): string {
  const v = typeof n === "bigint" ? Number(n) : Number(n ?? 0);
  if (!Number.isFinite(v)) return "0";
  if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(1)}M`;
  if (v >= 10_000) return `${(v / 1_000).toFixed(1)}k`;
  return String(v);
}

export function formatAudioBytes(n: number | string | bigint): string {
  const bytes = typeof n === "bigint" ? n : BigInt(String(n ?? 0));
  if (bytes < 1024n) return `${bytes} B`;
  if (bytes < 1024n * 1024n) return `${(Number(bytes) / 1024).toFixed(1)} KB`;
  return `${(Number(bytes) / (1024 * 1024)).toFixed(2)} MB`;
}

type UserUsageRow = Record<string, unknown>;

/** 用户列表「今日/本月」主行：对话轮次 + LLM token */
export function userUsagePrimaryLine(row: UserUsageRow, period: "today" | "month"): string {
  const count =
    period === "today" ? Number(row.todayUsageCount ?? 0) : Number(row.monthUsageCount ?? 0);
  const tok =
    period === "today" ? Number(row.todayTokenCount ?? 0) : Number(row.monthTokenCount ?? 0);
  return `${count} 次 · ${formatCompactNumber(tok)} tok`;
}

/** 用户列表副行：TTS / 翻译 / STT 细分（无数据时返回空串） */
export function userUsageDetailLine(row: UserUsageRow, period: "today" | "month"): string {
  const prefix = period === "today" ? "today" : "month";
  const parts: string[] = [];
  const ttsCount = Number(row[`${prefix}TtsCount`] ?? 0);
  const ttsChars = Number(row[`${prefix}TtsChars`] ?? 0);
  if (ttsCount > 0 || ttsChars > 0) {
    parts.push(`TTS ${ttsCount} · ${formatCompactNumber(ttsChars)} 字`);
  }
  const trCount = Number(row[`${prefix}TranslateCount`] ?? 0);
  const trChars = Number(row[`${prefix}TranslateChars`] ?? 0);
  if (trCount > 0 || trChars > 0) {
    parts.push(`翻译 ${trCount} · ${formatCompactNumber(trChars)} 字`);
  }
  const sttCount = Number(row[`${prefix}SttCount`] ?? 0);
  const sttBytes = row[`${prefix}SttAudioBytes`];
  if (sttCount > 0 || (sttBytes != null && String(sttBytes) !== "0")) {
    parts.push(`STT ${sttCount} · ${formatAudioBytes(sttBytes as string | number | bigint)}`);
  }
  return parts.join(" · ");
}

type ServiceUsage = {
  todayRequestCount?: number;
  todayUnitCount?: string;
  monthRequestCount?: number;
  monthUnitCount?: string;
  unitLabel?: string;
};

export function serviceUsageTodayLine(usage: ServiceUsage | undefined): string {
  if (!usage) return "—";
  const u = usage.unitLabel ?? "";
  return `${usage.todayRequestCount ?? 0} 次 / ${formatCompactNumber(usage.todayUnitCount ?? 0)} ${u}`;
}

export function serviceUsageMonthLine(usage: ServiceUsage | undefined): string {
  if (!usage) return "—";
  const u = usage.unitLabel ?? "";
  return `${usage.monthRequestCount ?? 0} 次 / ${formatCompactNumber(usage.monthUnitCount ?? 0)} ${u}`;
}
