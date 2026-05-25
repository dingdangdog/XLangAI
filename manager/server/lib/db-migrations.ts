import { createHash, randomUUID } from "node:crypto";
import { access, readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { Client } from "pg";

const MIGRATIONS_DIR = path.resolve(process.cwd(), "prisma", "migrations");
const MIGRATIONS_TABLE = "_prisma_migrations";
const LOCK_KEY_1 = 20260521;
const LOCK_KEY_2 = 1;

type MigrationFile = {
  name: string;
  sql: string;
  checksum: string;
};

type DatabaseTarget = {
  databaseUrl: string;
  adminUrl: string;
  databaseName: string;
  schemaName: string;
};

/** 与 Prisma migrate 一致：SHA-256 基于 LF 归一化后的 migration.sql */
export function normalizeMigrationSql(sql: string): string {
  return sql.replace(/\r\n/g, "\n").replace(/\r/g, "\n");
}

function getChecksum(sql: string): string {
  return createHash("sha256")
    .update(normalizeMigrationSql(sql))
    .digest("hex");
}

const EXPECTED_TABLES = [
  "sys_languages",
  "sys_llm_service_configs",
  "sys_stt_service_configs",
  "sys_translate_service_configs",
  "sys_object_storage_configs",
  "sys_system_settings",
  "sys_tts_service_configs",
  "sys_voice_roles",
  "sys_prompt_templates",
  "sys_membership_tiers",
  "sys_billing_products",
  "usr_users",
  "usr_store_transactions",
  "usr_user_usage",
  "usr_conversations",
  "usr_messages",
  "bak_usr_users",
  "bak_usr_conversations",
  "bak_usr_messages",
  "bak_usr_user_usage",
];

let migrationTask: Promise<void> | null = null;

function quoteIdentifier(value: string) {
  return `"${value.replaceAll(`"`, `""`)}"`;
}

function getDatabaseTarget(): DatabaseTarget {
  const databaseUrl = process.env.DATABASE_URL;
  if (!databaseUrl) {
    throw new Error("DATABASE_URL is required before applying migrations.");
  }

  const parsed = new URL(databaseUrl);
  const databaseName = decodeURIComponent(parsed.pathname.replace(/^\//, ""));
  if (!databaseName) {
    throw new Error("DATABASE_URL must include a database name.");
  }

  const schemaName =
    parsed.searchParams.get("schema") ||
    process.env.DATABASE_SCHEMA ||
    "public";

  const adminUrl =
    process.env.DATABASE_BOOTSTRAP_URL ||
    (() => {
      const bootstrap = new URL(databaseUrl);
      bootstrap.pathname = "/postgres";
      return bootstrap.toString();
    })();

  return {
    databaseUrl,
    adminUrl,
    databaseName,
    schemaName,
  };
}

async function readMigrationFiles(): Promise<MigrationFile[]> {
  try {
    await access(MIGRATIONS_DIR);
  } catch {
    return [];
  }

  const entries = await readdir(MIGRATIONS_DIR, { withFileTypes: true });
  const migrations = await Promise.all(
    entries
      .filter((entry) => entry.isDirectory())
      .sort((a, b) => a.name.localeCompare(b.name))
      .map(async (entry) => {
        const raw = await readFile(
          path.join(MIGRATIONS_DIR, entry.name, "migration.sql"),
          "utf8"
        );

        return {
          name: entry.name,
          sql: normalizeMigrationSql(raw),
          checksum: getChecksum(raw),
        };
      })
  );

  return migrations;
}

async function ensureMigrationsTable(client: Client) {
  await client.query(`
    CREATE TABLE IF NOT EXISTS "${MIGRATIONS_TABLE}" (
      "id" TEXT PRIMARY KEY,
      "checksum" TEXT NOT NULL,
      "finished_at" TIMESTAMPTZ,
      "migration_name" TEXT NOT NULL UNIQUE,
      "logs" TEXT,
      "rolled_back_at" TIMESTAMPTZ,
      "started_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      "applied_steps_count" INTEGER NOT NULL DEFAULT 0
    );
  `);
}

async function ensureSchema(client: Client, schemaName: string) {
  await client.query(
    `CREATE SCHEMA IF NOT EXISTS ${quoteIdentifier(schemaName)}`
  );
  await client.query(`SET search_path TO ${quoteIdentifier(schemaName)}`);
}

async function loadAppliedMigrations(client: Client) {
  const result = await client.query<{
    migration_name: string;
    checksum: string;
  }>(
    `SELECT "migration_name", "checksum"
     FROM "${MIGRATIONS_TABLE}"
     WHERE "rolled_back_at" IS NULL`
  );

  return new Map(
    result.rows.map((row) => [row.migration_name, row.checksum] as const)
  );
}

async function hasUserTables(client: Client) {
  const result = await client.query<{ exists: boolean }>(
    `
      SELECT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = current_schema()
          AND table_name = ANY($1::text[])
      ) AS "exists"
    `,
    [EXPECTED_TABLES]
  );

  return Boolean(result.rows[0]?.exists);
}

async function validateExistingSchemaIsLatest(client: Client) {
  const tableCountResult = await client.query<{ count: string }>(
    `
      SELECT COUNT(*)::text AS "count"
      FROM information_schema.tables
      WHERE table_schema = current_schema()
        AND table_name = ANY($1::text[])
    `,
    [EXPECTED_TABLES]
  );

  const existingTableCount = Number(tableCountResult.rows[0]?.count ?? 0);
  if (existingTableCount !== EXPECTED_TABLES.length) {
    return false;
  }

  const billingResult = await client.query<{ exists: boolean }>(
    `
      SELECT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'usr_users'
          AND column_name = 'token_balance'
      ) AS "exists"
    `
  );

  return Boolean(billingResult.rows[0]?.exists);
}

async function ensureDatabaseExists(target: DatabaseTarget) {
  const probeClient = new Client({
    connectionString: target.databaseUrl,
  });

  try {
    await probeClient.connect();
    await probeClient.end();
    return;
  } catch (error: unknown) {
    try {
      await probeClient.end();
    } catch {
      // ignore
    }

    if ((error as { code?: string })?.code !== "3D000") {
      throw error;
    }
  }

  const adminClient = new Client({
    connectionString: target.adminUrl,
  });

  await adminClient.connect();
  try {
    const existsResult = await adminClient.query<{ exists: boolean }>(
      "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1) AS exists",
      [target.databaseName]
    );

    if (!existsResult.rows[0]?.exists) {
      await adminClient.query(
        `CREATE DATABASE ${quoteIdentifier(target.databaseName)}`
      );
      console.log(`[db:migrate] created database ${target.databaseName}`);
    }
  } finally {
    await adminClient.end();
  }
}

async function baselineMigrations(
  client: Client,
  migrations: MigrationFile[],
  applied: Map<string, string>
) {
  if (applied.size > 0) {
    return;
  }

  const hasTables = await hasUserTables(client);
  if (!hasTables) {
    return;
  }

  const isLatestSchema = await validateExistingSchemaIsLatest(client);
  if (!isLatestSchema) {
    throw new Error(
      "Existing database schema detected without _prisma_migrations history, and the schema does not match the current expected version. Refusing automatic baseline to avoid duplicate execution or destructive drift."
    );
  }

  await client.query("BEGIN");
  try {
    for (const migration of migrations) {
      await client.query(
        `INSERT INTO "${MIGRATIONS_TABLE}" ("id", "checksum", "finished_at", "migration_name", "started_at", "applied_steps_count") VALUES ($1, $2, NOW(), $3, NOW(), 1)`,
        [randomUUID(), migration.checksum, migration.name]
      );
      applied.set(migration.name, migration.checksum);
    }
    await client.query("COMMIT");
    console.log("[db:migrate] baselined existing schema");
  } catch (error) {
    await client.query("ROLLBACK");
    throw error;
  }
}

async function applyPendingMigrations() {
  const target = getDatabaseTarget();

  const migrations = await readMigrationFiles();
  if (migrations.length < 1) {
    return;
  }

  await ensureDatabaseExists(target);

  const client = new Client({
    connectionString: target.databaseUrl,
  });

  await client.connect();

  try {
    await ensureSchema(client, target.schemaName);
    await client.query("SELECT pg_advisory_lock($1, $2)", [
      LOCK_KEY_1,
      LOCK_KEY_2,
    ]);
    await ensureMigrationsTable(client);

    const applied = await loadAppliedMigrations(client);
    await baselineMigrations(client, migrations, applied);

    for (const migration of migrations) {
      const appliedChecksum = applied.get(migration.name);

      if (appliedChecksum) {
        if (appliedChecksum !== migration.checksum) {
          throw new Error(
            `Migration checksum mismatch: ${migration.name}. Existing databases cannot safely apply a modified migration file.`
          );
        }
        continue;
      }

      await client.query("BEGIN");
      try {
        await client.query(migration.sql);
        await client.query(
          `INSERT INTO "${MIGRATIONS_TABLE}" ("id", "checksum", "finished_at", "migration_name", "started_at", "applied_steps_count") VALUES ($1, $2, NOW(), $3, NOW(), 1)`,
          [randomUUID(), migration.checksum, migration.name]
        );
        await client.query("COMMIT");
        console.log(`[db:migrate] applied ${migration.name}`);
      } catch (error) {
        await client.query("ROLLBACK");
        throw error;
      }
    }
  } finally {
    try {
      await client.query("SELECT pg_advisory_unlock($1, $2)", [
        LOCK_KEY_1,
        LOCK_KEY_2,
      ]);
    } catch {
      // ignore
    }
    await client.end();
  }
}

export async function ensureDatabaseMigrations() {
  const enabled = useRuntimeConfig().manager.databaseAutoMigrate;
  if (enabled === false || enabled === "false") {
    return;
  }

  migrationTask ??= applyPendingMigrations();
  return migrationTask;
}
