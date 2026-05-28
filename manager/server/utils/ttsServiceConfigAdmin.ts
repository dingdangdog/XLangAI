import { createError } from "h3";
import { isSupportedTtsProvider } from "../lib/serverServiceCatalog";

export function autoTtsConfigCode(provider: string): string {
  const p = (provider || "openai_rest").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
}

export async function prepareTtsServiceConfigWrite(
  data: Record<string, unknown>,
): Promise<Record<string, unknown>> {
  const next = { ...data };
  const provider = String(next.provider ?? "").trim().toLowerCase();

  if (!isSupportedTtsProvider(provider)) {
    throw createError({
      statusCode: 400,
      message: `不支持的 TTS provider: ${provider}（Go Server 未实现该厂商）`,
    });
  }

  if (!String(next.code ?? "").trim()) {
    next.code = autoTtsConfigCode(provider);
  }

  return next;
}
