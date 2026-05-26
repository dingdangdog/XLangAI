import { mkdir, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { randomUUID } from "node:crypto";
import { S3Client, PutObjectCommand } from "@aws-sdk/client-s3";
import prisma from "../prisma";
import { getAudioDir } from "../../utils/audioDir";

const AI_AUDIO_PREFIX = "aiaudio";

function audioDir(): string {
  return getAudioDir();
}

async function getAssistantTtsStorageMode(): Promise<string> {
  const row = await prisma.sysSystemSetting.findUnique({
    where: { key: "media.assistant_tts.storage" },
  });
  const mode = (row?.value ?? "server").trim().toLowerCase();
  return mode || "server";
}

async function getActiveObjectStorage() {
  return prisma.sysObjectStorageConfig.findFirst({
    where: { status: "active" },
    orderBy: [{ sortOrder: "asc" }, { createdAt: "asc" }],
  });
}

function safeAudioFilename(preferredBaseName: string | undefined, ext: string): string {
  const normExt = ext.startsWith(".") ? ext : `.${ext}`;
  const base = (preferredBaseName ?? "")
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .replace(/[^a-zA-Z0-9._-]+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^[-_.]+|[-_.]+$/g, "")
    .slice(0, 80);
  return `${base || randomUUID()}${normExt}`;
}

async function writeLocalCopy(
  data: Buffer,
  ext: string,
  preferredBaseName?: string,
): Promise<string> {
  const dir = audioDir();
  await mkdir(dir, { recursive: true });
  const filename = safeAudioFilename(preferredBaseName, ext);
  await writeFile(join(dir, filename), data);
  return filename;
}

async function uploadCloud(
  data: Buffer,
  ext: string,
  contentType: string,
  preferredBaseName?: string,
): Promise<{ url: string; localFilename: string }> {
  const cfg = await getActiveObjectStorage();
  if (!cfg) {
    throw new Error("media.assistant_tts.storage=cloud but no active object storage config");
  }
  const provider = (cfg.provider ?? "").trim().toLowerCase();
  const normExt = ext.startsWith(".") ? ext : `.${ext}`;
  const localFilename = await writeLocalCopy(data, normExt, preferredBaseName);
  const key = `${AI_AUDIO_PREFIX}/${localFilename}`;

  if (provider === "local") {
    const base = (cfg.baseUrl ?? cfg.publicBaseUrl ?? "").replace(/\/$/, "");
    const url = base ? `${base}/${key}` : `/api/v1/audio/${localFilename}`;
    return { url, localFilename };
  }

  if (provider === "cloudflare_r2" || provider === "aliyun_oss") {
    const endpoint = (cfg.baseUrl ?? "").trim().replace(/\/$/, "");
    const accessKey = (cfg.apiKey ?? "").trim();
    const secretKey = (cfg.secretKey ?? "").trim();
    const bucket = (cfg.bucket ?? "").trim();
    const publicBase = (cfg.publicBaseUrl ?? "").trim().replace(/\/$/, "");
    if (!endpoint || !accessKey || !secretKey || !bucket || !publicBase) {
      throw new Error("Object storage config incomplete (endpoint/keys/bucket/public_base_url)");
    }
    const client = new S3Client({
      region: (cfg.region ?? "auto").trim() || "auto",
      endpoint,
      credentials: { accessKeyId: accessKey, secretAccessKey: secretKey },
      forcePathStyle: true,
    });
    await client.send(
      new PutObjectCommand({
        Bucket: bucket,
        Key: key,
        Body: data,
        ContentType: contentType || "application/octet-stream",
      }),
    );
    return { url: `${publicBase}/${key}`, localFilename };
  }

  throw new Error(`Manager preview does not support object storage provider "${provider}"`);
}

/** Saves per media.assistant_tts.storage; cloud mode also writes a local copy under AUDIO_DIR. */
export async function saveAssistantPreviewAudio(
  data: Buffer,
  ext: string,
  contentType: string,
  preferredBaseName?: string,
): Promise<{ previewAudioUrl: string; previewLocalFilename: string }> {
  const mode = await getAssistantTtsStorageMode();
  const normExt = ext.startsWith(".") ? ext : `.${ext}`;

  if (mode === "cloud") {
    const { url, localFilename } = await uploadCloud(
      data,
      normExt,
      contentType,
      preferredBaseName,
    );
    return { previewAudioUrl: url, previewLocalFilename: localFilename };
  }

  const localFilename = await writeLocalCopy(data, normExt, preferredBaseName);
  return {
    previewAudioUrl: `/api/v1/audio/${localFilename}`,
    previewLocalFilename: localFilename,
  };
}
