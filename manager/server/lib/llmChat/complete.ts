type LlmConfigRow = {
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

async function chatOpenAiCompatible(
  input: LlmConfigRow,
  system: string,
  user: string,
): Promise<string> {
  const key = (input.apiKey ?? "").trim();
  if (!key) throw new Error("LLM 未配置 API Key");
  const model = input.modelCode.trim();
  if (!model) throw new Error("LLM 未配置 model_code");
  const root = openAiCompatibleRoot(input.baseUrl);
  const res = await fetch(`${root}/v1/chat/completions`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${key}`,
    },
    body: JSON.stringify({
      model,
      messages: [
        { role: "system", content: system },
        { role: "user", content: user },
      ],
      temperature: 0.7,
    }),
  });
  const text = await res.text();
  if (!res.ok) {
    throw new Error(`LLM HTTP ${res.status}: ${text.slice(0, 300)}`);
  }
  let parsed: { choices?: Array<{ message?: { content?: string } }> };
  try {
    parsed = JSON.parse(text) as typeof parsed;
  } catch {
    throw new Error("LLM 响应不是有效 JSON");
  }
  const content = parsed.choices?.[0]?.message?.content?.trim();
  if (!content) throw new Error("LLM 返回空内容");
  return content;
}

async function chatClaude(input: LlmConfigRow, system: string, user: string): Promise<string> {
  const key = (input.apiKey ?? "").trim();
  if (!key) throw new Error("LLM 未配置 API Key");
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
      max_tokens: 8192,
      system,
      messages: [{ role: "user", content: user }],
    }),
  });
  const text = await res.text();
  if (!res.ok) {
    throw new Error(`Claude HTTP ${res.status}: ${text.slice(0, 300)}`);
  }
  let parsed: { content?: Array<{ type?: string; text?: string }> };
  try {
    parsed = JSON.parse(text) as typeof parsed;
  } catch {
    throw new Error("Claude 响应不是有效 JSON");
  }
  const content = parsed.content
    ?.map((p) => (p.type === "text" ? p.text ?? "" : ""))
    .join("")
    .trim();
  if (!content) throw new Error("Claude 返回空内容");
  return content;
}

async function chatGemini(input: LlmConfigRow, system: string, user: string): Promise<string> {
  const key = (input.apiKey ?? "").trim();
  if (!key) throw new Error("LLM 未配置 API Key");
  const model = encodeURIComponent(input.modelCode.trim());
  const root =
    (input.baseUrl?.trim() || "https://generativelanguage.googleapis.com").replace(/\/$/, "");
  const url = `${root}/v1beta/models/${model}:generateContent?key=${encodeURIComponent(key)}`;
  const res = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      systemInstruction: { parts: [{ text: system }] },
      contents: [{ role: "user", parts: [{ text: user }] }],
      generationConfig: { temperature: 0.7 },
    }),
  });
  const text = await res.text();
  if (!res.ok) {
    throw new Error(`Gemini HTTP ${res.status}: ${text.slice(0, 300)}`);
  }
  let parsed: {
    candidates?: Array<{ content?: { parts?: Array<{ text?: string }> } }>;
  };
  try {
    parsed = JSON.parse(text) as typeof parsed;
  } catch {
    throw new Error("Gemini 响应不是有效 JSON");
  }
  const content = parsed.candidates?.[0]?.content?.parts
    ?.map((p) => p.text ?? "")
    .join("")
    .trim();
  if (!content) throw new Error("Gemini 返回空内容");
  return content;
}

/** 调用已配置的 LLM 服务，返回 assistant 文本。 */
export async function llmChatComplete(
  input: LlmConfigRow,
  system: string,
  user: string,
): Promise<string> {
  const protocol = (input.protocol ?? "openai").trim().toLowerCase();
  if (protocol === "claude" || protocol === "anthropic") {
    return chatClaude(input, system, user);
  }
  if (protocol === "gemini" || protocol === "google_gemini") {
    return chatGemini(input, system, user);
  }
  return chatOpenAiCompatible(input, system, user);
}
