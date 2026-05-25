<script setup lang="ts">
import UserUsagePanel from "~/components/admin/UserUsagePanel.vue";
import {
  userUsageDetailLine,
  userUsagePrimaryLine,
} from "~/utils/usageDisplay";

const API = "/api/admin/users";
const api = useAdminResourceApi(API);
const toast = useToast();
const { confirm } = useConfirm();

type Opt = { id: string; label: string };

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const showDeleted = ref(false);
const statusBusyId = ref<string | null>(null);

const tierOptions = ref<Opt[]>([]);
const langOptions = ref<Opt[]>([]);
const optionsLoading = ref(false);

const STATUS_LABELS: Record<string, string> = {
  active: "正常",
  inactive: "已禁用",
  banned: "已封禁",
};

const STATUS_BADGE: Record<string, "success" | "warning" | "danger" | "muted"> = {
  active: "success",
  inactive: "warning",
  banned: "danger",
};

function statusLabel(status: unknown): string {
  const s = String(status ?? "active");
  return STATUS_LABELS[s] ?? s;
}

function statusBadgeVariant(status: unknown): "success" | "warning" | "danger" | "muted" {
  const s = String(status ?? "active");
  return STATUS_BADGE[s] ?? "muted";
}

async function loadRefs() {
  optionsLoading.value = true;
  try {
    const [tr, lr] = await Promise.all([
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/membership-tiers", {
        query: { page: 1, pageSize: 200 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/languages", {
        query: { page: 1, pageSize: 500 },
      }),
    ]);
    tierOptions.value = tr.items.map((r) => ({
      id: String(r.id),
      label: `${r.code} · ${r.name}`,
    }));
    langOptions.value = lr.items.map((r) => ({
      id: String(r.id),
      label: `${r.code} · ${r.name}`,
    }));
  } catch (e) {
    toast.error("加载下拉数据失败");
    console.error(e);
  } finally {
    optionsLoading.value = false;
  }
}

async function load() {
  loading.value = true;
  try {
    const q: Record<string, string | number | boolean> = {
      page: page.value,
      pageSize: pageSize.value,
    };
    if (showDeleted.value) q.includeDeleted = 1;
    const res = await api.list(q);
    list.value = res.items;
    total.value = res.total;
  } catch (e) {
    toast.error("加载失败");
    console.error(e);
  } finally {
    loading.value = false;
  }
}

watch([page, pageSize, showDeleted], () => void load(), { immediate: true });

onMounted(() => void loadRefs());

const TIER_BADGE: Record<string, "default" | "success" | "warning" | "muted"> = {
  free: "muted",
  plus: "default",
  pro: "success",
};

function tierBadgeVariant(code: unknown): "default" | "success" | "warning" | "muted" {
  const c = String(code ?? "").toLowerCase();
  return TIER_BADGE[c] ?? "muted";
}

function tierSubline(row: Record<string, unknown>): string {
  const parts: string[] = [];
  const code = String(row.tierCode ?? "").trim();
  // if (code) parts.push(code);
  const daily = row.tierDailyLimit;
  const monthly = row.tierMonthlyLimit;
  if (daily != null && Number(daily) > 0) parts.push(`日限额 ${daily}`);
  if (monthly != null && Number(monthly) > 0) parts.push(`月限额 ${monthly}`);
  if (row.subscriptionExpiresAt) {
    parts.push(`订阅至 ${formatDate(row.subscriptionExpiresAt)}`);
  }
  const bal = row.tokenBalance;
  if (bal != null && String(bal) !== "0") parts.push(`Token 余额 ${bal}`);
  return parts.join(" · ") || "—";
}

const usageDialogVisible = ref(false);
const usageUser = ref<{ id: string; label: string } | null>(null);

function usageUserLabel(row: Record<string, unknown>): string {
  const nick = String(row.nickname ?? "").trim();
  const phone = String(row.phone ?? "").trim();
  const email = String(row.email ?? "").trim();
  return nick || phone || email || String(row.id ?? "");
}

function openUsage(row: Record<string, unknown>) {
  usageUser.value = { id: String(row.id), label: usageUserLabel(row) };
  usageDialogVisible.value = true;
}

watch(usageDialogVisible, (open) => {
  if (!open) usageUser.value = null;
});

const dialogVisible = ref(false);
const dialogMode = ref<"create" | "edit">("create");
const saving = ref(false);

const form = reactive({
  id: "",
  phone: "",
  email: "",
  nickname: "",
  password: "",
  avatarUrl: "",
  tierId: "",
  languageId: "",
  settings: "{}",
  status: "active",
  remark: "",
});

function resetForm() {
  form.id = "";
  form.phone = "";
  form.email = "";
  form.nickname = "";
  form.password = "";
  form.avatarUrl = "";
  form.tierId = "";
  form.languageId = "";
  form.settings = "{}";
  form.status = "active";
  form.remark = "";
}

function openCreate() {
  dialogMode.value = "create";
  resetForm();
  dialogVisible.value = true;
  void loadRefs();
}

function openEdit(row: Record<string, unknown>) {
  dialogMode.value = "edit";
  form.id = String(row.id ?? "");
  form.phone = String(row.phone ?? "");
  form.email = String(row.email ?? "");
  form.nickname = String(row.nickname ?? "");
  form.password = "";
  form.avatarUrl = String(row.avatarUrl ?? "");
  form.tierId = String(row.tierId ?? "");
  form.languageId = String(row.languageId ?? "");
  form.settings = row.settings != null ? String(row.settings) : "{}";
  form.status = String(row.status ?? "active");
  form.remark = String(row.remark ?? "");
  dialogVisible.value = true;
  void loadRefs();
}

async function submit() {
  const phone = form.phone.trim();
  const email = form.email.trim();
  if (!phone && !email) {
    toast.error("须填写手机或邮箱其一");
    return;
  }
  if (dialogMode.value === "create") {
    if (!form.password || form.password.length < 6) {
      toast.error("新建用户须设置至少 6 位密码");
      return;
    }
  } else if (form.password && form.password.length < 6) {
    toast.error("新密码至少 6 位");
    return;
  }

  let settingsStr = (form.settings ?? "").trim();
  if (settingsStr) {
    try {
      JSON.parse(settingsStr);
    } catch {
      toast.error("settings 须为合法 JSON");
      return;
    }
  } else {
    settingsStr = "{}";
  }

  const body: Record<string, unknown> = {
    phone: phone || null,
    email: email || null,
    nickname: form.nickname.trim() || null,
    avatarUrl: form.avatarUrl.trim() || null,
    tierId: form.tierId || null,
    languageId: form.languageId || null,
    settings: settingsStr,
    status: form.status,
    remark: form.remark.trim() || null,
  };
  if (form.password) {
    body.password = form.password;
  }

  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await api.create(body);
      toast.success("已创建，用户可使用手机/邮箱 + 密码登录");
    } else {
      await api.update(form.id, { id: form.id, ...body });
      toast.success(form.password ? "已保存（含新密码）" : "已保存");
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

async function setUserStatus(row: Record<string, unknown>, status: string, actionLabel: string) {
  const id = String(row.id);
  if (row.deletedAt) return;
  const ok = await confirm({
    message: `确认将用户设为「${STATUS_LABELS[status] ?? status}」？${status !== "active" ? "该用户将无法再登录 App。" : ""}`,
    danger: status !== "active",
    confirmLabel: actionLabel,
  });
  if (!ok) return;
  statusBusyId.value = id;
  try {
    await api.update(id, { status });
    toast.success(`已${actionLabel}`);
    await load();
  } catch (e) {
    toast.error("操作失败");
    console.error(e);
  } finally {
    statusBusyId.value = null;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: "确认软删除该用户？删除后无法登录，可在「含已删除」中查看。",
    danger: true,
    confirmLabel: "软删",
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success("已标记删除");
    await load();
  } catch (e) {
    toast.error("操作失败");
    console.error(e);
  }
}

const statusOptions = [
  { value: "active", label: "正常（可登录）" },
  { value: "inactive", label: "已禁用（不可登录）" },
  { value: "banned", label: "已封禁（不可登录）" },
];

const tierSelectOptions = computed(() => [
  { value: "", label: "（可选）" },
  ...tierOptions.value.map((t) => ({ value: t.id, label: t.label })),
]);

const langSelectOptions = computed(() => [
  { value: "", label: "（可选）" },
  ...langOptions.value.map((l) => ({ value: l.id, label: l.label })),
]);
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader title="用户账号"
        description="账号与会员等级、今日/本月用量（对话轮次、LLM Token、TTS/翻译/STT）；可在此新建账号、重置密码、管理按日用量记录，以及禁用或封禁账号。">
        <template #actions>
          <AdminCheckbox v-model="showDeleted" label="含已删除" />
          <AdminButton variant="primary" @click="openCreate">新建账号</AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>手机</AdminTh>
          <AdminTh>邮箱</AdminTh>
          <AdminTh>昵称</AdminTh>
          <AdminTh width="72px">密码</AdminTh>
          <AdminTh width="160px">会员</AdminTh>
          <AdminTh width="140px">今日用量</AdminTh>
          <AdminTh width="140px">本月用量</AdminTh>
          <AdminTh width="96px">状态</AdminTh>
          <AdminTh>最近登录</AdminTh>
          <AdminTh width="260px" align="right">操作</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd>{{ row.phone ?? "—" }}</AdminTd>
          <AdminTd>{{ row.email ?? "—" }}</AdminTd>
          <AdminTd>{{ row.nickname ?? "—" }}</AdminTd>
          <AdminTd>
            <AdminBadge :variant="row.hasPassword ? 'success' : 'muted'">
              {{ row.hasPassword ? "已设" : "无" }}
            </AdminBadge>
          </AdminTd>
          <AdminTd>
            <div v-if="row.tierLabel" class="flex flex-col gap-0.5">
              <AdminBadge :variant="tierBadgeVariant(row.tierCode)">
                {{ row.tierName || row.tierCode }}
              </AdminBadge>
              <span class="text-xs text-muted leading-snug">{{ tierSubline(row) }}</span>
            </div>
            <span v-else class="text-sm text-muted">未分配</span>
          </AdminTd>
          <AdminTd class="text-sm tabular-nums">
            <div>{{ userUsagePrimaryLine(row, "today") }}</div>
            <div v-if="userUsageDetailLine(row, 'today')" class="text-xs text-muted leading-snug">
              {{ userUsageDetailLine(row, "today") }}
            </div>
          </AdminTd>
          <AdminTd class="text-sm tabular-nums">
            <div>{{ userUsagePrimaryLine(row, "month") }}</div>
            <div v-if="userUsageDetailLine(row, 'month')" class="text-xs text-muted leading-snug">
              {{ userUsageDetailLine(row, "month") }}
            </div>
          </AdminTd>
          <AdminTd>
            <AdminBadge :variant="statusBadgeVariant(row.status)">
              {{ statusLabel(row.status) }}
            </AdminBadge>
          </AdminTd>
          <AdminTd nowrap>{{ formatDateTime(row.lastLoginAt) }}</AdminTd>
          <AdminTd align="right">
            <AdminButton variant="link" @click="openUsage(row)">用量</AdminButton>
            <AdminButton variant="link" @click="openEdit(row)">编辑</AdminButton>
            <template v-if="!row.deletedAt">
              <AdminButton v-if="row.status !== 'inactive'" variant="link" :disabled="statusBusyId === String(row.id)"
                @click="setUserStatus(row, 'inactive', '禁用')">
                禁用
              </AdminButton>
              <AdminButton v-if="row.status !== 'banned'" variant="link" class="!text-danger-600"
                :disabled="statusBusyId === String(row.id)" @click="setUserStatus(row, 'banned', '封禁')">
                封禁
              </AdminButton>
              <AdminButton v-if="row.status !== 'active'" variant="link" :disabled="statusBusyId === String(row.id)"
                @click="setUserStatus(row, 'active', '解禁')">
                解禁
              </AdminButton>
            </template>
            <AdminButton variant="link" class="!text-danger-600" :disabled="!!row.deletedAt" @click="removeRow(row)">
              软删
            </AdminButton>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog v-model="dialogVisible" :title="dialogMode === 'create' ? '新建账号' : '编辑账号'" width="lg">
      <AdminSkeleton v-if="optionsLoading" :rows="5" />
      <template v-else>
        <AdminFormField v-if="dialogMode === 'edit'" label="ID">
          <AdminInput v-model="form.id" disabled />
        </AdminFormField>
        <AdminFormField label="手机" hint="手机与邮箱至少填一项">
          <AdminInput v-model="form.phone" />
        </AdminFormField>
        <AdminFormField label="邮箱">
          <AdminInput v-model="form.email" type="email" />
        </AdminFormField>
        <AdminFormField label="昵称">
          <AdminInput v-model="form.nickname" />
        </AdminFormField>
        <AdminFormField :label="dialogMode === 'create' ? '登录密码' : '重置密码'" :hint="dialogMode === 'create'
            ? '至少 6 位，用于 App 密码登录'
            : '留空则不修改；填写则覆盖为新密码'
          ">
          <AdminInput v-model="form.password" type="password" autocomplete="new-password"
            :placeholder="dialogMode === 'create' ? '必填' : '留空不修改'" />
        </AdminFormField>
        <AdminFormField label="头像 URL">
          <AdminInput v-model="form.avatarUrl" />
        </AdminFormField>
        <AdminFormField label="会员等级">
          <AdminSelect v-model="form.tierId" :options="tierSelectOptions" />
        </AdminFormField>
        <AdminFormField label="母语" hint="用于客户端语音转文字双语辅助识别；与练习目标语言无关">
          <AdminSelect v-model="form.languageId" :options="langSelectOptions" />
        </AdminFormField>
        <AdminFormField label="settings">
          <AdminInput v-model="form.settings" type="textarea" :rows="4" class="font-mono text-sm" />
        </AdminFormField>
        <AdminFormField label="账号状态">
          <AdminSelect v-model="form.status" :options="statusOptions" />
        </AdminFormField>
        <AdminFormField label="备注">
          <AdminInput v-model="form.remark" type="textarea" :rows="2" />
        </AdminFormField>
      </template>
      <template #footer>
        <AdminButton @click="dialogVisible = false">取消</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">保存</AdminButton>
      </template>
    </AdminDialog>

    <AdminDialog v-model="usageDialogVisible" :title="usageUser ? `用量明细 · ${usageUser.label}` : '用量明细'" width="lg">
      <UserUsagePanel v-if="usageUser" :user-id="usageUser.id" :user-label="usageUser.label" @changed="load" />
    </AdminDialog>
  </AdminListPage>
</template>
