import { ensureDatabaseMigrations } from "../lib/db-migrations";

export default defineNitroPlugin(async () => {
  if (!process.env.DATABASE_URL?.trim()) {
    console.warn("[db:migrate] 跳过：未设置 DATABASE_URL");
    return;
  }

  await ensureDatabaseMigrations();
});
