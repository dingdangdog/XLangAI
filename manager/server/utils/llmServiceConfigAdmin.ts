import { createError } from "h3";
import {
  isSupportedLlmProtocol,
  llmStoredProtocol,
  type LlmOpenAiFlavor,
  type LlmProtocolFamily,
} from "../lib/serverServiceCatalog";

export function autoLlmConfigCode(protocol: string): string {
  const p = (protocol || "openai").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
}

export async function prepareLlmServiceConfigWrite(
  data: Record<string, unknown>,
): Promise<Record<string, unknown>> {
  const next = { ...data };
  const protocol = String(next.protocol ?? "openai").trim().toLowerCase();

  if (!isSupportedLlmProtocol(protocol)) {
    throw createError({
      statusCode: 400,
      message: `不支持的 LLM 协议: ${protocol}（仅允许 Server 已实现的协议族）`,
    });
  }

  if (!String(next.code ?? "").trim()) {
    next.code = autoLlmConfigCode(protocol);
  }

  return next;
}

/** 供前端 POST body 可选字段：protocolFamily + openAiFlavor 优先于 protocol */
export function resolveLlmProtocolFromBody(body: Record<string, unknown>): string {
  const family = String(body.protocolFamily ?? "").trim() as LlmProtocolFamily;
  if (family === "openai" || family === "claude" || family === "gemini") {
    const flavor = String(body.openAiFlavor ?? "generic").trim() as LlmOpenAiFlavor;
    return llmStoredProtocol(family, flavor === "azure" ? "azure" : "generic");
  }
  return String(body.protocol ?? "openai").trim().toLowerCase();
}
