import prisma from "~~/server/lib/prisma";
import { probeSmsConfig } from "~~/server/lib/serviceProbe/smsProbe";

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id?.trim()) {
    throw createError({ statusCode: 400, message: "缺少配置 id" });
  }
  const row = await prisma.sysSmsServiceConfig.findUnique({ where: { id } });
  if (!row) {
    throw createError({ statusCode: 404, message: "短信配置不存在" });
  }
  const body = (await readBody(event).catch(() => ({}))) as Record<string, unknown>;
  const result = await probeSmsConfig({
    provider: String(body.provider ?? row.provider),
    apiKey: body.apiKey != null ? String(body.apiKey) : row.apiKey,
    secretKey: body.secretKey != null ? String(body.secretKey) : row.secretKey,
    region: body.region != null ? String(body.region) : row.region,
    signName: body.signName != null ? String(body.signName) : row.signName,
    templateCode: body.templateCode != null ? String(body.templateCode) : row.templateCode,
    config: body.config != null ? String(body.config) : row.config,
  });
  return result;
});
