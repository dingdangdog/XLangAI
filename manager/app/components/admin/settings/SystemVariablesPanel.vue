<script setup lang="ts">
import { ArrowLeftIcon, ChevronDownIcon, ChevronRightIcon } from "@heroicons/vue/24/outline";
import { SYSTEM_SETTING_GROUPS } from "~/utils/systemSettingsUi";

const { t } = useI18n();
const { labelForKey, hintForKey } = useSystemSettingLabels();
const route = useRoute();
const router = useRouter();
const isMobile = useIsMobile();
const API = "/api/admin/system-settings";
const api = useAdminResourceApi(API);
const toast = useToast();

const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const saving = ref(false);
/** 手机端：是否进入编辑钻取页 */
const mobileShowEditor = ref(false);

const expandedGroups = ref<Record<string, boolean>>({
  auth: true,
  media: true,
});

const form = reactive({
  id: "",
  key: "",
  value: "",
  valueType: "bool",
  status: "active",
  description: "",
  updatedAt: "",
});

async function load() {
  loading.value = true;
  try {
    const res = await api.list({ page: 1, pageSize: 200 });
    list.value = res.items;
  } catch (e) {
    toast.error(t("toast.loadFailed"));
    console.error(e);
  } finally {
    loading.value = false;
  }
}

void load();

const byKey = computed(() => {
  const map = new Map<string, Record<string, unknown>>();
  for (const row of list.value) {
    map.set(String(row.key ?? ""), row);
  }
  return map;
});

const menuGroups = computed(() =>
  SYSTEM_SETTING_GROUPS.map((group) => ({
    id: group.id,
    label: t(group.labelKey),
    description: t(group.descriptionKey),
    items: group.keys.map((key) => ({
      key,
      label: labelForKey(key),
      hint: hintForKey(key),
      row: byKey.value.get(key) ?? null,
    })),
  })),
);

function displayValuePreview(row: Record<string, unknown>) {
  const value = String(row.value ?? "");
  if (row.valueType === "bool") {
    return value === "true" ? t("status.on") : t("status.off");
  }
  if (String(row.key ?? "").startsWith("media.")) {
    if (value === "client") return t("pages.systemSettings.storageClient");
    if (value === "server") return t("pages.systemSettings.storageServer");
    if (value === "cloud") return t("pages.systemSettings.storageCloud");
    return value;
  }
  return value;
}

const selectedKey = ref("");

onMounted(() => {
  if (!route.query.key) return;
  const q = { ...route.query };
  delete q.key;
  router.replace({ query: q });
});

watch(
  () => list.value,
  () => {
    if (selectedKey.value) return;
    const first = SYSTEM_SETTING_GROUPS[0]?.keys[0];
    if (first) selectedKey.value = first;
  },
  { immediate: true },
);

function fillFormFromRow(row: Record<string, unknown>) {
  form.id = String(row.id ?? "");
  form.key = String(row.key ?? "");
  form.value = String(row.value ?? "");
  form.valueType = String(row.valueType ?? "string");
  form.status = String(row.status ?? "active");
  form.description = String(row.description ?? "");
  form.updatedAt = String(row.updatedAt ?? "");
}

watch(
  [selectedKey, byKey],
  () => {
    const row = byKey.value.get(selectedKey.value);
    if (row) fillFormFromRow(row);
  },
  { immediate: true },
);

function selectKey(key: string) {
  selectedKey.value = key;
  if (isMobile.value) mobileShowEditor.value = true;
}

function backToList() {
  mobileShowEditor.value = false;
}

function toggleGroup(id: string) {
  expandedGroups.value[id] = !expandedGroups.value[id];
}

watch(isMobile, (mobile) => {
  if (!mobile) mobileShowEditor.value = false;
});

const storageOptionsAll = computed(() => [
  { value: "client", label: t("pages.systemSettings.storageClient") },
  { value: "server", label: t("pages.systemSettings.storageServer") },
  { value: "cloud", label: t("pages.systemSettings.storageCloud") },
]);

const storageOptionsNoClient = computed(() =>
  storageOptionsAll.value.filter((o) => o.value !== "client"),
);

function storageOptionsForKey(k: string) {
  if (k === "media.user_recording.storage") return storageOptionsAll.value;
  return storageOptionsNoClient.value;
}

const boolValueOptions = computed(() => [
  { value: "true", label: t("status.trueOn") },
  { value: "false", label: t("status.falseOff") },
]);

const statusOptions = computed(() => [
  { value: "active", label: t("status.active") },
  { value: "inactive", label: t("status.inactive") },
]);

const hasSelection = computed(() => {
  const row = byKey.value.get(selectedKey.value);
  return !!row && !!String(row.id ?? "");
});

const missingRow = computed(
  () => !!selectedKey.value && !loading.value && !hasSelection.value,
);

const selectedLabel = computed(() => labelForKey(selectedKey.value));
const selectedHint = computed(() => hintForKey(selectedKey.value));

async function submit() {
  if (!form.id) return;
  saving.value = true;
  try {
    await api.update(form.id, {
      id: form.id,
      value: String(form.value).trim(),
      status: form.status,
      description: form.description.trim() || null,
    });
    toast.success(t("toast.saved"));
    await load();
    const row = byKey.value.get(form.key);
    if (row) fillFormFromRow(row);
  } catch (e) {
    toast.error(t("toast.saveFailed"));
    console.error(e);
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <div class="min-h-0 flex-1 overflow-hidden">
    <div
      class="grid h-full min-h-0 w-full overflow-hidden rounded-xl border border-border bg-surface shadow-sm md:grid-cols-[minmax(260px,300px)_minmax(0,1fr)]"
    >
      <!-- 左：配置列表（手机钻取时隐藏） -->
      <section
        class="flex min-h-0 flex-col overflow-hidden border-border bg-surface-muted/30 md:border-r"
        :class="mobileShowEditor ? 'hidden md:flex' : 'flex'"
      >
        <header class="shrink-0 border-b border-border bg-surface px-4 py-3">
          <h2 class="text-sm font-semibold text-foreground">
            {{ $t("pages.systemSettings.listColumnTitle") }}
          </h2>
          <p class="mt-1 text-xs leading-relaxed text-muted">
            {{ $t("pages.systemSettings.listColumnHint") }}
          </p>
        </header>

        <div v-if="loading && !list.length" class="space-y-2 p-3">
          <AdminSkeleton v-for="i in 6" :key="i" class="h-12 rounded-lg" />
        </div>

        <nav v-else class="min-h-0 flex-1 overflow-y-auto p-2">
          <div v-for="group in menuGroups" :key="group.id" class="mb-2">
            <button
              type="button"
              class="flex w-full items-start gap-1 rounded-lg px-2 py-2 text-left hover:bg-surface"
              @click="toggleGroup(group.id)"
            >
              <ChevronDownIcon v-if="expandedGroups[group.id]" class="mt-0.5 h-4 w-4 shrink-0 text-muted" />
              <ChevronRightIcon v-else class="mt-0.5 h-4 w-4 shrink-0 text-muted" />
              <div class="min-w-0 flex-1">
                <p class="text-sm font-semibold text-foreground">{{ group.label }}</p>
                <p class="text-xs text-muted">{{ group.description }}</p>
              </div>
            </button>

            <ul
              v-show="expandedGroups[group.id]"
              class="ml-4 mt-0.5 space-y-0.5 border-l-2 border-primary-200 pl-2 dark:border-primary-800"
            >
              <li v-for="item in group.items" :key="item.key">
                <button
                  type="button"
                  class="w-full rounded-lg px-2.5 py-2 text-left transition-colors"
                  :class="
                    selectedKey === item.key
                      ? 'bg-primary-600 text-white shadow-sm'
                      : 'hover:bg-surface text-foreground'
                  "
                  @click="selectKey(item.key)"
                >
                  <p class="text-sm font-medium leading-snug">{{ item.label }}</p>
                  <p
                    v-if="item.hint"
                    class="mt-0.5 line-clamp-2 text-xs"
                    :class="selectedKey === item.key ? 'text-primary-100' : 'text-muted'"
                  >
                    {{ item.hint }}
                  </p>
                  <p
                    v-if="item.row"
                    class="mt-1 truncate text-xs font-medium"
                    :class="
                      selectedKey === item.key
                        ? 'text-primary-100'
                        : String(item.row.status) === 'active'
                          ? 'text-muted'
                          : 'text-amber-600'
                    "
                  >
                    {{ displayValuePreview(item.row) }}
                    <span v-if="String(item.row.status) !== 'active'">
                      · {{ $t("status.inactive") }}
                    </span>
                  </p>
                </button>
              </li>
            </ul>
          </div>
        </nav>
      </section>

      <!-- 右：编辑表单（手机未选中时隐藏） -->
      <section
        class="min-h-0 min-w-0 flex-col overflow-hidden bg-surface"
        :class="mobileShowEditor || !isMobile ? 'flex' : 'hidden md:flex'"
      >
        <header class="shrink-0 border-b border-border px-4 py-3 sm:px-5">
          <div class="flex items-center gap-2">
            <button
              v-if="isMobile"
              type="button"
              class="rounded-lg p-1.5 text-muted hover:bg-surface-muted md:hidden"
              :aria-label="$t('common.back')"
              @click="backToList"
            >
              <ArrowLeftIcon class="h-5 w-5" />
            </button>
            <h2 class="text-sm font-semibold text-foreground">
              {{ $t("pages.systemSettings.editorColumnTitle") }}
            </h2>
          </div>
        </header>

        <template v-if="hasSelection">
          <div class="shrink-0 border-b border-border px-4 py-3 sm:px-5 sm:py-4">
            <h3 class="text-base font-semibold text-foreground sm:text-lg">{{ selectedLabel }}</h3>
            <p v-if="selectedHint" class="mt-2 text-sm leading-relaxed text-muted">{{ selectedHint }}</p>
          </div>

          <div class="min-h-0 flex-1 overflow-y-auto px-4 py-3 sm:px-5 sm:py-4">
            <div class="flex flex-col gap-4">
              <AdminFormField :label="$t('common.status')">
                <AdminSelect v-model="form.status" :options="statusOptions" />
                <p class="mt-1 text-xs text-muted">{{ $t("pages.systemSettings.statusHint") }}</p>
              </AdminFormField>

              <AdminFormField v-if="form.valueType === 'bool'" :label="$t('common.value')">
                <AdminSelect v-model="form.value" :options="boolValueOptions" />
              </AdminFormField>
              <AdminFormField v-else-if="form.key.startsWith('media.')" :label="$t('common.value')">
                <AdminSelect v-model="form.value" :options="storageOptionsForKey(form.key)" />
              </AdminFormField>
              <AdminFormField v-else :label="$t('common.value')">
                <AdminInput v-model="form.value" />
              </AdminFormField>

              <AdminFormField :label="$t('common.remark')">
                <AdminInput v-model="form.description" type="textarea" :rows="2" />
              </AdminFormField>

              <AdminFormField :label="$t('common.updatedAt')">
                <p class="text-sm text-muted">{{ formatDateTime(form.updatedAt) }}</p>
              </AdminFormField>
            </div>
          </div>

          <div class="flex shrink-0 justify-end border-t border-border px-4 py-3 sm:px-5 sm:py-4">
            <AdminButton variant="primary" :loading="saving" class="w-full sm:w-auto" @click="submit">
              {{ $t("common.save") }}
            </AdminButton>
          </div>
        </template>

        <div
          v-else-if="missingRow"
          class="flex min-h-0 flex-1 items-center justify-center px-8 text-center text-sm text-muted"
        >
          {{ $t("pages.systemSettings.rowMissing") }}
        </div>

        <div v-else class="flex min-h-0 flex-1 items-center justify-center text-sm text-muted">
          {{ $t("pages.systemSettings.selectItem") }}
        </div>
      </section>
    </div>
  </div>
</template>
