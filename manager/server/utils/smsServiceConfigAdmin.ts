import prisma from "../lib/prisma";
import { createError } from "h3";
import { isSupportedSmsProvider, normalizeSmsProvider } from "../lib/serverServiceCatalog";

/** 启用一条短信配置时，将其余活跃配置设为 inactive（全局仅一条 active）。 */
export async function deactivateOtherSmsServiceConfigs(excludeId?: string) {
  const where: { status: string; id?: { not: string } } = { status: "active" };
  if (excludeId) {
    where.id = { not: excludeId };
  }
  await prisma.sysSmsServiceConfig.updateMany({
    where,
    data: { status: "inactive" },
  });
}

export async function prepareSmsServiceConfigWrite(
  data: Record<string, unknown>,
  id?: string,
): Promise<Record<string, unknown>> {
  const next = { ...data };
  const provider = String(next.provider ?? "aliyun").trim();
  if (!isSupportedSmsProvider(provider)) {
    throw createError({
      statusCode: 400,
      message: `不支持的短信 provider: ${provider}（仅 aliyun / tencent）`,
    });
  }
  next.provider = normalizeSmsProvider(provider);
  if (next.status === "active") {
    await deactivateOtherSmsServiceConfigs(id);
  }
  return next;
}
