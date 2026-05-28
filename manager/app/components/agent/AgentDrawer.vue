<script setup lang="ts">
defineProps<{
  open: boolean;
  title: string;
  subtitle?: string;
  wide?: boolean;
}>();

const emit = defineEmits<{ close: [] }>();
</script>

<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div v-if="open" class="fixed inset-0 z-[100] flex justify-end">
        <button
          type="button"
          class="absolute inset-0 bg-black/35 backdrop-blur-[2px]"
          aria-label="close"
          @click="emit('close')"
        />

        <Transition
          appear
          enter-active-class="transition duration-250 ease-out"
          enter-from-class="translate-x-full"
          enter-to-class="translate-x-0"
          leave-active-class="transition duration-200 ease-in"
          leave-from-class="translate-x-0"
          leave-to-class="translate-x-full"
        >
          <div
            v-if="open"
            class="relative flex h-full w-full flex-col border-l border-border/80 bg-surface shadow-2xl"
            :class="wide ? 'max-w-2xl' : 'max-w-lg'"
            role="dialog"
            aria-modal="true"
          >
            <header class="shrink-0 border-b border-border/60 px-6 py-5">
              <div class="flex items-start justify-between gap-4">
                <div class="min-w-0">
                  <h2 class="truncate text-lg font-semibold tracking-tight text-foreground">{{ title }}</h2>
                  <p v-if="subtitle" class="mt-1 text-sm text-muted">{{ subtitle }}</p>
                </div>
                <button
                  type="button"
                  class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-muted transition hover:bg-surface-muted hover:text-foreground"
                  @click="emit('close')"
                >
                  <span class="text-xl leading-none">&times;</span>
                </button>
              </div>
              <slot name="header-extra" />
            </header>

            <div class="min-h-0 flex-1 overflow-y-auto px-6 py-5">
              <slot />
            </div>

            <footer
              v-if="$slots.footer"
              class="shrink-0 border-t border-border/60 bg-surface-muted/30 px-6 py-4"
            >
              <slot name="footer" />
            </footer>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>
