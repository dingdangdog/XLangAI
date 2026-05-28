import { createError } from "h3";
import prisma from "../lib/prisma";
import { isSupportedTranslateProtocol, LLM_OPENAI_COMPAT_STORED_PROTOCOLS } from "../lib/serverServiceCatalog";

const OPENAI_COMPAT_PROTOCOLS = new Set<string>(["", ...LLM_OPENAI_COMPAT_STORED_PROTOCOLS]);

/** 库表 code 唯一、Go 不按 code 查；新建时由协议自动生成。 */
export function autoTranslateConfigCode(protocol: string): string {
  const p = (protocol || "openai").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
}

/** 启用一条翻译配置时，将其余活跃配置设为 inactive（全局仅一条 active）。 */
export async function deactivateOtherTranslateConfigs(excludeId?: string) {
  const where: { status: string; id?: { not: string } } = { status: "active" };
  if (excludeId) {
    where.id = { not: excludeId };
  }
  await prisma.sysTranslateServiceConfig.updateMany({
    where,
    data: { status: "inactive" },
  });
}

export async function prepareTranslateServiceConfigWrite(
  data: Record<string, unknown>,
  id?: string,
): Promise<Record<string, unknown>> {
  const next = { ...data };
  const protocol = String(next.protocol ?? "openai").trim().toLowerCase();

  if (!isSupportedTranslateProtocol(protocol)) {
    throw createError({
      statusCode: 400,
      message: `不支持的翻译协议: ${protocol}`,
    });
  }

  if (!String(next.code ?? "").trim()) {
    next.code = autoTranslateConfigCode(protocol);
  }

  const llmConfigId =
    typeof next.llmConfigId === "string" ? next.llmConfigId.trim() : "";
  if (protocol === "openai" && llmConfigId) {
    const llm = await prisma.sysLlmServiceConfig.findUnique({
      where: { id: llmConfigId },
    });
    if (!llm || llm.status !== "active") {
      throw createError({
        statusCode: 400,
        message: "所选 LLM 配置不存在或未启用",
      });
    }
    if (!OPENAI_COMPAT_PROTOCOLS.has(String(llm.protocol ?? "").trim().toLowerCase())) {
      throw createError({
        statusCode: 400,
        message: "关联的 LLM 须为 OpenAI 兼容协议",
      });
    }
    next.llmConfigId = llmConfigId;
    next.baseUrl = null;
    next.apiKey = null;
    if (!String(next.modelCode ?? "").trim() || next.modelCode === "-") {
      next.modelCode = llm.modelCode;
    }
  } else if (protocol === "openai") {
    next.llmConfigId = null;
  } else {
    next.llmConfigId = null;
    if (!String(next.modelCode ?? "").trim()) {
      next.modelCode = "-";
    }
  }

  if (next.status === "active") {
    await deactivateOtherTranslateConfigs(id);
  }
  return next;
}
