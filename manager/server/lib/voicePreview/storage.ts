import { mkdir, unlink, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { randomUUID } from "node:crypto";
import prisma from "../prisma";
import { getAudioDir } from "../../utils/audioDir";
import {
  deleteObjects,
  joinPublicUrl,
  parseObjectStorageConfig,
  uploadObject,
} from "../objectStorage";

/** 试听音频在对象存储桶内的目录前缀（与 R2/OSS 文件夹 demoaudio 一致）。 */
const DEMO_AUDIO_PREFIX = "demoaudio";
/** 历史试听音频曾使用 aiaudio 前缀，删除时兼容清理。 */
const LEGACY_DEMO_AUDIO_PREFIX = "aiaudio";

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

function keyFromPublicUrl(publicBaseUrl: string, objectUrl: string): string {
  const base = publicBaseUrl.trim().replace(/\/$/, "");
  const url = objectUrl.trim();
  if (!base || !url) return "";
  const prefixes = [`${base}/`];
  if (!base.startsWith("http://") && !base.startsWith("https://")) {
    prefixes.push(`https://${base}/`, `http://${base}/`);
  }
  for (const prefix of prefixes) {
    if (url.startsWith(prefix)) {
      return url.slice(prefix.length);
    }
  }
  return "";
}

function resolveCloudObjectKeys(args: {
  previewAudioUrl?: string | null;
  localFilename?: string | null;
  publicBaseUrl?: string | null;
}): string[] {
  const keys = new Set<string>();
  const local = (args.localFilename ?? "").trim();
  const url = (args.previewAudioUrl ?? "").trim();
  const publicBase = (args.publicBaseUrl ?? "").trim();

  if (url.startsWith("http://") || url.startsWith("https://")) {
    const fromUrl = keyFromPublicUrl(publicBase, url);
    if (fromUrl) keys.add(fromUrl);
  }

  if (local) {
    keys.add(`${DEMO_AUDIO_PREFIX}/${local}`);
    keys.add(`${LEGACY_DEMO_AUDIO_PREFIX}/${local}`);
  }

  return [...keys];
}

async function deleteLocalCopy(localFilename: string): Promise<boolean> {
  const name = localFilename.trim();
  if (!name || /[/\\]/.test(name)) return false;
  try {
    await unlink(join(audioDir(), name));
    return true;
  } catch (e) {
    const code = (e as NodeJS.ErrnoException).code;
    if (code === "ENOENT") return false;
    throw e;
  }
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
  const row = await getActiveObjectStorage();
  if (!row) {
    throw new Error("media.assistant_tts.storage=cloud but no active object storage config");
  }
  const cfg = parseObjectStorageConfig(row);
  const normExt = ext.startsWith(".") ? ext : `.${ext}`;
  const localFilename = await writeLocalCopy(data, normExt, preferredBaseName);
  const key = `${DEMO_AUDIO_PREFIX}/${localFilename}`;

  if (cfg.provider === "local") {
    const url = joinPublicUrl(cfg.endpoint || cfg.publicBaseUrl, key);
    return {
      url: url || `/api/v1/audio/${localFilename}`,
      localFilename,
    };
  }

  const { url } = await uploadObject(cfg, {
    key,
    body: data,
    contentType: contentType || "application/octet-stream",
  });
  return { url, localFilename };
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

export type DeleteAssistantPreviewAudioArgs = {
  previewAudioUrl?: string | null;
  previewLocalFilename?: string | null;
};

/** 删除试听音频：本地 AUDIO_DIR 备份与云端对象（若已配置）同步清理。 */
export async function deleteAssistantPreviewAudio(
  args: DeleteAssistantPreviewAudioArgs,
): Promise<{ localDeleted: boolean; cloudKeys: string[] }> {
  const localFilename = (args.previewLocalFilename ?? "").trim();
  const previewAudioUrl = (args.previewAudioUrl ?? "").trim();
  const mode = await getAssistantTtsStorageMode();

  let localDeleted = false;
  if (localFilename) {
    localDeleted = await deleteLocalCopy(localFilename);
  }

  let cloudKeys: string[] = [];
  if (mode === "cloud") {
    const row = await getActiveObjectStorage();
    if (row) {
      const cfg = parseObjectStorageConfig(row);
      cloudKeys = resolveCloudObjectKeys({
        previewAudioUrl,
        localFilename,
        publicBaseUrl: cfg.publicBaseUrl,
      });
      await deleteObjects(cfg, cloudKeys);
    }
  }

  return { localDeleted, cloudKeys };
}
