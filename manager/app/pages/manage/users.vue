<script setup lang="ts">
const { t } = useI18n();
import UserUsagePanel from "~/components/admin/UserUsagePanel.vue";

const { userUsageDetailLine, userUsagePrimaryLine } = useUsageDisplay();

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
const llmOptions = ref<Opt[]>([]);
const optionsLoading = ref(false);

const STATUS_LABELS = computed<Record<string, string>>(() => ({
  active: t("status.active"),
  inactive: t("status.inactive"),
  banned: t("status.banned"),
}));

const STATUS_BADGE: Record<string, "success" | "warning" | "danger" | "muted"> = {
  active: "success",
  inactive: "warning",
  banned: "danger",
};

function statusLabel(status: unknown): string {
  const s = String(status ?? "active");
  return STATUS_LABELS.value[s] ?? s;
}

function statusBadgeVariant(status: unknown): "success" | "warning" | "danger" | "muted" {
  const s = String(status ?? "active");
  return STATUS_BADGE[s] ?? "muted";
}

async function loadRefs() {
  optionsLoading.value = true;
  try {
    const [tr, lr, ar] = await Promise.all([
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/membership-tiers", {
        query: { page: 1, pageSize: 200 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/languages", {
        query: { page: 1, pageSize: 500 },
      }),
      $fetch<{ items: Record<string, unknown>[] }>("/api/admin/llm-service-configs", {
        query: { page: 1, pageSize: 200 },
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
    llmOptions.value = ar.items.map((r) => ({
      id: String(r.id),
      label: `${r.code} · ${r.name}`,
    }));
  } catch (e) {
    toast.error(t("toast.loadRefsFailed"));
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
    toast.error(t("toast.loadFailed"));
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
  if (daily != null && Number(daily) > 0) {
    parts.push(t("usage.dailyLimit", { value: daily }));
  }
  if (monthly != null && Number(monthly) > 0) {
    parts.push(t("usage.monthlyLimit", { value: monthly }));
  }
  if (row.subscriptionExpiresAt) {
    parts.push(
      t("usage.subscriptionUntil", { date: formatDate(row.subscriptionExpiresAt) }),
    );
  }
  const bal = row.tokenBalance;
  if (bal != null && String(bal) !== "0") {
    parts.push(t("usage.tokenBalance", { value: bal }));
  }
  const turns = row.turnBalance;
  if (turns != null) {
    parts.push(t("usage.turnBalance", { value: turns }));
  }
  return parts.join(t("usage.separator")) || t("common.emDash");
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

const quotaDialogVisible = ref(false);
const quotaUser = ref<Record<string, unknown> | null>(null);
const grantAmount = ref<number | null>(null);
const grantReason = ref("");
const quotaSaving = ref(false);

const grantAfterBalance = computed(() => {
  const current = Number(quotaUser.value?.turnBalance ?? 0);
  const amount = Number(grantAmount.value ?? 0);
  return current + (Number.isFinite(amount) && amount > 0 ? Math.floor(amount) : 0);
});

function openQuota(row: Record<string, unknown>) {
  quotaUser.value = row;
  grantAmount.value = null;
  grantReason.value = "";
  quotaDialogVisible.value = true;
}

async function grantQuota() {
  const user = quotaUser.value;
  const amount = Math.floor(Number(grantAmount.value));
  const reason = grantReason.value.trim();
  if (!user || !Number.isFinite(amount) || amount <= 0) {
    toast.warning(t("pages.users.grantAmountInvalid"));
    return;
  }
  if (!reason) {
    toast.warning(t("pages.users.grantReasonRequired"));
    return;
  }

  quotaSaving.value = true;
  try {
    await $fetch(`/api/admin/users/${String(user.id)}/grant-turns`, {
      method: "POST",
      body: { amount, reason },
    });
    toast.success(t("pages.users.grantSuccess", { amount }));
    quotaDialogVisible.value = false;
    await load();
  } catch (error) {
    toast.error(t("toast.operationFailed"));
    console.error(error);
  } finally {
    quotaSaving.value = false;
  }
}

type UserAction = {
  label: string;
  onClick: () => void;
  danger?: boolean;
  disabled?: boolean;
};

function userAccountActions(row: Record<string, unknown>): UserAction[] {
  const actions: UserAction[] = [];
  if (!row.deletedAt) {
    if (row.status !== "inactive") {
      actions.push({
        label: t("pages.users.disable"),
        onClick: () => setUserStatus(row, "inactive", t("pages.users.disable")),
      });
    }
    if (row.status !== "banned") {
      actions.push({
        label: t("pages.users.ban"),
        danger: true,
        onClick: () => setUserStatus(row, "banned", t("pages.users.ban")),
      });
    }
    if (row.status !== "active") {
      actions.push({
        label: t("pages.users.unban"),
        onClick: () => setUserStatus(row, "active", t("pages.users.unban")),
      });
    }
    actions.push({
      label: t("common.softDelete"),
      danger: true,
      onClick: () => removeRow(row),
    });
  }
  return actions;
}

function userMobileActions(row: Record<string, unknown>): UserAction[] {
  return [
    {
      label: t("pages.users.grantQuota"),
      onClick: () => openQuota(row),
      disabled: !!row.deletedAt,
    },
    { label: t("common.usage"), onClick: () => openUsage(row) },
    { label: t("common.edit"), onClick: () => openEdit(row) },
    ...userAccountActions(row),
  ];
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
  defaultLlmConfigId: "",
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
  form.defaultLlmConfigId = "";
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
  form.defaultLlmConfigId = String(row.defaultLlmConfigId ?? "");
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
    toast.error(t("validation.phoneOrEmail"));
    return;
  }
  if (dialogMode.value === "create") {
    if (!form.password || form.password.length < 6) {
      toast.error(t("validation.newUserPassword"));
      return;
    }
  } else if (form.password && form.password.length < 6) {
    toast.error(t("validation.passwordMinEdit"));
    return;
  }

  let settingsStr = (form.settings ?? "").trim();
  if (settingsStr) {
    try {
      JSON.parse(settingsStr);
    } catch {
      toast.error(t("validation.invalidJson", { field: "settings" }));
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
    defaultLlmConfigId: form.defaultLlmConfigId || null,
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
      toast.success(t("toast.userCreated"));
    } else {
      await api.update(form.id, { id: form.id, ...body });
      toast.success(form.password ? t("toast.savedWithPassword") : t("toast.saved"));
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

async function setUserStatus(row: Record<string, unknown>, status: string, actionLabel: string) {
  const id = String(row.id);
  if (row.deletedAt) return;
  const ok = await confirm({
    message: t("confirm.setUserStatus", {
      status: STATUS_LABELS.value[status] ?? status,
      extra: status !== "active" ? t("confirm.cannotLoginAfter") : "",
    }),
    danger: status !== "active",
    confirmLabel: actionLabel,
  });
  if (!ok) return;
  statusBusyId.value = id;
  try {
    await api.update(id, { status });
    toast.success(t("toast.statusSet", { action: actionLabel }));
    await load();
  } catch (e) {
    toast.error(t("toast.operationFailed"));
    console.error(e);
  } finally {
    statusBusyId.value = null;
  }
}

async function removeRow(row: Record<string, unknown>) {
  const ok = await confirm({
    message: t("confirm.softDeleteUser"),
    danger: true,
    confirmLabel: t("common.softDelete"),
  });
  if (!ok) return;
  try {
    await api.remove(String(row.id));
    toast.success(t("toast.markedDeleted"));
    await load();
  } catch (e) {
    toast.error(t("toast.operationFailed"));
    console.error(e);
  }
}

const statusOptions = computed(() => [
  { value: "active", label: t("status.activeLogin") },
  { value: "inactive", label: t("status.inactiveLogin") },
  { value: "banned", label: t("status.bannedLogin") },
]);

const tierSelectOptions = computed(() => [
  { value: "", label: t("common.optional") },
  ...tierOptions.value.map((tier) => ({ value: tier.id, label: tier.label })),
]);

const langSelectOptions = computed(() => [
  { value: "", label: t("common.optional") },
  ...langOptions.value.map((l) => ({ value: l.id, label: l.label })),
]);

const llmSelectOptions = computed(() => [
  { value: "", label: t("pages.users.defaultLlmGlobal") },
  ...llmOptions.value.map((l) => ({ value: l.id, label: l.label })),
]);
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.users.title')"
        :description="$t('pages.users.description')"
      >
        <template #actions>
          <AdminCheckbox v-model="showDeleted" :label="$t('common.includeDeleted')" />
          <AdminButton variant="primary" class="w-full sm:w-auto" @click="openCreate">
            {{ $t("pages.users.createAccount") }}
          </AdminButton>
        </template>
      </AdminPageHeader>
    </template>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh>{{ $t("fields.phone") }}</AdminTh>
          <AdminTh>{{ $t("fields.email") }}</AdminTh>
          <AdminTh>{{ $t("fields.nickname") }}</AdminTh>
          <AdminTh width="72px">{{ $t("fields.password") }}</AdminTh>
          <AdminTh width="160px">{{ $t("fields.tier") }}</AdminTh>
          <AdminTh width="160px">{{ $t("fields.defaultLlm") }}</AdminTh>
          <AdminTh width="140px">{{ $t("fields.todayUsageCol") }}</AdminTh>
          <AdminTh width="140px">{{ $t("fields.monthUsageCol") }}</AdminTh>
          <AdminTh width="96px">{{ $t("common.status") }}</AdminTh>
          <AdminTh>{{ $t("fields.lastLogin") }}</AdminTh>
          <AdminTh width="260px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd>{{ row.phone ?? t("common.emDash") }}</AdminTd>
          <AdminTd>{{ row.email ?? t("common.emDash") }}</AdminTd>
          <AdminTd>{{ row.nickname ?? t("common.emDash") }}</AdminTd>
          <AdminTd>
            <AdminBadge :variant="row.hasPassword ? 'success' : 'muted'">
              {{ row.hasPassword ? $t("common.set") : $t("common.unset") }}
            </AdminBadge>
          </AdminTd>
          <AdminTd>
            <div v-if="row.tierLabel" class="flex flex-col gap-0.5">
              <AdminBadge :variant="tierBadgeVariant(row.tierCode)">
                {{ row.tierName || row.tierCode }}
              </AdminBadge>
              <span class="text-xs text-muted leading-snug">{{ tierSubline(row) }}</span>
            </div>
            <span v-else class="text-sm text-muted">{{ $t("common.notAssigned") }}</span>
          </AdminTd>
          <AdminTd class="text-sm">
            <span v-if="row.defaultLlmLabel">{{ row.defaultLlmLabel }}</span>
            <span v-else class="text-muted">{{ $t("pages.users.defaultLlmGlobal") }}</span>
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
            <AdminButton variant="link" :disabled="!!row.deletedAt" @click="openQuota(row)">
              {{ $t("pages.users.grantQuota") }}
            </AdminButton>
            <AdminButton variant="link" @click="openUsage(row)">{{ $t("common.usage") }}</AdminButton>
            <AdminButton variant="link" @click="openEdit(row)">{{ $t("common.edit") }}</AdminButton>
            <AdminOverflowMenu
              v-if="userAccountActions(row).length"
              :actions="userAccountActions(row)"
            />
          </AdminTd>
        </AdminTr>
        <template #mobile>
          <p v-if="!list.length && !loading" class="py-12 text-center text-sm text-muted">
            {{ $t("table.noData") }}
          </p>
          <AdminMobileCard
            v-for="row in list"
            :key="String(row.id)"
            :title="usageUserLabel(row)"
            :subtitle="[row.phone, row.email].filter(Boolean).join(' · ') || undefined"
          >
            <template #badge>
              <AdminBadge :variant="statusBadgeVariant(row.status)">
                {{ statusLabel(row.status) }}
              </AdminBadge>
            </template>
            <template #menu>
              <AdminOverflowMenu :actions="userMobileActions(row)" />
            </template>
            <AdminMobileMeta :label="$t('fields.tier')">
              <span v-if="row.tierLabel">{{ row.tierName || row.tierCode }}</span>
              <span v-else>{{ $t("common.notAssigned") }}</span>
            </AdminMobileMeta>
            <AdminMobileMeta :label="$t('fields.defaultLlm')">
              {{ row.defaultLlmLabel || $t("pages.users.defaultLlmGlobal") }}
            </AdminMobileMeta>
            <AdminMobileMeta :label="$t('fields.todayUsageCol')">
              {{ userUsagePrimaryLine(row, "today") }}
            </AdminMobileMeta>
            <AdminMobileMeta :label="$t('fields.monthUsageCol')">
              {{ userUsagePrimaryLine(row, "month") }}
            </AdminMobileMeta>
            <AdminMobileMeta :label="$t('fields.lastLogin')">
              {{ formatDateTime(row.lastLoginAt) }}
            </AdminMobileMeta>
          </AdminMobileCard>
        </template>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? $t('pages.users.createDialog') : $t('pages.users.editDialog')"
      width="lg"
    >
      <AdminSkeleton v-if="optionsLoading" :rows="5" />
      <template v-else>
        <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
          <AdminInput v-model="form.id" disabled />
        </AdminFormField>
        <AdminFormField :label="$t('fields.phone')" :hint="$t('pages.users.phoneEmailHint')">
          <AdminInput v-model="form.phone" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.email')">
          <AdminInput v-model="form.email" type="email" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.nickname')">
          <AdminInput v-model="form.nickname" />
        </AdminFormField>
        <AdminFormField
          :label="dialogMode === 'create' ? $t('fields.loginPassword') : $t('fields.resetPassword')"
          :hint="
            dialogMode === 'create'
              ? $t('pages.users.passwordHintCreate')
              : $t('pages.users.passwordHintEdit')
          "
        >
          <AdminInput
            v-model="form.password"
            type="password"
            autocomplete="new-password"
            :placeholder="
              dialogMode === 'create'
                ? $t('pages.users.passwordPlaceholderCreate')
                : $t('pages.users.passwordPlaceholderEdit')
            "
          />
        </AdminFormField>
        <AdminFormField :label="$t('fields.avatarUrl')">
          <AdminInput v-model="form.avatarUrl" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.tierLevel')">
          <AdminSelect v-model="form.tierId" :options="tierSelectOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.nativeLanguage')" :hint="$t('pages.users.nativeLanguageHint')">
          <AdminSelect v-model="form.languageId" :options="langSelectOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.defaultLlm')" :hint="$t('pages.users.defaultLlmHint')">
          <AdminSelect v-model="form.defaultLlmConfigId" :options="llmSelectOptions" />
        </AdminFormField>
        <AdminFormField label="settings">
          <AdminInput v-model="form.settings" type="textarea" :rows="4" class="font-mono text-sm" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.accountStatus')">
          <AdminSelect v-model="form.status" :options="statusOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('common.remark')">
          <AdminInput v-model="form.remark" type="textarea" :rows="2" />
        </AdminFormField>
      </template>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submit">{{ $t("common.save") }}</AdminButton>
      </template>
    </AdminDialog>

    <AdminDialog
      v-model="quotaDialogVisible"
      :title="$t('pages.users.grantDialog')"
      width="md"
    >
      <div v-if="quotaUser" class="space-y-4">
        <div class="rounded-xl border border-border bg-surface-muted/40 p-4">
          <p class="text-sm font-semibold text-foreground">{{ usageUserLabel(quotaUser) }}</p>
          <p class="mt-1 text-xs text-muted">{{ quotaUser.phone || quotaUser.email || quotaUser.id }}</p>
          <dl class="mt-4 grid grid-cols-3 gap-3 text-center">
            <div>
              <dt class="text-xs text-muted">{{ $t("pages.users.grantMonthlyUsage") }}</dt>
              <dd class="mt-1 text-lg font-semibold tabular-nums text-foreground">
                {{ quotaUser.monthUsageCount ?? 0 }} / {{ quotaUser.tierMonthlyLimit || $t("common.emDash") }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-muted">{{ $t("pages.users.grantCurrentBalance") }}</dt>
              <dd class="mt-1 text-lg font-semibold tabular-nums text-foreground">
                {{ quotaUser.turnBalance ?? 0 }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-muted">{{ $t("pages.users.grantAfterBalance") }}</dt>
              <dd class="mt-1 text-lg font-semibold tabular-nums text-primary-600">
                {{ grantAfterBalance }}
              </dd>
            </div>
          </dl>
        </div>
        <AdminFormField :label="$t('pages.users.grantAmount')" required>
          <AdminInput
            v-model="grantAmount"
            type="number"
            :placeholder="$t('pages.users.grantAmountPlaceholder')"
          />
        </AdminFormField>
        <AdminFormField
          :label="$t('pages.users.grantReason')"
          :hint="$t('pages.users.grantReasonHint')"
          required
        >
          <AdminInput
            v-model="grantReason"
            type="textarea"
            :rows="2"
            :placeholder="$t('pages.users.grantReasonPlaceholder')"
          />
        </AdminFormField>
      </div>
      <template #footer>
        <AdminButton @click="quotaDialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="quotaSaving" @click="grantQuota">
          {{ $t("pages.users.grantSubmit") }}
        </AdminButton>
      </template>
    </AdminDialog>

    <AdminDialog
      v-model="usageDialogVisible"
      :title="
        usageUser
          ? $t('pages.users.usageDialogWithUser', { label: usageUser.label })
          : $t('pages.users.usageDialog')
      "
      width="lg"
    >
      <UserUsagePanel v-if="usageUser" :user-id="usageUser.id" :user-label="usageUser.label" />
    </AdminDialog>
  </AdminListPage>
</template>
