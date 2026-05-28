import prisma from "../../../lib/prisma";
import { probeLlmConfig } from "../../../lib/serviceProbe/llmProbe";

export default defineEventHandler(async (event) => {
  const id = getRouterParam(event, "id");
  if (!id?.trim()) {
    throw createError({ statusCode: 400, message: "缺少配置 id" });
  }
  const row = await prisma.sysLlmServiceConfig.findUnique({ where: { id } });
  if (!row) {
    throw createError({ statusCode: 404, message: "LLM 配置不存在" });
  }
  const body = (await readBody(event).catch(() => ({}))) as Record<string, unknown>;
  const result = await probeLlmConfig({
    protocol: String(body.protocol ?? row.protocol),
    baseUrl: body.baseUrl != null ? String(body.baseUrl) : row.baseUrl,
    apiKey: body.apiKey != null ? String(body.apiKey) : row.apiKey,
    modelCode: String(body.modelCode ?? row.modelCode),
    config: body.config != null ? String(body.config) : row.config,
  });
  return result;
});
