<script setup lang="ts">
import { computed, ref } from "vue";
import { EyeIcon, EyeSlashIcon } from "@heroicons/vue/24/outline";
import { maskKey } from "~/utils/maskKey";

const props = defineProps<{
  value?: string | null;
}>();

const revealed = ref(false);

const hasValue = computed(() => !!String(props.value ?? "").trim());

const display = computed(() => {
  const v = String(props.value ?? "").trim();
  if (!v) return "—";
  return revealed.value ? v : maskKey(v);
});
</script>

<template>
  <span class="inline-flex max-w-full items-center gap-1">
    <span class="break-all font-mono text-xs">{{ display }}</span>
    <button
      v-if="hasValue"
      type="button"
      class="shrink-0 rounded p-0.5 text-muted hover:bg-surface-muted hover:text-foreground"
      :aria-label="revealed ? '隐藏密钥' : '显示完整密钥'"
      :title="revealed ? '隐藏' : '显示完整密钥'"
      @click="revealed = !revealed"
    >
      <EyeSlashIcon v-if="revealed" class="h-3.5 w-3.5" />
      <EyeIcon v-else class="h-3.5 w-3.5" />
    </button>
  </span>
</template>
