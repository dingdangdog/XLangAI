import { createHmac } from "node:crypto";
import type { ObjectStorageRuntimeConfig, UploadObjectInput, UploadObjectResult } from "../types";
import { joinPublicUrl } from "../runtime";

function credsReady(cfg: ObjectStorageRuntimeConfig): boolean {
  return Boolean(cfg.accessKey && cfg.secretKey && cfg.bucket && cfg.publicBaseUrl);
}

function uploadHost(cfg: ObjectStorageRuntimeConfig): string {
  if (cfg.endpoint) return cfg.endpoint.replace(/\/$/, "");
  switch ((cfg.extra.zone ?? "z0").trim().toLowerCase()) {
    case "z1":
      return "https://upload-z1.qiniup.com";
    case "z2":
      return "https://upload-z2.qiniup.com";
    case "na0":
      return "https://upload-na0.qiniup.com";
    case "as0":
      return "https://upload-as0.qiniup.com";
    default:
      return "https://upload.qiniup.com";
  }
}

function rsHost(cfg: ObjectStorageRuntimeConfig): string {
  if (cfg.endpoint) {
    const u = new URL(cfg.endpoint.startsWith("http") ? cfg.endpoint : `https://${cfg.endpoint}`);
    return `${u.protocol}//rs.${u.hostname.replace(/^upload(-z[0-2]|-na0|-as0)?\./, "")}`;
  }
  switch ((cfg.extra.zone ?? "z0").trim().toLowerCase()) {
    case "z1":
      return "https://rs-z1.qiniup.com";
    case "z2":
      return "https://rs-z2.qiniup.com";
    case "na0":
      return "https://rs-na0.qiniup.com";
    case "as0":
      return "https://rs-as0.qiniup.com";
    default:
      return "https://rs.qiniu.com";
  }
}

function urlSafeBase64(input: string | Buffer): string {
  const b64 = Buffer.isBuffer(input) ? input.toString("base64") : Buffer.from(input).toString("base64");
  return b64.replace(/\+/g, "-").replace(/\//g, "_");
}

function uploadToken(cfg: ObjectStorageRuntimeConfig, key: string, contentType: string): string {
  const deadline = Math.floor(Date.now() / 1000) + 3600;
  const policy: Record<string, unknown> = {
    scope: `${cfg.bucket}:${key}`,
    deadline,
  };
  if (contentType) policy.mimeLimit = contentType;
  const encoded = urlSafeBase64(JSON.stringify(policy));
  const sign = createHmac("sha1", cfg.secretKey).update(encoded).digest("base64");
  const safeSign = sign.replace(/\+/g, "-").replace(/\//g, "_");
  return `${cfg.accessKey}:${safeSign}:${encoded}`;
}

function qboxAuth(cfg: ObjectStorageRuntimeConfig, data: string): string {
  const encoded = urlSafeBase64(data);
  const sign = createHmac("sha1", cfg.secretKey).update(encoded).digest("base64");
  const safeSign = sign.replace(/\+/g, "-").replace(/\//g, "_");
  return `QBox ${cfg.accessKey}:${safeSign}`;
}

export async function uploadQiniu(
  cfg: ObjectStorageRuntimeConfig,
  input: UploadObjectInput,
): Promise<UploadObjectResult> {
  if (!credsReady(cfg)) {
    throw new Error("Object storage config incomplete (keys/bucket/public_base_url)");
  }
  const token = uploadToken(cfg, input.key, input.contentType);
  const form = new FormData();
  form.set("token", token);
  form.set("key", input.key);
  form.set("file", new Blob([input.body], { type: input.contentType || "application/octet-stream" }));

  const resp = await fetch(uploadHost(cfg), { method: "POST", body: form });
  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(`Qiniu upload failed: ${resp.status} ${text}`);
  }
  return { url: joinPublicUrl(cfg.publicBaseUrl, input.key), key: input.key };
}

export async function deleteQiniu(cfg: ObjectStorageRuntimeConfig, keys: string[]): Promise<void> {
  if (!keys.length || !credsReady(cfg)) return;
  await Promise.all(
    keys.map(async (key) => {
      try {
        const entry = urlSafeBase64(`${cfg.bucket}:${key}`);
        const path = `/delete/${entry}`;
        const resp = await fetch(`${rsHost(cfg)}${path}`, {
          method: "POST",
          headers: { Authorization: qboxAuth(cfg, path + "\n") },
        });
        if (!resp.ok) {
          const text = await resp.text();
          console.warn(`[objectStorage] qiniu delete ${key} failed: ${resp.status} ${text}`);
        }
      } catch (e) {
        console.warn(`[objectStorage] qiniu delete ${key} error:`, e);
      }
    }),
  );
}
