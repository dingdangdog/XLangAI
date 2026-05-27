<script setup lang="ts">
const { t } = useI18n();

const API = "/api/admin/object-storage-configs";

const api = useAdminResourceApi(API);

const toast = useToast();

const { confirm } = useConfirm();



const PROVIDERS = [
  { value: "local", labelKey: "pages.objectStorage.providerLocal" },
  { value: "cloudflare_r2", labelKey: "pages.objectStorage.providerR2" },
  { value: "qiniu", labelKey: "pages.objectStorage.providerQiniu" },
  { value: "aliyun_oss", labelKey: "pages.objectStorage.providerAliyunOss" },
] as const;



const CONFIG_HINTS: Record<string, string> = {

  local: "{}",

  cloudflare_r2: '{"path_style":true}',

  qiniu: '{"zone":"z0"}',

  aliyun_oss: "{}",

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

  provider: "local",

  baseUrl: "",

  publicBaseUrl: "",

  apiKey: "",

  secretKey: "",

  bucket: "",

  region: "",

  config: "{}",

  status: "inactive",

  sortOrder: 0,

  remark: "",

});



function resetForm() {

  form.id = "";

  form.code = "";

  form.name = "";

  form.provider = "local";

  form.baseUrl = "";

  form.publicBaseUrl = "";

  form.apiKey = "";

  form.secretKey = "";

  form.bucket = "";

  form.region = "";

  form.config = CONFIG_HINTS.local ?? "{}";

  form.status = "inactive";

  form.sortOrder = 0;

  form.remark = "";

}



function onProviderChange(v: string) {

  form.config = CONFIG_HINTS[v] ?? "{}";

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

  form.provider = String(row.provider ?? "local");

  form.baseUrl = String(row.baseUrl ?? "");

  form.publicBaseUrl = String(row.publicBaseUrl ?? "");

  form.apiKey = String(row.apiKey ?? "");

  form.secretKey = String(row.secretKey ?? "");

  form.bucket = String(row.bucket ?? "");

  form.region = String(row.region ?? "");

  form.config = row.config != null ? String(row.config) : "{}";

  form.status = String(row.status ?? "inactive");

  form.sortOrder = Number(row.sortOrder ?? 0);

  form.remark = String(row.remark ?? "");

  dialogVisible.value = true;

}



/** 库表 code 唯一、Go 不按 code 查；新建时自动生成。 */

function autoStorageCode(provider: string) {

  const p = (provider || "local").trim().replace(/[^a-zA-Z0-9_-]/g, "_");

  return `${p}-${Date.now()}`;

}



async function submit() {

  if (!form.name.trim()) {

    toast.warning(t("validation.fillName"));

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

        ? autoStorageCode(form.provider)

        : form.code.trim(),

    name: form.name.trim(),

    provider: form.provider,

    baseUrl: form.baseUrl.trim() || null,

    publicBaseUrl: form.publicBaseUrl.trim() || null,

    apiKey: form.apiKey.trim() || null,

    secretKey: form.secretKey.trim() || null,

    bucket: form.bucket.trim() || null,

    region: form.region.trim() || null,

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

    message: t("confirm.deleteObjectStorage"),

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



const isR2 = computed(() => form.provider === "cloudflare_r2");

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
      {{ $t("pages.objectStorage.configAnomaly", { count: activeCount }) }}
    </AdminAlert>

    <AdminPanel>

      <AdminTable :loading="loading">

        <template #head>

          <AdminTh>{{ $t("common.name") }}</AdminTh>

          <AdminTh width="140px">{{ $t("common.provider") }}</AdminTh>

          <AdminTh>{{ $t("fields.bucket") }}</AdminTh>

          <AdminTh>{{ $t("fields.publicUrl") }}</AdminTh>

          <AdminTh width="88px">{{ $t("common.status") }}</AdminTh>

          <AdminTh width="72px">{{ $t("common.sort") }}</AdminTh>

          <AdminTh>{{ $t("common.updatedAt") }}</AdminTh>

          <AdminTh width="200px" align="right">{{ $t("common.actions") }}</AdminTh>

        </template>

        <AdminTr v-for="row in list" :key="String(row.id)">

          <AdminTd>{{ row.name }}</AdminTd>

          <AdminTd>{{ row.provider }}</AdminTd>

          <AdminTd>{{ row.bucket }}</AdminTd>

          <AdminTd class="max-w-[200px] truncate">{{ row.publicBaseUrl }}</AdminTd>

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

      :title="dialogMode === 'create' ? $t('pages.objectStorage.createDialog') : $t('pages.objectStorage.editDialog')"

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



      <template v-if="isR2">

        <AdminFormField :label="$t('fields.accessKeyId')" required>

          <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />

        </AdminFormField>

        <AdminFormField :label="$t('fields.secretAccessKey')" required>

          <AdminInput v-model="form.secretKey" type="password" autocomplete="off" />

        </AdminFormField>

        <AdminFormField :label="$t('fields.s3Endpoint')" required>

          <AdminInput v-model="form.baseUrl" />

        </AdminFormField>

        <AdminFormField :label="$t('fields.bucketName')" required>

          <AdminInput v-model="form.bucket" />

        </AdminFormField>

        <AdminFormField :label="$t('fields.publicAccessUrl')" required>

          <AdminInput v-model="form.publicBaseUrl" />

        </AdminFormField>

      </template>



      <template v-else>

        <AdminFormField :label="$t('fields.endpointBaseUrl')">

          <AdminInput v-model="form.baseUrl" />

        </AdminFormField>

        <AdminFormField :label="$t('fields.publicDomain')">

          <AdminInput v-model="form.publicBaseUrl" />

        </AdminFormField>

        <AdminFormField label="Access Key / API Key">

          <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />

        </AdminFormField>

        <AdminFormField label="Secret Key">

          <AdminInput v-model="form.secretKey" type="password" autocomplete="off" />

        </AdminFormField>

        <AdminFormField :label="$t('fields.bucket')">

          <AdminInput v-model="form.bucket" />

        </AdminFormField>

        <AdminFormField :label="$t('common.region')">

          <AdminInput v-model="form.region" />

        </AdminFormField>

      </template>



      <AdminFormField :label="$t('fields.extJson')">

        <AdminInput v-model="form.config" type="textarea" :rows="3" class="font-mono text-sm" />

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


