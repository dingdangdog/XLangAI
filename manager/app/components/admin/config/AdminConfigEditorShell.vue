<script setup lang="ts">
defineProps<{
  title: string;
  subtitle?: string;
  editorTab: "connection" | "advanced" | "raw";
  saving?: boolean;
  deleting?: boolean;
  activating?: boolean;
  canActivate?: boolean;
  canDelete?: boolean;
  dirty?: boolean;
}>();

const emit = defineEmits<{
  "update:editorTab": ["connection" | "advanced" | "raw"];
  save: [];
  cancel: [];
  delete: [];
  activate: [];
}>();

const tabs = computed(() => [
  { key: "connection" as const, label: "serviceConfig.tabs.connection" },
  { key: "advanced" as const, label: "serviceConfig.tabs.advanced" },
  { key: "raw" as const, label: "serviceConfig.tabs.raw" },
]);
</script>

<template>
  <AdminPanel :fill="false" class="flex min-h-[520px] flex-col">
    <div class="border-b border-border px-4 py-3">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h3 class="text-base font-semibold text-foreground">{{ title }}</h3>
          <p v-if="subtitle" class="mt-0.5 text-sm text-muted">{{ subtitle }}</p>
        </div>
        <div class="flex flex-wrap gap-2">
          <AdminButton v-if="canActivate" variant="secondary" size="sm" :loading="activating" @click="emit('activate')">
            {{ $t("common.enable") }}
          </AdminButton>
          <AdminButton v-if="canDelete" variant="ghost" size="sm" class="!text-danger-600" :loading="deleting" @click="emit('delete')">
            {{ $t("common.delete") }}
          </AdminButton>
        </div>
      </div>

      <AdminTabBar
        class="mt-3"
        :tabs="tabs.map((t) => ({ key: t.key, label: $t(t.label) }))"
        :active-tab="editorTab"
        @change="emit('update:editorTab', $event as 'connection' | 'advanced' | 'raw')"
      />
    </div>

    <div class="flex-1 overflow-y-auto px-4 py-4">
      <div v-show="editorTab === 'connection'">
        <slot name="connection" />
      </div>
      <div v-show="editorTab === 'advanced'">
        <slot name="advanced" />
      </div>
      <div v-show="editorTab === 'raw'">
        <slot name="raw" />
      </div>
      <div class="mt-4">
        <slot name="probe" />
      </div>
    </div>

    <div class="flex items-center justify-end gap-2 border-t border-border px-4 py-3">
      <AdminButton @click="emit('cancel')">{{ $t("common.cancel") }}</AdminButton>
      <AdminButton variant="primary" :loading="saving" :disabled="!dirty" @click="emit('save')">
        {{ $t("serviceConfig.saveAndApply") }}
      </AdminButton>
    </div>
  </AdminPanel>
</template>
