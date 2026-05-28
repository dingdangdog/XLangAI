<script setup lang="ts">
defineProps<{
  editorMode: "create" | "edit";
  setupPhase: "vendor" | "connect";
  setupStepNumber: number;
  saving: boolean;
  deleting: boolean;
  dirty: boolean;
  showDelete?: boolean;
}>();

const emit = defineEmits<{
  close: [];
  back: [];
  next: [];
  save: [];
  delete: [];
}>();

const { t } = useI18n();
</script>

<template>
  <div class="flex w-full flex-wrap items-center justify-between gap-3">
    <AdminButton
      v-if="showDelete && editorMode === 'edit'"
      variant="ghost"
      class="!text-danger-600"
      :loading="deleting"
      @click="emit('delete')"
    >
      {{ t("common.delete") }}
    </AdminButton>
    <div v-else />
    <div class="flex gap-2">
      <AdminButton @click="emit('close')">{{ t("common.cancel") }}</AdminButton>
      <AdminButton v-if="setupPhase === 'vendor'" variant="primary" @click="emit('next')">
        {{ t("agentHub.continue") }}
      </AdminButton>
      <AdminButton
        v-else
        variant="primary"
        :loading="saving"
        :disabled="editorMode === 'edit' && !dirty"
        @click="emit('save')"
      >
        {{ t("agentHub.saveAndEnable") }}
      </AdminButton>
    </div>
  </div>
</template>
