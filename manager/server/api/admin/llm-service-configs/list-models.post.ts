import prisma from "../../../lib/prisma";
import { listOpenAiCompatibleModels } from "../../../lib/llmModelsList";
import { getLlmVendorPreset, isOpenAiCompatPreset } from "../../../lib/llmVendorPresets";
import { isLlmOpenAiCompatibleProtocol } from "../../../lib/serverServiceCatalog";

export default defineEventHandler(async (event) => {
  const body = (await readBody(event).catch(() => ({}))) as Record<string, unknown>;
  const configId = String(body.configId ?? "").trim();

  let protocol = String(body.protocol ?? "").trim().toLowerCase();
  let baseUrl = body.baseUrl != null ? String(body.baseUrl) : "";
  let apiKey = body.apiKey != null ? String(body.apiKey) : "";

  if (configId) {
    const row = await prisma.sysLlmServiceConfig.findUnique({ where: { id: configId } });
    if (!row) {
      throw createError({ statusCode: 404, message: "LLM 配置不存在" });
    }
    protocol = String(body.protocol ?? row.protocol ?? "openai").trim().toLowerCase();
    if (body.baseUrl == null) baseUrl = String(row.baseUrl ?? "");
    if (body.apiKey == null) apiKey = String(row.apiKey ?? "");
  }

  const presetId = String(body.vendorPresetId ?? "").trim();
  const preset = presetId ? getLlmVendorPreset(presetId) : undefined;
  if (preset && !preset.isCustom && preset.baseUrl && !baseUrl.trim()) {
    baseUrl = preset.baseUrl;
  }

  if (protocol === "claude" || protocol === "anthropic" || protocol === "gemini" || protocol === "google_gemini") {
    return {
      models: [],
      manualOnly: true,
      message: "该协议暂不支持后台拉取模型列表，请手动输入 model",
    };
  }

  if (!isLlmOpenAiCompatibleProtocol(protocol) && protocol !== "openai") {
    return {
      models: [],
      manualOnly: true,
      message: "不支持的协议",
    };
  }

  if (preset && !isOpenAiCompatPreset(preset)) {
    return { models: [], manualOnly: true };
  }

  const result = await listOpenAiCompatibleModels({ baseUrl, apiKey });
  if (result.error && !result.models.length) {
    return {
      models: [],
      manualOnly: false,
      error: result.error,
    };
  }
  return {
    models: result.models,
    manualOnly: false,
    error: result.error,
  };
});
