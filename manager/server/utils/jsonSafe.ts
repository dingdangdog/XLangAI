/** 递归将 BigInt 转为字符串，避免 h3 JSON.stringify 抛出 Do not know how to serialize a BigInt */
export function jsonSafe<T>(value: T): T {
  if (value === null || value === undefined) return value;
  if (typeof value === "bigint") return value.toString() as T;
  if (Array.isArray(value)) return value.map((v) => jsonSafe(v)) as T;
  if (typeof value === "object") {
    if (value instanceof Date) return value;
    const out: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(value as Record<string, unknown>)) {
      out[k] = jsonSafe(v);
    }
    return out as T;
  }
  return value;
}
