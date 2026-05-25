import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { getAudioDir } from "../../../utils/audioDir";

function previewAudioMimeType(filename: string): string {
  const ext = filename.slice(filename.lastIndexOf(".")).toLowerCase();
  switch (ext) {
    case ".wav":
      return "audio/wav";
    case ".m4a":
    case ".mp4":
      return "audio/mp4";
    case ".webm":
      return "audio/webm";
    case ".ogg":
    case ".oga":
      return "audio/ogg";
    default:
      return "audio/mpeg";
  }
}

export default defineEventHandler(async (event) => {
  const filename = getRouterParam(event, "filename")?.trim() ?? "";
  if (!filename || filename.length > 100 || /[/\\]/.test(filename)) {
    throw createError({ statusCode: 400, message: "invalid filename" });
  }

  const filePath = join(getAudioDir(), filename);

  let data: Buffer;
  try {
    data = await readFile(filePath);
  } catch {
    throw createError({ statusCode: 404, message: "not found" });
  }

  setHeader(event, "Content-Type", previewAudioMimeType(filename));
  return data;
});
