<script setup lang="ts">
import type { ConfigFieldSchema } from "~/utils/serviceConfigSchemas";

const props = defineProps<{
  schema: ConfigFieldSchema[];
  modelValue: Record<string, string | number | boolean>;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: Record<string, string | number | boolean>];
}>();

const { t } = useI18n();

function updateField(key: string, value: string | number | boolean) {
  emit("update:modelValue", { ...props.modelValue, [key]: value });
}
</script>

<template>
  <div v-if="schema.length" class="grid gap-1 sm:grid-cols-2">
    <AdminFormField
      v-for="field in schema"
      :key="field.key"
      :label="t(field.labelKey)"
      :required="field.required"
      :hint="field.hintKey ? t(field.hintKey) : undefined"
    >
      <AdminSelect
        v-if="field.type === 'select'"
        :model-value="String(modelValue[field.key] ?? '')"
        :options="field.options ?? []"
        @update:model-value="updateField(field.key, $event)"
      />
      <label
        v-else-if="field.type === 'boolean'"
        class="inline-flex cursor-pointer items-center gap-2 text-sm text-foreground"
      >
        <input
          type="checkbox"
          class="h-4 w-4 rounded border-border text-primary-600 focus:ring-primary-500"
          :checked="Boolean(modelValue[field.key])"
          @change="updateField(field.key, ($event.target as HTMLInputElement).checked)"
        />
      </label>
      <AdminInput
        v-else
        :model-value="modelValue[field.key] ?? ''"
        :type="field.type === 'number' ? 'number' : 'text'"
        :placeholder="field.placeholder"
        @update:model-value="updateField(field.key, field.type === 'number' ? Number($event) : $event)"
      />
    </AdminFormField>
  </div>
  <p v-else class="text-sm text-muted">{{ $t("serviceConfig.noAdvancedFields") }}</p>
</template>
