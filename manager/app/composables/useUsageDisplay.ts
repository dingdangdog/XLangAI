import {
  formatAudioBytes,
  formatCompactNumber,
} from "~/utils/usageDisplay";

type UserUsageRow = Record<string, unknown>;

type ServiceUsage = {
  todayRequestCount?: number;
  todayUnitCount?: string;
  monthRequestCount?: number;
  monthUnitCount?: string;
  unitLabel?: string;
};

export function useUsageDisplay() {
  const { t } = useI18n();

  function userUsagePrimaryLine(row: UserUsageRow, period: "today" | "month"): string {
    const count =
      period === "today" ? Number(row.todayUsageCount ?? 0) : Number(row.monthUsageCount ?? 0);
    const tok =
      period === "today" ? Number(row.todayTokenCount ?? 0) : Number(row.monthTokenCount ?? 0);
    return t("usage.primaryLine", {
      count,
      tokens: formatCompactNumber(tok),
    });
  }

  function userUsageDetailLine(row: UserUsageRow, period: "today" | "month"): string {
    const prefix = period === "today" ? "today" : "month";
    const parts: string[] = [];
    const ttsCount = Number(row[`${prefix}TtsCount`] ?? 0);
    const ttsChars = Number(row[`${prefix}TtsChars`] ?? 0);
    if (ttsCount > 0 || ttsChars > 0) {
      parts.push(
        t("usage.ttsLine", {
          count: ttsCount,
          chars: formatCompactNumber(ttsChars),
        }),
      );
    }
    const trCount = Number(row[`${prefix}TranslateCount`] ?? 0);
    const trChars = Number(row[`${prefix}TranslateChars`] ?? 0);
    if (trCount > 0 || trChars > 0) {
      parts.push(
        t("usage.translateLine", {
          count: trCount,
          chars: formatCompactNumber(trChars),
        }),
      );
    }
    const sttCount = Number(row[`${prefix}SttCount`] ?? 0);
    const sttBytes = row[`${prefix}SttAudioBytes`];
    if (sttCount > 0 || (sttBytes != null && String(sttBytes) !== "0")) {
      parts.push(
        t("usage.sttLine", {
          count: sttCount,
          size: formatAudioBytes(sttBytes as string | number | bigint),
        }),
      );
    }
    return parts.join(t("usage.separator"));
  }

  function serviceUsageTodayLine(usage: ServiceUsage | undefined): string {
    if (!usage) return t("common.emDash");
    const u = usage.unitLabel ?? "";
    return t("usage.serviceLine", {
      count: usage.todayRequestCount ?? 0,
      units: formatCompactNumber(usage.todayUnitCount ?? 0),
      unitLabel: u,
    });
  }

  function serviceUsageMonthLine(usage: ServiceUsage | undefined): string {
    if (!usage) return t("common.emDash");
    const u = usage.unitLabel ?? "";
    return t("usage.serviceLine", {
      count: usage.monthRequestCount ?? 0,
      units: formatCompactNumber(usage.monthUnitCount ?? 0),
      unitLabel: u,
    });
  }

  function usageCountCharsLine(count: unknown, chars: unknown): string {
    return t("usage.countChars", {
      count: Number(count ?? 0),
      chars: Number(chars ?? 0),
    });
  }

  return {
    userUsagePrimaryLine,
    userUsageDetailLine,
    serviceUsageTodayLine,
    serviceUsageMonthLine,
    usageCountCharsLine,
    formatAudioBytes,
    formatCompactNumber,
  };
}
