<script setup lang="ts">
import { ChevronDownIcon, GlobeAltIcon } from "@heroicons/vue/24/outline";

type LocaleOption = {
  code: string;
  name?: string;
};

const { locale, locales } = useI18n();
const switchLocalePath = useSwitchLocalePath();
const isOpen = ref(false);
const buttonRef = ref<HTMLElement | null>(null);
const dropdownRef = ref<HTMLElement | null>(null);

const availableLocales = computed<LocaleOption[]>(() =>
  locales.value.map((item) =>
    typeof item === "string" ? { code: item, name: item } : item,
  ),
);

const currentLocale = computed(
  () =>
    availableLocales.value.find((item) => item.code === locale.value) ??
    availableLocales.value[0],
);

function closeDropdown() {
  isOpen.value = false;
}

async function switchLocale(code: string) {
  closeDropdown();
  if (code === locale.value) return;
  await navigateTo(switchLocalePath(code));
}

function handleClickOutside(event: MouseEvent) {
  if (!isOpen.value) return;

  const target = event.target as Node;
  if (buttonRef.value?.contains(target) || dropdownRef.value?.contains(target)) {
    return;
  }

  closeDropdown();
}

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onUnmounted(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>

<template>
  <div class="relative">
    <button ref="buttonRef" type="button"
      class="inline-flex w-full items-center gap-2 md:gap-3 rounded-lg px-3 py-2 text-sm font-medium text-muted transition-colors hover:bg-surface-muted hover:text-foreground"
      :aria-label="$t('language.switcher')" @click.stop="isOpen = !isOpen">
      <GlobeAltIcon class="h-5 w-5 shrink-0" />
      <span class="min-w-0 flex-1 truncate text-left">{{ currentLocale?.name }}</span>
      <ChevronDownIcon class="h-4 w-4 shrink-0 transition-transform" :class="{ 'rotate-180': isOpen }" />
    </button>

    <Transition enter-active-class="transition ease-out duration-100" enter-from-class="scale-95 opacity-0"
      enter-to-class="scale-100 opacity-100" leave-active-class="transition ease-in duration-75"
      leave-from-class="scale-100 opacity-100" leave-to-class="scale-95 opacity-0">
      <div v-if="isOpen" ref="dropdownRef"
        class="absolute bottom-full left-0 z-50 mb-2 w-full overflow-hidden rounded-xl border border-border bg-surface shadow-lg">
        <button v-for="item in availableLocales" :key="item.code" type="button"
          class="block w-full px-4 py-2 text-left text-sm text-muted transition-colors hover:bg-surface-muted hover:text-foreground"
          :class="{ 'bg-surface-muted text-foreground': item.code === locale }" @click="switchLocale(item.code)">
          {{ item.name ?? item.code }}
        </button>
      </div>
    </Transition>
  </div>
</template>
