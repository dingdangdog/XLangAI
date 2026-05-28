import prisma from "../../../lib/prisma";
import { probeTtsConfig } from "../../../lib/serviceProbe/ttsProbe";

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id?.trim()) {
    throw createError({ statusCode: 400, message: "缺少配置 id" });
  }
  const row = await prisma.ttsServiceConfig.findUnique({ where: { id } });
  if (!row) {
    throw createError({ statusCode: 404, message: "TTS 配置不存在" });
  }
  const body = (await readBody(event).catch(() => ({}))) as Record<string, unknown>;
  const result = await probeTtsConfig({
    provider: String(body.provider ?? row.provider),
    baseUrl: body.baseUrl != null ? String(body.baseUrl) : row.baseUrl,
    apiKey: body.apiKey != null ? String(body.apiKey) : row.apiKey,
    region: body.region != null ? String(body.region) : row.region,
    modelCode: String(body.modelCode ?? row.modelCode),
    config: body.config != null ? String(body.config) : row.config,
    voiceCode: body.voiceCode != null ? String(body.voiceCode) : null,
  });
  return result;
});
