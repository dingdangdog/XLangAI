const DEFAULT_AUDIO_DIR = "/app/storage/audio";

/** 与 Go API 共用 AUDIO_DIR，manager 写/读本地试听音频时使用。 */
export function getAudioDir(): string {
  return (process.env.AUDIO_DIR ?? DEFAULT_AUDIO_DIR).trim() || DEFAULT_AUDIO_DIR;
}
