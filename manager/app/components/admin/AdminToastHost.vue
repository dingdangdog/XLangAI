<script setup lang="ts">
import {
  CheckCircleIcon,
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  XMarkIcon,
} from "@heroicons/vue/24/outline";

const { toasts, dismiss } = useToast();

const typeStyles: Record<string, string> = {
  success: "border-primary-200 bg-primary-50 text-primary-800 dark:border-primary-800 dark:bg-primary-950/40 dark:text-primary-200",
  error: "border-danger-500/30 bg-danger-50 text-danger-600 dark:bg-danger-50/10 dark:text-danger-500",
  warning: "border-amber-200 bg-amber-50 text-amber-800 dark:border-amber-800 dark:bg-amber-950/30 dark:text-amber-200",
  info: "border-border bg-surface text-foreground",
};

const icons = {
  success: CheckCircleIcon,
  error: ExclamationCircleIcon,
  warning: ExclamationTriangleIcon,
  info: InformationCircleIcon,
};
</script>

<template>
  <div
    class="pointer-events-none fixed right-4 top-4 z-[100] flex w-full max-w-sm flex-col gap-2"
    aria-live="polite"
  >
    <TransitionGroup name="toast">
      <div
        v-for="t in toasts"
        :key="t.id"
        class="pointer-events-auto flex items-start gap-2 rounded-xl border px-4 py-3 text-sm shadow-lg"
        :class="typeStyles[t.type]"
      >
        <component :is="icons[t.type]" class="mt-0.5 h-5 w-5 shrink-0" />
        <span class="flex-1">{{ t.message }}</span>
        <button
          type="button"
          class="shrink-0 rounded p-0.5 opacity-70 hover:opacity-100"
          @click="dismiss(t.id)"
        >
          <XMarkIcon class="h-4 w-4" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.25s ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(1rem);
}
</style>
