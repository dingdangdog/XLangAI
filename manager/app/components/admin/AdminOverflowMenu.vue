<script setup lang="ts">
import { EllipsisVerticalIcon } from "@heroicons/vue/24/outline";

export type AdminMenuAction = {
  label: string;
  onClick: () => void;
  danger?: boolean;
  disabled?: boolean;
};

const props = defineProps<{
  actions: AdminMenuAction[];
}>();

const open = ref(false);
const root = ref<HTMLElement | null>(null);

function toggle() {
  open.value = !open.value;
}

function run(action: AdminMenuAction) {
  if (action.disabled) return;
  open.value = false;
  action.onClick();
}

function onDocClick(e: MouseEvent) {
  if (!open.value || !root.value) return;
  if (!root.value.contains(e.target as Node)) open.value = false;
}

onMounted(() => document.addEventListener("click", onDocClick));
onUnmounted(() => document.removeEventListener("click", onDocClick));

const visibleActions = computed(() => props.actions.filter(Boolean));
</script>

<template>
  <div ref="root" class="relative">
    <button
      type="button"
      class="flex h-8 w-8 items-center justify-center rounded-lg text-muted hover:bg-surface-muted hover:text-foreground"
      :aria-label="$t('common.actions')"
      :aria-expanded="open"
      @click.stop="toggle"
    >
      <EllipsisVerticalIcon class="h-5 w-5" />
    </button>
    <div
      v-if="open"
      class="absolute right-0 z-20 mt-1 min-w-[9.5rem] overflow-hidden rounded-xl border border-border bg-surface py-1 shadow-lg"
      role="menu"
    >
      <button
        v-for="(action, i) in visibleActions"
        :key="i"
        type="button"
        role="menuitem"
        class="flex w-full px-3 py-2 text-left text-sm transition-colors disabled:opacity-40"
        :class="
          action.danger
            ? 'text-danger-600 hover:bg-danger-50 dark:hover:bg-danger-50/10'
            : 'text-foreground hover:bg-surface-muted'
        "
        :disabled="action.disabled"
        @click.stop="run(action)"
      >
        {{ action.label }}
      </button>
    </div>
  </div>
</template>
