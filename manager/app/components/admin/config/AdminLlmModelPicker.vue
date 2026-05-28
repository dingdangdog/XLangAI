<script setup lang="ts">
const props = defineProps<{
  modelValue: string;
  options: { value: string; label: string }[];
  loading?: boolean;
  manualOnly?: boolean;
  error?: string | null;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
  refresh: [];
}>();

const { t } = useI18n();
const manual = ref(false);

watch(
  () => [props.manualOnly, props.options.length] as const,
  ([m, n]) => {
    manual.value = m || n === 0;
  },
  { immediate: true },
);

const selectOptions = computed(() => {
  const opts = props.options.map((o) => ({ value: o.value, label: o.label }));
  if (props.modelValue && !opts.some((o) => o.value === props.modelValue)) {
    opts.unshift({ value: props.modelValue, label: props.modelValue });
  }
  return opts;
});
</script>

<template>
  <div class="space-y-2">
    <div class="flex gap-2">
      <div class="min-w-0 flex-1">
        <AdminSelect
          v-if="!manual && selectOptions.length"
          :model-value="modelValue"
          :options="selectOptions"
          @update:model-value="emit('update:modelValue', $event)"
        />
        <AdminInput
          v-else
          :model-value="modelValue"
          :placeholder="t('pages.llmConfigs.modelPlaceholder')"
          @update:model-value="emit('update:modelValue', $event)"
        />
      </div>
      <button
        type="button"
        class="shrink-0 rounded-xl bg-surface-muted px-3 py-2 text-xs font-medium text-foreground transition hover:bg-primary-50 hover:text-primary-700 disabled:opacity-40"
        :disabled="disabled || loading"
        @click="emit('refresh')"
      >
        {{ loading ? "…" : t("serverCatalog.llm.fetchModels") }}
      </button>
    </div>
    <p v-if="error" class="text-xs text-danger-500">{{ error }}</p>
    <label v-if="!manualOnly && selectOptions.length" class="inline-flex items-center gap-2 text-xs text-muted">
      <input v-model="manual" type="checkbox" class="rounded border-border" />
      {{ t("serverCatalog.llm.manualModelInput") }}
    </label>
  </div>
</template>
