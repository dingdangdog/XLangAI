<script setup lang="ts">
const { t } = useI18n();

const API = "/api/admin/sms-service-configs";

const api = useAdminResourceApi(API);

const toast = useToast();

const { confirm } = useConfirm();

const PROVIDERS = [
  { value: "aliyun", labelKey: "pages.smsService.providerAliyun" },
  { value: "tencent", labelKey: "pages.smsService.providerTencent" },
] as const;

const CONFIG_HINTS: Record<string, string> = {
  aliyun: '{"endpoint":"dysmsapi.aliyuncs.com","template_param_key":"code"}',
  tencent: '{"sdk_app_id":"","endpoint":"sms.tencentcloudapi.com"}',
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
  provider: "aliyun",
  apiKey: "",
  secretKey: "",
  region: "cn-hangzhou",
  signName: "",
  templateCode: "",
  config: CONFIG_HINTS.aliyun,
  status: "inactive",
  sortOrder: 0,
  remark: "",
});

function resetForm() {
  form.id = "";
  form.code = "";
  form.name = "";
  form.provider = "aliyun";
  form.apiKey = "";
  form.secretKey = "";
  form.region = "cn-hangzhou";
  form.signName = "";
  form.templateCode = "";
  form.config = CONFIG_HINTS.aliyun;
  form.status = "inactive";
  form.sortOrder = 0;
  form.remark = "";
}

function onProviderChange(v: string) {
  form.config = CONFIG_HINTS[v] ?? "{}";
  if (v === "tencent" && (!form.region || form.region === "cn-hangzhou")) {
    form.region = "ap-guangzhou";
  }
  if (v === "aliyun" && form.region === "ap-guangzhou") {
    form.region = "cn-hangzhou";
  }
}

watch(() => form.provider, onProviderChange);

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  form.code = String(row.code ?? "");
  form.name = String(row.name ?? "");
  form.provider = String(row.provider ?? "aliyun");
  form.apiKey = String(row.apiKey ?? "");
  form.secretKey = String(row.secretKey ?? "");
  form.region = String(row.region ?? "");
  form.signName = String(row.signName ?? "");
  form.templateCode = String(row.templateCode ?? "");
  form.config = row.config != null ? String(row.config) : "{}";
  form.status = String(row.status ?? "inactive");
  form.sortOrder = Number(row.sortOrder ?? 0);
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
}

function autoSmsCode(provider: string) {
  const p = (provider || "aliyun").trim().replace(/[^a-zA-Z0-9_-]/g, "_");
  return `${p}-${Date.now()}`;
}

async function submit() {
  if (!form.name.trim()) {
    toast.warning(t("validation.fillName"));
    return;
  }
  if (!form.signName.trim() || !form.templateCode.trim()) {
    toast.warning(t("pages.smsService.fillSignAndTemplate"));
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
    code: dialogMode.value === "create" ? autoSmsCode(form.provider) : form.code.trim(),
    name: form.name.trim(),
    provider: form.provider,
    apiKey: form.apiKey.trim() || null,
    secretKey: form.secretKey.trim() || null,
    region: form.region.trim() || null,
    signName: form.signName.trim() || null,
    templateCode: form.templateCode.trim() || null,
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
    message: t("confirm.deleteSmsConfig"),
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

const providerOptions = computed(() =>
  PROVIDERS.map((p) => ({ value: p.value, label: t(p.labelKey) })),
);

const activeCount = computed(
  () => list.value.filter((r) => String(r.status) === "active").length,
);

const isTencent = computed(() => form.provider === "tencent");

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

    <AdminAlert v-if="activeCount > 1" :title="$t('pages.smsService.configAnomalyTitle')" variant="warning">
      {{ $t("pages.smsService.configAnomaly", { count: activeCount }) }}
    </AdminAlert>

    <AdminAlert :title="$t('pages.smsService.hintTitle')" variant="info">
      {{ $t("pages.smsService.hintBody") }}
    </AdminAlert>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("common.name") }}</AdminTh>
          <AdminTh width="120px">{{ $t("common.provider") }}</AdminTh>
          <AdminTh>{{ $t("pages.smsService.signName") }}</AdminTh>
          <AdminTh>{{ $t("pages.smsService.templateCode") }}</AdminTh>
          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>
          <AdminTh width="72px">{{ $t("common.sort") }}</AdminTh>
          <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>
          <AdminTh width="200px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd>{{ row.name }}</AdminTd>
          <AdminTd>{{ row.provider }}</AdminTd>
          <AdminTd>{{ row.signName }}</AdminTd>
          <AdminTd class="max-w-[160px] truncate font-mono text-sm">{{ row.templateCode }}</AdminTd>
          <AdminTd>
            <AdminBadge :variant="row.status === 'active' ? 'success' : 'default'">
              {{ row.status }}
            </AdminBadge>
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
      :title="dialogMode === 'create' ? $t('pages.smsService.createDialog') : $t('pages.smsService.editDialog')"
      width="lg"
    >
      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField :label="$t('common.name')" required>
        <AdminInput v-model="form.name" />
      </AdminFormField>
      <AdminFormField :label="$t('common.provider')">
        <AdminSelect v-model="form.provider" :options="providerOptions" />
      </AdminFormField>
      <AdminFormField :label="$t('pages.smsService.accessKey')" required>
        <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />
      </AdminFormField>
      <AdminFormField :label="$t('pages.smsService.secretKey')" required>
        <AdminInput v-model="form.secretKey" type="password" autocomplete="off" />
      </AdminFormField>
      <AdminFormField :label="$t('common.region')">
        <AdminInput v-model="form.region" :placeholder="isTencent ? 'ap-guangzhou' : 'cn-hangzhou'" />
      </AdminFormField>
      <AdminFormField :label="$t('pages.smsService.signName')" required>
        <AdminInput v-model="form.signName" />
      </AdminFormField>
      <AdminFormField :label="$t('pages.smsService.templateCode')" required>
        <AdminInput v-model="form.templateCode" class="font-mono text-sm" />
      </AdminFormField>
      <AdminFormField :label="$t('fields.extJson')">
        <AdminInput v-model="form.config" type="textarea" :rows="3" class="font-mono text-sm" />
        <p v-if="isTencent" class="mt-1 text-xs text-gray-500">
          {{ $t("pages.smsService.tencentConfigHint") }}
        </p>
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
