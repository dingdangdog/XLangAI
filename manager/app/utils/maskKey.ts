/** API Key 列表展示脱敏 */
export function maskKey(key: string | null | undefined): string {
  if (!key || key.length < 8) return key ? "••••" : "—";
  return `${key.slice(0, 4)}••••${key.slice(-4)}`;
}
