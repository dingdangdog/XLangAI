export type LlmModelOption = {
  id: string;
  label?: string;
};

function normalizeOpenAiRoot(baseUrl: string | null | undefined): string {
  const root = (baseUrl?.trim() || "https://api.openai.com").replace(/\/$/, "");
  return root.endsWith("/v1") ? root.slice(0, -3) : root;
}

function isVolcengineArkBaseUrl(baseUrl: string): boolean {
  const u = baseUrl.trim().toLowerCase();
  return u.includes("volces.com") || u.includes("volcengine.com");
}

function openAiModelsUrl(baseUrl: string | null | undefined): string {
  const root = normalizeOpenAiRoot(baseUrl);
  if (isVolcengineArkBaseUrl(root)) {
    return `${root}/models`;
  }
  return `${root}/v1/models`;
}

/** OpenAI 兼容 GET /v1/models（方舟为 /api/v3/models） */
export async function listOpenAiCompatibleModels(args: {
  baseUrl: string | null;
  apiKey: string | null;
}): Promise<{ models: LlmModelOption[]; error?: string }> {
  const key = (args.apiKey ?? "").trim();
  if (!key) {
    return { models: [], error: "缺少 API Key" };
  }
  const url = openAiModelsUrl(args.baseUrl);
  try {
    const res = await fetch(url, {
      headers: { Authorization: `Bearer ${key}` },
    });
    const text = await res.text();
    if (!res.ok) {
      return {
        models: [],
        error: `HTTP ${res.status}: ${text.slice(0, 200)}`,
      };
    }
    let parsed: { data?: Array<{ id?: string }> };
    try {
      parsed = JSON.parse(text) as { data?: Array<{ id?: string }> };
    } catch {
      return { models: [], error: "模型列表响应不是 JSON" };
    }
    const models = (parsed.data ?? [])
      .map((m) => String(m.id ?? "").trim())
      .filter(Boolean)
      .sort((a, b) => a.localeCompare(b))
      .map((id) => ({ id }));
    return { models };
  } catch (e) {
    return {
      models: [],
      error: e instanceof Error ? e.message : "拉取模型失败",
    };
  }
}
