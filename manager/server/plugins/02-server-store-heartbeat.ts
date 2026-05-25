import { getLocalServerStoreConfig, sendServerStoreHeartbeat } from "../utils/serverStoreConfig";

declare global {
  // eslint-disable-next-line no-var
  var __xlangaiServerStoreHeartbeatTimer:
    | ReturnType<typeof setInterval>
    | undefined;
}

async function heartbeatOnce() {
  const config = await getLocalServerStoreConfig();
  if (!config.enabled) return;

  try {
    await sendServerStoreHeartbeat();
  } catch (error) {
    console.warn(
      "[server-store] heartbeat failed:",
      error instanceof Error ? error.message : error,
    );
  }
}

export default defineNitroPlugin(() => {
  if (globalThis.__xlangaiServerStoreHeartbeatTimer) return;

  const timer = setInterval(() => {
    void heartbeatOnce();
  }, 5 * 60 * 1000);
  timer.unref?.();
  globalThis.__xlangaiServerStoreHeartbeatTimer = timer;

  setTimeout(() => {
    void heartbeatOnce();
  }, 30 * 1000).unref?.();
});
