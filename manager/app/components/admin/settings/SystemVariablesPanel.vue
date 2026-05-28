<script setup lang="ts">
import { ChevronDownIcon, ChevronRightIcon } from "@heroicons/vue/24/outline";
import { SYSTEM_SETTING_GROUPS } from "~/utils/systemSettingsUi";

const { t, te } = useI18n();
const route = useRoute();
const router = useRouter();
const API = "/api/admin/system-settings";
const api = useAdminResourceApi(API);
const toast = useToast();

const list = ref<Record<string, unknown>[]>([]);
const loading = ref(false);
const saving = ref(false);

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

function labelForKey(k: string) {
  const i18nKey = `pages.systemSettings.keys.${k}`;
  return te(i18nKey) ? t(i18nKey) : k;
}

function hintForKey(k: string) {
  const i18nKey = `pages.systemSettings.hints.${k}`;
  return te(i18nKey) ? t(i18nKey) : "";
}

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

const selectedKey = computed({
  get: () => String(route.query.key ?? ""),
  set: (key: string) => {
    router.replace({ query: { ...route.query, key: key || undefined } });
  },
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
}

function toggleGroup(id: string) {
  expandedGroups.value[id] = !expandedGroups.value[id];
}

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

const selectedRow = computed(() => byKey.value.get(selectedKey.value) ?? null);
const hasSelection = computed(() => !!selectedRow.value && !!form.id);

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
  <div class="flex min-h-0 flex-1 gap-4">
    <aside
      class="flex w-full shrink-0 flex-col gap-2 overflow-hidden rounded-xl border border-border bg-surface lg:w-72"
    >
      <div class="border-b border-border px-3 py-2.5">
        <p class="text-sm font-medium text-foreground">{{ $t("pages.systemSettings.menuTitle") }}</p>
        <p class="mt-0.5 text-xs text-muted">{{ $t("pages.systemSettings.description") }}</p>
      </div>

      <div v-if="loading && !list.length" class="space-y-2 p-3">
        <AdminSkeleton v-for="i in 6" :key="i" class="h-10 rounded-lg" />
      </div>

      <nav v-else class="min-h-0 flex-1 overflow-y-auto p-2">
        <div v-for="group in SYSTEM_SETTING_GROUPS" :key="group.id" class="mb-1">
          <button
            type="button"
            class="flex w-full items-start gap-1 rounded-lg px-2 py-2 text-left hover:bg-surface-muted"
            @click="toggleGroup(group.id)"
          >
            <ChevronDownIcon v-if="expandedGroups[group.id]" class="mt-0.5 h-4 w-4 shrink-0 text-muted" />
            <ChevronRightIcon v-else class="mt-0.5 h-4 w-4 shrink-0 text-muted" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium text-foreground">{{ $t(group.labelKey) }}</p>
              <p class="text-xs text-muted">{{ $t(group.descriptionKey) }}</p>
            </div>
          </button>

          <ul v-show="expandedGroups[group.id]" class="ml-5 space-y-0.5 border-l border-border pl-2">
            <li v-for="key in group.keys" :key="key">
              <button
                type="button"
                class="w-full rounded-lg px-2 py-2 text-left transition-colors"
                :class="
                  selectedKey === key
                    ? 'bg-primary-50 text-primary-700 dark:bg-primary-950/40 dark:text-primary-300'
                    : 'hover:bg-surface-muted text-foreground'
                "
                @click="selectKey(key)"
              >
                <p class="text-sm font-medium leading-snug">{{ labelForKey(key) }}</p>
                <p v-if="hintForKey(key)" class="mt-0.5 line-clamp-2 text-xs text-muted">
                  {{ hintForKey(key) }}
                </p>
                <p
                  v-if="byKey.get(key)"
                  class="mt-1 truncate text-xs"
                  :class="String(byKey.get(key)?.status) === 'active' ? 'text-muted' : 'text-warning-600'"
                >
                  {{ displayValuePreview(byKey.get(key)!) }}
                  <span v-if="String(byKey.get(key)?.status) !== 'active'" class="ml-1">
                    · {{ $t("status.inactive") }}
                  </span>
                </p>
              </button>
            </li>
          </ul>
        </div>
      </nav>
    </aside>

    <main class="min-h-0 min-w-0 flex-1">
      <AdminPanel v-if="hasSelection" class="flex min-h-0 flex-1 flex-col">
        <div class="border-b border-border px-4 py-3">
          <h3 class="text-base font-semibold text-foreground">{{ labelForKey(form.key) }}</h3>
          <p v-if="hintForKey(form.key)" class="mt-1 text-sm text-muted">{{ hintForKey(form.key) }}</p>
          <code class="mt-2 inline-block rounded bg-surface-muted px-2 py-0.5 text-xs">{{ form.key }}</code>
        </div>

        <div class="flex flex-1 flex-col gap-4 p-4">
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

          <div class="mt-auto flex justify-end border-t border-border pt-4">
            <AdminButton variant="primary" :loading="saving" @click="submit">
              {{ $t("common.save") }}
            </AdminButton>
          </div>
        </div>
      </AdminPanel>

      <AdminPanel
        v-else
        class="flex min-h-[240px] flex-1 items-center justify-center text-sm text-muted"
      >
        {{ $t("pages.systemSettings.selectItem") }}
      </AdminPanel>
    </main>
  </div>
</template>
