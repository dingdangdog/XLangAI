<script setup lang="ts">
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

const statLabels: Record<string, string> = {
  users: "用户总数",
  activeUsers: "活跃用户",
  languages: "语言",
  voiceRoles: "语音角色",
  conversations: "会话",
  messages: "消息",
  llmConfigs: "LLM 配置",
  sttConfigs: "STT 配置",
  ttsConfigs: "TTS 配置",
  translateConfigs: "翻译配置",
};

const stats = computed(() => {
  if (!form.stats) return [];
  return Object.entries(form.stats).map(([key, value]) => ({
    label: statLabels[key] ?? key,
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
    toast.error("加载服务器商店配置失败");
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
    toast.success("已保存服务器商店配置");
  } catch (error) {
    toast.error("保存失败");
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
    toast.success("已上传到官网服务器商店");
  } catch (error) {
    toast.error("上传失败，请检查服务器地址、官网地址和网络连通性");
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
    toast.success("心跳已发送");
  } catch (error) {
    toast.error("心跳发送失败");
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
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        title="服务器商店"
        description="维护本服务器公开信息，并同步到官网服务器商店。"
      >
        <template #actions>
          <a
            :href="form.officialStoreUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center rounded-lg border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-surface-muted"
          >
            打开官网商店
          </a>
        </template>
      </AdminPageHeader>
    </template>

    <AdminSkeleton v-if="loading" :rows="8" />

    <div v-else class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_360px]">
      <AdminPanel>
        <div class="p-4 md:p-5">
        <AdminAlert title="同步说明">
          上传到官网的是本服务器公开资料和可选统计数据，不包含用户资料、会话内容、消息内容或密钥。
          开启公开后，后台会定时向官网发送心跳，用于判断服务器是否活跃。
        </AdminAlert>

        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="开放到官网服务器商店">
            <AdminCheckbox v-model="form.enabled" label="允许官网展示本服务器" />
          </AdminFormField>
          <AdminFormField label="上传统计数据">
            <AdminCheckbox v-model="form.uploadStats" label="上传聚合统计数量" />
          </AdminFormField>
          <AdminFormField label="本服务器地址" required>
            <AdminInput
              v-model="form.serverAddress"
              placeholder="如 https://api.example.com 或 https://example.com"
            />
          </AdminFormField>
          <AdminFormField label="官网地址">
            <AdminInput v-model="form.homepageUrl" placeholder="如 https://example.com" />
          </AdminFormField>
          <AdminFormField label="服务器名称" required>
            <AdminInput v-model="form.name" placeholder="如 某某学习社区服务器" />
          </AdminFormField>
          <AdminFormField label="地区">
            <AdminInput v-model="form.region" placeholder="如 中国大陆 / Singapore" />
          </AdminFormField>
          <AdminFormField label="联系邮箱">
            <AdminInput v-model="form.contactEmail" type="email" placeholder="admin@example.com" />
          </AdminFormField>
          <AdminFormField label="版本">
            <AdminInput v-model="form.version" placeholder="如 1.0.0" />
          </AdminFormField>
          <AdminFormField label="Logo 地址">
            <AdminInput v-model="form.logoUrl" placeholder="https://example.com/logo.png" />
          </AdminFormField>
          <AdminFormField label="心跳间隔（秒）">
            <AdminInput v-model="form.heartbeatIntervalSeconds" type="number" />
          </AdminFormField>
          <AdminFormField class="md:col-span-2" label="一句话简介">
            <AdminInput
              v-model="form.summary"
              type="textarea"
              :rows="2"
              placeholder="展示在官网列表中的简短介绍"
            />
          </AdminFormField>
          <AdminFormField class="md:col-span-2" label="详细简介">
            <AdminInput
              v-model="form.description"
              type="textarea"
              :rows="5"
              placeholder="服务器面向人群、使用说明、限制条件等"
            />
          </AdminFormField>
        </div>

        <div class="mt-6 flex flex-wrap gap-3">
          <AdminButton variant="secondary" :loading="saving" @click="save">保存配置</AdminButton>
          <AdminButton variant="primary" :loading="publishing" @click="publish">
            保存并上传到官网
          </AdminButton>
          <AdminButton variant="secondary" :loading="heartbeating" @click="heartbeat">
            立即发送心跳
          </AdminButton>
        </div>
        </div>
      </AdminPanel>

      <div class="space-y-4">
        <AdminPanel>
          <div class="p-4 md:p-5">
          <h2 class="text-base font-semibold text-foreground">官网状态</h2>
          <dl class="mt-4 space-y-3 text-sm">
            <div>
              <dt class="text-muted">官网域名</dt>
              <dd class="mt-1 break-all text-foreground">{{ form.officialHomeUrl }}</dd>
            </div>
            <div>
              <dt class="text-muted">商店链接</dt>
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
              <dt class="text-muted">官网服务器 ID</dt>
              <dd class="mt-1 break-all text-foreground">
                {{ form.officialServerId || "尚未上传" }}
              </dd>
            </div>
            <div>
              <dt class="text-muted">最近同步</dt>
              <dd class="mt-1 text-foreground">
                {{ form.lastSyncAt ? formatDateTime(form.lastSyncAt) : "暂无" }}
              </dd>
            </div>
            <div>
              <dt class="text-muted">最近心跳</dt>
              <dd class="mt-1 text-foreground">
                {{ form.lastHeartbeatAt ? formatDateTime(form.lastHeartbeatAt) : "暂无" }}
              </dd>
            </div>
          </dl>
          <AdminAlert v-if="form.lastError" class="mt-4 !mb-0" variant="warning" title="最近错误">
            {{ form.lastError }}
          </AdminAlert>
          </div>
        </AdminPanel>

        <AdminPanel>
          <div class="p-4 md:p-5">
          <h2 class="text-base font-semibold text-foreground">将上传的统计数据</h2>
          <p class="mt-2 text-sm text-muted">
            仅在开启“上传统计数据”后发送到官网，内容为聚合数量。
          </p>
          <div v-if="stats.length" class="mt-4 grid grid-cols-2 gap-3">
            <div v-for="stat in stats" :key="stat.label" class="rounded-lg bg-surface-muted p-3">
              <div class="text-xs text-muted">{{ stat.label }}</div>
              <div class="mt-1 text-lg font-semibold text-foreground">{{ stat.value }}</div>
            </div>
          </div>
          <p v-else class="mt-4 text-sm text-muted">当前未开启统计上传。</p>
          </div>
        </AdminPanel>
      </div>
    </div>
  </AdminListPage>
</template>
