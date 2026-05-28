export type ObjectStorageProvider = "local" | "cloudflare_r2" | "qiniu" | "aliyun_oss";

export type ObjectStorageRuntimeConfig = {
  id: string;
  provider: ObjectStorageProvider;
  endpoint: string;
  publicBaseUrl: string;
  accessKey: string;
  secretKey: string;
  bucket: string;
  region: string;
  extra: Record<string, string>;
};

export type UploadObjectInput = {
  key: string;
  body: Buffer;
  contentType: string;
};

export type UploadObjectResult = {
  url: string;
  key: string;
};
