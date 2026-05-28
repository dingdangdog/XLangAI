<script setup lang="ts">
import { CheckCircleIcon, XCircleIcon } from "@heroicons/vue/24/solid";
const { t } = useI18n();
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
    toast.error(t("toast.loadFailed"));
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
    toast.error(t("toast.loadOptionsFailed"));
    console.error(e);
  } finally {
    optionsLoading.value = false;
  }
}

watch([page, pageSize], () => void load(), { immediate: true });

const VOICE_ROLE_TABLE_COLUMN_KEYS = [
  { prop: "name", labelKey: "fields.voiceRoleName" },
  { prop: "voiceCode", labelKey: "fields.voiceCode" },
  { prop: "languageLabel", labelKey: "fields.language" },
  { prop: "ttsConfigLabel", labelKey: "fields.ttsConfig" },
  { prop: "gender", labelKey: "fields.gender" },
  { prop: "previewAudioUrl", labelKey: "fields.preview" },
  { prop: "status", labelKey: "common.status" },
  { prop: "sortOrder", labelKey: "common.sort" },
  { prop: "remark", labelKey: "common.remark" },
  { prop: "createdAt", labelKey: "common.createdAt" },
  { prop: "updatedAt", labelKey: "common.updatedAt" },
] as const;

const tableColumns = computed(() => {
  const columns =
    !list.value.length
      ? VOICE_ROLE_TABLE_COLUMN_KEYS
      : VOICE_ROLE_TABLE_COLUMN_KEYS.filter((c) =>
        Object.keys(list.value[0] as object).includes(c.prop),
      );
  return columns.map((c) => ({ prop: c.prop, label: t(c.labelKey) }));
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
  rolePrompt: "",
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
    return t("pages.voiceRoles.hintAzure");
  }
  if (p === "openai_rest") {
    return t("pages.voiceRoles.hintOpenai");
  }
  if (!p) {
    return t("pages.voiceRoles.hintNoTts");
  }
  return t("pages.voiceRoles.hintProvider", { provider: p });
});

function resetVoiceForm() {
  voiceForm.id = "";
  voiceForm.languageId = "";
  voiceForm.ttsServiceConfigId = "";
  voiceForm.voiceCode = "";
  voiceForm.name = "";
  voiceForm.gender = "female";
  voiceForm.rolePrompt = "";
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
  voiceForm.rolePrompt = r.rolePrompt != null ? String(r.rolePrompt) : "";
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
    rolePrompt: voiceForm.rolePrompt.trim() || null,
    sortOrder: voiceForm.sortOrder,
    status: voiceForm.status,
    remark: voiceForm.remark.trim() || null,
    config: voiceForm.config.trim() || null,
  };
}

async function submitVoice() {
  if (!voiceForm.languageId) {
    toast.warning(t("validation.selectTargetLanguage"));
    return;
  }
  if (!voiceForm.ttsServiceConfigId) {
    toast.warning(t("validation.selectTtsConfig"));
    return;
  }
  if (!voiceForm.voiceCode.trim()) {
    toast.warning(t("validation.fillVoiceCode"));
    return;
  }
  if (!voiceForm.name.trim()) {
    toast.warning(t("validation.fillDisplayName"));
    return;
  }
  const payload = buildPayload();
  saving.value = true;
  try {
    if (dialogMode.value === "create") {
      await $fetch("/api/admin/voice-roles", { method: "POST", body: payload });
      toast.success(t("toast.created"));
    } else {
      await $fetch(`/api/admin/voice-roles/${voiceForm.id}`, {
        method: "PUT",
        body: { id: voiceForm.id, ...payload },
      });
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
    message: t("confirm.deleteVoiceRole"),
    danger: true,
    confirmLabel: t("common.delete"),
  });
  if (!ok) return;
  try {
    await $fetch(`/api/admin/voice-roles/${row.id as string}`, { method: "DELETE" });
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

const langSelectOptions = computed(() =>
  langOptions.value.map((l) => ({ value: l.id, label: `${l.code} · ${l.name}` })),
);

const ttsSelectOptions = computed(() =>
  ttsOptions.value.map((t) => ({ value: t.id, label: `${t.name} (${t.provider})` })),
);

const genderOptions = computed(() => [
  { value: "", label: t("common.optionalParen") },
  { value: "female", label: t("status.female") },
  { value: "male", label: t("status.male") },
]);

const { activateRow, activatingId } = useActivateConfigRow({
  api,
  getList: () => list.value,
  exclusivity: "multi-active",
  reload: load,
});

const previewGeneratingId = ref<string | null>(null);
const previewDeletingId = ref<string | null>(null);
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
    toast.warning(t("validation.previewNotGenerated"));
    return;
  }
  let el = previewAudioEl.value;
  if (!el) {
    el = new Audio();
    previewAudioEl.value = el;
  }
  el.src = url;
  void el.play().catch((e) => {
    toast.error(t("toast.playFailed"));
    console.error(e);
  });
}

async function generatePreview(row: Record<string, unknown>) {
  const id = String(row.id ?? "");
  if (!id) return;
  previewGeneratingId.value = id;
  try {
    await $fetch(`/api/admin/voice-roles/${id}/generate-preview`, { method: "POST" });
    toast.success(t("toast.previewGenerated"));
    await load();
  } catch (e) {
    toast.error(t("toast.previewGenerateFailed"));
    console.error(e);
  } finally {
    previewGeneratingId.value = null;
  }
}

async function deletePreview(row: Record<string, unknown>) {
  const id = String(row.id ?? "");
  if (!id || !row.previewAudioUrl) return;
  const ok = await confirm({
    message: t("confirm.deleteVoiceRolePreview"),
    danger: true,
    confirmLabel: t("common.deletePreview"),
  });
  if (!ok) return;
  stopPreviewPlayback();
  previewDeletingId.value = id;
  try {
    await $fetch(`/api/admin/voice-roles/${id}/delete-preview`, { method: "POST" });
    toast.success(t("toast.previewDeleted"));
    await load();
  } catch (e) {
    toast.error(t("toast.previewDeleteFailed"));
    console.error(e);
  } finally {
    previewDeletingId.value = null;
  }
}
</script>

<template>
  <div class="flex min-h-0 flex-1 flex-col gap-4">
    <div class="flex justify-end">
      <AdminButton variant="primary" @click="openCreate">{{ $t("common.create") }}</AdminButton>
    </div>

    <AdminAlert :title="$t('pages.voiceRoles.usageAlertTitle')">
      {{ $t("pages.voiceRoles.usageAlert") }}
    </AdminAlert>

    <AdminPanel>
      <AdminTable :loading="loading">
        <template #head>
          <AdminTh v-for="col in tableColumns" :key="col.prop">{{ col.label }}</AdminTh>
          <AdminTh width="300px" align="right">{{ $t("common.actions") }}</AdminTh>
        </template>
        <AdminTr v-for="row in list" :key="String(row.id)">
          <AdminTd v-for="col in tableColumns" :key="col.prop">
            <template v-if="col.prop === 'previewAudioUrl'">
              <span v-if="row.previewAudioUrl" class="inline-flex" :title="$t('common.generated')">
                <CheckCircleIcon class="h-5 w-5 text-success-600" aria-hidden="true" />
                <span class="sr-only">{{ $t("common.generated") }}</span>
              </span>
              <span v-else class="inline-flex" :title="$t('common.notGenerated')">
                <XCircleIcon class="h-5 w-5 text-surface-400" aria-hidden="true" />
                <span class="sr-only">{{ $t("common.notGenerated") }}</span>
              </span>
            </template>
            <template v-else>
              <AdminCellText :title="cellValue(row, col.prop)">
                {{ cellValue(row, col.prop) }}
              </AdminCellText>
            </template>
          </AdminTd>
          <AdminTd align="right">
            <div class="flex w-[300px] flex-col gap-0.5">
              <div class="flex flex-nowrap items-center justify-end gap-x-2">
                <AdminButton v-if="String(row.status) !== 'active'" variant="link" size="sm"
                  :loading="activatingId === String(row.id)" @click="activateRow(row)">
                  {{ $t("common.enable") }}
                </AdminButton>
                <AdminButton variant="link" size="sm" :disabled="!row.previewAudioUrl" @click="playPreview(row)">
                  {{ $t("common.play") }}
                </AdminButton>
                <AdminButton variant="link" size="sm" @click="openEdit(row)">
                  {{ $t("common.edit") }}
                </AdminButton>
                <AdminButton variant="link" size="sm" class="!text-danger-600" @click="removeRow(row)">
                  {{ $t("common.delete") }}
                </AdminButton>
              </div>
              <div class="flex flex-nowrap items-center justify-end gap-x-2">
                <AdminButton variant="link" size="sm" :loading="previewGeneratingId === String(row.id)"
                  @click="generatePreview(row)">
                  {{ $t("common.generatePreview") }}
                </AdminButton>
                <AdminButton variant="link" size="sm" class="!text-danger-600" :disabled="!row.previewAudioUrl"
                  :loading="previewDeletingId === String(row.id)" @click="deletePreview(row)">
                  {{ $t("common.deletePreview") }}
                </AdminButton>
              </div>
            </div>
          </AdminTd>
        </AdminTr>
      </AdminTable>
      <AdminPagination v-model:page="page" v-model:page-size="pageSize" :total="total" />
    </AdminPanel>

    <AdminDialog v-model="dialogVisible"
      :title="dialogMode === 'create' ? $t('pages.voiceRoles.createDialog') : $t('pages.voiceRoles.editDialog')"
      width="lg">
      <AdminSkeleton v-if="optionsLoading" :rows="6" />
      <template v-else>
        <AdminFormField v-if="dialogMode === 'edit'" :label="$t('common.id')">
          <AdminInput v-model="voiceForm.id" disabled />
        </AdminFormField>
        <AdminFormField :label="$t('fields.language')" required>
          <AdminSelect v-model="voiceForm.languageId" :options="langSelectOptions"
            :placeholder="$t('pages.voiceRoles.selectLanguage')" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.ttsConfig')" required>
          <AdminSelect v-model="voiceForm.ttsServiceConfigId" :options="ttsSelectOptions"
            :placeholder="$t('pages.voiceRoles.selectTts')" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.voiceCode')" required :hint="voiceCodeHint">
          <AdminInput v-model="voiceForm.voiceCode" :placeholder="$t('pages.voiceRoles.voiceCodePlaceholder')" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.voiceRoleName')" required>
          <AdminInput v-model="voiceForm.name" :placeholder="$t('pages.voiceRoles.namePlaceholder')" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.rolePrompt')" :hint="$t('pages.voiceRoles.rolePromptHint')">
          <AdminInput v-model="voiceForm.rolePrompt" type="textarea" :rows="5"
            :placeholder="$t('pages.voiceRoles.rolePromptPlaceholder')" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.gender')">
          <AdminSelect v-model="voiceForm.gender" :options="genderOptions"
            :placeholder="$t('pages.voiceRoles.genderPlaceholder')" />
        </AdminFormField>
        <AdminFormField :label="$t('common.sort')">
          <AdminInput v-model="voiceForm.sortOrder" type="number" />
        </AdminFormField>
        <AdminFormField :label="$t('common.status')">
          <AdminSelect v-model="voiceForm.status" :options="statusOptions" />
        </AdminFormField>
        <AdminFormField :label="$t('common.remark')">
          <AdminInput v-model="voiceForm.remark" type="textarea" :rows="2" />
        </AdminFormField>
        <AdminFormField :label="$t('fields.extConfig')" :hint="$t('pages.voiceRoles.extConfigHint')">
          <AdminInput v-model="voiceForm.config" type="textarea" :rows="2" />
        </AdminFormField>
      </template>
      <template #footer>
        <AdminButton @click="dialogVisible = false">{{ $t("common.cancel") }}</AdminButton>
        <AdminButton variant="primary" :loading="saving" @click="submitVoice">{{ $t("common.save") }}</AdminButton>
      </template>
    </AdminDialog>
  </div>
</template>
