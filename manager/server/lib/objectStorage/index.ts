import type { ObjectStorageRuntimeConfig, UploadObjectInput, UploadObjectResult } from "./types";
import { deleteS3Compatible, uploadS3Compatible } from "./providers/s3Compatible";
import { deleteQiniu, uploadQiniu } from "./providers/qiniu";
import { joinPublicUrl, parseObjectStorageConfig } from "./runtime";

export type { ObjectStorageProvider, ObjectStorageRuntimeConfig, UploadObjectInput, UploadObjectResult } from "./types";
export { joinPublicUrl, parseObjectStorageConfig };

export async function uploadObject(
  cfg: ObjectStorageRuntimeConfig,
  input: UploadObjectInput,
): Promise<UploadObjectResult> {
  switch (cfg.provider) {
    case "cloudflare_r2":
    case "aliyun_oss":
      return uploadS3Compatible(cfg, input);
    case "qiniu":
      return uploadQiniu(cfg, input);
    case "local": {
      const url = joinPublicUrl(cfg.endpoint || cfg.publicBaseUrl, input.key);
      return { url: url || `/api/v1/audio/${input.key.split("/").pop()}`, key: input.key };
    }
    default:
      throw new Error(`Unsupported object storage provider "${cfg.provider}"`);
  }
}

export async function deleteObjects(cfg: ObjectStorageRuntimeConfig, keys: string[]): Promise<void> {
  if (!keys.length) return;
  switch (cfg.provider) {
    case "cloudflare_r2":
    case "aliyun_oss":
      await deleteS3Compatible(cfg, keys);
      return;
    case "qiniu":
      await deleteQiniu(cfg, keys);
      return;
    case "local":
      return;
    default:
      console.warn(`[objectStorage] delete skipped for provider ${cfg.provider}`);
  }
}
