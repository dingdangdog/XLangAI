import prisma from "../lib/prisma";

const PROVIDERS = ["local", "cloudflare_r2", "qiniu", "aliyun_oss"] as const;

function str(v: unknown): string {
  return typeof v === "string" ? v.trim() : "";
}

function parseExtra(raw: unknown): Record<string, string> {
  const text = str(raw);
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

function missingFields(provider: string, data: Record<string, unknown>): string[] {
  const baseUrl = str(data.baseUrl);
  const publicBaseUrl = str(data.publicBaseUrl);
  const apiKey = str(data.apiKey);
  const secretKey = str(data.secretKey);
  const bucket = str(data.bucket);
  const missing: string[] = [];

  switch (provider) {
    case "local":
      return missing;
    case "cloudflare_r2":
      if (!apiKey) missing.push("apiKey");
      if (!secretKey) missing.push("secretKey");
      if (!baseUrl) missing.push("baseUrl");
      if (!bucket) missing.push("bucket");
      if (!publicBaseUrl) missing.push("publicBaseUrl");
      return missing;
    case "qiniu":
      if (!apiKey) missing.push("apiKey");
      if (!secretKey) missing.push("secretKey");
      if (!bucket) missing.push("bucket");
      if (!publicBaseUrl) missing.push("publicBaseUrl");
      return missing;
    case "aliyun_oss":
      if (!apiKey) missing.push("apiKey");
      if (!secretKey) missing.push("secretKey");
      if (!baseUrl) missing.push("baseUrl");
      if (!bucket) missing.push("bucket");
      if (!publicBaseUrl) missing.push("publicBaseUrl");
      return missing;
    default:
      return ["provider"];
  }
}

/** 启用一条对象存储配置时，将其余活跃配置设为 inactive（全局仅一条 active）。 */
export async function deactivateOtherObjectStorageConfigs(excludeId?: string) {
  const where: { status: string; id?: { not: string } } = { status: "active" };
  if (excludeId) {
    where.id = { not: excludeId };
  }
  await prisma.sysObjectStorageConfig.updateMany({
    where,
    data: { status: "inactive" },
  });
}

export async function prepareObjectStorageServiceConfigWrite(
  data: Record<string, unknown>,
  id?: string,
): Promise<Record<string, unknown>> {
  const next = { ...data };
  const provider = str(next.provider).toLowerCase() || "local";
  if (!PROVIDERS.includes(provider as (typeof PROVIDERS)[number])) {
    throw createError({
      statusCode: 400,
      statusMessage: `unsupported object storage provider: ${provider}`,
    });
  }
  next.provider = provider;

  const configRaw = str(next.config);
  if (configRaw) {
    try {
      JSON.parse(configRaw);
    } catch {
      throw createError({ statusCode: 400, statusMessage: "config must be valid JSON" });
    }
  } else {
    next.config = "{}";
  }

  if (next.status === "active" && provider !== "local") {
    const missing = missingFields(provider, next);
    if (missing.length) {
      throw createError({
        statusCode: 400,
        statusMessage: `object storage config incomplete for ${provider}: missing ${missing.join(", ")}`,
      });
    }
  }

  parseExtra(next.config);

  if (next.status === "active") {
    await deactivateOtherObjectStorageConfigs(id);
  }
  return next;
}
