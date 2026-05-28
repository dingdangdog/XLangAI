import { DeleteObjectCommand, PutObjectCommand, S3Client } from "@aws-sdk/client-s3";
import type { ObjectStorageRuntimeConfig, UploadObjectInput, UploadObjectResult } from "../types";
import { joinPublicUrl } from "../runtime";

function s3Client(cfg: ObjectStorageRuntimeConfig): S3Client {
  const endpoint = cfg.endpoint.replace(/\/$/, "");
  return new S3Client({
    region: cfg.region || "auto",
    endpoint,
    credentials: { accessKeyId: cfg.accessKey, secretAccessKey: cfg.secretKey },
    forcePathStyle: cfg.extra.path_style !== "false",
  });
}

function credsReady(cfg: ObjectStorageRuntimeConfig): boolean {
  return Boolean(cfg.endpoint && cfg.accessKey && cfg.secretKey && cfg.bucket && cfg.publicBaseUrl);
}

export async function uploadS3Compatible(
  cfg: ObjectStorageRuntimeConfig,
  input: UploadObjectInput,
): Promise<UploadObjectResult> {
  if (!credsReady(cfg)) {
    throw new Error("Object storage config incomplete (endpoint/keys/bucket/public_base_url)");
  }
  const client = s3Client(cfg);
  await client.send(
    new PutObjectCommand({
      Bucket: cfg.bucket,
      Key: input.key,
      Body: input.body,
      ContentType: input.contentType || "application/octet-stream",
    }),
  );
  return { url: joinPublicUrl(cfg.publicBaseUrl, input.key), key: input.key };
}

export async function deleteS3Compatible(cfg: ObjectStorageRuntimeConfig, keys: string[]): Promise<void> {
  if (!keys.length || !credsReady(cfg)) return;
  const client = s3Client(cfg);
  await Promise.all(
    keys.map(async (key) => {
      try {
        await client.send(new DeleteObjectCommand({ Bucket: cfg.bucket, Key: key }));
      } catch (e) {
        console.warn(`[objectStorage] failed to delete ${key}:`, e);
      }
    }),
  );
}
