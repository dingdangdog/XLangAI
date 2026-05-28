import prisma from "../lib/prisma";

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
  if (next.status === "active") {
    await deactivateOtherSmsServiceConfigs(id);
  }
  return next;
}
