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

const tabs = computed(() =>
  TAB_KEYS.map((key) => ({
    key,
    icon: TAB_META[key].icon,
    label: t(TAB_META[key].labelKey),
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
  <AdminListPage fill>
    <template #header>
      <AdminPageHeader :title="$t('pages.systemSettingsHub.title')" />

      <AgentSegmentNav
        :items="tabs"
        :active="activeTab"
        @change="setTab($event as TabKey)"
      />
    </template>

    <div class="flex min-h-0 flex-1 flex-col">
      <KeepAlive>
        <component
          :is="TAB_PANEL[activeTab]"
          :key="activeTab"
          class="flex min-h-0 flex-1 flex-col"
        />
      </KeepAlive>
    </div>
  </AdminListPage>
</template>
