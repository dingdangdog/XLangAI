<script setup lang="ts">
import BackupConversationsPanel from "~/components/admin/backup/BackupConversationsPanel.vue";
import BackupMessagesPanel from "~/components/admin/backup/BackupMessagesPanel.vue";
import BackupUserUsagePanel from "~/components/admin/backup/BackupUserUsagePanel.vue";
import BackupUsersPanel from "~/components/admin/backup/BackupUsersPanel.vue";

const { t } = useI18n();

const TAB_KEYS = ["users", "conversations", "messages", "usage"] as const;

type TabKey = (typeof TAB_KEYS)[number];

const tabLabelKeys: Record<TabKey, string> = {
  users: "pages.backups.tabUsers",
  conversations: "pages.backups.tabConversations",
  messages: "pages.backups.tabMessages",
  usage: "pages.backups.tabUsage",
};

const TABS = computed(() =>
  TAB_KEYS.map((key) => ({
    key,
    label: t(tabLabelKeys[key]),
  })),
);

const route = useRoute();
const router = useRouter();

const activeTab = computed<TabKey>(() => {
  const tab = String(route.query.tab ?? "users");
  return TAB_KEYS.includes(tab as TabKey) ? (tab as TabKey) : "users";
});

function setTab(key: TabKey) {
  if (activeTab.value === key) return;
  router.replace({ query: { ...route.query, tab: key } });
}

const TAB_PANEL = {
  users: BackupUsersPanel,
  conversations: BackupConversationsPanel,
  messages: BackupMessagesPanel,
  usage: BackupUserUsagePanel,
} as const;
</script>

<template>
  <AdminListPage>
    <template #header>
      <AdminPageHeader
        :title="$t('pages.backups.title')"
        :description="$t('pages.backups.description')"
      />

      <AdminAlert :title="$t('pages.backups.readonlyAlert')" />

      <div
        class="flex flex-wrap gap-1 rounded-lg border border-border bg-surface-muted p-1"
        role="tablist"
      >
        <button
          v-for="tab in TABS"
          :key="tab.key"
          type="button"
          role="tab"
          class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors"
          :class="
            activeTab === tab.key
              ? 'bg-surface text-foreground shadow-sm'
              : 'text-muted hover:text-foreground'
          "
          :aria-selected="activeTab === tab.key"
          @click="setTab(tab.key)"
        >
          {{ tab.label }}
        </button>
      </div>
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
