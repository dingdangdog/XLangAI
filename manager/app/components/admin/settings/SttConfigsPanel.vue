<script setup lang="ts">
const { t } = useI18n();
const API = "/api/admin/stt-service-configs";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

const PROTOCOLS = [
  { value: "openai", labelKey: "pages.sttConfigs.protocolOpenai" },
  { value: "azure_speech_rest", labelKey: "pages.sttConfigs.protocolAzure" },
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

/** 库表 code 唯一；Go 取 sort_order 最小的 active，不按 code 查。新建时自动生成。 */
function autoSttCode(protocol: string) {
  const p = (protocol || "openai").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
}

async function submit() {
  if (!form.name.trim()) {
    toast.warning(t("validation.fillName"));
    return;
  }
  if (form.protocol === "openai" && !form.modelCode.trim()) {
    toast.warning(t("validation.sttModelRequired"));
    return;
  }
  if (form.protocol === "azure_speech_rest" && !form.modelCode.trim()) {
    form.modelCode = "-";
  }
  const bu = form.baseUrl.trim().toLowerCase();
  if (form.protocol === "openai" && bu.includes(NVIDIA_INTEGRATE_HOST)) {
    toast.error(t("pages.sttConfigs.nvidiaWarning"));
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
        ? autoSttCode(form.protocol)
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
    message: t("confirm.deleteSttConfig"),
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

const protocolOptions = computed(() =>
  PROTOCOLS.map((p) => ({ value: p.value, label: t(p.labelKey) })),
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

    <AdminAlert :title="$t('pages.sttConfigs.configAlertTitle')">
      {{ $t("pages.sttConfigs.configAlert") }}
      <code>{{ NVIDIA_INTEGRATE_HOST }}</code>
    </AdminAlert>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("common.code") }}</AdminTh>
          <AdminTh>{{ $t("common.name") }}</AdminTh>
          <AdminTh width="100px">{{ $t("fields.protocol") }}</AdminTh>
          <AdminTh>{{ $t("common.model") }}</AdminTh>
          <AdminTh>{{ $t("fields.apiKey") }}</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
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
      :title="dialogMode === 'create' ? $t('pages.sttConfigs.createDialog') : $t('pages.sttConfigs.editDialog')"
      width="lg"
    >
      <AdminAlert :title="$t('pages.sttConfigs.dialogAlertTitle')">
        <template v-if="form.protocol === 'openai'">
          {{ $t("pages.sttConfigs.dialogAlertOpenai") }}
        </template>
        <template v-else>
          {{ $t("pages.sttConfigs.dialogAlertAzure") }}
        </template>
      </AdminAlert>
      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField :label="$t('common.name')" required>
        <AdminInput v-model="form.name" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.protocol')">
        <AdminSelect v-model="form.protocol" :options="protocolOptions" />
      </AdminFormField>
      <AdminFormField v-if="form.protocol === 'openai'" :label="$t('fields.baseUrl')">
        <AdminInput
          v-model="form.baseUrl"
          :placeholder="$t('pages.sttConfigs.baseUrlPlaceholder')"
        />
      </AdminFormField>
      <AdminFormField :label="$t('fields.apiKey')">
        <AdminInput v-model="form.apiKey" type="password" />
      </AdminFormField>
      <AdminFormField
        :label="form.protocol === 'azure_speech_rest' ? $t('pages.sttConfigs.modelAzure') : $t('pages.sttConfigs.modelOpenai')"
        :required="form.protocol === 'openai'"
      >
        <AdminInput
          v-model="form.modelCode"
          :placeholder="
            form.protocol === 'azure_speech_rest'
              ? $t('pages.sttConfigs.modelPlaceholderAzure')
              : $t('pages.sttConfigs.modelPlaceholderOpenai')
          "
        />
      </AdminFormField>
      <AdminFormField :label="$t('fields.extJson')">
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
