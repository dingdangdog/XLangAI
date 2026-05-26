import { spawn } from "node:child_process";

/** 与 Go `tts_loudness.go` 默认一致。 */
export const TTS_LOUDNESS_DEFAULT_I = -14.0;
export const TTS_LOUDNESS_DEFAULT_TP = -1.0;
export const TTS_LOUDNESS_DEFAULT_LRA = 8.0;

export class FFmpegNotFoundError extends Error {
  constructor() {
    super("ffmpeg not found: install ffmpeg or set XLANGAI_FFMPEG_PATH");
    this.name = "FFmpegNotFoundError";
  }
}

function ffmpegBin(): string {
  return (process.env.XLANGAI_FFMPEG_PATH ?? "ffmpeg").trim() || "ffmpeg";
}

function loudnessEnabled(): boolean {
  const raw = (process.env.XLANGAI_TTS_LOUDNESS_NORM ?? "1").trim().toLowerCase();
  return raw !== "0" && raw !== "false" && raw !== "off";
}

function targetLufs(): number {
  const raw = (process.env.XLANGAI_TTS_TARGET_LUFS ?? "").trim();
  if (!raw) return TTS_LOUDNESS_DEFAULT_I;
  const n = Number(raw);
  return Number.isFinite(n) ? n : TTS_LOUDNESS_DEFAULT_I;
}

function outputArgs(mimeType: string): { format: string; codecArgs: string[]; outMime: string } {
  const m = (mimeType ?? "").trim().toLowerCase();
  if (m.includes("wav")) {
    return { format: "wav", codecArgs: ["-acodec", "pcm_s16le"], outMime: "audio/wav" };
  }
  if (m.includes("ogg")) {
    return { format: "ogg", codecArgs: ["-c:a", "libvorbis"], outMime: "audio/ogg" };
  }
  return { format: "mp3", codecArgs: ["-c:a", "libmp3lame", "-q:a", "2"], outMime: "audio/mpeg" };
}

async function runFfmpeg(args: string[], input: Buffer): Promise<Buffer> {
  return new Promise((resolve, reject) => {
    const child = spawn(ffmpegBin(), args, { stdio: ["pipe", "pipe", "pipe"] });
    const stdout: Buffer[] = [];
    const stderr: Buffer[] = [];
    child.stdout.on("data", (chunk: Buffer) => stdout.push(chunk));
    child.stderr.on("data", (chunk: Buffer) => stderr.push(chunk));
    child.on("error", (err: NodeJS.ErrnoException) => {
      if (err.code === "ENOENT") {
        reject(new FFmpegNotFoundError());
        return;
      }
      reject(err);
    });
    child.on("close", (code) => {
      const out = Buffer.concat(stdout);
      if (code === 0 && out.length > 0) {
        resolve(out);
        return;
      }
      const detail = Buffer.concat(stderr).toString("utf8").trim();
      reject(new Error(`ffmpeg exit ${code ?? "?"}${detail ? `: ${detail}` : ""}`));
    });
    child.stdin.end(input);
  });
}

/** 与 Go `NormalizeTTSLoudness` 相同滤镜链。 */
export async function normalizeTtsLoudness(
  audio: Buffer,
  mimeType: string,
): Promise<{ data: Buffer; mimeType: string }> {
  if (!audio.length) {
    throw new Error("tts loudness: empty audio");
  }
  const i = targetLufs();
  const filter = `highpass=f=80,volume=4dB,loudnorm=I=${i}:TP=${TTS_LOUDNESS_DEFAULT_TP}:LRA=${TTS_LOUDNESS_DEFAULT_LRA}:print_format=none,alimiter=limit=0.97`;
  const { format, codecArgs, outMime } = outputArgs(mimeType);
  const args = [
    "-hide_banner",
    "-loglevel",
    "error",
    "-i",
    "pipe:0",
    "-af",
    filter,
    ...codecArgs,
    "-f",
    format,
    "pipe:1",
  ];
  const data = await runFfmpeg(args, audio);
  return { data, mimeType: outMime };
}

/** 与 Go `BoostTTSVolumeFallback` 相同。 */
export async function boostTtsVolumeFallback(
  audio: Buffer,
  mimeType: string,
): Promise<{ data: Buffer; mimeType: string }> {
  if (!audio.length) {
    throw new Error("tts boost: empty audio");
  }
  const filter = "highpass=f=80,volume=12dB,alimiter=limit=0.97";
  const { format, codecArgs, outMime } = outputArgs(mimeType);
  const args = [
    "-hide_banner",
    "-loglevel",
    "error",
    "-i",
    "pipe:0",
    "-af",
    filter,
    ...codecArgs,
    "-f",
    format,
    "pipe:1",
  ];
  const data = await runFfmpeg(args, audio);
  return { data, mimeType: outMime };
}

/**
 * 试听/助手 TTS 统一响度：优先 loudnorm，失败则 fallback 增益，与 Go 对话 TTS 一致。
 * ffmpeg 不可用时返回原始音频（并打日志）。
 */
export async function applyAssistantTtsLoudness(
  audio: Buffer,
  mimeType: string,
): Promise<{ data: Buffer; mimeType: string; ext: string }> {
  if (!loudnessEnabled()) {
    return { data: audio, mimeType, ext: extFromMime(mimeType) };
  }
  try {
    const norm = await normalizeTtsLoudness(audio, mimeType);
    return { ...norm, ext: extFromMime(norm.mimeType) };
  } catch (e) {
    if (e instanceof FFmpegNotFoundError) {
      console.warn(
        "[voicePreview] WARNING: ffmpeg not found — preview audio is not loudness-normalized. Install ffmpeg or set XLANGAI_FFMPEG_PATH",
      );
      return { data: audio, mimeType, ext: extFromMime(mimeType) };
    }
    console.warn("[voicePreview] loudnorm failed, trying boost fallback:", e);
    try {
      const boosted = await boostTtsVolumeFallback(audio, mimeType);
      return { ...boosted, ext: extFromMime(boosted.mimeType) };
    } catch (fallbackErr) {
      if (fallbackErr instanceof FFmpegNotFoundError) {
        console.warn("[voicePreview] WARNING: ffmpeg not found — using raw TTS audio");
      } else {
        console.warn("[voicePreview] boost fallback failed, using raw TTS audio:", fallbackErr);
      }
      return { data: audio, mimeType, ext: extFromMime(mimeType) };
    }
  }
}

function extFromMime(mimeType: string): string {
  const m = (mimeType ?? "").trim().toLowerCase();
  if (m.includes("wav")) return ".wav";
  if (m.includes("ogg")) return ".ogg";
  return ".mp3";
}
