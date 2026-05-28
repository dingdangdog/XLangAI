<script setup lang="ts">
import type { ServiceProbeResult } from "~/composables/useServiceConfigProbe";

defineProps<{
  loading?: boolean;
  result?: ServiceProbeResult | null;
  disabled?: boolean;
  hint?: string;
}>();

const emit = defineEmits<{ probe: [] }>();
</script>

<template>
  <div class="rounded-xl border border-border bg-surface-muted/40 p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h4 class="text-sm font-medium text-foreground">{{ $t("serviceConfig.probeTitle") }}</h4>
        <p class="mt-1 text-xs text-muted">
          {{ hint ?? $t("serviceConfig.probeHint") }}
        </p>
      </div>
      <AdminButton variant="secondary" size="sm" :loading="loading" :disabled="disabled" @click="emit('probe')">
        {{ $t("serviceConfig.probeAction") }}
      </AdminButton>
    </div>

    <div
      v-if="result"
      class="mt-3 rounded-lg border px-3 py-2.5 text-sm"
      :class="
        result.ok
          ? 'border-primary-300/70 bg-primary-50/60 text-primary-900 dark:border-primary-500/40 dark:bg-primary-950/30 dark:text-primary-100'
          : 'border-danger-300/70 bg-danger-50/50 text-danger-800 dark:border-danger-500/40 dark:bg-danger-950/20 dark:text-danger-100'
      "
    >
      <div class="flex flex-wrap items-center gap-2">
        <span class="font-medium">{{ result.ok ? $t("serviceConfig.probeOk") : $t("serviceConfig.probeFailed") }}</span>
        <span class="text-xs tabular-nums opacity-80">{{ result.latencyMs }}ms</span>
      </div>
      <p class="mt-1">{{ result.message }}</p>
      <p v-if="result.detail" class="mt-1 break-all font-mono text-xs opacity-80">{{ result.detail }}</p>
    </div>
  </div>
</template>
