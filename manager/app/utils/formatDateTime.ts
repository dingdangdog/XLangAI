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

function intlLocale(code: string): string {
  if (code === "zh") return "zh-CN";
  if (code === "ja") return "ja-JP";
  return "en-US";
}

export function useFormatDateTime() {
  const { locale, t } = useI18n();
  const dateTimeLocale = computed(() => intlLocale(locale.value));

  function formatDateTime(value: unknown): string {
    const d = toDate(value);
    if (!d) {
      return value == null || value === "" ? t("common.emDash") : String(value);
    }
    return new Intl.DateTimeFormat(dateTimeLocale.value, {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    }).format(d);
  }

  function formatDate(value: unknown): string {
    const d = toDate(value);
    if (!d) {
      return value == null || value === "" ? t("common.emDash") : String(value);
    }
    return new Intl.DateTimeFormat(dateTimeLocale.value, {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    }).format(d);
  }

  return { formatDateTime, formatDate };
}

/** @deprecated 请使用 useFormatDateTime() */
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

/** @deprecated 请使用 useFormatDateTime() */
export function formatDate(value: unknown): string {
  const d = toDate(value);
  if (!d) return value == null || value === "" ? "—" : String(value);
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).format(d);
}
