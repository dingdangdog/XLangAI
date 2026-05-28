import crypto from "node:crypto";
import type { ServiceProbeResult } from "./types";

type SmsProbeInput = {
  provider: string;
  apiKey: string | null;
  secretKey: string | null;
  region: string | null;
  signName: string | null;
  templateCode: string | null;
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

function percentEncode(value: string): string {
  return encodeURIComponent(value)
    .replace(/\+/g, "%20")
    .replace(/\*/g, "%2A")
    .replace(/%7E/g, "~");
}

function aliyunRpcSign(params: Record<string, string>, secret: string): string {
  const sorted = Object.keys(params)
    .sort()
    .map((k) => `${percentEncode(k)}=${percentEncode(params[k] ?? "")}`)
    .join("&");
  const stringToSign = `POST&${percentEncode("/")}&${percentEncode(sorted)}`;
  const hmac = crypto.createHmac("sha1", `${secret}&`);
  hmac.update(stringToSign);
  return hmac.digest("base64");
}

async function probeAliyun(input: SmsProbeInput): Promise<ServiceProbeResult> {
  const start = Date.now();
  const accessKeyId = (input.apiKey ?? "").trim();
  const accessKeySecret = (input.secretKey ?? "").trim();
  const signName = (input.signName ?? "").trim();
  const templateCode = (input.templateCode ?? "").trim();
  if (!accessKeyId || !accessKeySecret) {
    return { ok: false, latencyMs: 0, message: "缺少 AccessKey 或 Secret" };
  }
  if (!signName || !templateCode) {
    return { ok: false, latencyMs: 0, message: "缺少签名或模板 ID" };
  }
  const cfg = parseConfig(input.config);
  const endpoint = String(cfg.endpoint ?? "dysmsapi.aliyuncs.com");
  const params: Record<string, string> = {
    AccessKeyId: accessKeyId,
    Action: "QuerySmsTemplate",
    Format: "JSON",
    RegionId: (input.region ?? "cn-hangzhou").trim() || "cn-hangzhou",
    SignatureMethod: "HMAC-SHA1",
    SignatureNonce: crypto.randomUUID(),
    SignatureVersion: "1.0",
    TemplateCode: templateCode,
    Timestamp: new Date().toISOString().replace(/\.\d{3}Z$/, "Z"),
    Version: "2017-05-25",
  };
  params.Signature = aliyunRpcSign(params, accessKeySecret);
  const body = new URLSearchParams(params).toString();
  const res = await fetch(`https://${endpoint}/`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
  });
  const latencyMs = Date.now() - start;
  const text = await res.text();
  let parsed: Record<string, unknown> = {};
  try {
    parsed = JSON.parse(text) as Record<string, unknown>;
  } catch {
    return {
      ok: false,
      latencyMs,
      message: `HTTP ${res.status}`,
      detail: text.slice(0, 400),
    };
  }
  const code = String(parsed.Code ?? parsed.code ?? "");
  if (code && code !== "OK") {
    return {
      ok: false,
      latencyMs,
      message: code,
      detail: String(parsed.Message ?? parsed.message ?? text.slice(0, 200)),
    };
  }
  return {
    ok: true,
    latencyMs,
    message: "模板校验通过",
    detail: `sign=${signName}, template=${templateCode}`,
  };
}

async function probeTencent(input: SmsProbeInput): Promise<ServiceProbeResult> {
  const start = Date.now();
  const secretId = (input.apiKey ?? "").trim();
  const secretKey = (input.secretKey ?? "").trim();
  const templateCode = (input.templateCode ?? "").trim();
  const cfg = parseConfig(input.config);
  const sdkAppId = String(cfg.sdk_app_id ?? "").trim();
  if (!secretId || !secretKey) {
    return { ok: false, latencyMs: 0, message: "缺少 SecretId 或 SecretKey" };
  }
  if (!sdkAppId) {
    return { ok: false, latencyMs: 0, message: "config 中缺少 sdk_app_id" };
  }
  if (!templateCode) {
    return { ok: false, latencyMs: 0, message: "缺少模板 ID" };
  }

  const region = (input.region ?? "ap-guangzhou").trim() || "ap-guangzhou";
  const endpoint = String(cfg.endpoint ?? "sms.tencentcloudapi.com");
  const timestamp = Math.floor(Date.now() / 1000);
  const payload = JSON.stringify({
    SmsSdkAppId: sdkAppId,
    TemplateIdSet: [Number.isNaN(Number(templateCode)) ? templateCode : Number(templateCode)],
  });
  const canonicalRequest = [
    "POST",
    "/",
    "",
    "content-type:application/json; charset=utf-8",
    `host:${endpoint}`,
    "",
    "content-type;host",
    crypto.createHash("sha256").update(payload).digest("hex"),
  ].join("\n");
  const date = new Date(timestamp * 1000).toISOString().slice(0, 10);
  const credentialScope = `${date}/sms/tc3_request`;
  const stringToSign = [
    "TC3-HMAC-SHA256",
    String(timestamp),
    credentialScope,
    crypto.createHash("sha256").update(canonicalRequest).digest("hex"),
  ].join("\n");
  const secretDate = crypto.createHmac("sha256", `TC3${secretKey}`).update(date).digest();
  const secretService = crypto.createHmac("sha256", secretDate).update("sms").digest();
  const secretSigning = crypto.createHmac("sha256", secretService).update("tc3_request").digest();
  const signature = crypto.createHmac("sha256", secretSigning).update(stringToSign).digest("hex");
  const authorization = `TC3-HMAC-SHA256 Credential=${secretId}/${credentialScope}, SignedHeaders=content-type;host, Signature=${signature}`;

  const res = await fetch(`https://${endpoint}`, {
    method: "POST",
    headers: {
      Authorization: authorization,
      "Content-Type": "application/json; charset=utf-8",
      Host: endpoint,
      "X-TC-Action": "DescribeSmsTemplateList",
      "X-TC-Region": region,
      "X-TC-Timestamp": String(timestamp),
      "X-TC-Version": "2021-01-11",
    },
    body: payload,
  });
  const latencyMs = Date.now() - start;
  const text = await res.text();
  let parsed: Record<string, unknown> = {};
  try {
    parsed = JSON.parse(text) as Record<string, unknown>;
  } catch {
    return {
      ok: false,
      latencyMs,
      message: `HTTP ${res.status}`,
      detail: text.slice(0, 400),
    };
  }
  const resp = parsed.Response as Record<string, unknown> | undefined;
  if (resp?.Error) {
    const err = resp.Error as Record<string, unknown>;
    return {
      ok: false,
      latencyMs,
      message: String(err.Code ?? "Error"),
      detail: String(err.Message ?? text.slice(0, 200)),
    };
  }
  return {
    ok: true,
    latencyMs,
    message: "模板校验通过",
    detail: `sdk_app_id=${sdkAppId}, template=${templateCode}`,
  };
}

export async function probeSmsConfig(input: SmsProbeInput): Promise<ServiceProbeResult> {
  const provider = (input.provider ?? "aliyun").trim().toLowerCase();
  try {
    if (provider === "tencent") return await probeTencent(input);
    return await probeAliyun(input);
  } catch (e) {
    return {
      ok: false,
      latencyMs: 0,
      message: e instanceof Error ? e.message : "短信探活失败",
    };
  }
}
