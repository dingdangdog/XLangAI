import type { ServiceProbeResult } from "./types";

type LlmProbeInput = {
  protocol: string;
  baseUrl: string | null;
  apiKey: string | null;
  modelCode: string;
  config: string | null;
};

function parseConfig(raw: string | null): Record<string, unknown> {
  if (!raw?.trim()) return {};
  try {
    return JSON.parse(raw) as Record<string, unknown>;
  } catch {
    return {};
  }
}

function openAiCompatibleRoot(baseUrl: string | null): string {
  const root = (baseUrl?.trim() || "https://api.openai.com").replace(/\/$/, "");
  return root.endsWith("/v1") ? root.slice(0, -3) : root;
}

async function probeOpenAiCompatible(input: LlmProbeInput): Promise<ServiceProbeResult> {
  const start = Date.now();
  const key = (input.apiKey ?? "").trim();
  if (!key) {
    return { ok: false, latencyMs: 0, message: "缺少 API Key" };
  }
  const root = openAiCompatibleRoot(input.baseUrl);
  const res = await fetch(`${root}/v1/chat/completions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${key}`,
    },
    body: JSON.stringify({
      model: input.modelCode.trim(),
      messages: [{ role: "user", content: "ping" }],
      max_tokens: 5,
    }),
  });
  const latencyMs = Date.now() - start;
  if (!res.ok) {
    const body = await res.text();
    return {
      ok: false,
      latencyMs,
      message: `HTTP ${res.status}`,
      detail: body.slice(0, 400),
    };
  }
  return {
    ok: true,
    latencyMs,
    message: "模型响应正常",
    detail: `endpoint: ${root}/v1/chat/completions`,
  };
}

async function probeClaude(input: LlmProbeInput): Promise<ServiceProbeResult> {
  const start = Date.now();
  const key = (input.apiKey ?? "").trim();
  if (!key) {
    return { ok: false, latencyMs: 0, message: "缺少 API Key" };
  }
  const cfg = parseConfig(input.config);
  const root = (input.baseUrl?.trim() || "https://api.anthropic.com").replace(/\/$/, "");
  const version = String(cfg.anthropic_version ?? "2023-06-01");
  const res = await fetch(`${root}/v1/messages`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "x-api-key": key,
      "anthropic-version": version,
    },
    body: JSON.stringify({
      model: input.modelCode.trim(),
      max_tokens: 8,
      messages: [{ role: "user", content: "ping" }],
    }),
  });
  const latencyMs = Date.now() - start;
  if (!res.ok) {
    const body = await res.text();
    return {
      ok: false,
      latencyMs,
      message: `HTTP ${res.status}`,
      detail: body.slice(0, 400),
    };
  }
  return { ok: true, latencyMs, message: "Claude 响应正常" };
}

async function probeGemini(input: LlmProbeInput): Promise<ServiceProbeResult> {
  const start = Date.now();
  const key = (input.apiKey ?? "").trim();
  if (!key) {
    return { ok: false, latencyMs: 0, message: "缺少 API Key" };
  }
  const model = encodeURIComponent(input.modelCode.trim());
  const root =
    (input.baseUrl?.trim() || "https://generativelanguage.googleapis.com").replace(/\/$/, "");
  const url = `${root}/v1beta/models/${model}:generateContent?key=${encodeURIComponent(key)}`;
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      contents: [{ role: "user", parts: [{ text: "ping" }] }],
      generationConfig: { maxOutputTokens: 8 },
    }),
  });
  const latencyMs = Date.now() - start;
  if (!res.ok) {
    const body = await res.text();
    return {
      ok: false,
      latencyMs,
      message: `HTTP ${res.status}`,
      detail: body.slice(0, 400),
    };
  }
  return { ok: true, latencyMs, message: "Gemini 响应正常" };
}

export async function probeLlmConfig(input: LlmProbeInput): Promise<ServiceProbeResult> {
  const protocol = (input.protocol ?? "openai").trim().toLowerCase();
  if (!input.modelCode?.trim()) {
    return { ok: false, latencyMs: 0, message: "缺少 model_code" };
  }
  try {
    if (protocol === "claude") return await probeClaude(input);
    if (protocol === "gemini") return await probeGemini(input);
    return await probeOpenAiCompatible(input);
  } catch (e) {
    return {
      ok: false,
      latencyMs: 0,
      message: e instanceof Error ? e.message : "探活失败",
    };
  }
}
