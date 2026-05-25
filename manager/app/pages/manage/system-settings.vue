<script setup lang="ts">
const API = "/api/admin/system-settings";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

const KEY_LABELS: Record<string, string> = {
  "auth.password.enabled": "账号密码登录",
  "auth.password.register_enabled": "账号密码注册",
  "auth.sms.enabled": "短信验证码登录",
  "auth.sms.register_enabled": "短信验证码注册",
  "auth.google.enabled": "Google 登录",
  "auth.google.register_enabled": "Google 自动注册",
  "auth.apple.enabled": "Apple 登录",
  "auth.apple.register_enabled": "Apple 自动注册",
  "media.user_recording.storage": "用户录音存储",
  "media.assistant_tts.storage": "AI 回复音频存储",
  "media.avatar.storage": "头像存储",
  "official_server_store.config": "官网服务器商店配置",
};

const STORAGE_OPTIONS = [
  { value: "client", label: "仅客户端" },
  { value: "server", label: "服务器本地" },
  { value: "cloud", label: "云端对象存储" },
];
const STORAGE_OPTIONS_NO_CLIENT = STORAGE_OPTIONS.filter((o) => o.value !== "client");

const page = ref(1);
const pageSize = ref(50);
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
  key: "",
  value: "",
  valueType: "bool",
  description: "",
});

function labelForKey(k: string) {
  return KEY_LABELS[k] ?? k;
}

function storageOptionsForKey(k: string) {
  if (k === "media.user_recording.storage") return STORAGE_OPTIONS;
  return STORAGE_OPTIONS_NO_CLIENT;
}

function resetForm() {
  form.id = "";
  form.key = "";
  form.value = "true";
  form.valueType = "bool";
  form.description = "";
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  form.key = String(row.key ?? "");
  form.value = String(row.value ?? "");
  form.valueType = String(row.valueType ?? "string");
  form.description = String(row.description ?? "");
  dialogVisible.value = true;
}

function payload() {
  return {
    key: form.key.trim(),
    value: String(form.value).trim(),
    valueType: form.valueType,
    description: form.description.trim() || null,
  };
}

async function submit() {
  if (dialogMode.value === "create" && !form.key.trim()) {
    toast.warning("请填写 key");
    return;
  }
  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await api.create(payload());
      toast.success("已创建");
    } else {
      await api.update(form.id, { id: form.id, ...payload(), key: form.key });
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
    message: `确认删除系统变量 ${String(row.key ?? "")}？`,
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

const valueTypeOptions = [
  { value: "bool", label: "bool" },
  { value: "string", label: "string" },
  { value: "json", label: "json" },
];

const boolValueOptions = [
  { value: "true", label: "true（开启）" },
  { value: "false", label: "false（关闭）" },
];
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        title="系统变量"
        description="登录开关、媒体存储策略等 KEY-VALUE。与 LLM / 对象存储等厂商密钥表无关。"
      >
        <template #actions>
          <AdminButton variant="primary" @click="openCreate">新建</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>Key</AdminTh>
          <AdminTh>说明</AdminTh>
          <AdminTh>值</AdminTh>
          <AdminTh width="72px">类型</AdminTh>
          <AdminTh>备注</AdminTh>
          <AdminTh>更新时间</AdminTh>
          <AdminTh width="140px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd nowrap><code class="text-xs">{{ row.key }}</code></AdminTd>
          <AdminTd>{{ labelForKey(String(row.key ?? '')) }}</AdminTd>
          <AdminTd>
            <span v-if="row.valueType === 'bool'">{{ row.value === 'true' ? '开启' : '关闭' }}</span>
            <span v-else>{{ row.value }}</span>
          </AdminTd>
          <AdminTd>{{ row.valueType }}</AdminTd>
          <AdminTd>{{ row.description ?? "—" }}</AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.updatedAt) }}</AdminTd>
          <AdminTd align="right">
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
      :title="dialogMode === 'create' ? '新建系统变量' : '编辑系统变量'"
    >
      <AdminFormField label="Key" required>
        <AdminInput
          v-model="form.key"
          :disabled="dialogMode === 'edit'"
          placeholder="如 auth.sms.enabled"
        />
      </AdminFormField>
      <AdminFormField label="类型">
        <AdminSelect
          v-model="form.valueType"
          :options="valueTypeOptions"
          :disabled="dialogMode === 'edit'"
        />
      </AdminFormField>
      <AdminFormField v-if="form.valueType === 'bool'" label="值">
        <AdminSelect v-model="form.value" :options="boolValueOptions" />
      </AdminFormField>
      <AdminFormField v-else-if="form.key.startsWith('media.')" label="值">
        <AdminSelect
          v-model="form.value"
          :options="storageOptionsForKey(form.key)"
        />
      </AdminFormField>
      <AdminFormField v-else label="值">
        <AdminInput v-model="form.value" />
      </AdminFormField>
      <AdminFormField label="备注">
        <AdminInput v-model="form.description" type="textarea" :rows="2" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">取消</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">保存</AdminButton>
      </template>
    </AdminDialog>
  </AdminListPage>
</template>
