import bcrypt from "bcryptjs";
import type { AppPrismaClient } from "./prisma";

/** 与 Go `bcrypt.DefaultCost`、`userAdmin` 一致 */
const BCRYPT_ROUNDS = 10;

/** 标记运营后台 bootstrap 账号，便于后续鉴权或排查 */
export const MANAGER_ADMIN_REMARK = "manager_admin";

export type ManagerAdminBootstrapConfig = {
  username: string;
  password: string;
  nickname: string;
};

/**
 * 从环境变量读取运营管理员初始化配置。
 * 须同时设置 `MANAGER_ADMIN_USERNAME` 与 `MANAGER_ADMIN_PASSWORD`（明文，入库前 bcrypt）。
 * `MANAGER_ADMIN_SEED=false` 时跳过；未配置账号密码时不创建（不自动生成随机密码）。
 */
export function readManagerAdminBootstrapConfig(): ManagerAdminBootstrapConfig | null {
  const managerCfg = useRuntimeConfig().manager;
  if (managerCfg.adminSeed === "false") {
    return null;
  }
  const username = managerCfg.adminUsername?.trim() ?? "";
  const password = String(managerCfg.adminPassword ?? "");
  const nickname = managerCfg.adminNickname?.trim() || "管理员";

  if (!username && !password.trim()) {
    return null;
  }

  if (!username || !password.trim()) {
    console.warn(
      "[data-seed] 须同时设置 MANAGER_ADMIN_USERNAME 与 MANAGER_ADMIN_PASSWORD，跳过运营管理员初始化",
    );
    return null;
  }

  if (password.trim().length < 6) {
    console.warn("[data-seed] MANAGER_ADMIN_PASSWORD 至少 6 位，跳过运营管理员初始化");
    return null;
  }

  return { username, password: password.trim(), nickname };
}

function maskLoginId(id: string): string {
  if (id.includes("@")) {
    const at = id.indexOf("@");
    const local = id.slice(0, at);
    const domain = id.slice(at + 1);
    if (!local || !domain) return id;
    const vis = local.length <= 2 ? `${local[0] ?? ""}*` : `${local.slice(0, 2)}***`;
    return `${vis}@${domain}`;
  }
  if (id.length <= 4) return "***";
  return `${id.slice(0, 3)}****${id.slice(-2)}`;
}

/**
 * 幂等写入运营管理员（usr_users）：按手机或邮箱查重，密码 bcrypt 哈希后入库。
 */
export async function ensureManagerAdminUser(
  db: AppPrismaClient,
  config: ManagerAdminBootstrapConfig,
): Promise<void> {
  const isEmail = config.username.includes("@");
  const phone = isEmail ? null : config.username;
  const email = isEmail ? config.username : null;

  const rows = isEmail
    ? await db.$queryRaw<{ id: string }[]>`
        SELECT id FROM usr_users WHERE email = ${email} AND deleted_at IS NULL LIMIT 1
      `
    : await db.$queryRaw<{ id: string }[]>`
        SELECT id FROM usr_users WHERE phone = ${phone} AND deleted_at IS NULL LIMIT 1
      `;

  if (rows.length > 0) {
    const row = rows[0]!;
    console.info(
      `[data-seed] 运营管理员 ${maskLoginId(config.username)} 已存在（id=${row.id}），跳过插入`,
    );
    return;
  }

  const freeTier = await db.membershipTier.findUnique({ where: { code: "free" } });
  const passwordHash = await bcrypt.hash(config.password, BCRYPT_ROUNDS);

  await db.user.create({
    data: {
      phone,
      email,
      passwordHash,
      nickname: config.nickname,
      tierId: freeTier?.id ?? null,
      status: "active",
      remark: MANAGER_ADMIN_REMARK,
    },
  });

  console.info(
    `[data-seed] 已初始化运营管理员 ${maskLoginId(config.username)}（密码已 bcrypt 哈希入库，请勿在日志中输出明文）`,
  );
}
