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
  <div class="flex w-full flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:justify-between sm:gap-3">
    <AdminButton
      v-if="showDelete && editorMode === 'edit'"
      variant="ghost"
      class="!text-danger-600 w-full sm:w-auto"
      :loading="deleting"
      @click="emit('delete')"
    >
      {{ t("common.delete") }}
    </AdminButton>
    <div v-else class="hidden sm:block" />
    <div class="flex w-full flex-col-reverse gap-2 sm:w-auto sm:flex-row">
      <AdminButton class="w-full sm:w-auto" @click="emit('close')">{{ t("common.cancel") }}</AdminButton>
      <AdminButton
        v-if="setupPhase === 'vendor'"
        variant="primary"
        class="w-full sm:w-auto"
        @click="emit('next')"
      >
        {{ t("agentHub.continue") }}
      </AdminButton>
      <AdminButton
        v-else
        variant="primary"
        class="w-full sm:w-auto"
        :loading="saving"
        :disabled="editorMode === 'edit' && !dirty"
        @click="emit('save')"
      >
        {{ t("agentHub.saveAndEnable") }}
      </AdminButton>
    </div>
  </div>
</template>
