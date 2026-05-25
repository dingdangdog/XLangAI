import prisma from "../lib/prisma";

/** 启用一条 STT 配置时，将其余活跃配置设为 inactive（与 Go GetActive 单条语义一致）。 */
export async function deactivateOtherSttConfigs(excludeId?: string) {
  const where: { status: string; id?: { not: string } } = { status: "active" };
  if (excludeId) {
    where.id = { not: excludeId };
  }
  await prisma.sysSttServiceConfig.updateMany({
    where,
    data: { status: "inactive" },
  });
}

export async function prepareSttServiceConfigWrite(
  data: Record<string, unknown>,
  id?: string,
): Promise<Record<string, unknown>> {
  const next = { ...data };
  if (next.status === "active") {
    await deactivateOtherSttConfigs(id);
  }
  return next;
}
