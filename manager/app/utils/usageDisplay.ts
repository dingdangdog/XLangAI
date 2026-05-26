/** 后台用量展示辅助（纯数值格式化，文案见 useUsageDisplay） */

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
