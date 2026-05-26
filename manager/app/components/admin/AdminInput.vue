<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { EyeIcon, EyeSlashIcon } from "@heroicons/vue/24/outline";

const { t } = useI18n();

const model = defineModel<string | number | null>({ required: true });

const props = withDefaults(
  defineProps<{
    type?: "text" | "password" | "email" | "number" | "textarea";
    placeholder?: string;
    disabled?: boolean;
    rows?: number;
    autocomplete?: string;
    /** type=password 时是否显示「显示/隐藏」按钮 */
    revealable?: boolean;
  }>(),
  { type: "text", rows: 3, disabled: false, revealable: true }
);

const revealed = ref(false);

watch(
  () => props.type,
  (t) => {
    if (t !== "password") revealed.value = false;
  }
);

const canReveal = computed(() => props.type === "password" && props.revealable);

const effectiveType = computed(() => {
  if (canReveal.value && revealed.value) return "text";
  return props.type;
});

const inputClass =
  "w-full rounded-lg border border-border bg-surface-muted/50 px-3 py-2 text-sm text-foreground placeholder:text-muted focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 disabled:cursor-not-allowed disabled:opacity-60";

const fieldClass = computed(() => {
  const mono = props.type === "password" ? " font-mono" : "";
  const pad = canReveal.value ? " pr-10" : "";
  return inputClass + mono + pad;
});
</script>

<template>
  <textarea
    v-if="type === 'textarea'"
    v-model="model as string"
    :rows="rows"
    :placeholder="placeholder"
    :disabled="disabled"
    :class="inputClass"
  />
  <div v-else-if="canReveal" class="relative">
    <input
      v-model="model"
      :type="effectiveType"
      :placeholder="placeholder"
      :disabled="disabled"
      :autocomplete="autocomplete"
      :class="fieldClass"
    />
    <button
      type="button"
      class="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted hover:bg-surface-muted hover:text-foreground"
      :aria-label="revealed ? t('security.hide') : t('security.show')"
      :title="revealed ? t('security.hideKey') : t('security.showKey')"
      tabindex="-1"
      @click="revealed = !revealed"
    >
      <EyeSlashIcon v-if="revealed" class="h-4 w-4" />
      <EyeIcon v-else class="h-4 w-4" />
    </button>
  </div>
  <input
    v-else
    v-model="model"
    :type="type"
    :placeholder="placeholder"
    :disabled="disabled"
    :autocomplete="autocomplete"
    :class="inputClass"
  />
</template>
