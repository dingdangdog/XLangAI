/**
 * 一次性执行：启用德/法/西、趣味人设角色、跟读场景词汇。
 * 用法（在 servers/manager 下）：
 *   $env:DATABASE_URL="..."; npx --yes tsx scripts/run-content-expand-seed.ts
 */
import { PrismaClient } from "../prisma/generated/client.ts";
import { PrismaPg } from "@prisma/adapter-pg";
import pg from "pg";
import { runBusinessDataSeed } from "../server/lib/businessDataSeed.ts";

async function main() {
  if (!process.env.DATABASE_URL?.trim()) {
    throw new Error("DATABASE_URL is required");
  }
  const pool = new pg.Pool({ connectionString: process.env.DATABASE_URL });
  const db = new PrismaClient({ adapter: new PrismaPg(pool) });
  try {
    console.info("[seed] running business + read-aloud content expand…");
    await runBusinessDataSeed(db as never);
    console.info("[seed] done");
  } finally {
    await db.$disconnect();
    await pool.end();
  }
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
