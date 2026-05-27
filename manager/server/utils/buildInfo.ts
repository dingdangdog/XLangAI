function readPublicBuildInfo() {
  const config = useRuntimeConfig();
  return {
    version: String(config.public.appVersion || "dev").trim() || "dev",
    sha: String(config.public.buildSha || "").trim(),
  };
}

/** Semver tag from CI or git, e.g. v1.2.3 or dev */
export function getAppVersion() {
  return readPublicBuildInfo().version;
}

/** Version without leading v, for server store display */
export function getAppVersionDisplay() {
  const version = getAppVersion();
  return version.startsWith("v") ? version.slice(1) : version;
}

export function getBuildSha() {
  return readPublicBuildInfo().sha;
}

export function getBuildInfo() {
  const { version, sha } = readPublicBuildInfo();
  return {
    version,
    displayVersion: version.startsWith("v") ? version.slice(1) : version,
    sha,
  };
}
