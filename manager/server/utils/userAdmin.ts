import { createError } from "h3";
import bcrypt from "bcryptjs";
import prisma from "../lib/prisma";

/** 与 Go `bcrypt.DefaultCost`、种子数据一致 */
const BCRYPT_ROUNDS = 10;

export function redactUserRecord(row: Record<string, unknown>): Record<string, unknown> {
  const { passwordHash, password, ...rest } = row;
  return {
    ...rest,
    hasPassword: typeof passwordHash === "string" && passwordHash.length > 0,
  };
}

async function normalizeUserDefaultLlmConfigId(
  value: unknown,
): Promise<string | null | undefined> {
  if (value === undefined) {
    return undefined;
  }
  if (value === null || value === "") {
    return null;
  }
  const id = String(value).trim();
  if (!id) {
    return null;
  }
  const row = await prisma.sysLlmServiceConfig.findFirst({
    where: { id, status: "active" },
    select: { id: true },
  });
  if (!row) {
    throw createError({
      statusCode: 400,
      message: "指定的 LLM 配置不存在或已停用",
    });
  }
  return id;
}

export async function prepareUserAdminWriteData(
  body: Record<string, unknown>,
  mode: "create" | "update",
): Promise<Record<string, unknown>> {
  const data = { ...body };
  delete data.password;
  delete data.hasPassword;

  const plain = typeof body.password === "string" ? body.password.trim() : "";

  if (mode === "create") {
    if (!plain) {
      throw createError({ statusCode: 400, message: "创建用户须设置密码（至少 6 位）" });
    }
    if (plain.length < 6) {
      throw createError({ statusCode: 400, message: "密码至少 6 位" });
    }
    const phone = data.phone != null ? String(data.phone).trim() : "";
    const email = data.email != null ? String(data.email).trim() : "";
    if (!phone && !email) {
      throw createError({ statusCode: 400, message: "须填写手机或邮箱其一" });
    }
    data.passwordHash = await bcrypt.hash(plain, BCRYPT_ROUNDS);
  } else if (plain) {
    if (plain.length < 6) {
      throw createError({ statusCode: 400, message: "密码至少 6 位" });
    }
    data.passwordHash = await bcrypt.hash(plain, BCRYPT_ROUNDS);
  }

  if ("defaultLlmConfigId" in body) {
    data.defaultLlmConfigId = await normalizeUserDefaultLlmConfigId(body.defaultLlmConfigId);
  }

  // 运营加次：优先于绝对写入 turnBalance
  if (mode === "update" && "addTurnBalance" in body) {
    const delta = Math.floor(Number(body.addTurnBalance));
    delete data.addTurnBalance;
    if (Number.isFinite(delta) && delta > 0) {
      delete data.turnBalance;
      data.turnBalance = { increment: delta };
      return data;
    }
  }

  if ("turnBalance" in body || mode === "create") {
    data.turnBalance = await normalizeTurnBalance(body.turnBalance, mode);
  }

  return data;
}

async function normalizeTurnBalance(
  value: unknown,
  mode: "create" | "update",
): Promise<number> {
  if (value === undefined || value === null || value === "") {
    if (mode === "create") {
      const setting = await prisma.sysSystemSetting.findUnique({
        where: { key: "quota.signup_turn_grant" },
        select: { value: true, status: true },
      });
      const raw =
        setting && (!setting.status || setting.status === "active")
          ? String(setting.value ?? "").trim()
          : "20";
      const n = Number.parseInt(raw || "20", 10);
      return Number.isFinite(n) && n >= 0 ? n : 20;
    }
    throw createError({ statusCode: 400, message: "turnBalance 无效" });
  }
  const n = Math.floor(Number(value));
  if (!Number.isFinite(n) || n < 0) {
    throw createError({ statusCode: 400, message: "turnBalance 须为非负整数" });
  }
  return n;
}

function utcDateOnly(d = new Date()): Date {
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate()));
}

/** 用户列表：会员等级展示字段 + 今日/本月用量摘要 */
export async function attachUserListFields(
  rows: Record<string, unknown>[],
): Promise<Record<string, unknown>[]> {
  if (rows.length === 0) return rows;

  const tierIds = [
    ...new Set(
      rows
        .map((r) => r.tierId)
        .filter((x): x is string => typeof x === "string" && x.length > 0),
    ),
  ];
  const langIds = [
    ...new Set(
      rows
        .map((r) => r.languageId)
        .filter((x): x is string => typeof x === "string" && x.length > 0),
    ),
  ];
  const llmIds = [
    ...new Set(
      rows
        .map((r) => r.defaultLlmConfigId)
        .filter((x): x is string => typeof x === "string" && x.length > 0),
    ),
  ];
  const userIds = rows.map((r) => String(r.id ?? "")).filter((id) => id.length > 0);

  const today = utcDateOnly();
  const monthStart = new Date(Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), 1));

  const [tiers, langs, llms, todayRows, monthGrouped] = await Promise.all([
    tierIds.length
      ? prisma.membershipTier.findMany({ where: { id: { in: tierIds } } })
      : Promise.resolve([]),
    langIds.length
      ? prisma.language.findMany({ where: { id: { in: langIds } } })
      : Promise.resolve([]),
    llmIds.length
      ? prisma.sysLlmServiceConfig.findMany({ where: { id: { in: llmIds } } })
      : Promise.resolve([]),
    userIds.length
      ? prisma.userUsage.findMany({
          where: { userId: { in: userIds }, date: today },
        })
      : Promise.resolve([] as Awaited<ReturnType<typeof prisma.userUsage.findMany>>),
    userIds.length
      ? prisma.userUsage.groupBy({
          by: ["userId"],
          where: { userId: { in: userIds }, date: { gte: monthStart } },
          _sum: {
            usageCount: true,
            tokenCount: true,
            translateCount: true,
            translateChars: true,
            ttsCount: true,
            ttsChars: true,
            sttCount: true,
            sttAudioBytes: true,
          },
        })
      : Promise.resolve([] as Awaited<ReturnType<typeof prisma.userUsage.groupBy>>),
  ]);

  const tierById = new Map(tiers.map((t) => [t.id, t]));
  const langById = new Map(langs.map((l) => [l.id, l]));
  const llmById = new Map(llms.map((l) => [l.id, l]));
  const todayByUser = new Map(todayRows.map((u) => [u.userId, u]));
  const monthByUser = new Map(monthGrouped.map((g) => [g.userId, g]));

  return rows.map((r) => {
    const base = redactUserRecord(r);
    const tierId = r.tierId;
    const tier = typeof tierId === "string" ? tierById.get(tierId) : undefined;
    const langId = r.languageId;
    const lang = typeof langId === "string" ? langById.get(langId) : undefined;
    const llmId = r.defaultLlmConfigId;
    const llm = typeof llmId === "string" ? llmById.get(llmId) : undefined;
    const uid = String(r.id ?? "");
    const todayU = todayByUser.get(uid);
    const monthU = monthByUser.get(uid);
    const tierCode = tier?.code ?? "";
    const tierName = tier?.name ?? "";
    return {
      ...base,
      tierCode,
      tierName,
      tierLabel: tier ? `${tier.code} · ${tier.name}` : "",
      nativeLanguageCode: lang?.code ?? "",
      nativeLanguageLabel: lang ? `${lang.code} · ${lang.name}` : "",
      defaultLlmLabel: llm ? `${llm.code} · ${llm.name}` : "",
      tierDailyLimit: tier?.dailyLimit ?? null,
      tierMonthlyLimit: tier?.monthlyLimit ?? null,
      todayUsageCount: todayU?.usageCount ?? 0,
      todayTokenCount: todayU?.tokenCount ?? 0,
      todayTranslateCount: todayU?.translateCount ?? 0,
      todayTranslateChars: todayU?.translateChars ?? 0,
      todayTtsCount: todayU?.ttsCount ?? 0,
      todayTtsChars: todayU?.ttsChars ?? 0,
      todaySttCount: todayU?.sttCount ?? 0,
      todaySttAudioBytes: String(todayU?.sttAudioBytes ?? BigInt(0)),
      monthUsageCount: monthU?._sum.usageCount ?? 0,
      monthTokenCount: monthU?._sum.tokenCount ?? 0,
      monthTranslateCount: monthU?._sum.translateCount ?? 0,
      monthTranslateChars: monthU?._sum.translateChars ?? 0,
      monthTtsCount: monthU?._sum.ttsCount ?? 0,
      monthTtsChars: monthU?._sum.ttsChars ?? 0,
      monthSttCount: monthU?._sum.sttCount ?? 0,
      monthSttAudioBytes: String(monthU?._sum.sttAudioBytes ?? BigInt(0)),
    };
  });
}
