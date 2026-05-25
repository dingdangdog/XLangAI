<script setup lang="ts">
import { XMarkIcon } from "@heroicons/vue/24/outline";

const { visible, options, resolveConfirm, loading } = useConfirm();
</script>

<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="visible"
        class="fixed inset-0 z-[90] flex items-center justify-center bg-black/50 p-4"
        @click.self="resolveConfirm(false)"
      >
        <div
          class="w-full max-w-md rounded-2xl border border-border bg-surface p-6 shadow-xl"
          @click.stop
        >
          <div class="mb-4 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-foreground">{{ options.title }}</h3>
            <button
              type="button"
              class="rounded-lg p-1 text-muted hover:bg-surface-muted hover:text-foreground"
              @click="resolveConfirm(false)"
            >
              <XMarkIcon class="h-5 w-5" />
            </button>
          </div>
          <p class="text-sm text-muted">{{ options.message }}</p>
          <div class="mt-6 flex justify-end gap-2">
            <button
              type="button"
              class="rounded-lg border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-surface-muted"
              :disabled="loading"
              @click="resolveConfirm(false)"
            >
              {{ options.cancelLabel }}
            </button>
            <button
              type="button"
              class="rounded-lg px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
              :class="
                options.danger
                  ? 'bg-danger-600 hover:bg-danger-500'
                  : 'bg-primary-600 hover:bg-primary-700'
              "
              :disabled="loading"
              @click="resolveConfirm(true)"
            >
              {{ options.confirmLabel }}
            </button>
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
