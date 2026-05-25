const DATE_TIME_PROPS = new Set([
  "createdAt",
  "updatedAt",
  "deletedAt",
  "cancelledAt",
  "lastLoginAt",
]);

/** 表格列是否为常见时间戳字段 */
export function isDateTimeField(prop: string): boolean {
  return DATE_TIME_PROPS.has(prop);
}

function toDate(value: unknown): Date | null {
  if (value == null || value === "") return null;
  if (value instanceof Date) {
    return Number.isNaN(value.getTime()) ? null : value;
  }
  if (typeof value === "number") {
    const d = new Date(value);
    return Number.isNaN(d.getTime()) ? null : d;
  }
  const s = String(value).trim();
  if (!s) return null;
  const d = new Date(s);
  return Number.isNaN(d.getTime()) ? null : d;
}

/** 日期时间：本地时区，如 2026-05-13 11:32:25 */
export function formatDateTime(value: unknown): string {
  const d = toDate(value);
  if (!d) return value == null || value === "" ? "—" : String(value);
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  }).format(d);
}

/** 仅日期：如 2026-05-13 */
export function formatDate(value: unknown): string {
  const d = toDate(value);
  if (!d) return value == null || value === "" ? "—" : String(value);
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(d);
}
