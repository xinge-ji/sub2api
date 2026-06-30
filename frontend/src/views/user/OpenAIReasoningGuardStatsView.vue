<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading && !stats" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else>
        <div class="card">
          <div class="bg-gradient-to-r from-sky-50 via-white to-amber-50 px-5 py-5 dark:from-dark-900 dark:via-dark-800 dark:to-dark-900">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div class="space-y-2">
                <div class="flex items-center gap-2">
                  <span :class="['badge text-xs', selectedCombo ? 'badge-primary' : 'badge-gray']">
                    {{ t('usage.currentView') }}
                  </span>
                  <span class="text-xs text-gray-500 dark:text-gray-400">
                    {{ displayedRange }}
                  </span>
                </div>
                <div>
                  <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                    {{ selectedComboLabel }}
                  </h2>
                  <p v-if="selectionDescription" class="mt-1 text-sm text-gray-600 dark:text-gray-300">
                    {{ selectionDescription }}
                  </p>
                </div>
              </div>

              <div class="flex flex-wrap items-center gap-3">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t('usage.timeRange') }}:
                  </span>
                  <DateRangePicker
                    v-model:start-date="startDate"
                    v-model:end-date="endDate"
                    @change="onDateRangeChange"
                  />
                </div>

                <button @click="loadStats" :disabled="loading" class="btn btn-secondary">
                  {{ t('common.refresh') }}
                </button>

                <div class="flex items-center gap-2 lg:ml-2">
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t('usage.granularity') }}:
                  </span>
                  <div class="w-28">
                    <Select
                      v-model="granularity"
                      :options="granularityOptions"
                      @change="loadStats"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-sky-100 p-2 dark:bg-sky-900/30">
                <Icon name="chart" size="md" class="text-sky-600 dark:text-sky-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  {{ t('usage.requestCount') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatNumber(effectiveSummary.request_count) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ displayedRange }}
                </p>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
                <Icon name="shield" size="md" class="text-amber-600 dark:text-amber-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  {{ t('usage.matchCount') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatNumber(effectiveSummary.match_count) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ selectedCombo ? t('usage.selectionScoped') : t('usage.inSelectedRange') }}
                </p>
              </div>
            </div>
          </div>

          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-rose-100 p-2 dark:bg-rose-900/30">
                <Icon name="bolt" size="md" class="text-rose-600 dark:text-rose-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  {{ t('usage.matchRatio') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ ratioText }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ granularityLabel }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <div class="card p-4">
          <div class="mb-4 flex flex-wrap items-start justify-between gap-3">
            <div>
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ t('usage.guardStatus') }}
              </h3>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                {{ runtime.enabled ? t('usage.guardEnabled') : t('usage.guardDisabled') }}
              </p>
            </div>
            <div class="rounded-lg bg-gray-50 px-3 py-2 text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300">
              {{ t('usage.guardStatusCode') }}: {{ runtime.intercept_status_code || 502 }}
            </div>
          </div>

          <div v-if="runtime.rules.length" class="overflow-x-auto rounded-xl border border-gray-100 dark:border-gray-700">
            <table class="w-full text-xs">
              <thead class="bg-gray-50 dark:bg-dark-800">
                <tr class="text-gray-500 dark:text-gray-400">
                  <th class="px-3 py-3 text-left">{{ t('admin.settings.openaiReasoningGuard.model') }}</th>
                  <th class="px-3 py-3 text-left">{{ t('admin.settings.openaiReasoningGuard.codes') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="rule in runtime.rules"
                  :key="`${rule.model}-${rule.codes.join(',')}`"
                  class="border-t border-gray-100 dark:border-gray-700"
                >
                  <td class="px-3 py-3 font-medium text-gray-900 dark:text-white">{{ rule.model }}</td>
                  <td class="px-3 py-3 text-gray-600 dark:text-gray-300">{{ rule.codes.join(', ') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div
            v-else
            class="rounded-xl border border-dashed border-gray-200 px-4 py-6 text-sm text-gray-500 dark:border-gray-700 dark:text-gray-400"
          >
            {{ t('usage.guardRulesEmpty') }}
          </div>
        </div>

        <div class="grid grid-cols-1 gap-6 xl:grid-cols-5">
          <div class="card p-4 xl:col-span-3">
            <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
              <div>
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ t('usage.interceptTrend') }}
                </h3>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ granularityLabel }} · {{ displayedRange }}
                </p>
              </div>

              <div class="flex flex-wrap items-center gap-2">
                <span :class="['badge text-xs', selectedCombo ? 'badge-primary' : 'badge-gray']">
                  {{ selectedComboLabel }}
                </span>
                <button
                  v-if="selectedCombo"
                  type="button"
                  class="text-xs font-medium text-primary-600 transition-colors hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
                  @click="clearSelection"
                >
                  {{ t('common.clearSelection') }}
                </button>
              </div>
            </div>

            <div v-if="loading" class="flex h-72 items-center justify-center">
              <LoadingSpinner size="md" />
            </div>
            <div v-else-if="trendChartData" class="h-72 overflow-hidden">
              <Line
                :data="trendChartData"
                :options="trendChartOptions"
              />
            </div>
            <div
              v-else
              class="flex h-72 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
            >
              {{ t('admin.dashboard.noDataAvailable') }}
            </div>
          </div>

          <div class="card p-4 xl:col-span-2">
            <div class="mb-4">
              <div class="flex items-center justify-between gap-3">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                  {{ t('usage.modelEffortBreakdown') }}
                </h3>
                <span class="text-xs text-gray-500 dark:text-gray-400">
                  {{ formatNumber(modelEffortRows.length) }}
                </span>
              </div>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                {{ t('usage.modelEffortBreakdownHint') }}
              </p>
            </div>

            <div v-if="modelEffortRows.length > 0 && modelEffortChartData" class="space-y-5">
              <div class="mx-auto h-52 w-52 overflow-hidden">
                <Doughnut :data="modelEffortChartData" :options="doughnutOptions" />
              </div>

              <div class="max-h-80 overflow-y-auto rounded-xl border border-gray-100 dark:border-gray-700">
                <table class="w-full text-xs">
                  <thead class="sticky top-0 bg-gray-50/95 dark:bg-dark-800/95">
                    <tr class="text-gray-500 dark:text-gray-400">
                      <th class="px-3 pb-2 pt-3 text-left">{{ t('usage.requestedModel') }}</th>
                      <th class="px-3 pb-2 pt-3 text-left">{{ t('usage.reasoningEffort') }}</th>
                      <th class="px-3 pb-2 pt-3 text-right">{{ t('usage.requestCount') }}</th>
                      <th class="px-3 pb-2 pt-3 text-right">{{ t('usage.matchCount') }}</th>
                      <th class="px-3 pb-2 pt-3 text-right">{{ t('usage.matchRatio') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="item in modelEffortRows"
                      :key="item.key"
                      class="cursor-pointer border-t border-gray-100 transition-colors hover:bg-gray-50 dark:border-gray-700 dark:hover:bg-dark-700/40"
                      :class="selectedComboKey === item.key ? 'bg-primary-50 dark:bg-primary-900/10' : ''"
                      @click="toggleCombo(item.key)"
                    >
                      <td class="max-w-[180px] truncate px-3 py-3 font-medium text-gray-900 dark:text-white" :title="item.model">
                        {{ item.model }}
                      </td>
                      <td class="px-3 py-3 text-gray-600 dark:text-gray-300">
                        {{ item.reasoning_effort }}
                      </td>
                      <td class="px-3 py-3 text-right text-gray-600 dark:text-gray-400">
                        {{ formatNumber(item.request_count) }}
                      </td>
                      <td class="px-3 py-3 text-right text-gray-600 dark:text-gray-400">
                        {{ formatNumber(item.match_count) }}
                      </td>
                      <td class="px-3 py-3 text-right font-medium text-gray-900 dark:text-white">
                        {{ formatRatio(item.match_ratio) }}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
            <div
              v-else
              class="flex h-72 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
            >
              {{ t('admin.dashboard.noDataAvailable') }}
            </div>
          </div>
        </div>

        <div class="card overflow-hidden">
          <div class="flex flex-col gap-4 border-b border-gray-100 bg-gradient-to-r from-sky-50 via-white to-white px-5 py-5 dark:border-gray-700 dark:from-dark-900 dark:via-dark-800 dark:to-dark-800">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div class="space-y-2">
                <div class="flex items-center gap-2">
                  <span class="badge badge-gray text-xs">
                    Codex Radar
                  </span>
                  <span class="text-xs text-gray-500 dark:text-gray-400">
                    {{ t('usage.externalRadarStaticHint') }}
                  </span>
                </div>
                <div>
                  <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                    {{ t('usage.externalRadarTitle') }}
                  </h3>
                  <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">
                    {{ t('usage.externalRadarDescription') }}
                  </p>
                </div>
              </div>

              <div class="flex flex-wrap items-center gap-3">
                <button @click="loadCodexRadar" :disabled="codexRadarLoading" class="btn btn-secondary">
                  {{ t('common.refresh') }}
                </button>
                <a
                  :href="CODEX_RADAR_URL"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="btn btn-secondary"
                >
                  <Icon name="externalLink" size="sm" class="mr-1.5" :stroke-width="2" />
                  {{ t('usage.externalRadarOpenSource') }}
                </a>
              </div>
            </div>
          </div>

          <div v-if="codexRadarLoading" class="flex h-[720px] items-center justify-center">
            <LoadingSpinner size="md" />
          </div>

          <div
            v-else-if="codexRadarError"
            class="flex h-[720px] flex-col items-center justify-center gap-4 px-6 text-center"
          >
            <div class="rounded-full bg-amber-100 p-3 dark:bg-amber-900/30">
              <Icon name="exclamationTriangle" size="lg" class="text-amber-600 dark:text-amber-400" :stroke-width="2" />
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900 dark:text-white">
                {{ t('usage.externalRadarLoadFailed') }}
              </p>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ codexRadarError }}
              </p>
            </div>
          </div>

          <iframe
            v-else-if="codexRadarSrcdoc"
            ref="codexRadarFrame"
            :srcdoc="codexRadarSrcdoc"
            sandbox="allow-same-origin"
            loading="lazy"
            referrerpolicy="no-referrer"
            class="w-full border-0 bg-transparent"
            :style="{ height: `${codexRadarFrameHeight}px` }"
            :title="t('usage.externalRadarTitle')"
            @load="syncCodexRadarFrameHeight"
          ></iframe>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  ArcElement,
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  Tooltip,
} from 'chart.js'
import { Doughnut, Line } from 'vue-chartjs'
import AppLayout from '@/components/layout/AppLayout.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import * as usageAPI from '@/api/usage'
import type {
  OpenAIReasoningGuardModelEffortItem,
  OpenAIReasoningGuardModelEffortTrendPoint,
  OpenAIReasoningGuardStatsResponse,
  OpenAIReasoningGuardTrendPoint,
} from '@/api/usage'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'

ChartJS.register(
  ArcElement,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler,
)

const { t } = useI18n()
const appStore = useAppStore()

type GuardGranularity = 'day' | 'hour'
type GuardTrendPoint = OpenAIReasoningGuardTrendPoint | OpenAIReasoningGuardModelEffortTrendPoint

const stats = ref<OpenAIReasoningGuardStatsResponse | null>(null)
const loading = ref(false)
const granularity = ref<GuardGranularity>('day')
const selectedComboKey = ref<string | null>(null)
const chartEvents: Array<keyof HTMLElementEventMap> = ['click', 'mousemove', 'mouseout']
const codexRadarLoading = ref(false)
const codexRadarError = ref('')
const codexRadarSrcdoc = ref('')
const codexRadarFrame = ref<HTMLIFrameElement | null>(null)
const codexRadarFrameHeight = ref(560)

const MAX_MODEL_EFFORT_SEGMENTS = 8
const CODEX_RADAR_URL = 'https://codexradar.com/'

const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const today = new Date()
const defaultEnd = formatLocalDate(today)
const defaultStartDate = (() => {
  const start = new Date(today)
  start.setDate(start.getDate() - 6)
  return formatLocalDate(start)
})()

const startDate = ref(defaultStartDate)
const endDate = ref(defaultEnd)

const granularityOptions = computed(() => [
  { value: 'day', label: t('usage.day') },
  { value: 'hour', label: t('usage.hour') },
])

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))

const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb',
  primary: '#0ea5e9',
  secondary: '#f59e0b',
  accent: '#ef4444',
  ring: ['#0ea5e9', '#14b8a6', '#f59e0b', '#f97316', '#ef4444', '#8b5cf6', '#ec4899', '#84cc16'],
  stroke: isDarkMode.value ? '#0f172a' : '#ffffff',
}))

const modelEffortRows = computed(() => stats.value?.model_efforts || [])

const modelEffortChartRows = computed<OpenAIReasoningGuardModelEffortItem[]>(() => {
  if (modelEffortRows.value.length <= MAX_MODEL_EFFORT_SEGMENTS) {
    return modelEffortRows.value
  }
  const head = modelEffortRows.value.slice(0, MAX_MODEL_EFFORT_SEGMENTS - 1)
  const tail = modelEffortRows.value.slice(MAX_MODEL_EFFORT_SEGMENTS - 1)
  const tailRequestCount = tail.reduce((sum, item) => sum + item.request_count, 0)
  const tailMatchCount = tail.reduce((sum, item) => sum + item.match_count, 0)
  const tailInterceptCount = tail.reduce((sum, item) => sum + item.intercept_count, 0)
  return [
    ...head,
    {
      model: t('admin.ops.charts.other'),
      reasoning_effort: t('admin.ops.charts.other'),
      key: '__other__',
      request_count: tailRequestCount,
      match_count: tailMatchCount,
      match_ratio: tailRequestCount > 0 ? tailMatchCount / tailRequestCount : 0,
      intercept_count: tailInterceptCount,
      intercept_ratio: tailRequestCount > 0 ? tailInterceptCount / tailRequestCount : 0,
    },
  ]
})

const selectedCombo = computed(() => {
  if (!selectedComboKey.value) return null
  return modelEffortRows.value.find((item) => item.key === selectedComboKey.value) || null
})

const effectiveSummary = computed(() => {
  if (selectedCombo.value) {
    return {
      request_count: selectedCombo.value.request_count,
      match_count: selectedCombo.value.match_count,
      match_ratio: selectedCombo.value.match_ratio,
      intercept_count: selectedCombo.value.intercept_count,
      intercept_ratio: selectedCombo.value.intercept_ratio,
    }
  }
  return stats.value?.summary || {
    request_count: 0,
    match_count: 0,
    match_ratio: 0,
    intercept_count: 0,
    intercept_ratio: 0,
  }
})

const ratioText = computed(() => formatRatio(effectiveSummary.value.match_ratio))
const granularityLabel = computed(() => granularity.value === 'hour' ? t('usage.hour') : t('usage.day'))
const runtime = computed(() => {
  const value = stats.value?.runtime
  return {
    enabled: value?.enabled ?? false,
    intercept_status_code: value?.intercept_status_code || 502,
    rules: Array.isArray(value?.rules)
      ? value!.rules.map((rule) => ({
          model: typeof rule?.model === 'string' ? rule.model : '',
          codes: Array.isArray(rule?.codes) ? rule.codes : [],
        }))
      : [],
  }
})

const displayedRange = computed(() => {
  const start = stats.value?.start_date || startDate.value
  const end = stats.value?.end_date || endDate.value
  return `${start} ~ ${end}`
})

const selectedComboLabel = computed(() => {
  if (!selectedCombo.value) return t('usage.overallView')
  return `${selectedCombo.value.model} / ${selectedCombo.value.reasoning_effort}`
})

const selectionDescription = computed(() => {
  if (selectedCombo.value) return t('usage.selectionScoped')
  return ''
})

const effectiveTrend = computed<GuardTrendPoint[]>(() => {
  if (!selectedComboKey.value) return stats.value?.trend || []
  return (stats.value?.model_effort_trend || []).filter((item) => item.key === selectedComboKey.value)
})

const trendChartData = computed(() => {
  if (!effectiveTrend.value.length) return null
  return {
    labels: effectiveTrend.value.map((item) => item.date),
    datasets: [
      {
        label: t('usage.requestCount'),
        data: effectiveTrend.value.map((item) => item.request_count),
        borderColor: chartColors.value.primary,
        backgroundColor: `${chartColors.value.primary}20`,
        fill: true,
        tension: 0.3,
      },
      {
        label: t('usage.matchCount'),
        data: effectiveTrend.value.map((item) => item.match_count),
        borderColor: chartColors.value.secondary,
        backgroundColor: `${chartColors.value.secondary}18`,
        fill: true,
        tension: 0.3,
      },
      {
        label: t('usage.matchRatio'),
        data: effectiveTrend.value.map((item) => Number((item.match_ratio * 100).toFixed(2))),
        borderColor: chartColors.value.accent,
        backgroundColor: `${chartColors.value.accent}20`,
        borderDash: [5, 5],
        fill: false,
        tension: 0.3,
        yAxisID: 'yPercent',
      },
    ],
  }
})

const trendChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: {
    duration: 0,
  },
  normalized: true,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
  resizeDelay: 120,
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 14,
        font: {
          size: 11,
        },
      },
    },
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid,
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10,
        },
      },
    },
    y: {
      grid: {
        color: chartColors.value.grid,
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10,
        },
        callback: (value: string | number) => formatCompactNumber(Number(value)),
      },
    },
    yPercent: {
      position: 'right' as const,
      min: 0,
      max: 100,
      grid: {
        drawOnChartArea: false,
      },
      ticks: {
        color: chartColors.value.accent,
        font: {
          size: 10,
        },
        callback: (value: string | number) => `${value}%`,
      },
    },
  },
}))

function buildPalette(length: number): string[] {
  return Array.from({ length }, (_, index) => chartColors.value.ring[index % chartColors.value.ring.length])
}

const modelEffortChartData = computed(() => {
  if (!modelEffortChartRows.value.length) return null
  const palette = buildPalette(modelEffortChartRows.value.length)
  return {
    labels: modelEffortChartRows.value.map((item) => `${item.model} / ${item.reasoning_effort}`),
    datasets: [
      {
        data: modelEffortChartRows.value.map((item) => item.request_count),
        backgroundColor: palette,
        borderColor: modelEffortChartRows.value.map((item) =>
          item.key !== '__other__' && item.key === selectedComboKey.value ? chartColors.value.accent : chartColors.value.stroke
        ),
        borderWidth: modelEffortChartRows.value.map((item) =>
          item.key !== '__other__' && item.key === selectedComboKey.value ? 3 : 1
        ),
      },
    ],
  }
})

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  animation: {
    duration: 0,
  },
  resizeDelay: 120,
  events: chartEvents,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const item = modelEffortChartRows.value[context.dataIndex] as OpenAIReasoningGuardModelEffortItem | undefined
          if (!item) return ''
          return `${item.model} / ${item.reasoning_effort}: ${formatNumber(item.request_count)}`
        },
      },
    },
  },
  cutout: '68%',
}))

function formatNumber(value?: number | null): string {
  return Number(value || 0).toLocaleString()
}

function formatCompactNumber(value: number): string {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1_000) return `${(value / 1_000).toFixed(1)}K`
  return value.toLocaleString()
}

function formatRatio(value: number): string {
  return `${(value * 100).toFixed(2)}%`
}

function syncGranularityFromRange(nextStart = startDate.value, nextEnd = endDate.value) {
  const start = new Date(nextStart)
  const end = new Date(nextEnd)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))
  granularity.value = daysDiff <= 1 ? 'hour' : 'day'
}

async function loadStats() {
  loading.value = true
  try {
    const nextStats = await usageAPI.getOpenAIReasoningGuardStats({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
    })
    stats.value = nextStats
    if (selectedComboKey.value) {
      const stillExists = (nextStats.model_efforts || []).some((item) => item.key === selectedComboKey.value)
      if (!stillExists) {
        selectedComboKey.value = null
      }
    }
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    loading.value = false
  }
}

function onDateRangeChange(range: { startDate: string; endDate: string; preset: string | null }) {
  syncGranularityFromRange(range.startDate, range.endDate)
  void loadStats()
}

function toggleCombo(key: string) {
  selectedComboKey.value = selectedComboKey.value === key ? null : key
}

function clearSelection() {
  selectedComboKey.value = null
}

function buildCodexRadarSrcdoc(html: string): string {
  const doc = new DOMParser().parseFromString(html, 'text/html')
  const modelSection = doc.querySelector('section.model-iq')
  const scoreArticle =
    modelSection?.querySelector('div.model-iq-grid article.model-iq-score') ||
    doc.querySelector('article.model-iq-score')

  if (!scoreArticle) {
    throw new Error('codexradar snippet not found')
  }

  const styles = Array.from(doc.querySelectorAll('style'))
    .map((style) => style.textContent || '')
    .join('\n')

  return `<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <base href="${CODEX_RADAR_URL}">
    <style>
${styles}
html, body {
  margin: 0;
  padding: 0;
  min-height: auto;
  background: transparent !important;
  overflow: hidden;
}
body {
  overflow: hidden;
}
.codexradar-embed-root {
  padding: 12px 16px 16px;
}
.codexradar-embed-root .shell {
  width: auto !important;
  margin: 0 !important;
  padding: 0 !important;
}
.codexradar-embed-root .model-iq-score {
  max-width: 920px;
  margin: 0 auto;
  box-shadow: none !important;
}
.codexradar-embed-root .model-iq-chart-view[hidden] {
  display: none !important;
}
.codexradar-embed-root .model-iq-chart-view:not([hidden]) {
  display: block !important;
}
.codexradar-embed-root .model-iq-chart svg {
  display: block;
  width: 100% !important;
  height: auto !important;
}
.codexradar-embed-root .model-iq-score-chip,
.codexradar-embed-root .model-iq-chart-controls {
  pointer-events: none;
}
    </style>
  </head>
  <body>
    <div class="codexradar-embed-root">
      ${scoreArticle.outerHTML}
    </div>
  </body>
</html>`
}

function syncCodexRadarFrameHeight() {
  requestAnimationFrame(() => {
    const frame = codexRadarFrame.value
    const doc = frame?.contentDocument
    if (!frame || !doc) return
    const root = doc.documentElement
    const body = doc.body
    const nextHeight = Math.max(
      root?.scrollHeight || 0,
      body?.scrollHeight || 0,
      root?.offsetHeight || 0,
      body?.offsetHeight || 0,
    )
    codexRadarFrameHeight.value = Math.max(Math.ceil(nextHeight), 420)
  })
}

async function loadCodexRadar() {
  codexRadarLoading.value = true
  codexRadarError.value = ''
  try {
    const response = await fetch(CODEX_RADAR_URL, {
      headers: {
        Accept: 'text/html',
      },
    })
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    const html = await response.text()
    codexRadarSrcdoc.value = buildCodexRadarSrcdoc(html)
  } catch (error) {
    codexRadarSrcdoc.value = ''
    codexRadarError.value = extractApiErrorMessage(error, t('common.error'))
  } finally {
    codexRadarLoading.value = false
  }
}

onMounted(() => {
  syncGranularityFromRange()
  void loadStats()
  void loadCodexRadar()
  window.addEventListener('resize', syncCodexRadarFrameHeight)
})

onUnmounted(() => {
  window.removeEventListener('resize', syncCodexRadarFrameHeight)
})
</script>
