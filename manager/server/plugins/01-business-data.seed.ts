import prisma from "../lib/prisma";
import { ensureTestSeedUser, runBusinessDataSeed } from "../lib/businessDataSeed";
import {
  ensureManagerAdminUser,
  readManagerAdminBootstrapConfig,
} from "../lib/managerAdminSeed";

/**
 * 启动时顺序执行（在 00-db-migrate 之后）：
 * 1. MANAGER_ADMIN_* 配置时初始化运营管理员（与 MANAGER_AUTO_SEED 无关）；
 * 2. MANAGER_AUTO_SEED !== "false" 时跑业务种子；
 * 3. MANAGER_TEST_ACCOUNT_SEED !== "false" 时再跑测试账号。
 */
export default defineNitroPlugin(async () => {
  if (!process.env.DATABASE_URL?.trim()) {
    console.warn("[data-seed] 跳过：未设置 DATABASE_URL");
    return;
  }

  const adminCfg = readManagerAdminBootstrapConfig();
  if (adminCfg) {
    try {
      await ensureManagerAdminUser(prisma, adminCfg);
    } catch (e) {
      console.error("[data-seed] 运营管理员初始化失败（不影响服务启动）:", e);
    }
  }
  const managerCfg = useRuntimeConfig().manager;
  const wantBiz = managerCfg.autoSeed !== "false";
  const wantTest = managerCfg.testAccountSeed !== "false";

  if (wantBiz) {
    try {
      await runBusinessDataSeed(prisma);
    } catch (e) {
      console.error("[data-seed] 业务数据初始化失败（不影响服务启动）:", e);
    }
  }

  if (!wantTest) return;

  try {
    await ensureTestSeedUser(prisma);
  } catch (e) {
    console.error("[data-seed] 测试账号初始化失败（不影响服务启动）:", e);
  }
});
