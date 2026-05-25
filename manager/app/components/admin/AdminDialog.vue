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
        class="fixed inset-0 z-[80] flex items-center justify-center bg-black/50 p-4"
      >
        <div
          class="flex max-h-[90vh] w-full flex-col rounded-2xl border border-border bg-surface shadow-xl"
          :class="widths[width ?? 'md']"
          @click.stop
        >
          <div
            class="flex shrink-0 items-center justify-between border-b border-border px-5 py-4"
          >
            <h3 class="text-lg font-semibold text-foreground">{{ title }}</h3>
            <button
              type="button"
              class="rounded-lg p-1 text-muted hover:bg-surface-muted"
              @click="close"
            >
              <XMarkIcon class="h-5 w-5" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto px-5 py-4">
            <slot />
          </div>
          <div
            v-if="$slots.footer"
            class="flex shrink-0 justify-end gap-2 border-t border-border px-5 py-4"
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
