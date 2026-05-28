<script setup lang="ts">
import { ArrowTopRightOnSquareIcon } from "@heroicons/vue/24/outline";

const { t } = useI18n();
type ServerStoreConfig = {
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
  officialHomeUrl: string;
  officialStoreUrl: string;
  stats: Record<string, number> | null;
  systemVersion: string;
};

const toast = useToast();
const loading = ref(false);
const saving = ref(false);
const publishing = ref(false);
const heartbeating = ref(false);

const form = reactive<ServerStoreConfig>({
  enabled: false,
  uploadStats: false,
  serverAddress: "",
  homepageUrl: "",
  name: "",
  summary: "",
  description: "",
  contactEmail: "",
  logoUrl: "",
  region: "",
  version: "",
  officialServerId: "",
  officialServerToken: "",
  heartbeatIntervalSeconds: 300,
  lastSyncAt: null,
  lastHeartbeatAt: null,
  lastError: "",
  officialHomeUrl: "https://xlangai.com",
  officialStoreUrl: "https://xlangai.com/servers",
  stats: null,
  systemVersion: "",
});

const statLabelKeys: Record<string, string> = {
  users: "pages.serverStore.statUsers",
  activeUsers: "pages.serverStore.statActiveUsers",
  languages: "pages.serverStore.statLanguages",
  voiceRoles: "pages.serverStore.statVoiceRoles",
  conversations: "pages.serverStore.statConversations",
  messages: "pages.serverStore.statMessages",
  llmConfigs: "pages.serverStore.statLlmConfigs",
  sttConfigs: "pages.serverStore.statSttConfigs",
  ttsConfigs: "pages.serverStore.statTtsConfigs",
  translateConfigs: "pages.serverStore.statTranslateConfigs",
};

const stats = computed(() => {
  if (!form.stats) return [];
  return Object.entries(form.stats).map(([key, value]) => ({
    label: statLabelKeys[key] ? t(statLabelKeys[key]) : key,
    value,
  }));
});

const isRegistered = computed(() => !!form.officialServerId);

function applyState(state: ServerStoreConfig) {
  Object.assign(form, state);
}

function payload() {
  return {
    enabled: form.enabled,
    uploadStats: form.uploadStats,
    serverAddress: form.serverAddress,
    homepageUrl: form.homepageUrl,
    name: form.name,
    summary: form.summary,
    description: form.description,
    contactEmail: form.contactEmail,
    logoUrl: form.logoUrl,
    region: form.region,
    officialServerId: form.officialServerId,
    officialServerToken: form.officialServerToken,
  };
}

async function load() {
  loading.value = true;
  try {
    const state = await $fetch<ServerStoreConfig>("/api/admin/server-store/config");
    applyState(state);
  } catch (error) {
    toast.error(t("toast.loadServerStoreFailed"));
    console.error(error);
  } finally {
    loading.value = false;
  }
}

async function save() {
  saving.value = true;
  try {
    const state = await $fetch<ServerStoreConfig>("/api/admin/server-store/config", {
      method: "PUT",
      body: payload(),
    });
    applyState(state);
    toast.success(t("toast.serverStoreSaved"));
  } catch (error) {
    toast.error(t("toast.saveFailed"));
    console.error(error);
  } finally {
    saving.value = false;
  }
}

async function publish() {
  publishing.value = true;
  try {
    const state = await $fetch<ServerStoreConfig>("/api/admin/server-store/publish", {
      method: "POST",
      body: payload(),
    });
    applyState(state);
    toast.success(t("toast.uploadedToOfficial"));
  } catch (error) {
    toast.error(t("toast.uploadFailed"));
    console.error(error);
  } finally {
    publishing.value = false;
  }
}

async function heartbeat() {
  heartbeating.value = true;
  try {
    const state = await $fetch<ServerStoreConfig>("/api/admin/server-store/heartbeat", {
      method: "POST",
    });
    applyState(state);
    toast.success(t("toast.heartbeatSent"));
  } catch (error) {
    toast.error(t("toast.heartbeatFailed"));
    console.error(error);
  } finally {
    heartbeating.value = false;
  }
}

onMounted(() => {
  void load();
});
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col overflow-hidden">
    <AdminSkeleton v-if="loading" :rows="8" />

    <div
      v-else
      class="grid min-h-0 flex-1 gap-4 overflow-y-auto xl:grid-cols-[minmax(0,1fr)_380px] xl:overflow-hidden"
    >
      <AdminPanel :fill="false" class="flex min-h-0 min-w-0 flex-col xl:overflow-hidden">
        <div class="shrink-0 border-b border-border px-4 py-4 md:px-5">
          <h2 class="text-base font-semibold text-foreground">
            {{ $t("pages.serverStore.configTitle") }}
          </h2>
          <p class="mt-1 text-sm text-muted">{{ $t("pages.serverStore.description") }}</p>
        </div>

        <div class="min-h-0 flex-1 space-y-8 p-4 md:p-5 xl:overflow-y-auto xl:pr-3">
          <AdminAlert :title="$t('pages.serverStore.syncAlertTitle')">
            {{ $t("pages.serverStore.syncAlert") }}
          </AdminAlert>

          <section class="space-y-4">
            <div>
              <h3 class="text-sm font-semibold text-foreground">
                {{ $t("pages.serverStore.sections.sync") }}
              </h3>
              <p class="mt-1 text-xs text-muted">{{ $t("pages.serverStore.sections.syncDesc") }}</p>
            </div>
            <div class="grid gap-3 sm:grid-cols-2">
              <label
                class="flex cursor-pointer items-center justify-between gap-4 rounded-2xl border border-border/70 px-4 py-3.5 transition hover:border-primary-300/50"
              >
                <div class="min-w-0">
                  <p class="text-sm font-medium text-foreground">
                    {{ $t("pages.serverStore.enableOnOfficial") }}
                  </p>
                  <p class="mt-0.5 text-xs leading-relaxed text-muted">
                    {{ $t("pages.serverStore.enableHint") }}
                  </p>
                </div>
                <input
                  v-model="form.enabled"
                  type="checkbox"
                  class="h-5 w-5 shrink-0 rounded border-border text-primary-600"
                />
              </label>
              <label
                class="flex cursor-pointer items-center justify-between gap-4 rounded-2xl border border-border/70 px-4 py-3.5 transition hover:border-primary-300/50"
              >
                <div class="min-w-0">
                  <p class="text-sm font-medium text-foreground">
                    {{ $t("pages.serverStore.uploadStats") }}
                  </p>
                  <p class="mt-0.5 text-xs leading-relaxed text-muted">
                    {{ $t("pages.serverStore.uploadStatsHint") }}
                  </p>
                </div>
                <input
                  v-model="form.uploadStats"
                  type="checkbox"
                  class="h-5 w-5 shrink-0 rounded border-border text-primary-600"
                />
              </label>
            </div>
          </section>

          <section class="space-y-4">
            <div>
              <h3 class="text-sm font-semibold text-foreground">
                {{ $t("pages.serverStore.sections.identity") }}
              </h3>
              <p class="mt-1 text-xs text-muted">{{ $t("pages.serverStore.sections.identityDesc") }}</p>
            </div>
            <div class="grid gap-x-4 gap-y-0 md:grid-cols-2">
              <AdminFormField :label="$t('pages.serverStore.serverName')" required>
                <AdminInput
                  v-model="form.name"
                  :placeholder="$t('pages.serverStore.serverNamePlaceholder')"
                />
              </AdminFormField>
              <AdminFormField :label="$t('pages.serverStore.region')">
                <AdminInput v-model="form.region" :placeholder="$t('pages.serverStore.regionPlaceholder')" />
              </AdminFormField>
              <AdminFormField :label="$t('pages.serverStore.serverUrl')" required>
                <AdminInput
                  v-model="form.serverAddress"
                  :placeholder="$t('pages.serverStore.serverUrlPlaceholder')"
                />
              </AdminFormField>
              <AdminFormField :label="$t('pages.serverStore.homepageUrl')">
                <AdminInput v-model="form.homepageUrl" :placeholder="$t('pages.serverStore.homepagePlaceholder')" />
              </AdminFormField>
              <AdminFormField :label="$t('pages.serverStore.logoUrl')">
                <AdminInput v-model="form.logoUrl" placeholder="https://example.com/logo.png" />
              </AdminFormField>
              <AdminFormField :label="$t('pages.serverStore.version')">
                <p class="rounded-lg border border-border bg-surface-muted px-3 py-2 text-sm text-foreground">
                  {{ form.systemVersion || $t("common.emDash") }}
                </p>
                <p class="mt-1 text-xs text-muted">{{ $t("pages.serverStore.versionAutoHint") }}</p>
              </AdminFormField>
            </div>
          </section>

          <section class="space-y-4">
            <div>
              <h3 class="text-sm font-semibold text-foreground">
                {{ $t("pages.serverStore.sections.presentation") }}
              </h3>
              <p class="mt-1 text-xs text-muted">
                {{ $t("pages.serverStore.sections.presentationDesc") }}
              </p>
            </div>
            <div class="space-y-0">
              <AdminFormField :label="$t('pages.serverStore.tagline')">
                <AdminInput
                  v-model="form.summary"
                  type="textarea"
                  :rows="2"
                  :placeholder="$t('pages.serverStore.taglinePlaceholder')"
                />
              </AdminFormField>
              <AdminFormField class="!mb-0" :label="$t('pages.serverStore.descriptionField')">
                <AdminInput
                  v-model="form.description"
                  type="textarea"
                  :rows="5"
                  :placeholder="$t('pages.serverStore.descriptionPlaceholder')"
                />
              </AdminFormField>
            </div>
          </section>

          <section class="space-y-4">
            <div>
              <h3 class="text-sm font-semibold text-foreground">
                {{ $t("pages.serverStore.sections.contact") }}
              </h3>
              <p class="mt-1 text-xs text-muted">{{ $t("pages.serverStore.sections.contactDesc") }}</p>
            </div>
            <AdminFormField class="!mb-0" :label="$t('pages.serverStore.contactEmail')">
              <AdminInput v-model="form.contactEmail" type="email" placeholder="admin@example.com" />
            </AdminFormField>
          </section>
        </div>

        <div class="shrink-0 border-t border-border px-4 py-4 md:px-5">
          <AdminButton variant="secondary" :loading="saving" @click="save">
            {{ $t("pages.serverStore.saveConfig") }}
          </AdminButton>
        </div>
      </AdminPanel>

      <div class="flex min-h-0 flex-col gap-4 xl:overflow-y-auto xl:pr-1">
        <AdminPanel :fill="false">
          <div
            class="flex flex-wrap items-start justify-between gap-3 border-b border-border px-4 py-4 md:px-5"
          >
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="text-base font-semibold text-foreground">
                  {{ $t("pages.serverStore.officialStatus") }}
                </h2>
                <AdminBadge :variant="form.enabled ? 'success' : 'muted'">
                  {{
                    form.enabled
                      ? $t("pages.serverStore.listedOnOfficial")
                      : $t("pages.serverStore.notListed")
                  }}
                </AdminBadge>
                <AdminBadge v-if="isRegistered" variant="default">
                  {{ $t("pages.serverStore.registered") }}
                </AdminBadge>
              </div>
              <p class="mt-1 text-sm text-muted">{{ form.officialHomeUrl }}</p>
            </div>
            <a
              :href="form.officialStoreUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex shrink-0 items-center gap-1.5 rounded-lg border border-border px-3 py-2 text-sm font-medium text-foreground transition hover:bg-surface-muted"
            >
              {{ $t("common.openOfficialStore") }}
              <ArrowTopRightOnSquareIcon class="h-4 w-4 text-muted" />
            </a>
          </div>

          <div class="space-y-4 p-4 md:p-5">
            <dl class="grid gap-3 text-sm sm:grid-cols-2">
              <div class="rounded-xl bg-surface-muted/70 px-3 py-2.5">
                <dt class="text-xs text-muted">{{ $t("pages.serverStore.storeLink") }}</dt>
                <dd class="mt-1 break-all">
                  <a
                    :href="form.officialStoreUrl"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="font-medium text-primary-600 hover:text-primary-700"
                  >
                    {{ form.officialStoreUrl }}
                  </a>
                </dd>
              </div>
              <div class="rounded-xl bg-surface-muted/70 px-3 py-2.5">
                <dt class="text-xs text-muted">{{ $t("pages.serverStore.officialServerId") }}</dt>
                <dd class="mt-1 break-all font-medium text-foreground">
                  {{ form.officialServerId || $t("common.notYet") }}
                </dd>
              </div>
              <div class="rounded-xl bg-surface-muted/70 px-3 py-2.5">
                <dt class="text-xs text-muted">{{ $t("pages.serverStore.lastSync") }}</dt>
                <dd class="mt-1 font-medium text-foreground">
                  {{ form.lastSyncAt ? formatDateTime(form.lastSyncAt) : $t("common.none") }}
                </dd>
              </div>
              <div class="rounded-xl bg-surface-muted/70 px-3 py-2.5">
                <dt class="text-xs text-muted">{{ $t("pages.serverStore.lastHeartbeat") }}</dt>
                <dd class="mt-1 font-medium text-foreground">
                  {{ form.lastHeartbeatAt ? formatDateTime(form.lastHeartbeatAt) : $t("common.none") }}
                </dd>
              </div>
              <div class="rounded-xl bg-surface-muted/70 px-3 py-2.5">
                <dt class="text-xs text-muted">{{ $t("pages.serverStore.heartbeatInterval") }}</dt>
                <dd class="mt-1 font-medium text-foreground">
                  {{
                    form.enabled
                      ? $t("pages.serverStore.heartbeatIntervalValue", {
                          seconds: form.heartbeatIntervalSeconds,
                        })
                      : $t("common.emDash")
                  }}
                </dd>
                <p class="mt-1 text-xs text-muted">{{ $t("pages.serverStore.heartbeatIntervalHint") }}</p>
              </div>
            </dl>

            <AdminAlert
              v-if="form.lastError"
              class="!mb-0"
              variant="warning"
              :title="$t('pages.serverStore.lastError')"
            >
              {{ form.lastError }}
            </AdminAlert>

            <div class="border-t border-border pt-4">
              <p class="text-xs font-medium text-muted">{{ $t("pages.serverStore.syncActions") }}</p>
              <div class="mt-3 flex flex-col gap-2 sm:flex-row">
                <AdminButton class="flex-1" variant="primary" :loading="publishing" @click="publish">
                  {{ $t("pages.serverStore.saveAndUpload") }}
                </AdminButton>
                <AdminButton class="flex-1" variant="secondary" :loading="heartbeating" @click="heartbeat">
                  {{ $t("pages.serverStore.sendHeartbeat") }}
                </AdminButton>
              </div>
            </div>
          </div>
        </AdminPanel>

        <AdminPanel :fill="false">
          <div class="p-4 md:p-5">
            <h2 class="text-base font-semibold text-foreground">{{ $t("pages.serverStore.statsToUpload") }}</h2>
            <p class="mt-2 text-sm text-muted">
              {{ $t("pages.serverStore.statsHint") }}
            </p>
            <div v-if="stats.length" class="mt-4 grid grid-cols-2 gap-3">
              <div v-for="stat in stats" :key="stat.label" class="rounded-lg bg-surface-muted p-3">
                <div class="text-xs text-muted">{{ stat.label }}</div>
                <div class="mt-1 text-lg font-semibold text-foreground">{{ stat.value }}</div>
              </div>
            </div>
            <p v-else class="mt-4 text-sm text-muted">{{ $t("pages.serverStore.statsDisabled") }}</p>
          </div>
        </AdminPanel>
      </div>
    </div>
  </div>
</template>
