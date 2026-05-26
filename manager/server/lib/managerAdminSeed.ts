import bcrypt from "bcryptjs";
import type { AppPrismaClient } from "./prisma";

/** Matches Go bcrypt.DefaultCost and userAdmin */
const BCRYPT_ROUNDS = 10;

/** Marks manager bootstrap account for auth and troubleshooting */
export const MANAGER_ADMIN_REMARK = "manager_admin";

export type ManagerAdminBootstrapConfig = {
  username: string;
  password: string;
  nickname: string;
};

/**
 * Reads manager admin bootstrap config from env.
 * Requires both MANAGER_ADMIN_USERNAME and MANAGER_ADMIN_PASSWORD (hashed with bcrypt before insert).
 * Skips when MANAGER_ADMIN_SEED=false or credentials are incomplete (no random password).
 */
export function readManagerAdminBootstrapConfig(): ManagerAdminBootstrapConfig | null {
  const managerCfg = useRuntimeConfig().manager;
  if (managerCfg.adminSeed === "false") {
    return null;
  }
  const username = managerCfg.adminUsername?.trim() ?? "";
  const password = String(managerCfg.adminPassword ?? "");
  const nickname = managerCfg.adminNickname?.trim() || "Admin";

  if (!username && !password.trim()) {
    return null;
  }

  if (!username || !password.trim()) {
    console.warn(
      "[data-seed] set both MANAGER_ADMIN_USERNAME and MANAGER_ADMIN_PASSWORD; skip manager admin seed",
    );
    return null;
  }

  if (password.trim().length < 6) {
    console.warn("[data-seed] MANAGER_ADMIN_PASSWORD must be at least 6 characters; skip manager admin seed");
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

/** Idempotent manager admin user (usr_users); dedupe by phone or email; password stored as bcrypt hash. */
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
      `[data-seed] manager admin ${maskLoginId(config.username)} already exists (id=${row.id}), skip insert`,
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
    `[data-seed] initialized manager admin ${maskLoginId(config.username)} (password stored as bcrypt hash)`,
  );
}
