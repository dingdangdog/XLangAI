<script setup lang="ts">
import type { Component } from "vue";

export type UsageTrendChartPoint = {
  date: string;
  requestCount: number;
  unitCount: number;
};

const props = withDefaults(
  defineProps<{
    title: string;
    subtitle?: string;
    icon?: Component;
    points: UsageTrendChartPoint[];
    unitLabel: string;
    requestLegend: string;
    unitLegend: string;
    showUnits?: boolean;
    formatUnit?: (value: number) => string;
  }>(),
  { showUnits: true },
);

const { formatCompactNumber } = useUsageDisplay();

function displayUnit(value: number): string {
  return props.formatUnit ? props.formatUnit(value) : formatCompactNumber(value);
}

const hoveredIndex = ref<number | null>(null);

const chartWidth = 480;
const chartHeight = 140;
const padLeft = 4;
const padRight = 4;
const padTop = 8;
const padBottom = 22;
const innerW = chartWidth - padLeft - padRight;
const innerH = chartHeight - padTop - padBottom;

const maxRequest = computed(() =>
  Math.max(1, ...props.points.map((p) => p.requestCount)),
);
const maxUnit = computed(() =>
  Math.max(1, ...props.points.map((p) => p.unitCount)),
);

function xAt(i: number): number {
  const n = props.points.length;
  if (n <= 1) return padLeft + innerW / 2;
  return padLeft + (i / (n - 1)) * innerW;
}

function yRequest(v: number): number {
  return padTop + innerH - (v / maxRequest.value) * innerH;
}

function yUnit(v: number): number {
  return padTop + innerH - (v / maxUnit.value) * innerH;
}

const requestLine = computed(() =>
  props.points
    .map((p, i) => `${i === 0 ? "M" : "L"} ${xAt(i).toFixed(2)} ${yRequest(p.requestCount).toFixed(2)}`)
    .join(" "),
);

const requestArea = computed(() => {
  if (props.points.length === 0) return "";
  const baseline = padTop + innerH;
  const line = props.points
    .map((p, i) => `${i === 0 ? "M" : "L"} ${xAt(i).toFixed(2)} ${yRequest(p.requestCount).toFixed(2)}`)
    .join(" ");
  const lastX = xAt(props.points.length - 1).toFixed(2);
  const firstX = xAt(0).toFixed(2);
  return `${line} L ${lastX} ${baseline} L ${firstX} ${baseline} Z`;
});

const unitLine = computed(() =>
  props.points
    .map((p, i) => `${i === 0 ? "M" : "L"} ${xAt(i).toFixed(2)} ${yUnit(p.unitCount).toFixed(2)}`)
    .join(" "),
);

const xLabels = computed(() => {
  const n = props.points.length;
  if (n === 0) return [];
  const step = n <= 7 ? 1 : n <= 14 ? 2 : 5;
  return props.points
    .map((p, i) => ({ ...p, i }))
    .filter((_, i) => i === 0 || i === n - 1 || i % step === 0);
});

function formatAxisDate(iso: string): string {
  const [, month, day] = iso.split("-");
  return `${month}/${day}`;
}

const activePoint = computed(() => {
  const i = hoveredIndex.value;
  if (i == null || i < 0 || i >= props.points.length) return null;
  return props.points[i]!;
});

function onPointerMove(event: MouseEvent) {
  const target = event.currentTarget as SVGSVGElement;
  const rect = target.getBoundingClientRect();
  const ratio = (event.clientX - rect.left) / rect.width;
  const idx = Math.round(ratio * (props.points.length - 1));
  hoveredIndex.value = Math.max(0, Math.min(props.points.length - 1, idx));
}

function onPointerLeave() {
  hoveredIndex.value = null;
}
</script>

<template>
  <div class="rounded-xl border border-border bg-surface-muted/40 p-4">
    <div class="flex items-start gap-3">
      <div
        v-if="icon"
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary-600 dark:bg-primary-950 dark:text-primary-400"
      >
        <component :is="icon" class="h-4 w-4" aria-hidden="true" />
      </div>
      <div class="min-w-0 flex-1">
        <h3 class="text-sm font-semibold text-foreground">{{ title }}</h3>
        <p v-if="subtitle" class="mt-0.5 text-xs text-muted">{{ subtitle }}</p>
      </div>
    </div>

    <div
      v-if="activePoint"
      class="mt-3 rounded-lg border border-border bg-surface px-3 py-2 text-xs text-foreground"
    >
      <span class="font-medium">{{ activePoint.date }}</span>
      <span class="mx-2 text-muted">·</span>
      <span>{{ requestLegend }} {{ activePoint.requestCount }}</span>
      <template v-if="showUnits">
        <span class="mx-2 text-muted">·</span>
        <span>{{ unitLegend }} {{ displayUnit(activePoint.unitCount) }} {{ unitLabel }}</span>
      </template>
    </div>
    <div v-else class="mt-3 h-[34px]" aria-hidden="true" />

    <div class="relative mt-1">
      <svg
        :viewBox="`0 0 ${chartWidth} ${chartHeight}`"
        class="h-36 w-full touch-none select-none"
        role="img"
        :aria-label="title"
        @mousemove="onPointerMove"
        @mouseleave="onPointerLeave"
      >
        <line
          :x1="padLeft"
          :y1="padTop + innerH"
          :x2="padLeft + innerW"
          :y2="padTop + innerH"
          class="stroke-border"
          stroke-width="1"
        />

        <path
          v-if="requestArea"
          :d="requestArea"
          class="fill-primary-500/15"
        />
        <path
          :d="requestLine"
          fill="none"
          class="stroke-primary-500"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <path
          v-if="showUnits"
          :d="unitLine"
          fill="none"
          class="stroke-accent-500"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-dasharray="4 3"
        />

        <template v-for="(p, i) in points" :key="p.date">
          <circle
            v-if="hoveredIndex === i"
            :cx="xAt(i)"
            :cy="yRequest(p.requestCount)"
            r="4"
            class="fill-primary-500 stroke-surface"
            stroke-width="2"
          />
        </template>

        <text
          v-for="item in xLabels"
          :key="item.date"
          :x="xAt(item.i)"
          :y="chartHeight - 4"
          text-anchor="middle"
          class="fill-muted text-[10px]"
        >
          {{ formatAxisDate(item.date) }}
        </text>
      </svg>
    </div>

    <div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted">
      <span class="inline-flex items-center gap-1.5">
        <span class="inline-block h-0.5 w-4 rounded bg-primary-500" />
        {{ requestLegend }}
      </span>
      <span v-if="showUnits" class="inline-flex items-center gap-1.5">
        <span class="inline-block h-0.5 w-4 rounded border border-dashed border-accent-500" />
        {{ unitLegend }}
      </span>
    </div>
  </div>
</template>
