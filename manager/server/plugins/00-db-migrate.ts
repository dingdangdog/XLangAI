import { ensureDatabaseMigrations } from "../lib/db-migrations";

export default defineNitroPlugin(async () => {
  if (!process.env.DATABASE_URL?.trim()) {
    console.warn("[db:migrate] skip: DATABASE_URL not set");
    return;
  }

  await ensureDatabaseMigrations();
});
