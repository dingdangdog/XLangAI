import prisma from "../lib/prisma";

export type UsageTrendPoint = {
  date: string;
  requestCount: number;
  unitCount: string;
};

export type UsageTrendService = {
  unitLabel: string;
  points: UsageTrendPoint[];
  totalRequests: number;
  totalUnits: string;
};

export type UsageTrendResponse = {
  days: 7 | 30;
  from: string;
  to: string;
  services: {
    llm: UsageTrendService;
    tts: UsageTrendService;
    translate: UsageTrendService;
    stt: UsageTrendService;
  };
  conversations: {
    points: { date: string; count: number }[];
    total: number;
  };
};

const SERVICE_TYPES = ["llm", "tts", "translate", "stt"] as const;
type ServiceType = (typeof SERVICE_TYPES)[number];

const UNIT_LABELS: Record<ServiceType, string> = {
  llm: "tok",
  tts: "字",
  translate: "字",
  stt: "B",
};

function utcDateOnly(d = new Date()): Date {
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate()));
}

function formatDateISO(d: Date): string {
  return d.toISOString().slice(0, 10);
}

function buildDateRange(days: 7 | 30): Date[] {
  const end = utcDateOnly();
  const result: Date[] = [];
  for (let i = days - 1; i >= 0; i--) {
    const d = new Date(end);
    d.setUTCDate(d.getUTCDate() - i);
    result.push(d);
  }
  return result;
}

function emptyService(type: ServiceType, dates: Date[]): UsageTrendService {
  return {
    unitLabel: UNIT_LABELS[type],
    points: dates.map((d) => ({
      date: formatDateISO(d),
      requestCount: 0,
      unitCount: "0",
    })),
    totalRequests: 0,
    totalUnits: "0",
  };
}

export async function fetchDashboardUsageTrends(days: 7 | 30): Promise<UsageTrendResponse> {
  const dates = buildDateRange(days);
  const from = dates[0]!;
  const to = dates[dates.length - 1]!;

  const [serviceRows, conversationRows] = await Promise.all([
    prisma.sysServiceUsageDaily.groupBy({
      by: ["date", "serviceType"],
      where: {
        date: { gte: from, lte: to },
        serviceType: { in: [...SERVICE_TYPES] },
      },
      _sum: { requestCount: true, unitCount: true },
    }),
    prisma.userUsage.groupBy({
      by: ["date"],
      where: { date: { gte: from, lte: to } },
      _sum: { usageCount: true },
    }),
  ]);

  const serviceMap = new Map<string, { requestCount: number; unitCount: bigint }>();
  for (const row of serviceRows) {
    const key = `${formatDateISO(row.date)}:${row.serviceType}`;
    serviceMap.set(key, {
      requestCount: row._sum.requestCount ?? 0,
      unitCount: row._sum.unitCount ?? BigInt(0),
    });
  }

  const conversationMap = new Map<string, number>();
  for (const row of conversationRows) {
    conversationMap.set(formatDateISO(row.date), row._sum.usageCount ?? 0);
  }

  const services = {} as UsageTrendResponse["services"];
  for (const type of SERVICE_TYPES) {
    let totalRequests = 0;
    let totalUnits = BigInt(0);
    const points: UsageTrendPoint[] = dates.map((d) => {
      const date = formatDateISO(d);
      const agg = serviceMap.get(`${date}:${type}`);
      const requestCount = agg?.requestCount ?? 0;
      const unitCount = agg?.unitCount ?? BigInt(0);
      totalRequests += requestCount;
      totalUnits += unitCount;
      return {
        date,
        requestCount,
        unitCount: String(unitCount),
      };
    });
    services[type] = {
      unitLabel: UNIT_LABELS[type],
      points,
      totalRequests,
      totalUnits: String(totalUnits),
    };
  }

  let conversationTotal = 0;
  const conversationPoints = dates.map((d) => {
    const date = formatDateISO(d);
    const count = conversationMap.get(date) ?? 0;
    conversationTotal += count;
    return { date, count };
  });

  return {
    days,
    from: formatDateISO(from),
    to: formatDateISO(to),
    services,
    conversations: {
      points: conversationPoints,
      total: conversationTotal,
    },
  };
}
