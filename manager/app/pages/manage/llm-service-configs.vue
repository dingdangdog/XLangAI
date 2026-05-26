<script setup lang="ts">
const { t } = useI18n();

const { serviceUsageMonthLine, serviceUsageTodayLine } = useUsageDisplay();

const API = "/api/admin/llm-service-configs";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

const PROTOCOLS = [
  { value: "openai", label: "OpenAI 官方 / 兼容网关" },
  { value: "azure_openai", label: "Azure OpenAI（Chat Completions）" },
  { value: "ollama", label: "Ollama（OpenAI 兼容）" },
  { value: "deepseek", label: "DeepSeek（OpenAI 兼容）" },
  { value: "openrouter", label: "OpenRouter（OpenAI 兼容）" },
  { value: "groq", label: "Groq（OpenAI 兼容）" },
  { value: "together", label: "Together（OpenAI 兼容）" },
  { value: "zhipu", label: "智谱（OpenAI 兼容）" },
  { value: "moonshot", label: "Moonshot（OpenAI 兼容）" },
  { value: "siliconflow", label: "SiliconFlow（OpenAI 兼容）" },
  { value: "nvidia_nim", label: "NVIDIA NIM（OpenAI 兼容 Chat）" },
  { value: "claude", label: "Anthropic Claude（Messages API）" },
  { value: "gemini", label: "Google Gemini（generateContent）" },
] as const;

const CONFIG_HINTS: Record<string, string> = {
  openai: "{}",
  azure_openai: '{"api_version":"2024-02-15-preview"}',
  ollama: "{}",
  deepseek: "{}",
  openrouter: "{}",
  groq: "{}",
  together: "{}",
  zhipu: "{}",
  moonshot: "{}",
  siliconflow: "{}",
  nvidia_nim: "{}",
  claude: '{"anthropic_version":"2023-06-01","max_tokens":4096,"temperature":0.7}',
  gemini: '{"max_tokens":8192,"temperature":0.7}',
};

const DEFAULT_MODEL: Record<string, string> = {
  openai: "gpt-4o-mini",
  claude: "claude-3-5-sonnet-20241022",
  gemini: "gemini-1.5-flash",
  deepseek: "deepseek-chat",
  ollama: "llama3",
};

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);

async function load() {
  loading.value = true;
  try {
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

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

const form = reactive({
  id: "",
  code: "",
  name: "",
  protocol: "openai",
  baseUrl: "",
  apiKey: "",
  modelCode: "",
  config: "{}",
  status: "active",
  sortOrder: 0,
  remark: "",
});

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.protocol = "openai";
  form.baseUrl = "";
  form.apiKey = "";
  form.modelCode = "";
  form.config = CONFIG_HINTS.openai ?? "{}";
  form.status = "active";
  form.sortOrder = 0;
  form.remark = "";
}

function onProtocolChange(v: string) {
  form.config = CONFIG_HINTS[v] ?? "{}";
  const def = DEFAULT_MODEL[v];
  if (def && (dialogMode.value === "create" || !form.modelCode.trim())) {
    form.modelCode = def;
  }
}

watch(() => form.protocol, onProtocolChange);

/** 库表 code 唯一；Go 按 id 或 sort_order 取 active，不按 code 查。新建时自动生成。 */
function autoLlmCode(protocol: string) {
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
  form.modelCode = String(row.modelCode ?? "");
  form.config = row.config != null ? String(row.config) : "{}";
  form.status = String(row.status ?? "active");
  form.sortOrder = Number(row.sortOrder ?? 0);
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
}

function isOpenAICompatible(protocol: string): boolean {
  const p = protocol.trim().toLowerCase();
  return (
    p === "" ||
    p === "openai" ||
    [
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
    ].includes(p)
  );
}

async function submit() {
  if (!form.name.trim() || !form.modelCode.trim()) {
    toast.warning("请填写名称与模型 code");
    return;
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
  const body = {
    code:
      dialogMode.value === "create"
        ? autoLlmCode(form.protocol)
        : form.code.trim(),
    name: form.name.trim(),
    protocol: form.protocol,
    baseUrl: form.baseUrl.trim() || null,
    apiKey: form.apiKey.trim() || null,
    modelCode: form.modelCode.trim(),
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
    message: "确认删除该 LLM 配置？",
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

const statusOptions = [
  { value: "active", label: "active" },
  { value: "inactive", label: "inactive" },
];

const protocolOptions = PROTOCOLS.map((p) => ({ value: p.value, label: p.label }));

const { activateRow, activatingId } = useActivateConfigRow({
  api,
  getList: () => list.value,
  exclusivity: "multi-active",
  reload: load,
});
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        title="LLM 服务配置"
        description="Go 按 id 或 sort_order 最小的 active 记录选用，不按编码查找；新建时编码自动生成。"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
        </template>
      </AdminPageHeader>

      <AdminAlert title="协议说明">
        <ul class="list-disc pl-5 text-sm space-y-1">
          <li>
            <strong>OpenAI 兼容类</strong>（openai、azure_openai、ollama、deepseek 等）：Base URL 填根地址（勿带
            /v1）；API Key 须在后台填写。
          </li>
          <li><strong>claude</strong>：默认 <code>https://api.anthropic.com</code>；config 可设
            <code>anthropic_version</code>、<code>max_tokens</code>。</li>
          <li><strong>gemini</strong>：默认 Google Generative Language API；model 如
            <code>gemini-1.5-flash</code>。</li>
          <li>语音转写（STT）使用独立的 STT 服务配置，不会复用 LLM 配置或环境变量。</li>
        </ul>
      </AdminAlert>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>编码</AdminTh>
          <AdminTh>{{ $t("common.name") }}</AdminTh>
          <AdminTh width="140px">协议</AdminTh>
          <AdminTh>模型</AdminTh>
          <AdminTh>API Key</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh width="120px">今日用量</AdminTh>
          <AdminTh width="120px">本月用量</AdminTh>
          <AdminTh width="72px">{{ $t("common.sort") }}</AdminTh>
          <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>
          <AdminTh width="200px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap>{{ row.code }}</AdminTd>
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.protocol }}</AdminTd>
          <AdminTd>{{ row.modelCode }}</AdminTd>
          <AdminTd><AdminMaskedKey :value="row.apiKey as string | null" /></AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
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
              启用
            </AdminButton>
            <AdminButton variant="link" @click="openEdit(row)">{{ $t("common.edit") }}</AdminButton>
            <AdminButton variant="link" class="!text-danger-600" @click="removeRow(row)">
              删除
            </AdminButton>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建 LLM 配置' : '编辑 LLM 配置'"
      width="lg"
    >
      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField :label="$t('common.name')" required>
        <AdminInput v-model="form.name" />
      </AdminFormField>
      <AdminFormField label="协议">
        <AdminSelect v-model="form.protocol" :options="protocolOptions" />
      </AdminFormField>
      <AdminFormField
        label="Base URL"
        :hint="
          isOpenAICompatible(form.protocol)
            ? 'OpenAI 兼容：根地址（不要带 /v1）'
            : form.protocol === 'claude'
              ? '默认 https://api.anthropic.com'
              : '默认 Google Generative Language API'
        "
      >
        <AdminInput v-model="form.baseUrl" placeholder="可选" />
      </AdminFormField>
      <AdminFormField
        label="API Key"
        hint="必填"
      >
        <AdminInput v-model="form.apiKey" type="password" />
      </AdminFormField>
      <AdminFormField label="模型 code" required>
        <AdminInput v-model="form.modelCode" placeholder="如 gpt-4o-mini、claude-3-5-sonnet-20241022" />
      </AdminFormField>
      <AdminFormField label="扩展 JSON" :hint="CONFIG_HINTS[form.protocol]">
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
  </AdminListPage>
</template>
