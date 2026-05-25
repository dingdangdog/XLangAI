<script setup lang="ts">
import BackupConversationsPanel from "~/components/admin/backup/BackupConversationsPanel.vue";
import BackupMessagesPanel from "~/components/admin/backup/BackupMessagesPanel.vue";
import BackupUserUsagePanel from "~/components/admin/backup/BackupUserUsagePanel.vue";
import BackupUsersPanel from "~/components/admin/backup/BackupUsersPanel.vue";

const TABS = [
  { key: "users", label: "用户" },
  { key: "conversations", label: "会话" },
  { key: "messages", label: "消息" },
  { key: "usage", label: "用量" },
] as const;

type TabKey = (typeof TABS)[number]["key"];

const route = useRoute();
const router = useRouter();

const activeTab = computed<TabKey>(() => {
  const t = String(route.query.tab ?? "users");
  return TABS.some((x) => x.key === t) ? (t as TabKey) : "users";
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
      <AdminPageHeader title="备份归档" description="用户注销等流程产生的只读历史快照" />

      <AdminAlert title="只读归档">
        以下数据仅供审计与恢复参考，不可在此增删改。
      </AdminAlert>

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
        <component :is="TAB_PANEL[activeTab]" :key="activeTab" class="flex min-h-0 flex-1 flex-col" />
      </KeepAlive>
    </div>
  </AdminListPage>
</template>
