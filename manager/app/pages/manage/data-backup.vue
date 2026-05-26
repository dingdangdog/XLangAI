<script setup lang="ts">
type BackupInfo = {
  format: string;
  version: number;
  totalRows: number;
  tables: Record<string, number>;
};

type ImportResult = {
  mode: "merge" | "replace";
  imported: Record<string, number>;
  totalImported: number;
  backupExportedAt: string;
};

const { t } = useI18n();
const toast = useToast();
const { confirm } = useConfirm();

const loading = ref(false);
const exporting = ref(false);
const importing = ref(false);
const info = ref<BackupInfo | null>(null);
const importMode = ref<"merge" | "replace">("merge");
const selectedFile = ref<File | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);

const tableRows = computed(() => {
  if (!info.value) return [];
  return Object.entries(info.value.tables).map(([key, count]) => ({
    key,
    label: t(`pages.dataBackup.tableLabels.${key}`, key),
    count,
  }));
});

async function loadInfo() {
  loading.value = true;
  try {
    info.value = await $fetch<BackupInfo>("/api/admin/database-backup/info");
  } catch (error) {
    toast.error(t("toast.loadFailed"));
    console.error(error);
  } finally {
    loading.value = false;
  }
}

async function exportBackup() {
  exporting.value = true;
  try {
    const response = await fetch("/api/admin/database-backup/export", {
      credentials: "include",
    });
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }

    const blob = await response.blob();
    const disposition = response.headers.get("Content-Disposition") ?? "";
    const match = disposition.match(/filename="([^"]+)"/);
    const filename = match?.[1] ?? `xlangai-db-backup-${Date.now()}.json`;

    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    URL.revokeObjectURL(url);

    toast.success(t("toast.dbBackupExported"));
  } catch (error) {
    toast.error(t("toast.dbBackupExportFailed"));
    console.error(error);
  } finally {
    exporting.value = false;
  }
}

function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  selectedFile.value = input.files?.[0] ?? null;
}

function clearFile() {
  selectedFile.value = null;
  if (fileInputRef.value) {
    fileInputRef.value.value = "";
  }
}

async function importBackup() {
  if (!selectedFile.value) {
    toast.error(t("toast.dbBackupFileRequired"));
    return;
  }

  const message =
    importMode.value === "replace"
      ? t("confirm.importDbBackupReplace")
      : t("confirm.importDbBackupMerge");

  const ok = await confirm({
    message,
    danger: importMode.value === "replace",
  });
  if (!ok) return;

  importing.value = true;
  try {
    const form = new FormData();
    form.append("file", selectedFile.value);
    form.append("mode", importMode.value);

    const result = await $fetch<ImportResult>("/api/admin/database-backup/import", {
      method: "POST",
      body: form,
    });

    toast.success(
      t("toast.dbBackupImported", {
        count: result.totalImported,
        mode: t(`pages.dataBackup.mode.${result.mode}`),
      }),
    );
    clearFile();
    await loadInfo();
  } catch (error) {
    toast.error(t("toast.dbBackupImportFailed"));
    console.error(error);
  } finally {
    importing.value = false;
  }
}

onMounted(() => {
  void loadInfo();
});
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.dataBackup.title')"
        :description="$t('pages.dataBackup.description')"
      />

      <AdminAlert :title="$t('pages.dataBackup.alertTitle')">
        {{ $t("pages.dataBackup.alertBody") }}
      </AdminAlert>
    </template>

    <AdminSkeleton v-if="loading" :rows="10" />

    <div v-else class="flex min-h-0 flex-1 flex-col gap-4">
      <div class="grid shrink-0 gap-4 lg:grid-cols-2">
        <AdminPanel>
          <div class="flex h-full flex-col p-4 md:p-5">
            <h2 class="text-base font-semibold text-foreground">
              {{ $t("pages.dataBackup.exportTitle") }}
            </h2>
            <p class="mt-1 text-sm text-muted">{{ $t("pages.dataBackup.exportHint") }}</p>
            <p class="mt-3 text-sm text-foreground">
              {{ $t("pages.dataBackup.totalRows", { total: info?.totalRows ?? 0 }) }}
            </p>
            <div class="mt-4 flex flex-wrap gap-3">
              <AdminButton variant="primary" :loading="exporting" @click="exportBackup">
                {{ $t("pages.dataBackup.export") }}
              </AdminButton>
            </div>
          </div>
        </AdminPanel>

        <AdminPanel>
          <div class="flex h-full flex-col p-4 md:p-5">
            <h2 class="text-base font-semibold text-foreground">
              {{ $t("pages.dataBackup.importTitle") }}
            </h2>
            <p class="mt-1 text-sm text-muted">{{ $t("pages.dataBackup.importHint") }}</p>

            <div class="mt-4 space-y-4">
              <AdminFormField :label="$t('pages.dataBackup.importMode')">
                <div class="space-y-2">
                  <label class="flex items-start gap-2 text-sm">
                    <input v-model="importMode" type="radio" value="merge" class="mt-1" />
                    <span>
                      <span class="font-medium text-foreground">
                        {{ $t("pages.dataBackup.mode.merge") }}
                      </span>
                      <span class="mt-0.5 block text-muted">
                        {{ $t("pages.dataBackup.mode.mergeHint") }}
                      </span>
                    </span>
                  </label>
                  <label class="flex items-start gap-2 text-sm">
                    <input v-model="importMode" type="radio" value="replace" class="mt-1" />
                    <span>
                      <span class="font-medium text-foreground">
                        {{ $t("pages.dataBackup.mode.replace") }}
                      </span>
                      <span class="mt-0.5 block text-muted">
                        {{ $t("pages.dataBackup.mode.replaceHint") }}
                      </span>
                    </span>
                  </label>
                </div>
              </AdminFormField>

              <AdminFormField :label="$t('pages.dataBackup.selectFile')">
                <input
                  ref="fileInputRef"
                  type="file"
                  accept="application/json,.json"
                  class="block w-full text-sm text-foreground file:mr-3 file:rounded-md file:border file:border-border file:bg-surface-muted file:px-3 file:py-1.5 file:text-sm file:font-medium"
                  @change="onFileChange"
                />
                <p v-if="selectedFile" class="mt-2 text-xs text-muted">
                  {{ selectedFile.name }} ({{ Math.ceil(selectedFile.size / 1024) }} KB)
                </p>
              </AdminFormField>

              <div class="flex flex-wrap gap-3">
                <AdminButton
                  variant="primary"
                  :loading="importing"
                  :disabled="!selectedFile"
                  @click="importBackup"
                >
                  {{ $t("pages.dataBackup.import") }}
                </AdminButton>
                <AdminButton
                  variant="secondary"
                  :disabled="!selectedFile || importing"
                  @click="clearFile"
                >
                  {{ $t("common.clear") }}
                </AdminButton>
              </div>
            </div>
          </div>
        </AdminPanel>
      </div>

      <AdminPanel class="flex min-h-0 min-w-0 flex-1 flex-col">
        <div
          class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-border p-4 md:px-5 md:py-4"
        >
          <div>
            <h2 class="text-base font-semibold text-foreground">
              {{ $t("pages.dataBackup.currentData") }}
            </h2>
            <p class="mt-1 text-sm text-muted">{{ $t("pages.dataBackup.currentDataHint") }}</p>
          </div>
          <AdminButton variant="secondary" @click="loadInfo">
            {{ $t("common.refresh") }}
          </AdminButton>
        </div>

        <AdminTable>
          <template #head>
            <AdminTh>{{ $t("pages.dataBackup.tableName") }}</AdminTh>
            <AdminTh width="120px">{{ $t("pages.dataBackup.rowCount") }}</AdminTh>
          </template>
          <AdminTr v-for="row in tableRows" :key="row.key">
            <AdminTd>{{ row.label }}</AdminTd>
            <AdminTd>{{ row.count }}</AdminTd>
          </AdminTr>
        </AdminTable>
      </AdminPanel>
    </div>
  </AdminListPage>
</template>
