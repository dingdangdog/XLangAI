<script setup lang="ts">
const { t } = useI18n();

const { serviceUsageMonthLine, serviceUsageTodayLine } = useUsageDisplay();

const API = "/api/admin/translate-service-configs";
const LLM_API = "/api/admin/llm-service-configs";
const api = useAdminResourceApi(API);
const llmApi = useAdminResourceApi(LLM_API);
const toast = useToast();
const { confirm } = useConfirm();

const PROTOCOLS = [
  { value: "openai", label: "OpenAI 兼容（LLM Chat 翻译）" },
  { value: "azure_translator", label: "Azure 翻译（Text API v3）" },
  { value: "deepl", label: "DeepL" },
  { value: "google_translate", label: "Google Cloud Translation" },
  { value: "baidu_translate", label: "百度翻译" },
  { value: "tencent_translate", label: "腾讯云翻译（TMT）" },
  { value: "aliyun_translate", label: "阿里云机器翻译" },
] as const;

const CONFIG_HINTS: Record<string, string> = {
  openai: '{"system_prompt_template":"... %s ...","max_chars":12000}',
  azure_translator: '{"region":"eastasia","api_version":"3.0"}',
  deepl: '{"use_free_api":true}',
  google_translate: "{}",
  baidu_translate: "{}",
  tencent_translate: '{"region":"ap-guangzhou"}',
  aliyun_translate: '{"region":"cn-hangzhou"}',
};

const OPENAI_COMPAT = new Set([
  "openai",
  "azure_openai",
  "ollama",
  "deepseek",
  "openrouter",
  "groq",
  "together",
  "zhipu",
  "moonshot",
  "siliconflow",
  "nvidia_nim",
  "mistral",
]);

function isOpenAICompatible(protocol: string): boolean {
  return OPENAI_COMPAT.has(protocol.trim().toLowerCase());
}

function needsApiSecret(protocol: string): boolean {
  return ["tencent_translate", "baidu_translate", "aliyun_translate"].includes(protocol);
}

function apiKeyLabel(protocol: string): string {
  switch (protocol) {
    case "tencent_translate":
      return "SecretId";
    case "baidu_translate":
      return "App ID";
    case "aliyun_translate":
      return t("fields.accessKeyId");
    case "google_translate":
      return t("fields.apiKey");
    case "deepl":
      return "Auth Key";
    case "azure_translator":
      return t("pages.translateConfigs.subscriptionKey");
    default:
      return t("fields.apiKey");
  }
}

function apiSecretLabel(protocol: string): string {
  switch (protocol) {
    case "tencent_translate":
      return "SecretKey";
    case "baidu_translate":
      return t("pages.translateConfigs.secretKey");
    case "aliyun_translate":
      return t("fields.secretAccessKey");
    default:
      return "API Secret";
  }
}

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const llmConfigs = ref<Record<string, unknown>[]>([]);

async function loadLlmConfigs() {
  try {
    const res = await llmApi.list({ page: 1, pageSize: 200 });
    llmConfigs.value = res.items.filter(
      (r) =>
        String(r.status) === "active" &&
        isOpenAICompatible(String(r.protocol ?? "openai")),
    );
  } catch (e) {
    console.error(e);
  }
}

function llmLinkDisplay(id: unknown): string {
  const s = String(id ?? "").trim();
  if (!s) return "";
  const row = llmConfigs.value.find((r) => String(r.id) === s);
  return row ? String(row.name ?? row.code) : `LLM ${s.slice(0, 8)}…`;
}

async function load() {
  loading.value = true;
  try {
    await loadLlmConfigs();
    const res = await api.list({ page: page.value, pageSize: pageSize.value });
    list.value = res.items;
    total.value = res.total;
  } catch (e) {
    toast.error(t("toast.loadFailed"));
    console.error(e);
  } finally {
    loading.value = false;
  }
}

watch([page, pageSize], () => void load(), { immediate: true });
onMounted(() => void loadLlmConfigs());

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);
const useLlmLink = ref(false);

const form = reactive({
  id: "",
  code: "",
  name: "",
  protocol: "openai",
  baseUrl: "",
  apiKey: "",
  apiSecret: "",
  modelCode: "",
  llmConfigId: "",
  config: "{}",
  status: "inactive",
  sortOrder: 0,
  remark: "",
});

const llmConfigOptions = computed(() =>
  llmConfigs.value.map((r) => ({
    value: String(r.id),
    label: `${r.name} · ${r.protocol} · ${r.modelCode}`,
  })),
);

const llmLinkLabel = computed(() => {
  const id = form.llmConfigId.trim();
  if (!id) return "";
  const row = llmConfigs.value.find((r) => String(r.id) === id);
  return row ? `${row.name} (${row.protocol})` : id;
});

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.protocol = "openai";
  form.baseUrl = "";
  form.apiKey = "";
  form.apiSecret = "";
  form.modelCode = "";
  form.llmConfigId = "";
  form.config = CONFIG_HINTS.openai ?? "{}";
  form.status = "inactive";
  form.sortOrder = 0;
  form.remark = "";
  useLlmLink.value = false;
}

function onProtocolChange(v: string) {
  form.config = CONFIG_HINTS[v] ?? "{}";
  if (v === "openai") {
    if (!useLlmLink.value && !form.modelCode.trim()) {
      form.modelCode = "gpt-4o-mini";
    }
  } else {
    useLlmLink.value = false;
    form.llmConfigId = "";
    if (!form.modelCode.trim() || form.modelCode === "gpt-4o-mini") {
      form.modelCode = "-";
    }
  }
}

watch(() => form.protocol, onProtocolChange);

watch(useLlmLink, (linked) => {
  if (!linked) {
    form.llmConfigId = "";
    if (form.protocol === "openai" && !form.modelCode.trim()) {
      form.modelCode = "gpt-4o-mini";
    }
  } else {
    form.baseUrl = "";
    form.apiKey = "";
    if (!form.llmConfigId && llmConfigs.value.length === 1) {
      form.llmConfigId = String(llmConfigs.value[0]?.id ?? "");
    }
  }
});

function autoTranslateCode(protocol: string) {
  const p = (protocol || "openai").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  form.modelCode = "gpt-4o-mini";
  dialogVisible.value = true;
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  form.code = String(row.code ?? "");
  form.name = String(row.name ?? "");
  form.protocol = String(row.protocol ?? "openai");
  form.baseUrl = String(row.baseUrl ?? "");
  form.apiKey = String(row.apiKey ?? "");
  form.apiSecret = String(row.apiSecret ?? "");
  form.modelCode = String(row.modelCode ?? "");
  form.llmConfigId = String(row.llmConfigId ?? "");
  form.config = row.config != null ? String(row.config) : "{}";
  form.status = String(row.status ?? "inactive");
  form.sortOrder = Number(row.sortOrder ?? 0);
  form.remark = String(row.remark ?? "");
  useLlmLink.value =
    form.protocol === "openai" && form.llmConfigId.trim().length > 0;
  dialogVisible.value = true;
}

function modelRequired(): boolean {
  return form.protocol === "openai" && !useLlmLink.value;
}

async function submit() {
  if (!form.name.trim()) {
    toast.warning(t("validation.fillName"));
    return;
  }
  if (form.protocol === "openai" && useLlmLink.value) {
    if (!form.llmConfigId.trim()) {
      toast.warning(t("validation.selectLlmConfig"));
      return;
    }
  } else if (modelRequired() && !form.modelCode.trim()) {
    toast.warning(t("validation.fillModelOrLinkLlm"));
    return;
  }
  if (!modelRequired() && !form.modelCode.trim()) {
    form.modelCode = "-";
  }
  let configStr = (form.config ?? "").trim();
  if (configStr) {
    try {
      JSON.parse(configStr);
    } catch {
      toast.error(t("validation.invalidJson", { field: t("fields.extJson") }));
      return;
    }
  } else {
    configStr = "{}";
  }

  const linked = form.protocol === "openai" && useLlmLink.value;
  const body = {
    code:
      dialogMode.value === "create"
        ? autoTranslateCode(form.protocol)
        : form.code.trim(),
    name: form.name.trim(),
    protocol: form.protocol,
    baseUrl: linked ? null : form.baseUrl.trim() || null,
    apiKey: linked ? null : form.apiKey.trim() || null,
    apiSecret: needsApiSecret(form.protocol) ? form.apiSecret.trim() || null : null,
    modelCode: form.modelCode.trim(),
    llmConfigId: linked ? form.llmConfigId.trim() : null,
    config: configStr,
    status: form.status,
    sortOrder: form.sortOrder,
    remark: form.remark.trim() || null,
  };

  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await api.create(body);
      toast.success(t("toast.created"));
    } else {
      await api.update(form.id, { id: form.id, ...body });
      toast.success(t("toast.saved"));
    }
    dialogVisible.value = false;
    await load();
  } catch (e) {
    toast.error(t("toast.saveFailed"));
    console.error(e);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: t("confirm.deleteTranslateConfig"),
    danger: true,
    confirmLabel: t("common.delete"),
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success(t("toast.deleted"));
    await load();
  } catch (e) {
    toast.error(t("toast.deleteFailed"));
    console.error(e);
  }
}

const statusOptions = computed(() => [
  { value: "active", label: t("pages.translateConfigs.statusActive") },
  { value: "inactive", label: "inactive" },
]);

const protocolOptions = PROTOCOLS.map((p) => ({ value: p.value, label: p.label }));

const activeCount = computed(
  () => list.value.filter((r) => String(r.status) === "active").length,
);

const { activateRow, activatingId } = useActivateConfigRow({
  api,
  getList: () => list.value,
  exclusivity: "single-active",
  reload: load,
});
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col gap-4">
    <div class="flex justify-end">
      <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
    </div>

    <AdminAlert v-if="activeCount > 1" :title="$t('pages.objectStorage.configAnomalyTitle')" variant="warning">
      {{ $t("pages.translateConfigs.configAnomaly", { count: activeCount }) }}
    </AdminAlert>

    <AdminAlert :title="$t('pages.translateConfigs.protocolAlert')">
      <p class="whitespace-pre-line text-sm">{{ $t("pages.translateConfigs.protocolAlertDetails") }}</p>
    </AdminAlert>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("common.name") }}</AdminTh>
          <AdminTh width="140px">{{ $t("fields.protocol") }}</AdminTh>
          <AdminTh>{{ $t("fields.modelOrLink") }}</AdminTh>
          <AdminTh>{{ $t("fields.secret") }}</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh width="120px">{{ $t("common.todayUsage") }}</AdminTh>
          <AdminTh width="120px">{{ $t("common.monthUsage") }}</AdminTh>
          <AdminTh width="72px">{{ $t("common.sort") }}</AdminTh>
          <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>
          <AdminTh width="200px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.protocol }}</AdminTd>
          <AdminTd>
            <span v-if="row.llmConfigId">{{ $t("pages.translateConfigs.linkedPrefix", { label: llmLinkDisplay(row.llmConfigId) }) }}</span>
            <span v-else>{{ row.modelCode }}</span>
          </AdminTd>
          <AdminTd class="space-y-1">
            <AdminMaskedKey :value="row.apiKey as string | null" />
            <AdminMaskedKey v-if="row.apiSecret" :value="row.apiSecret as string" />
          </AdminTd>
          <AdminTd>
            <AdminBadge :variant="row.status === 'active' ? 'success' : 'default'">
              {{ row.status }}
            </AdminBadge>
          </AdminTd>
          <AdminTd class="text-sm tabular-nums">
            <div>{{ serviceUsageTodayLine(row.usage as Record<string, unknown> | undefined) }}</div>
          </AdminTd>
          <AdminTd class="text-sm tabular-nums">
            <div>{{ serviceUsageMonthLine(row.usage as Record<string, unknown> | undefined) }}</div>
          </AdminTd>
          <AdminTd>{{ row.sortOrder }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
          <AdminTd align="right" class="whitespace-nowrap">
            <AdminButton
              v-if="String(row.status) !== 'active'"
              variant="link"
              :loading="activatingId === String(row.id)"
              @click="activateRow(row)"
            >
              {{ $t("common.enable") }}
            </AdminButton>
            <AdminButton variant="link" @click="openEdit(row)">{{ $t("common.edit") }}</AdminButton>
            <AdminButton variant="link" class="!text-danger-600" @click="removeRow(row)">
              {{ $t("common.delete") }}
            </AdminButton>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? $t('pages.translateConfigs.createDialog') : $t('pages.translateConfigs.editDialog')"
      width="lg"
    >
      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField :label="$t('common.name')" required>
        <AdminInput v-model="form.name" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.protocol')">
        <AdminSelect v-model="form.protocol" :options="protocolOptions" />
      </AdminFormField>

      <template v-if="form.protocol === 'openai'">
        <AdminFormField :label="$t('pages.translateConfigs.llmSource')">
          <label class="inline-flex items-center gap-2 text-sm cursor-pointer">
            <input v-model="useLlmLink" type="checkbox" class="rounded border-border" />
            {{ $t("pages.translateConfigs.linkLlmRecommended") }}
          </label>
        </AdminFormField>
        <AdminFormField v-if="useLlmLink" :label="$t('fields.linkedLlmShort')" required>
          <AdminSelect
            v-model="form.llmConfigId"
            :options="llmConfigOptions"
            :placeholder="$t('pages.translateConfigs.selectActiveLlm')"
          />
          <p v-if="llmConfigOptions.length === 0" class="mt-1 text-xs text-muted">
            {{ $t("pages.translateConfigs.createLlmFirst") }}
          </p>
          <p v-else-if="llmLinkLabel" class="mt-1 text-xs text-muted">
            {{ $t("pages.translateConfigs.selectedLlm", { label: llmLinkLabel }) }}
          </p>
        </AdminFormField>
        <template v-else>
          <AdminFormField :label="$t('fields.baseUrl')" :hint="$t('pages.translateConfigs.baseUrlHint')">
            <AdminInput v-model="form.baseUrl" :placeholder="$t('pages.translateConfigs.baseUrlPlaceholder')" />
          </AdminFormField>
          <AdminFormField :label="$t('fields.apiKey')" :hint="$t('pages.translateConfigs.apiKeyRequired')">
            <AdminInput v-model="form.apiKey" type="password" />
          </AdminFormField>
          <AdminFormField :label="$t('fields.modelCode')" required>
            <AdminInput v-model="form.modelCode" placeholder="gpt-4o-mini" />
          </AdminFormField>
        </template>
        <AdminFormField
          v-if="useLlmLink"
          :label="`${$t('fields.modelCode')} (${$t('common.optional')})`"
          :hint="$t('pages.translateConfigs.modelOverrideHint')"
        >
          <AdminInput v-model="form.modelCode" :placeholder="$t('pages.translateConfigs.modelOverride')" />
        </AdminFormField>
      </template>

      <template v-else>
        <AdminFormField
          v-if="form.protocol === 'azure_translator' || form.protocol === 'deepl'"
          :label="$t('fields.baseUrl')"
          :hint="$t('pages.translateConfigs.customEndpointHint')"
        >
          <AdminInput v-model="form.baseUrl" :placeholder="$t('pages.llmConfigs.optionalPlaceholder')" />
        </AdminFormField>
        <AdminFormField :label="apiKeyLabel(form.protocol)">
          <AdminInput v-model="form.apiKey" type="password" />
        </AdminFormField>
        <AdminFormField v-if="needsApiSecret(form.protocol)" :label="apiSecretLabel(form.protocol)">
          <AdminInput v-model="form.apiSecret" type="password" />
        </AdminFormField>
      </template>

      <AdminFormField :label="$t('fields.extJson')" :hint="CONFIG_HINTS[form.protocol]">
        <AdminInput v-model="form.config" type="textarea" :rows="4" class="font-mono text-sm" />
      </AdminFormField>
      <AdminFormField :label="$t('common.status')">
        <AdminSelect v-model="form.status" :options="statusOptions" />
      </AdminFormField>
      <AdminFormField :label="$t('common.sort')">
        <AdminInput v-model="form.sortOrder" type="number" />
      </AdminFormField>
      <AdminFormField :label="$t('common.remark')">
        <AdminInput v-model="form.remark" type="textarea" :rows="2" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">{{ $t("common.save") }}</AdminButton>
      </template>
    </AdminDialog>
  </div>
</template>
