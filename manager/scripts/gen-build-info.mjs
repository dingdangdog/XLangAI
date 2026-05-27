import { execSync } from "node:child_process";
import { writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const rootDir = join(dirname(fileURLToPath(import.meta.url)), "..");
const outPath = join(rootDir, ".build-info.json");

function runGit(command) {
  try {
    return execSync(command, {
      cwd: rootDir,
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
  } catch {
    return "";
  }
}

const version =
  process.env.NUXT_PUBLIC_APP_VERSION ||
  runGit("git describe --tags --exact-match") ||
  runGit("git describe --tags --always") ||
  "dev";

const sha = process.env.NUXT_PUBLIC_BUILD_SHA || runGit("git rev-parse --short HEAD") || "";

writeFileSync(outPath, `${JSON.stringify({ version, sha }, null, 2)}\n`);
console.log(`[build-info] version=${version} sha=${sha || "(none)"}`);
