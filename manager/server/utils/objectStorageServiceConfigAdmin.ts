import prisma from "../lib/prisma";

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
  if (next.status === "active") {
    await deactivateOtherObjectStorageConfigs(id);
  }
  return next;
}
