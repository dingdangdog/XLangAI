<script setup lang="ts">
import type { Component } from "vue";
import {
  ArrowDownTrayIcon,
  ChatBubbleBottomCenterTextIcon,
  Cog6ToothIcon,
  ServerIcon,
} from "@heroicons/vue/24/outline";
import DataBackupPanel from "~/components/admin/settings/DataBackupPanel.vue";
import ServerStorePanel from "~/components/admin/settings/ServerStorePanel.vue";
import SmsServiceConfigsPanel from "~/components/admin/settings/SmsServiceConfigsPanel.vue";
import SystemVariablesPanel from "~/components/admin/settings/SystemVariablesPanel.vue";

const { t } = useI18n();

const TAB_KEYS = ["variables", "sms", "server-store", "data-backup"] as const;

type TabKey = (typeof TAB_KEYS)[number];

const TAB_META: Record<
  TabKey,
  { icon: Component; labelKey: string; descriptionKey: string }
> = {
  variables: {
    icon: Cog6ToothIcon,
    labelKey: "pages.systemSettingsHub.tabs.variables",
    descriptionKey: "pages.systemSettings.description",
  },
  sms: {
    icon: ChatBubbleBottomCenterTextIcon,
    labelKey: "pages.systemSettingsHub.tabs.sms",
    descriptionKey: "pages.smsService.description",
  },
  "server-store": {
    icon: ServerIcon,
    labelKey: "pages.systemSettingsHub.tabs.serverStore",
    descriptionKey: "pages.serverStore.description",
  },
  "data-backup": {
    icon: ArrowDownTrayIcon,
    labelKey: "pages.systemSettingsHub.tabs.dataBackup",
    descriptionKey: "pages.dataBackup.description",
  },
};

const navItems = computed(() =>
  TAB_KEYS.map((key) => ({
    key,
    icon: TAB_META[key].icon,
    label: t(TAB_META[key].labelKey),
    description: t(TAB_META[key].descriptionKey),
  })),
);

const { activeTab, setTab } = useAdminTabRoute<TabKey>({
  tabKeys: TAB_KEYS,
  defaultTab: "variables",
});

const TAB_PANEL = {
  variables: SystemVariablesPanel,
  sms: SmsServiceConfigsPanel,
  "server-store": ServerStorePanel,
  "data-backup": DataBackupPanel,
} as const;
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader :title="$t('pages.systemSettingsHub.title')" />
    </template>

    <div class="flex min-h-0 flex-1 gap-4">
      <aside
        class="flex w-full shrink-0 flex-col gap-1 overflow-hidden rounded-xl border border-border bg-surface lg:w-64"
      >
        <div class="border-b border-border px-3 py-2.5">
          <p class="text-xs font-medium uppercase tracking-wide text-muted">
            {{ $t("pages.systemSettingsHub.navTitle") }}
          </p>
        </div>
        <nav class="min-h-0 flex-1 overflow-y-auto p-2">
          <button
            v-for="item in navItems"
            :key="item.key"
            type="button"
            class="mb-1 flex w-full items-start gap-3 rounded-lg px-3 py-2.5 text-left transition-colors"
            :class="
              activeTab === item.key
                ? 'bg-primary-50 text-primary-800 dark:bg-primary-950/40 dark:text-primary-200'
                : 'hover:bg-surface-muted text-foreground'
            "
            @click="setTab(item.key)"
          >
            <component
              :is="item.icon"
              class="mt-0.5 h-5 w-5 shrink-0"
              :class="activeTab === item.key ? 'text-primary-600' : 'text-muted'"
            />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-medium leading-snug">{{ item.label }}</p>
              <p class="mt-0.5 line-clamp-3 text-xs text-muted">{{ item.description }}</p>
            </div>
          </button>
        </nav>
      </aside>

      <main class="flex min-h-0 min-w-0 flex-1 flex-col">
        <KeepAlive>
          <component
            :is="TAB_PANEL[activeTab]"
            :key="activeTab"
            class="flex min-h-0 flex-1 flex-col"
          />
        </KeepAlive>
      </main>
    </div>
  </AdminListPage>
</template>
