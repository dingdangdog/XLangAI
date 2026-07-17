<script setup lang="ts">
import { XMarkIcon } from "@heroicons/vue/24/outline";

const model = defineModel<boolean>({ required: true });

defineProps<{
  title: string;
  width?: "sm" | "md" | "lg";
}>();

const widths = {
  sm: "max-w-md",
  md: "max-w-lg",
  lg: "max-w-2xl",
};

function close() {
  model.value = false;
}
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="model"
        class="fixed inset-0 z-[80] flex items-end justify-center bg-black/50 p-0 sm:items-center sm:p-4"
      >
        <div
          class="flex max-h-[min(96dvh,100%)] w-full flex-col rounded-t-2xl border border-border bg-surface shadow-xl sm:max-h-[90vh] sm:rounded-2xl"
          :class="widths[width ?? 'md']"
          @click.stop
        >
          <div
            class="flex shrink-0 items-center justify-between border-b border-border px-4 py-3 sm:px-5 sm:py-4"
          >
            <h3 class="pr-2 text-base font-semibold text-foreground sm:text-lg">{{ title }}</h3>
            <button
              type="button"
              class="rounded-lg p-1 text-muted hover:bg-surface-muted"
              :aria-label="$t('common.cancel')"
              @click="close"
            >
              <XMarkIcon class="h-5 w-5" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto px-4 py-3 sm:px-5 sm:py-4">
            <slot />
          </div>
          <div
            v-if="$slots.footer"
            class="flex shrink-0 flex-wrap justify-end gap-2 border-t border-border px-4 py-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] sm:px-5 sm:py-4"
          >
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
