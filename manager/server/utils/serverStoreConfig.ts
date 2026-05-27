import { createError } from "h3";
import prisma from "../lib/prisma";
import { getAppVersionDisplay } from "./buildInfo";

const CONFIG_KEY = "official_server_store.config";
const DEFAULT_OFFICIAL_HOME_URL = "https://xlangai.com";
const DEFAULT_HEARTBEAT_SECONDS = 5 * 60;

export type LocalServerStoreConfig = {
  enabled: boolean;
  uploadStats: boolean;
  serverAddress: string;
  homepageUrl: string;
  name: string;
  summary: string;
  description: string;
  contactEmail: string;
  logoUrl: string;
  region: string;
  version: string;
  officialServerId: string;
  officialServerToken: string;
  heartbeatIntervalSeconds: number;
  lastSyncAt: string | null;
  lastHeartbeatAt: string | null;
  lastError: string;
};

type OfficialRegisterResponse = {
  entry: {
    id: string;
    heartbeatIntervalSeconds?: number;
  };
  token?: string;
  officialStoreUrl?: string;
};

function officialHomeUrl() {
  return String(
    useRuntimeConfig().public.officialHomeUrl || DEFAULT_OFFICIAL_HOME_URL,
  ).replace(/\/+$/, "");
}

function defaultConfig(): LocalServerStoreConfig {
  return {
    enabled: false,
    uploadStats: false,
    serverAddress: "",
    homepageUrl: "",
    name: "小浪AI 自托管服务器",
    summary: "",
    description: "",
    contactEmail: "",
    logoUrl: "",
    region: "",
    version: "",
    officialServerId: "",
    officialServerToken: "",
    heartbeatIntervalSeconds: DEFAULT_HEARTBEAT_SECONDS,
    lastSyncAt: null,
    lastHeartbeatAt: null,
    lastError: "",
  };
}

function asBool(value: unknown) {
  return value === true;
}

function asText(value: unknown, maxLength: number) {
  return typeof value === "string" ? value.trim().slice(0, maxLength) : "";
}

function normalizeUrl(value: unknown, label: string, required = false) {
  const raw = asText(value, 512);
  if (!raw) {
    if (required) throw createError({ statusCode: 400, message: `请填写${label}` });
    return "";
  }

  let parsed: URL;
  try {
    parsed = new URL(raw);
  } catch {
    throw createError({ statusCode: 400, message: `${label}格式不正确` });
  }
  if (!["http:", "https:"].includes(parsed.protocol)) {
    throw createError({ statusCode: 400, message: `${label}必须以 http:// 或 https:// 开头` });
  }
  parsed.hash = "";
  return parsed.toString().replace(/\/$/, "");
}

function parseConfig(raw: string | null | undefined): LocalServerStoreConfig {
  if (!raw) return defaultConfig();
  try {
    const parsed = JSON.parse(raw) as Partial<LocalServerStoreConfig>;
    return {
      ...defaultConfig(),
      ...parsed,
      enabled: parsed.enabled === true,
      uploadStats: parsed.uploadStats === true,
      heartbeatIntervalSeconds: Math.max(
        60,
        Math.min(Number(parsed.heartbeatIntervalSeconds) || DEFAULT_HEARTBEAT_SECONDS, 24 * 60 * 60),
      ),
      lastSyncAt: parsed.lastSyncAt ?? null,
      lastHeartbeatAt: parsed.lastHeartbeatAt ?? null,
    };
  } catch {
    return defaultConfig();
  }
}

function normalizeConfig(input: Partial<LocalServerStoreConfig>, previous: LocalServerStoreConfig) {
  const enabled = asBool(input.enabled);
  const uploadStats = asBool(input.uploadStats);

  return {
    ...previous,
    enabled,
    uploadStats,
    serverAddress: normalizeUrl(input.serverAddress, "本服务器地址", enabled),
    homepageUrl: normalizeUrl(input.homepageUrl, "官网地址"),
    name: asText(input.name, 120) || previous.name || "小浪AI 自托管服务器",
    summary: asText(input.summary, 500),
    description: asText(input.description, 5000),
    contactEmail: asText(input.contactEmail, 255),
    logoUrl: normalizeUrl(input.logoUrl, "Logo 地址"),
    region: asText(input.region, 100),
    version: getAppVersionDisplay(),
    officialServerId: asText(input.officialServerId, 100) || previous.officialServerId,
    officialServerToken: asText(input.officialServerToken, 200) || previous.officialServerToken,
    heartbeatIntervalSeconds: Math.max(
      60,
      Math.min(Number(input.heartbeatIntervalSeconds) || previous.heartbeatIntervalSeconds, 24 * 60 * 60),
    ),
  };
}

export function getOfficialServerStoreUrl(id?: string) {
  return id ? `${officialHomeUrl()}/servers/${id}` : `${officialHomeUrl()}/servers`;
}

export async function getLocalServerStoreConfig() {
  const row = await prisma.sysSystemSetting.findUnique({ where: { key: CONFIG_KEY } });
  return parseConfig(row?.value);
}

export async function saveLocalServerStoreConfig(input: Partial<LocalServerStoreConfig>) {
  const previous = await getLocalServerStoreConfig();
  const next = normalizeConfig(input, previous);
  await prisma.sysSystemSetting.upsert({
    where: { key: CONFIG_KEY },
    create: {
      key: CONFIG_KEY,
      value: JSON.stringify(next),
      valueType: "json",
      description: "本服务器同步到官网服务器商店的配置",
    },
    update: {
      value: JSON.stringify(next),
      valueType: "json",
      description: "本服务器同步到官网服务器商店的配置",
    },
  });
  return next;
}

async function patchLocalServerStoreConfig(patch: Partial<LocalServerStoreConfig>) {
  const previous = await getLocalServerStoreConfig();
  const next = { ...previous, ...patch };
  await prisma.sysSystemSetting.upsert({
    where: { key: CONFIG_KEY },
    create: {
      key: CONFIG_KEY,
      value: JSON.stringify(next),
      valueType: "json",
      description: "本服务器同步到官网服务器商店的配置",
    },
    update: {
      value: JSON.stringify(next),
      valueType: "json",
      description: "本服务器同步到官网服务器商店的配置",
    },
  });
  return next;
}

export async function collectServerStoreStats() {
  const [
    users,
    activeUsers,
    languages,
    voiceRoles,
    conversations,
    messages,
    llmConfigs,
    sttConfigs,
    ttsConfigs,
    translateConfigs,
  ] = await Promise.all([
    prisma.user.count({ where: { deletedAt: null } }),
    prisma.user.count({ where: { deletedAt: null, status: "active" } }),
    prisma.language.count({ where: { status: "active" } }),
    prisma.voiceRole.count({ where: { status: "active" } }),
    prisma.conversation.count({ where: { deletedAt: null } }),
    prisma.message.count({ where: { deletedAt: null } }),
    prisma.sysLlmServiceConfig.count({ where: { status: "active" } }),
    prisma.sysSttServiceConfig.count({ where: { status: "active" } }),
    prisma.ttsServiceConfig.count({ where: { status: "active" } }),
    prisma.sysTranslateServiceConfig.count({ where: { status: "active" } }),
  ]);

  return {
    users,
    activeUsers,
    languages,
    voiceRoles,
    conversations,
    messages,
    llmConfigs,
    sttConfigs,
    ttsConfigs,
    translateConfigs,
  };
}

function publicConfig(config: LocalServerStoreConfig, stats?: Awaited<ReturnType<typeof collectServerStoreStats>>) {
  return {
    ...config,
    officialHomeUrl: officialHomeUrl(),
    officialStoreUrl: getOfficialServerStoreUrl(config.officialServerId || undefined),
    stats: config.uploadStats ? stats ?? null : null,
  };
}

function buildOfficialPayload(
  config: LocalServerStoreConfig,
  stats?: Awaited<ReturnType<typeof collectServerStoreStats>>,
) {
  return {
    officialServerId: config.officialServerId || undefined,
    token: config.officialServerToken || undefined,
    serverAddress: config.serverAddress,
    homepageUrl: config.homepageUrl || undefined,
    name: config.name,
    summary: config.summary || undefined,
    description: config.description || undefined,
    contactEmail: config.contactEmail || undefined,
    logoUrl: config.logoUrl || undefined,
    region: config.region || undefined,
    version: getAppVersionDisplay(),
    isPublic: config.enabled,
    uploadStats: config.uploadStats,
    stats: config.uploadStats ? stats : undefined,
    heartbeatIntervalSeconds: config.heartbeatIntervalSeconds,
  };
}

export async function getServerStoreAdminState() {
  const config = await getLocalServerStoreConfig();
  const stats = await collectServerStoreStats();
  return {
    ...publicConfig(config, stats),
    systemVersion: getAppVersionDisplay(),
  };
}

export async function publishServerStoreConfig(input?: Partial<LocalServerStoreConfig>) {
  const saved = input ? await saveLocalServerStoreConfig(input) : await getLocalServerStoreConfig();
  if (!saved.enabled) {
    throw createError({ statusCode: 400, message: "请先开启“开放到官网服务器商店”" });
  }

  const stats = saved.uploadStats ? await collectServerStoreStats() : undefined;
  const response = await $fetch<OfficialRegisterResponse>(`${officialHomeUrl()}/api/server-store/register`, {
    method: "POST",
    body: buildOfficialPayload(saved, stats),
  });

  const now = new Date().toISOString();
  const next = await patchLocalServerStoreConfig({
    officialServerId: response.entry.id,
    officialServerToken: response.token || saved.officialServerToken,
    heartbeatIntervalSeconds: response.entry.heartbeatIntervalSeconds || saved.heartbeatIntervalSeconds,
    lastSyncAt: now,
    lastHeartbeatAt: now,
    lastError: "",
  });

  return publicConfig(next, stats);
}

export async function sendServerStoreHeartbeat() {
  const config = await getLocalServerStoreConfig();
  if (!config.enabled) return publicConfig(config);
  if (!config.officialServerId || !config.officialServerToken) {
    return publishServerStoreConfig();
  }

  const stats = config.uploadStats ? await collectServerStoreStats() : undefined;
  try {
    const response = await $fetch<{ nextHeartbeatSeconds?: number }>(
      `${officialHomeUrl()}/api/server-store/heartbeat`,
      {
        method: "POST",
        body: {
          officialServerId: config.officialServerId,
          token: config.officialServerToken,
          healthStatus: "online",
          stats: config.uploadStats ? stats : undefined,
        },
      },
    );
    const now = new Date().toISOString();
    const next = await patchLocalServerStoreConfig({
      heartbeatIntervalSeconds: response.nextHeartbeatSeconds || config.heartbeatIntervalSeconds,
      lastHeartbeatAt: now,
      lastError: "",
    });
    return publicConfig(next, stats);
  } catch (error) {
    const message = error instanceof Error ? error.message : "心跳发送失败";
    const next = await patchLocalServerStoreConfig({ lastError: message });
    throw createError({ statusCode: 502, message });
  }
}
