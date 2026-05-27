<script setup lang="ts">
import type { Component } from "vue";
import {
  CloudArrowUpIcon,
  CpuChipIcon,
  GlobeAltIcon,
  MicrophoneIcon,
  ServerIcon,
  SparklesIcon,
} from "@heroicons/vue/24/outline";
import LlmConfigsPanel from "~/components/admin/settings/LlmConfigsPanel.vue";
import ObjectStorageConfigsPanel from "~/components/admin/settings/ObjectStorageConfigsPanel.vue";
import SttConfigsPanel from "~/components/admin/settings/SttConfigsPanel.vue";
import TranslateConfigsPanel from "~/components/admin/settings/TranslateConfigsPanel.vue";
import TtsConfigsPanel from "~/components/admin/settings/TtsConfigsPanel.vue";
import VoiceRolesPanel from "~/components/admin/settings/VoiceRolesPanel.vue";

const { t } = useI18n();

const TAB_KEYS = [
  "llm",
  "stt",
  "tts",
  "translate",
  "voice-roles",
  "object-storage",
] as const;

type TabKey = (typeof TAB_KEYS)[number];

const TAB_META: Record<
  TabKey,
  { icon: Component; labelKey: string; descriptionKey: string }
> = {
  llm: {
    icon: CpuChipIcon,
    labelKey: "pages.aiSettings.tabs.llm",
    descriptionKey: "pages.llmConfigs.description",
  },
  stt: {
    icon: ServerIcon,
    labelKey: "pages.aiSettings.tabs.stt",
    descriptionKey: "pages.sttConfigs.description",
  },
  tts: {
    icon: MicrophoneIcon,
    labelKey: "pages.aiSettings.tabs.tts",
    descriptionKey: "pages.ttsConfigs.description",
  },
  translate: {
    icon: GlobeAltIcon,
    labelKey: "pages.aiSettings.tabs.translate",
    descriptionKey: "pages.translateConfigs.description",
  },
  "voice-roles": {
    icon: SparklesIcon,
    labelKey: "pages.aiSettings.tabs.voiceRoles",
    descriptionKey: "pages.voiceRoles.title",
  },
  "object-storage": {
    icon: CloudArrowUpIcon,
    labelKey: "pages.aiSettings.tabs.objectStorage",
    descriptionKey: "pages.objectStorage.description",
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
  defaultTab: "llm",
});

const tabDescription = computed(() => t(TAB_META[activeTab.value].descriptionKey));

const TAB_PANEL = {
  llm: LlmConfigsPanel,
  stt: SttConfigsPanel,
  tts: TtsConfigsPanel,
  translate: TranslateConfigsPanel,
  "voice-roles": VoiceRolesPanel,
  "object-storage": ObjectStorageConfigsPanel,
} as const;
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.aiSettings.title')"
        :description="tabDescription"
      />

      <AdminTabBar
        :tabs="tabs"
        :active-tab="activeTab"
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
