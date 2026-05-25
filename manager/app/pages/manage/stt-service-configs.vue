<script setup lang="ts">
const API = "/api/admin/stt-service-configs";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

const PROTOCOLS = [
  { value: "openai", label: "OpenAI 兼容（Whisper 等）" },
  { value: "azure_speech_rest", label: "Azure 语音（REST 短音频）" },
] as const;
const NVIDIA_INTEGRATE_HOST = "integrate.api.nvidia.com";

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
    toast.error("加载失败");
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
  form.config = "{}";
  form.status = "active";
  form.sortOrder = 0;
  form.remark = "";
}

function onProtocolChange(v: string) {
  if (v === "azure_speech_rest" && (!form.modelCode.trim() || form.modelCode === "whisper-1")) {
    form.modelCode = "-";
  }
  if (v === "openai" && form.modelCode === "-") {
    form.modelCode = "whisper-1";
  }
}

watch(() => form.protocol, onProtocolChange);

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  form.modelCode = "whisper-1";
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

async function submit() {
  if (!form.code.trim() || !form.name.trim()) {
    toast.warning("请填写编码与名称");
    return;
  }
  if (form.protocol === "openai" && !form.modelCode.trim()) {
    toast.warning("OpenAI 兼容协议须填写模型 code（如 whisper-1）");
    return;
  }
  if (form.protocol === "azure_speech_rest" && !form.modelCode.trim()) {
    form.modelCode = "-";
  }
  const bu = form.baseUrl.trim().toLowerCase();
  if (form.protocol === "openai" && bu.includes(NVIDIA_INTEGRATE_HOST)) {
    toast.error(
      "STT 不能使用 NVIDIA integrate 网关地址：该地址不提供 OpenAI 兼容的 /v1/audio/transcriptions。请填写实际提供语音转写 API 的根地址。",
    );
    return;
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
  const body = {
    code: form.code.trim(),
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
    message: "确认删除该 STT 配置？",
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
  { value: "active", label: "active" },
  { value: "inactive", label: "inactive" },
];

const protocolOptions = PROTOCOLS.map((p) => ({ value: p.value, label: p.label }));

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
        title="STT 服务配置"
        description="数据表 sys_stt_service_configs。Go 读取 status = active 中 sort_order 最小的一条。OpenAI 兼容须实现 POST /v1/audio/transcriptions；Azure 使用认知服务短音频 REST，服务端需安装 ffmpeg。"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">新建</AdminButton>
        </template>
      </AdminPageHeader>

      <AdminAlert title="配置说明">
        OpenAI 兼容：Base URL 填根地址（不要带 /v1）。勿填
        <code>{{ NVIDIA_INTEGRATE_HOST }}</code>。Azure：在扩展 JSON 中配置 region（及可选 locale），API Key 须在后台填写。
      </AdminAlert>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>编码</AdminTh>
          <AdminTh>名称</AdminTh>
          <AdminTh width="100px">协议</AdminTh>
          <AdminTh>模型</AdminTh>
          <AdminTh>API Key</AdminTh>
          <AdminTh width="88px">状态</AdminTh>
          <AdminTh width="72px">排序</AdminTh>
          <AdminTh>更新时间</AdminTh>
          <AdminTh width="200px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap>{{ row.code }}</AdminTd>
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.protocol }}</AdminTd>
          <AdminTd>{{ row.modelCode }}</AdminTd>
          <AdminTd><AdminMaskedKey :value="row.apiKey as string | null" /></AdminTd>
          <AdminTd><AdminBadge>{{ row.status }}</AdminBadge></AdminTd>
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
      :title="dialogMode === 'create' ? '新建 STT 配置' : '编辑 STT 配置'"
      width="lg"
    >
      <AdminAlert title="语音转写 STT">
        <template v-if="form.protocol === 'openai'">
          Base URL 填<strong>根地址</strong>（服务端会请求 /v1/audio/transcriptions），末尾不要带 /v1。未填 Base URL
          时，Go 端默认使用 https://api.openai.com。
        </template>
        <template v-else>
          Azure 短音频识别：API Key 填 Speech 资源密钥；扩展 JSON 至少包含 "region":"eastasia"（与资源一致）。可选
          locale（如 zh-CN），不填则按会话目标语言映射。模型 code 可填 -。
        </template>
      </AdminAlert>
      <AdminFormField v-if="dialogMode === 'edit'" label="ID">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField label="编码" required>
        <AdminInput v-model="form.code" :disabled="dialogMode === 'edit'" placeholder="唯一标识" />
      </AdminFormField>
      <AdminFormField label="名称" required>
        <AdminInput v-model="form.name" />
      </AdminFormField>
      <AdminFormField label="协议">
        <AdminSelect v-model="form.protocol" :options="protocolOptions" />
      </AdminFormField>
      <AdminFormField v-if="form.protocol === 'openai'" label="Base URL">
        <AdminInput
          v-model="form.baseUrl"
          placeholder="https://api.openai.com 或其它提供转写 REST 的根地址"
        />
      </AdminFormField>
      <AdminFormField label="API Key">
        <AdminInput v-model="form.apiKey" type="password" />
      </AdminFormField>
      <AdminFormField
        :label="form.protocol === 'azure_speech_rest' ? '模型 code（可 -）' : '模型 code'"
        :required="form.protocol === 'openai'"
      >
        <AdminInput
          v-model="form.modelCode"
          :placeholder="
            form.protocol === 'azure_speech_rest' ? 'Azure 可填 -' : '如 whisper-1（按厂商文档填写）'
          "
        />
      </AdminFormField>
      <AdminFormField label="扩展 JSON">
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
