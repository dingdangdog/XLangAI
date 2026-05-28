import type { ObjectStorageProvider, ObjectStorageRuntimeConfig } from "./types";

type DbRow = {
  id: string;
  provider?: string | null;
  baseUrl?: string | null;
  publicBaseUrl?: string | null;
  apiKey?: string | null;
  secretKey?: string | null;
  bucket?: string | null;
  region?: string | null;
  config?: string | null;
};

const PROVIDERS: ObjectStorageProvider[] = ["local", "cloudflare_r2", "qiniu", "aliyun_oss"];

function parseExtra(raw?: string | null): Record<string, string> {
  const text = (raw ?? "").trim();
  if (!text) return {};
  try {
    const parsed = JSON.parse(text) as Record<string, unknown>;
    const out: Record<string, string> = {};
    for (const [k, v] of Object.entries(parsed)) {
      if (v != null) out[k] = String(v);
    }
    return out;
  } catch {
    return {};
  }
}

function normalizeProvider(raw?: string | null): ObjectStorageProvider {
  const v = (raw ?? "local").trim().toLowerCase();
  return PROVIDERS.includes(v as ObjectStorageProvider) ? (v as ObjectStorageProvider) : "local";
}

export function parseObjectStorageConfig(row: DbRow): ObjectStorageRuntimeConfig {
  return {
    id: row.id,
    provider: normalizeProvider(row.provider),
    endpoint: (row.baseUrl ?? "").trim(),
    publicBaseUrl: (row.publicBaseUrl ?? "").trim().replace(/\/$/, ""),
    accessKey: (row.apiKey ?? "").trim(),
    secretKey: (row.secretKey ?? "").trim(),
    bucket: (row.bucket ?? "").trim(),
    region: (row.region ?? "").trim(),
    extra: parseExtra(row.config),
  };
}

export function joinPublicUrl(base: string, key: string): string {
  const normBase = base.trim().replace(/\/$/, "");
  const normKey = key.trim().replace(/^\//, "");
  if (!normBase) return normKey;
  if (!normKey) return normBase;
  if (normBase.startsWith("http://") || normBase.startsWith("https://")) {
    return `${normBase}/${normKey}`;
  }
  return `https://${normBase}/${normKey}`;
}
