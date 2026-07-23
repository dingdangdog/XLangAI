<script setup lang="ts">
import {
  ChatBubbleLeftIcon,
  CpuChipIcon,
  DocumentTextIcon,
  GlobeAltIcon,
  MicrophoneIcon,
  ServerIcon,
  SparklesIcon,
  UserGroupIcon,
  UsersIcon,
} from "@heroicons/vue/24/outline";
type UsageTrendService = {
  unitLabel: string;
  points: { date: string; requestCount: number; unitCount: string }[];
  totalRequests: number;
  totalUnits: string;
};

type UsageTrendResponse = {
  days: 7 | 30;
  from: string;
  to: string;
  services: {
    llm: UsageTrendService;
    tts: UsageTrendService;
    translate: UsageTrendService;
    stt: UsageTrendService;
  };
  conversations: {
    points: { date: string; count: number }[];
    total: number;
  };
};

const { t } = useI18n();
const localePath = useLocalePath();
const { serviceUsageMonthLine, formatAudioBytes } = useUsageDisplay();

const trendDays = ref<7 | 30>(7);

const { data: stats, pending: statsPending } = await useFetch("/api/admin/stats");

const {
  data: trends,
  pending: trendsPending,
} = await useFetch<UsageTrendResponse>("/api/admin/stats/usage-trends", {
  query: computed(() => ({ days: trendDays.value })),
  watch: [trendDays],
});

type EntityCard = {
  label: string;
  value: string | number;
  to: string;
  icon: typeof UsersIcon;
};

const entityGroups = computed(() => {
  const s = stats.value;
  if (!s) return [];
  return [
    {
      title: t("pages.dashboard.groups.business"),
      cards: [
        {
          label: t("pages.dashboard.users"),
          value: s.users,
          to: localePath("/manage/users"),
          icon: UsersIcon,
        },
        {
          label: t("pages.dashboard.conversations"),
          value: s.conversations,
          to: localePath("/manage/conversations"),
          icon: ChatBubbleLeftIcon,
        },
        {
          label: t("pages.dashboard.messages"),
          value: s.messages,
          to: localePath("/manage/messages"),
          icon: ChatBubbleLeftIcon,
        },
        {
          label: t("pages.dashboard.tiers"),
          value: s.tiers,
          to: localePath("/manage/membership-tiers"),
          icon: UserGroupIcon,
        },
      ] satisfies EntityCard[],
    },
    {
      title: t("pages.dashboard.groups.content"),
      cards: [
        {
          label: t("pages.dashboard.languages"),
          value: s.languages,
          to: localePath("/manage/languages"),
          icon: GlobeAltIcon,
        },
        {
          label: t("pages.dashboard.voiceRoles"),
          value: s.voiceRoles,
          to: `${localePath("/manage/ai-settings")}?tab=voice-roles`,
          icon: SparklesIcon,
        },
        {
          label: t("pages.dashboard.promptTemplates"),
          value: s.promptTemplates,
          to: localePath("/manage/prompt-templates"),
          icon: DocumentTextIcon,
        },
      ] satisfies EntityCard[],
    },
    {
      title: t("pages.dashboard.groups.services"),
      cards: [
        {
          label: t("pages.dashboard.llmConfigs"),
          value: s.llmConfigs,
          to: `${localePath("/manage/ai-settings")}?tab=llm`,
          icon: CpuChipIcon,
        },
        {
          label: t("pages.dashboard.sttConfigs"),
          value: s.sttConfigs,
          to: `${localePath("/manage/ai-settings")}?tab=stt`,
          icon: ServerIcon,
        },
        {
          label: t("pages.dashboard.ttsConfigs"),
          value: s.ttsConfigs,
          to: `${localePath("/manage/ai-settings")}?tab=tts`,
          icon: MicrophoneIcon,
        },
        {
          label: t("pages.dashboard.serverStore"),
          value: t("common.config"),
          to: `${localePath("/manage/system-settings")}?tab=server-store`,
          icon: ServerIcon,
        },
      ] satisfies EntityCard[],
    },
  ];
});

function serviceSummary(service: UsageTrendResponse["services"]["llm"] | undefined): string {
  if (!service) return t("common.emDash");
  return serviceUsageMonthLine({
    todayRequestCount: service.totalRequests,
    todayUnitCount: service.totalUnits,
    monthRequestCount: service.totalRequests,
    monthUnitCount: service.totalUnits,
    unitLabel: service.unitLabel,
  });
}

function sttSummary(service: UsageTrendResponse["services"]["stt"] | undefined): string {
  if (!service) return t("common.emDash");
  return t("pages.dashboard.sttPeriodTotal", {
    count: service.totalRequests,
    size: formatAudioBytes(service.totalUnits),
  });
}

const usageCharts = computed(() => {
  const tr = trends.value;
  if (!tr) return [];
  return [
    {
      key: "llm",
      title: t("pages.dashboard.usageLlm"),
      subtitle: serviceSummary(tr.services.llm),
      icon: CpuChipIcon,
      points: tr.services.llm.points.map((p) => ({
        date: p.date,
        requestCount: p.requestCount,
        unitCount: Number(p.unitCount),
      })),
      unitLabel: tr.services.llm.unitLabel,
    },
    {
      key: "tts",
      title: t("pages.dashboard.usageTts"),
      subtitle: serviceSummary(tr.services.tts),
      icon: MicrophoneIcon,
      points: tr.services.tts.points.map((p) => ({
        date: p.date,
        requestCount: p.requestCount,
        unitCount: Number(p.unitCount),
      })),
      unitLabel: tr.services.tts.unitLabel,
    },
    {
      key: "translate",
      title: t("pages.dashboard.usageTranslate"),
      subtitle: serviceSummary(tr.services.translate),
      icon: GlobeAltIcon,
      points: tr.services.translate.points.map((p) => ({
        date: p.date,
        requestCount: p.requestCount,
        unitCount: Number(p.unitCount),
      })),
      unitLabel: tr.services.translate.unitLabel,
    },
    {
      key: "stt",
      title: t("pages.dashboard.usageStt"),
      subtitle: sttSummary(tr.services.stt),
      icon: ServerIcon,
      points: tr.services.stt.points.map((p) => ({
        date: p.date,
        requestCount: p.requestCount,
        unitCount: Number(p.unitCount),
      })),
      unitLabel: tr.services.stt.unitLabel,
      formatUnit: (n: number) => formatAudioBytes(n),
    },
    {
      key: "conversations",
      title: t("pages.dashboard.usageConversations"),
      subtitle: t("pages.dashboard.conversationTotal", { count: tr.conversations.total }),
      icon: ChatBubbleLeftIcon,
      points: tr.conversations.points.map((p) => ({
        date: p.date,
        requestCount: p.count,
        unitCount: 0,
      })),
      unitLabel: "",
      showUnits: false,
      requestLegend: t("pages.dashboard.legendTurns"),
    },
  ];
});

const loading = computed(() => statsPending.value || trendsPending.value);
</script>

<template>
  <AdminListPage :fill="false">
    <template #header>
      <AdminPageHeader :title="$t('pages.dashboard.title')" :description="$t('pages.dashboard.description')" />
    </template>

    <AdminSkeleton v-if="loading" :rows="10" />

    <div v-else class="flex flex-col gap-2 pb-1">
      <AdminPanel :fill="false">
        <div class="border-b border-border px-4 py-2 md:px-5">
          <h2 class="text-sm text-foreground">{{ $t("pages.dashboard.entitySection") }}</h2>
        </div>
        <div class="px-3 py-2 md:px-4">
          <section v-for="group in entityGroups" :key="group.title" class="mb-2 last:mb-0">
            <p class="mb-1 text-xs text-muted">{{ group.title }}</p>
            <div class="grid grid-cols-2 gap-1.5 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
              <AdminDashboardStatCard class="cursor-pointer" v-for="card in group.cards" :key="card.label"
                :label="card.label" :value="card.value" :to="card.to" :icon="card.icon" />
            </div>
          </section>
        </div>
      </AdminPanel>

      <AdminPanel :fill="false">
        <div class="flex flex-wrap items-center justify-between gap-2 border-b border-border px-4 py-2 md:px-5">
          <div>
            <h2 class="text-sm text-foreground">{{ $t("pages.dashboard.usageSection") }}</h2>
            <p class="mt-0.5 text-xs text-muted">
              {{ $t("pages.dashboard.usageSectionHint") }}
              <span v-if="trends">· {{ trends.from }} — {{ trends.to }} (UTC)</span>
            </p>
          </div>
          <div class="inline-flex rounded-md border border-border bg-surface p-0.5 text-xs" role="group"
            :aria-label="$t('pages.dashboard.periodLabel')">
            <button type="button" class="rounded px-2 py-1 transition-colors" :class="trendDays === 7
              ? 'bg-primary-500 text-white'
              : 'text-muted hover:text-foreground'
              " @click="trendDays = 7">
              {{ $t("pages.dashboard.last7Days") }}
            </button>
            <button type="button" class="rounded px-2 py-1 transition-colors" :class="trendDays === 30
              ? 'bg-primary-500 text-white'
              : 'text-muted hover:text-foreground'
              " @click="trendDays = 30">
              {{ $t("pages.dashboard.last30Days") }}
            </button>
          </div>
        </div>
        <div class="grid gap-2 p-1 md:grid-cols-2 md:p-2 xl:grid-cols-3">
          <AdminUsageTrendChart v-for="chart in usageCharts" :key="chart.key" :title="chart.title"
            :subtitle="chart.subtitle" :icon="chart.icon" :points="chart.points" :unit-label="chart.unitLabel"
            :show-units="chart.showUnits !== false"
            :request-legend="chart.requestLegend ?? $t('pages.dashboard.legendRequests')"
            :unit-legend="$t('pages.dashboard.legendUnits')" :format-unit="chart.formatUnit" />
        </div>
      </AdminPanel>
    </div>
  </AdminListPage>
</template>
