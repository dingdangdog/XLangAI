import prisma from "../lib/prisma";

export type ServiceUsageType = "llm" | "tts" | "translate" | "stt";

function utcDateOnly(d = new Date()): Date {
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate()));
}

function monthStart(today: Date): Date {
  return new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), 1));
}

export type ServiceUsageSummary = {
  todayRequestCount: number;
  todayUnitCount: string;
  monthRequestCount: number;
  monthUnitCount: string;
  unitLabel: string;
};

const UNIT_LABELS: Record<ServiceUsageType, string> = {
  llm: "tok",
  tts: "字",
  translate: "字",
  stt: "B",
};

/** 为服务配置列表附加今日/本月请求次数与计量单位合计。 */
export async function attachServiceConfigUsageFields(
  rows: Record<string, unknown>[],
  serviceType: ServiceUsageType,
): Promise<Record<string, unknown>[]> {
  if (rows.length === 0) return rows;

  const configIds = rows
    .map((r) => String(r.id ?? ""))
    .filter((id) => id.length > 0);
  if (configIds.length === 0) return rows;

  const today = utcDateOnly();
  const monthFrom = monthStart(today);
  const unitLabel = UNIT_LABELS[serviceType];

  const [todayRows, monthGrouped] = await Promise.all([
    configIds.length
      ? prisma.sysServiceUsageDaily.findMany({
          where: {
            serviceType,
            configId: { in: configIds },
            date: today,
          },
        })
      : Promise.resolve(
          [] as Awaited<ReturnType<typeof prisma.sysServiceUsageDaily.findMany>>,
        ),
    configIds.length
      ? prisma.sysServiceUsageDaily.groupBy({
          by: ["configId"],
          where: {
            serviceType,
            configId: { in: configIds },
            date: { gte: monthFrom },
          },
          _sum: { requestCount: true, unitCount: true },
        })
      : Promise.resolve(
          [] as Awaited<ReturnType<typeof prisma.sysServiceUsageDaily.groupBy>>,
        ),
  ]);

  const todayByConfig = new Map(todayRows.map((r) => [r.configId, r]));
  const monthByConfig = new Map(monthGrouped.map((g) => [g.configId, g]));

  return rows.map((r) => {
    const id = String(r.id ?? "");
    const todayU = todayByConfig.get(id);
    const monthU = monthByConfig.get(id);
    const usage: ServiceUsageSummary = {
      todayRequestCount: todayU?.requestCount ?? 0,
      todayUnitCount: String(todayU?.unitCount ?? BigInt(0)),
      monthRequestCount: monthU?._sum.requestCount ?? 0,
      monthUnitCount: String(monthU?._sum.unitCount ?? BigInt(0)),
      unitLabel,
    };
    return { ...r, usage };
  });
}

/** 格式化 STT 音频字节为可读大小。 */
export function formatAudioBytes(n: number | bigint | string): string {
  const bytes = typeof n === "bigint" ? n : BigInt(String(n ?? 0));
  const kb = BigInt(1024);
  if (bytes < kb) return `${bytes} B`;
  if (bytes < kb * kb) return `${(Number(bytes) / 1024).toFixed(1)} KB`;
  return `${(Number(bytes) / (1024 * 1024)).toFixed(2)} MB`;
}
