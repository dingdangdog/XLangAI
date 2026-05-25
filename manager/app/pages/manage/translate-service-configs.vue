<script setup lang="ts">
import {
  serviceUsageMonthLine,
  serviceUsageTodayLine,
} from "~/utils/usageDisplay";

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
      return "AccessKey ID";
    case "google_translate":
      return "API Key";
    case "deepl":
      return "Auth Key";
    case "azure_translator":
      return "订阅密钥";
    default:
      return "API Key";
  }
}

function apiSecretLabel(protocol: string): string {
  switch (protocol) {
    case "tencent_translate":
      return "SecretKey";
    case "baidu_translate":
      return "密钥（Secret Key）";
    case "aliyun_translate":
      return "AccessKey Secret";
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
    toast.error("加载失败");
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
    toast.warning("请填写名称");
    return;
  }
  if (form.protocol === "openai" && useLlmLink.value) {
    if (!form.llmConfigId.trim()) {
      toast.warning("请选择要关联的 LLM 配置");
      return;
    }
  } else if (modelRequired() && !form.modelCode.trim()) {
    toast.warning("请填写模型 code，或改为关联已有 LLM 配置");
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
      toast.error("扩展配置须为合法 JSON");
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
      toast.success("已创建");
    } else {
      await api.update(form.id, { id: form.id, ...body });
      toast.success("已保存");
    }
    dialogVisible.value = false;
    await load();
  } catch (e) {
    toast.error("保存失败");
    console.error(e);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: "确认删除该翻译配置？",
    danger: true,
    confirmLabel: "删除",
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success("已删除");
    await load();
  } catch (e) {
    toast.error("删除失败");
    console.error(e);
  }
}

const statusOptions = [
  { value: "active", label: "active（启用，将自动停用其它配置）" },
  { value: "inactive", label: "inactive" },
];

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
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        title="翻译服务配置"
        description="全局仅一条 active。列表「今日/本月用量」为成功翻译次数与源文本字符数（UTC 自然日）。"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">新建</AdminButton>
        </template>
      </AdminPageHeader>

      <AdminAlert v-if="activeCount > 1" title="配置异常" variant="warning">
        当前列表中有 {{ activeCount }} 条 active 记录，应仅保留一条。请编辑并将多余项设为 inactive。
      </AdminAlert>

      <AdminAlert title="协议说明">
        <ul class="list-disc pl-5 text-sm space-y-1">
          <li>
            <strong>openai（LLM 翻译）</strong>：可<strong>关联</strong>「LLM 服务配置」中已启用的 OpenAI
            兼容项，复用 Base URL / API Key / 模型，无需重复填写；也可单独自定义。
          </li>
          <li><strong>azure_translator</strong>：config 中配置 region；订阅密钥填在 API Key。</li>
          <li><strong>deepl</strong>：Auth Key；<code>use_free_api</code> 可选。</li>
          <li><strong>google_translate</strong>：Google API Key。</li>
          <li><strong>baidu_translate</strong>：App ID + 密钥（双字段）。</li>
          <li><strong>tencent_translate</strong>：SecretId + SecretKey（双字段，与 loden 一致）；region 写在扩展 JSON。</li>
          <li><strong>aliyun_translate</strong>：AccessKey ID + AccessKey Secret；region 写在扩展 JSON。</li>
        </ul>
      </AdminAlert>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>名称</AdminTh>
          <AdminTh width="140px">协议</AdminTh>
          <AdminTh>模型 / 关联</AdminTh>
          <AdminTh>密钥</AdminTh>
          <AdminTh width="88px">状态</AdminTh>
          <AdminTh width="120px">今日用量</AdminTh>
          <AdminTh width="120px">本月用量</AdminTh>
          <AdminTh width="72px">排序</AdminTh>
          <AdminTh>更新时间</AdminTh>
          <AdminTh width="200px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.protocol }}</AdminTd>
          <AdminTd>
            <span v-if="row.llmConfigId">关联 {{ llmLinkDisplay(row.llmConfigId) }}</span>
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
              启用
            </AdminButton>
            <AdminButton variant="link" @click="openEdit(row)">编辑</AdminButton>
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
      :title="dialogMode === 'create' ? '新建翻译配置' : '编辑翻译配置'"
      width="lg"
    >
      <AdminFormField v-if="dialogMode === 'edit'" label="ID">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField label="名称" required>
        <AdminInput v-model="form.name" />
      </AdminFormField>
      <AdminFormField label="协议">
        <AdminSelect v-model="form.protocol" :options="protocolOptions" />
      </AdminFormField>

      <template v-if="form.protocol === 'openai'">
        <AdminFormField label="LLM 配置来源">
          <label class="inline-flex items-center gap-2 text-sm cursor-pointer">
            <input v-model="useLlmLink" type="checkbox" class="rounded border-border" />
            关联已有 LLM 服务配置（推荐，避免重复填 Key）
          </label>
        </AdminFormField>
        <AdminFormField v-if="useLlmLink" label="关联 LLM" required>
          <AdminSelect
            v-model="form.llmConfigId"
            :options="llmConfigOptions"
            placeholder="选择已启用的 OpenAI 兼容 LLM"
          />
          <p v-if="llmConfigOptions.length === 0" class="mt-1 text-xs text-muted">
            请先在「LLM 服务配置」中创建并启用一条 OpenAI 兼容配置。
          </p>
          <p v-else-if="llmLinkLabel" class="mt-1 text-xs text-muted">已选：{{ llmLinkLabel }}</p>
        </AdminFormField>
        <template v-else>
          <AdminFormField label="Base URL" hint="OpenAI 兼容根地址（勿带 /v1）">
            <AdminInput v-model="form.baseUrl" placeholder="可选，默认 OpenAI 官方" />
          </AdminFormField>
          <AdminFormField label="API Key" hint="必填">
            <AdminInput v-model="form.apiKey" type="password" />
          </AdminFormField>
          <AdminFormField label="模型 code" required>
            <AdminInput v-model="form.modelCode" placeholder="gpt-4o-mini" />
          </AdminFormField>
        </template>
        <AdminFormField
          v-if="useLlmLink"
          label="模型 code（可选）"
          hint="留空则使用关联 LLM 的 model_code"
        >
          <AdminInput v-model="form.modelCode" placeholder="可选覆盖" />
        </AdminFormField>
      </template>

      <template v-else>
        <AdminFormField
          v-if="form.protocol === 'azure_translator' || form.protocol === 'deepl'"
          label="Base URL"
          hint="可选；厂商自定义端点"
        >
          <AdminInput v-model="form.baseUrl" placeholder="可选" />
        </AdminFormField>
        <AdminFormField :label="apiKeyLabel(form.protocol)">
          <AdminInput v-model="form.apiKey" type="password" />
        </AdminFormField>
        <AdminFormField v-if="needsApiSecret(form.protocol)" :label="apiSecretLabel(form.protocol)">
          <AdminInput v-model="form.apiSecret" type="password" />
        </AdminFormField>
      </template>

      <AdminFormField label="扩展 JSON" :hint="CONFIG_HINTS[form.protocol]">
        <AdminInput v-model="form.config" type="textarea" :rows="4" class="font-mono text-sm" />
      </AdminFormField>
      <AdminFormField label="状态">
        <AdminSelect v-model="form.status" :options="statusOptions" />
      </AdminFormField>
      <AdminFormField label="排序">
        <AdminInput v-model="form.sortOrder" type="number" />
      </AdminFormField>
      <AdminFormField label="备注">
        <AdminInput v-model="form.remark" type="textarea" :rows="2" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">取消</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">保存</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
