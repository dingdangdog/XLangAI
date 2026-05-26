import prisma from "../lib/prisma";
import { ensureTestSeedUser, runBusinessDataSeed } from "../lib/businessDataSeed";
import {
  ensureManagerAdminUser,
  readManagerAdminBootstrapConfig,
} from "../lib/managerAdminSeed";

/**
 * Startup order (after 00-db-migrate):
 * 1. MANAGER_ADMIN_* → bootstrap manager admin (independent of MANAGER_AUTO_SEED);
 * 2. MANAGER_AUTO_SEED !== "false" → business seed;
 * 3. MANAGER_TEST_ACCOUNT_SEED !== "false" → test account.
 */
export default defineNitroPlugin(async () => {
  if (!process.env.DATABASE_URL?.trim()) {
    console.warn("[data-seed] skip: DATABASE_URL not set");
    return;
  }

  const adminCfg = readManagerAdminBootstrapConfig();
  if (adminCfg) {
    try {
      await ensureManagerAdminUser(prisma, adminCfg);
    } catch (e) {
      console.error("[data-seed] manager admin seed failed (non-fatal):", e);
    }
  }
  const managerCfg = useRuntimeConfig().manager;
  const wantBiz = managerCfg.autoSeed !== "false";
  const wantTest = managerCfg.testAccountSeed !== "false";

  if (wantBiz) {
    try {
      await runBusinessDataSeed(prisma);
    } catch (e) {
      console.error("[data-seed] business seed failed (non-fatal):", e);
    }
  }

  if (!wantTest) return;

  try {
    await ensureTestSeedUser(prisma);
  } catch (e) {
    console.error("[data-seed] test account seed failed (non-fatal):", e);
  }
});
