<script setup lang="ts">
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
    version: form.version,
    officialServerId: form.officialServerId,
    officialServerToken: form.officialServerToken,
    heartbeatIntervalSeconds: Number(form.heartbeatIntervalSeconds) || 300,
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
  <div class="flex min-h-0 flex-1 flex-col gap-4">
    <div class="flex justify-end">
      <a
        :href="form.officialStoreUrl"
        target="_blank"
        rel="noopener noreferrer"
        class="inline-flex items-center rounded-lg border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-surface-muted"
      >
        {{ $t("common.openOfficialStore") }}
      </a>
    </div>

    <AdminSkeleton v-if="loading" :rows="8" />

    <div v-else class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
      <AdminPanel>
        <div class="p-4 md:p-5">
        <AdminAlert :title="$t('pages.serverStore.syncAlertTitle')">
          {{ $t("pages.serverStore.syncAlert") }}
        </AdminAlert>

        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField :label="$t('pages.serverStore.enableOnOfficial')">
            <AdminCheckbox v-model="form.enabled" :label="$t('pages.serverStore.allowOfficialDisplay')" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.serverStore.uploadStats')">
            <AdminCheckbox v-model="form.uploadStats" :label="$t('pages.serverStore.uploadAggregatedStats')" />
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
          <AdminFormField :label="$t('pages.serverStore.serverName')" required>
            <AdminInput v-model="form.name" :placeholder="$t('pages.serverStore.serverNamePlaceholder')" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.serverStore.region')">
            <AdminInput v-model="form.region" :placeholder="$t('pages.serverStore.regionPlaceholder')" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.serverStore.contactEmail')">
            <AdminInput v-model="form.contactEmail" type="email" placeholder="admin@example.com" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.serverStore.version')">
            <AdminInput v-model="form.version" :placeholder="$t('pages.serverStore.versionPlaceholder')" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.serverStore.logoUrl')">
            <AdminInput v-model="form.logoUrl" placeholder="https://example.com/logo.png" />
          </AdminFormField>
          <AdminFormField :label="$t('pages.serverStore.heartbeatInterval')">
            <AdminInput v-model="form.heartbeatIntervalSeconds" type="number" />
          </AdminFormField>
          <AdminFormField class="md:col-span-2" :label="$t('pages.serverStore.tagline')">
            <AdminInput
              v-model="form.summary"
              type="textarea"
              :rows="2"
              :placeholder="$t('pages.serverStore.taglinePlaceholder')"
            />
          </AdminFormField>
          <AdminFormField class="md:col-span-2" :label="$t('pages.serverStore.descriptionField')">
            <AdminInput
              v-model="form.description"
              type="textarea"
              :rows="5"
              :placeholder="$t('pages.serverStore.descriptionPlaceholder')"
            />
          </AdminFormField>
        </div>

        <div class="mt-6 flex flex-wrap gap-3">
          <AdminButton variant="secondary" :loading="saving" @click="save">{{ $t("pages.serverStore.saveConfig") }}</AdminButton>
          <AdminButton variant="primary" :loading="publishing" @click="publish">
            {{ $t("pages.serverStore.saveAndUpload") }}
          </AdminButton>
          <AdminButton variant="secondary" :loading="heartbeating" @click="heartbeat">
            {{ $t("pages.serverStore.sendHeartbeat") }}
          </AdminButton>
        </div>
        </div>
      </AdminPanel>

      <div class="space-y-4">
        <AdminPanel>
          <div class="p-4 md:p-5">
          <h2 class="text-base font-semibold text-foreground">{{ $t("pages.serverStore.officialStatus") }}</h2>
          <dl class="mt-4 space-y-3 text-sm">
            <div>
              <dt class="text-muted">{{ $t("pages.serverStore.officialDomain") }}</dt>
              <dd class="mt-1 break-all text-foreground">{{ form.officialHomeUrl }}</dd>
            </div>
            <div>
              <dt class="text-muted">{{ $t("pages.serverStore.storeLink") }}</dt>
              <dd class="mt-1 break-all">
                <a
                  :href="form.officialStoreUrl"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="text-primary-600 hover:text-primary-700"
                >
                  {{ form.officialStoreUrl }}
                </a>
              </dd>
            </div>
            <div>
              <dt class="text-muted">{{ $t("pages.serverStore.officialServerId") }}</dt>
              <dd class="mt-1 break-all text-foreground">
                {{ form.officialServerId || $t("common.notYet") }}
              </dd>
            </div>
            <div>
              <dt class="text-muted">{{ $t("pages.serverStore.lastSync") }}</dt>
              <dd class="mt-1 text-foreground">
                {{ form.lastSyncAt ? formatDateTime(form.lastSyncAt) : $t("common.none") }}
              </dd>
            </div>
            <div>
              <dt class="text-muted">{{ $t("pages.serverStore.lastHeartbeat") }}</dt>
              <dd class="mt-1 text-foreground">
                {{ form.lastHeartbeatAt ? formatDateTime(form.lastHeartbeatAt) : $t("common.none") }}
              </dd>
            </div>
          </dl>
          <AdminAlert v-if="form.lastError" class="mt-4 !mb-0" variant="warning" :title="$t('pages.serverStore.lastError')">
            {{ form.lastError }}
          </AdminAlert>
          </div>
        </AdminPanel>

        <AdminPanel>
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
