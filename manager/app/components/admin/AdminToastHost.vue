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
  <Teleport to="body">
    <div
      class="pointer-events-none fixed inset-x-3 top-[max(0.75rem,env(safe-area-inset-top))] z-[100] flex max-w-sm flex-col gap-2 sm:inset-x-auto sm:right-4 sm:top-4"
      aria-live="polite"
    >
      <TransitionGroup name="toast" tag="div" class="flex flex-col gap-2">
        <div
          v-for="item in toasts"
          :key="item.id"
          class="pointer-events-auto relative flex items-start gap-2 rounded-xl border px-4 py-3 text-sm shadow-lg"
          :class="typeStyles[item.type]"
        >
          <component :is="icons[item.type]" class="mt-0.5 h-5 w-5 shrink-0" />
          <span class="flex-1">{{ item.message }}</span>
          <button
            type="button"
            class="relative z-10 -mr-1 shrink-0 rounded p-1 opacity-70 hover:opacity-100"
            aria-label="关闭"
            @click.stop="dismiss(item.id)"
          >
            <XMarkIcon class="h-4 w-4" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-move,
.toast-enter-active,
.toast-leave-active {
  transition: all 0.25s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(1rem);
}

.toast-leave-active {
  position: absolute;
  right: 0;
  left: 0;
}
</style>
