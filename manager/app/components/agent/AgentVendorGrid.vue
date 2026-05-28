<script setup lang="ts">
import { vendorTheme } from "~/utils/vendorTheme";

export type VendorGridItem = {
  id: string;
  label: string;
  description?: string;
  group: string;
};

const props = defineProps<{
  items: VendorGridItem[];
  modelValue: string;
}>();

const emit = defineEmits<{ "update:modelValue": [id: string] }>();

const { t } = useI18n();
const query = ref("");

const grouped = computed(() => {
  const q = query.value.trim().toLowerCase();
  const filtered = props.items.filter(
    (i) =>
      !q ||
      i.label.toLowerCase().includes(q) ||
      i.id.toLowerCase().includes(q) ||
      i.group.toLowerCase().includes(q),
  );
  const map = new Map<string, VendorGridItem[]>();
  for (const item of filtered) {
    const g = item.group || "";
    if (!map.has(g)) map.set(g, []);
    map.get(g)!.push(item);
  }
  return [...map.entries()];
});
</script>

<template>
  <div class="space-y-4">
    <div class="relative">
      <input
        v-model="query"
        type="search"
        :placeholder="t('agentHub.searchVendor')"
        class="w-full rounded-xl border-0 bg-surface-muted/80 py-2.5 pl-4 pr-4 text-sm text-foreground placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-primary-500/40"
      />
    </div>

    <div v-for="[group, rows] in grouped" :key="group" class="space-y-2.5">
      <p v-if="group" class="text-xs font-medium uppercase tracking-wider text-muted">{{ group }}</p>
      <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
        <button
          v-for="item in rows"
          :key="item.id"
          type="button"
          class="group relative flex w-full items-center gap-3 rounded-2xl border px-3.5 py-3 text-left transition-all duration-200"
          :class="
            modelValue === item.id
              ? 'border-primary-400/80 bg-primary-50/50 shadow-md shadow-primary-500/10 ring-2 ring-primary-500/25 dark:bg-primary-950/20'
              : 'border-border/70 bg-surface hover:border-primary-300/60 hover:shadow-sm'
          "
          @click="emit('update:modelValue', item.id)"
        >
          <div
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br text-xs font-bold text-white shadow-sm"
            :class="vendorTheme(item.id).gradient"
          >
            {{ vendorTheme(item.id).abbr }}
          </div>
          <div class="min-w-0 flex-1 overflow-hidden">
            <p class="truncate text-sm font-medium leading-tight text-foreground">{{ item.label }}</p>
            <p
              v-if="item.description"
              class="mt-0.5 truncate text-xs leading-tight text-muted"
            >
              {{ item.description }}
            </p>
          </div>
        </button>
      </div>
    </div>
  </div>
</template>
