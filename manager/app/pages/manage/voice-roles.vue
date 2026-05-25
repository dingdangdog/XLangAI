<script setup lang="ts">
import { stripVoiceRoleVirtualFields } from "~/utils/voiceRoleUi";

const API = "/api/admin/voice-roles";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

type LangOpt = { id: string; code: string; name: string };
type TtsOpt = { id: string; code: string; name: string; provider: string };

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);

const langOptions = ref<LangOpt[]>([]);
const ttsOptions = ref<TtsOpt[]>([]);
const optionsLoading = ref(false);

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

async function loadOptions() {
  optionsLoading.value = true;
  try {
    const [lr, tr] = await Promise.all([
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/languages", {
        query: { page: 1, pageSize: 500 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/tts-service-configs", {
        query: { page: 1, pageSize: 500 },
      }),
    ]);
    langOptions.value = lr.items
      .filter((r) => String(r.status ?? "active") === "active")
      .map((r) => ({
        id: String(r.id),
        code: String(r.code ?? ""),
        name: String(r.name ?? ""),
      }));
    ttsOptions.value = tr.items
      .filter((r) => String(r.status ?? "active") === "active")
      .map((r) => ({
        id: String(r.id),
        code: String(r.code ?? ""),
        name: String(r.name ?? ""),
        provider: String(r.provider ?? ""),
      }));
  } catch (e) {
    toast.error("加载下拉选项失败");
    console.error(e);
  } finally {
    optionsLoading.value = false;
  }
}

watch([page, pageSize], () => void load(), { immediate: true });

const VOICE_ROLE_TABLE_COLUMNS: { prop: string; label: string }[] = [
  { prop: "name", label: "角色名称" },
  { prop: "voiceCode", label: "音色代码" },
  { prop: "languageLabel", label: "语言" },
  { prop: "ttsConfigLabel", label: "TTS 配置" },
  { prop: "gender", label: "性别" },
  { prop: "previewAudioUrl", label: "试听" },
  { prop: "status", label: "状态" },
  { prop: "sortOrder", label: "排序" },
  { prop: "remark", label: "备注" },
  { prop: "createdAt", label: "创建时间" },
  { prop: "updatedAt", label: "更新时间" },
];

const tableColumns = computed(() => {
  if (!list.value.length) {
    return VOICE_ROLE_TABLE_COLUMNS;
  }
  const keys = new Set(Object.keys(list.value[0] as object));
  return VOICE_ROLE_TABLE_COLUMNS.filter((c) => keys.has(c.prop));
});

function cellValue(row: Record<string, unknown>, key: string) {
  const v = row[key];
  if (v === null || v === undefined) return "";
  if (isDateTimeField(key)) return formatDateTime(v);
  if (typeof v === "object") return JSON.stringify(v);
  return String(v);
}

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

const voiceForm = reactive({
  id: "",
  languageId: "",
  ttsServiceConfigId: "",
  voiceCode: "",
  name: "",
  gender: "female",
  sortOrder: 0,
  status: "active",
  remark: "",
  config: "",
});

const selectedTts = computed(() =>
  ttsOptions.value.find((t) => t.id === voiceForm.ttsServiceConfigId),
);

const voiceCodeHint = computed(() => {
  const p = (selectedTts.value?.provider ?? "").trim();
  if (p === "azure_speech_rest") {
    return "Azure 神经语音短名，例如 zh-CN-XiaoxiaoNeural、en-US-JennyNeural；须与语音资源区域能力一致。";
  }
  if (p === "openai_rest") {
    return "OpenAI TTS 的 voice 名称，如 alloy、nova、echo 等（以当前 API 文档为准）。";
  }
  if (!p) {
    return "未选择 TTS 配置时无法判断协议；请先选择配置。空 provider 时服务端可能按 OpenAI 处理。";
  }
  return `当前 provider 为「${p}」，音色代码须与服务端对该协议的约定一致。`;
});

function resetVoiceForm() {
  voiceForm.id = "";
  voiceForm.languageId = "";
  voiceForm.ttsServiceConfigId = "";
  voiceForm.voiceCode = "";
  voiceForm.name = "";
  voiceForm.gender = "female";
  voiceForm.sortOrder = 0;
  voiceForm.status = "active";
  voiceForm.remark = "";
  voiceForm.config = "";
}

function openCreate() {
  dialogMode.value = "create";
  resetVoiceForm();
  dialogVisible.value = true;
  void loadOptions();
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  const r = stripVoiceRoleVirtualFields(row);
  voiceForm.id = String(r.id ?? "");
  voiceForm.languageId = String(r.languageId ?? "");
  voiceForm.ttsServiceConfigId = String(r.ttsServiceConfigId ?? "");
  voiceForm.voiceCode = String(r.voiceCode ?? "");
  voiceForm.name = String(r.name ?? "");
  voiceForm.gender = String(r.gender ?? "female") || "female";
  voiceForm.sortOrder = Number(r.sortOrder ?? 0);
  voiceForm.status = String(r.status ?? "active");
  voiceForm.remark = String(r.remark ?? "");
  voiceForm.config = r.config != null ? String(r.config) : "";
  dialogVisible.value = true;
  void loadOptions();
}

function buildPayload(): Record<string, unknown> {
  return {
    languageId: voiceForm.languageId || null,
    ttsServiceConfigId: voiceForm.ttsServiceConfigId || null,
    voiceCode: voiceForm.voiceCode.trim(),
    name: voiceForm.name.trim(),
    gender: voiceForm.gender || null,
    sortOrder: voiceForm.sortOrder,
    status: voiceForm.status,
    remark: voiceForm.remark.trim() || null,
    config: voiceForm.config.trim() || null,
  };
}

async function submitVoice() {
  if (!voiceForm.languageId) {
    toast.warning("请选择语言");
    return;
  }
  if (!voiceForm.ttsServiceConfigId) {
    toast.warning("请选择 TTS 服务配置（决定厂商与协议）");
    return;
  }
  if (!voiceForm.voiceCode.trim()) {
    toast.warning("请填写音色代码");
    return;
  }
  if (!voiceForm.name.trim()) {
    toast.warning("请填写显示名称");
    return;
  }
  const payload = buildPayload();
  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await $fetch("/api/admin/voice-roles", { method: "POST", body: payload });
      toast.success("已创建");
    } else {
      await $fetch(`/api/admin/voice-roles/${voiceForm.id}`, {
        method: "PUT",
        body: { id: voiceForm.id, ...payload },
      });
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
    message: "确认删除该语音角色？",
    danger: true,
    confirmLabel: "删除",
  });
  if (!ok) return;
  try {
    await $fetch(`/api/admin/voice-roles/${row.id as string}`, { method: "DELETE" });
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

const langSelectOptions = computed(() =>
  langOptions.value.map((l) => ({ value: l.id, label: `${l.code} · ${l.name}` })),
);

const ttsSelectOptions = computed(() =>
  ttsOptions.value.map((t) => ({ value: t.id, label: `${t.name} (${t.provider})` })),
);

const genderOptions = [
  { value: "", label: "（不选）" },
  { value: "female", label: "女" },
  { value: "male", label: "男" },
];

const { activateRow, activatingId } = useActivateConfigRow({
  api,
  getList: () => list.value,
  exclusivity: "multi-active",
  reload: load,
});

const previewGeneratingId = ref<string | null>(null);
const previewAudioEl = ref<HTMLAudioElement | null>(null);

function previewPlayUrl(row: Record<string, unknown>): string | null {
  const raw = String(row.previewAudioUrl ?? "").trim();
  if (!raw) return null;
  if (raw.startsWith("http://") || raw.startsWith("https://")) return raw;

  const local = String(row.previewLocalFilename ?? "").trim();
  if (local) {
    return `/api/admin/preview-audio/${encodeURIComponent(local)}`;
  }

  const match = raw.match(/\/api\/v1\/audio\/([^/?#]+)/);
  if (match?.[1]) {
    return `/api/admin/preview-audio/${encodeURIComponent(match[1])}`;
  }

  return null;
}

function previewAudioLabel(row: Record<string, unknown>): string {
  const local = String(row.previewLocalFilename ?? "").trim();
  if (local) return local;
  const raw = String(row.previewAudioUrl ?? "").trim();
  if (!raw) return "";
  return raw.split("/").filter(Boolean).pop() ?? raw;
}

function stopPreviewPlayback() {
  const el = previewAudioEl.value;
  if (el) {
    el.pause();
    el.currentTime = 0;
  }
}

function playPreview(row: Record<string, unknown>) {
  const url = previewPlayUrl(row);
  if (!url) {
    toast.warning("尚未生成试听，请先点击「生成试听」");
    return;
  }
  let el = previewAudioEl.value;
  if (!el) {
    el = new Audio();
    previewAudioEl.value = el;
  }
  el.src = url;
  void el.play().catch((e) => {
    toast.error("播放失败");
    console.error(e);
  });
}

async function generatePreview(row: Record<string, unknown>) {
  const id = String(row.id ?? "");
  if (!id) return;
  previewGeneratingId.value = id;
  try {
    await $fetch(`/api/admin/voice-roles/${id}/generate-preview`, { method: "POST" });
    toast.success("试听已生成");
    await load();
  } catch (e) {
    toast.error("生成试听失败");
    console.error(e);
  } finally {
    previewGeneratingId.value = null;
  }
}
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader title="语音角色">
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">新建</AdminButton>
        </template>
      </AdminPageHeader>

      <AdminAlert title="使用说明">
        试听在本管理端生成：语言页配置试听文案模板 → 本页「生成试听」→ 按 media.assistant_tts.storage 写库与存储。Go API 只把 preview_audio_url 返回给客户端播放。
      </AdminAlert>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh v-for="col in tableColumns" :key="col.prop">{{ col.label }}</AdminTh>
          <AdminTh width="200px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd v-for="col in tableColumns" :key="col.prop">
            <template v-if="col.prop === 'previewAudioUrl'">
              <span v-if="row.previewAudioUrl" class="inline-flex flex-col gap-0.5">
                <span class="text-success-600">已生成</span>
                <span class="max-w-[220px] truncate text-xs text-surface-500">
                  {{ previewAudioLabel(row) }}
                </span>
              </span>
              <span v-else class="text-surface-500">未生成</span>
            </template>
            <template v-else>
              {{ cellValue(row, col.prop) }}
            </template>
          </AdminTd>
          <AdminTd align="right" class="whitespace-nowrap">
            <AdminButton v-if="String(row.status) !== 'active'" variant="link"
              :loading="activatingId === String(row.id)" @click="activateRow(row)">
              启用
            </AdminButton>
            <AdminButton variant="link" :disabled="!row.previewAudioUrl" @click="playPreview(row)">
              播放
            </AdminButton>
            <AdminButton variant="link" :loading="previewGeneratingId === String(row.id)" @click="generatePreview(row)">
              生成试听
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

    <AdminDialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新建语音角色' : '编辑语音角色'" width="lg">
      <AdminSkeleton v-if="optionsLoading" :rows="6" />
      <template v-else>
        <AdminFormField v-if="dialogMode === 'edit'" label="id">
          <AdminInput v-model="voiceForm.id" disabled />
        </AdminFormField>
        <AdminFormField label="语言" required>
          <AdminSelect v-model="voiceForm.languageId" :options="langSelectOptions" placeholder="选择语言" />
        </AdminFormField>
        <AdminFormField label="TTS 配置" required>
          <AdminSelect v-model="voiceForm.ttsServiceConfigId" :options="ttsSelectOptions"
            placeholder="选择 TTS 服务配置（含厂商）" />
        </AdminFormField>
        <AdminFormField label="音色代码" required :hint="voiceCodeHint">
          <AdminInput v-model="voiceForm.voiceCode" placeholder="与所选厂商协议一致" />
        </AdminFormField>
        <AdminFormField label="角色名称" required>
          <AdminInput v-model="voiceForm.name" placeholder="列表与客户端展示用" />
        </AdminFormField>
        <AdminFormField label="性别">
          <AdminSelect v-model="voiceForm.gender" :options="genderOptions" placeholder="可选" />
        </AdminFormField>
        <AdminFormField label="排序">
          <AdminInput v-model="voiceForm.sortOrder" type="number" />
        </AdminFormField>
        <AdminFormField label="状态">
          <AdminSelect v-model="voiceForm.status" :options="statusOptions" />
        </AdminFormField>
        <AdminFormField label="备注">
          <AdminInput v-model="voiceForm.remark" type="textarea" :rows="2" />
        </AdminFormField>
        <AdminFormField label="扩展 config" hint="可选 JSON 字符串">
          <AdminInput v-model="voiceForm.config" type="textarea" :rows="2" />
        </AdminFormField>
      </template>
      <template #footer>
        <AdminButton @click="dialogVisible = false">取消</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submitVoice">保存</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
