<script setup lang="ts">
import { formatAudioBytes } from "~/utils/usageDisplay";

const props = defineProps<{
  userId: string;
  userLabel?: string;
}>();

const API = "/api/admin/user-usage";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

const page = ref(1);
const pageSize = ref(10);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);

async function load() {
  if (!props.userId) return;
  loading.value = true;
  try {
    const res = await api.list({
      page: page.value,
      pageSize: pageSize.value,
      userId: props.userId,
    });
    list.value = res.items;
    total.value = res.total;
  } catch (e) {
    toast.error("加载用量失败");
    console.error(e);
  } finally {
    loading.value = false;
  }
}

watch(
  () => props.userId,
  () => {
    page.value = 1;
    void load();
  },
  { immediate: true },
);
watch([page, pageSize], () => void load());

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

const form = reactive({
  id: "",
  date: "",
  usageCount: 0,
  tokenCount: 0,
});

function resetForm() {
  form.id = "";
  form.date = new Date().toISOString().slice(0, 10);
  form.usageCount = 0;
  form.tokenCount = 0;
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  const d = row.date;
  form.date =
    typeof d === "string"
      ? d.slice(0, 10)
      : d instanceof Date
        ? d.toISOString().slice(0, 10)
        : String(d ?? "").slice(0, 10);
  form.usageCount = Number(row.usageCount ?? 0);
  form.tokenCount = Number(row.tokenCount ?? 0);
  dialogVisible.value = true;
}

async function submit() {
  if (!form.date) {
    toast.warning("请选择日期");
    return;
  }
  const body = {
    userId: props.userId,
    date: new Date(`${form.date}T12:00:00.000Z`).toISOString(),
    usageCount: form.usageCount,
    tokenCount: form.tokenCount,
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
    emit("changed");
  } catch (e) {
    toast.error("保存失败");
    console.error(e);
  } finally {
    saving.value = false;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: "确认删除该用量记录？",
    danger: true,
    confirmLabel: "删除",
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success("已删除");
    await load();
    emit("changed");
  } catch (e) {
    toast.error("删除失败");
    console.error(e);
  }
}

const emit = defineEmits<{ changed: [] }>();
</script>

<template>
  <div>
    <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
      <p class="text-sm text-muted">
        用户
        <span class="font-medium text-foreground">{{ userLabel || userId }}</span>
        · 按 UTC 自然日聚合：对话轮次、LLM Token、翻译/TTS 次数与字符、STT 音频量
      </p>
      <AdminButton variant="primary" size="sm" @click="openCreate">新建记录</AdminButton>
    </div>

    <AdminTable :loading="loading">
      <template #head>
        <AdminTh width="120px">日期</AdminTh>
        <AdminTh width="80px">调用</AdminTh>
        <AdminTh width="80px">Token</AdminTh>
        <AdminTh width="100px">翻译</AdminTh>
        <AdminTh width="100px">TTS</AdminTh>
        <AdminTh width="100px">STT</AdminTh>
        <AdminTh>更新时间</AdminTh>
        <AdminTh width="140px" align="right">操作</AdminTh>
      </template>
      <AdminTr v-for="row in list" :key="String(row.id)">
        <AdminTd>{{ formatDate(row.date) }}</AdminTd>
        <AdminTd>{{ row.usageCount }}</AdminTd>
        <AdminTd>{{ row.tokenCount }}</AdminTd>
        <AdminTd class="text-xs">
          {{ row.translateCount ?? 0 }} 次 · {{ row.translateChars ?? 0 }} 字
        </AdminTd>
        <AdminTd class="text-xs">
          {{ row.ttsCount ?? 0 }} 次 · {{ row.ttsChars ?? 0 }} 字
        </AdminTd>
        <AdminTd class="text-xs">
          {{ row.sttCount ?? 0 }} 次 · {{ formatAudioBytes(String(row.sttAudioBytes ?? 0)) }}
        </AdminTd>
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

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建用量记录' : '编辑用量记录'"
    >
      <AdminFormField v-if="dialogMode === 'edit'" label="ID">
        <AdminInput v-model="form.id" disabled />
      </AdminFormField>
      <AdminFormField label="日期" required>
        <AdminInput
          v-model="form.date"
          type="text"
          placeholder="YYYY-MM-DD"
          :disabled="dialogMode === 'edit'"
        />
      </AdminFormField>
      <AdminFormField label="调用次数">
        <AdminInput v-model="form.usageCount" type="number" />
      </AdminFormField>
      <AdminFormField label="Token">
        <AdminInput v-model="form.tokenCount" type="number" />
      </AdminFormField>
      <template #footer>
        <AdminButton @click="dialogVisible = false">取消</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">保存</AdminButton>
      </template>
    </AdminDialog>
  </div>
</template>
