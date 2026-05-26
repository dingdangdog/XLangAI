<script setup lang="ts">
const { t } = useI18n();

const API = "/api/admin/object-storage-configs";

const api = useAdminResourceApi(API);

const toast = useToast();

const { confirm } = useConfirm();



const PROVIDERS = [

  { value: "local", label: "本地目录（服务器磁盘）" },

  { value: "cloudflare_r2", label: "Cloudflare R2（S3 兼容）" },

  { value: "qiniu", label: "七牛云对象存储" },

  { value: "aliyun_oss", label: "阿里云 OSS" },

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

    message: "确认删除该对象存储配置？",

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



const providerOptions = PROVIDERS.map((p) => ({ value: p.value, label: p.label }));



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
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        title="对象存储 / 图床"
        description="全局仅一条 active；Go 只读启用中的那条，不按编码查找。"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
        </template>
      </AdminPageHeader>

      <AdminAlert v-if="activeCount > 1" title="配置异常" variant="warning">
        当前列表中有 {{ activeCount }} 条 active 记录，应仅保留一条。
      </AdminAlert>
    </template>

    <AdminPanel>

      <AdminTable :loading="loading">

        <template #head>

          <AdminTh>{{ $t("common.name") }}</AdminTh>

          <AdminTh width="140px">Provider</AdminTh>

          <AdminTh>Bucket</AdminTh>

          <AdminTh>公网 URL</AdminTh>

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

      :title="dialogMode === 'create' ? '新建对象存储配置' : '编辑对象存储配置'"

      width="lg"

    >

      <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">

        <AdminInput v-model="form.id" disabled />

      </AdminFormField>

      <AdminFormField :label="$t('common.name')" required>

        <AdminInput v-model="form.name" />

      </AdminFormField>

      <AdminFormField label="Provider">

        <AdminSelect v-model="form.provider" :options="providerOptions" />

      </AdminFormField>



      <template v-if="isR2">

        <AdminFormField label="访问密钥 ID" required>

          <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />

        </AdminFormField>

        <AdminFormField label="机密访问密钥" required>

          <AdminInput v-model="form.secretKey" type="password" autocomplete="off" />

        </AdminFormField>

        <AdminFormField label="S3 API 终结点" required>

          <AdminInput v-model="form.baseUrl" />

        </AdminFormField>

        <AdminFormField label="存储桶名称" required>

          <AdminInput v-model="form.bucket" />

        </AdminFormField>

        <AdminFormField label="公共访问 URL" required>

          <AdminInput v-model="form.publicBaseUrl" />

        </AdminFormField>

      </template>



      <template v-else>

        <AdminFormField label="Endpoint / Base URL">

          <AdminInput v-model="form.baseUrl" />

        </AdminFormField>

        <AdminFormField label="公网访问域名">

          <AdminInput v-model="form.publicBaseUrl" />

        </AdminFormField>

        <AdminFormField label="Access Key / API Key">

          <AdminInput v-model="form.apiKey" type="password" autocomplete="off" />

        </AdminFormField>

        <AdminFormField label="Secret Key">

          <AdminInput v-model="form.secretKey" type="password" autocomplete="off" />

        </AdminFormField>

        <AdminFormField label="Bucket">

          <AdminInput v-model="form.bucket" />

        </AdminFormField>

        <AdminFormField label="Region">

          <AdminInput v-model="form.region" />

        </AdminFormField>

      </template>



      <AdminFormField label="扩展 JSON">

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
  </AdminListPage>
</template>


