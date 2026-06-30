<template>
  <div class="card p-4">
    <div class="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
      <div>
        <div class="flex items-center gap-2">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">
            {{ t('admin.usage.openaiReasoningGuardModelEffortTitle') }}
          </h3>
          <span
            :class="[
              'inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium',
              runtime?.enabled
                ? 'bg-rose-100 text-rose-700 dark:bg-rose-900/30 dark:text-rose-300'
                : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
            ]"
          >
            {{ runtime?.enabled ? t('usage.guardEnabled') : t('usage.guardDisabled') }}
          </span>
        </div>
      </div>
      <div class="text-xs text-gray-500 dark:text-gray-400">
        {{ t('usage.guardStatusCode') }}: {{ runtime?.intercept_status_code || 502 }}
      </div>
    </div>

    <div class="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.6fr)_minmax(320px,1fr)]">
      <div class="min-w-0">
        <div v-if="loading" class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="chartData" class="h-48">
          <Line :data="chartData" :options="chartOptions" />
        </div>
        <div
          v-else
          class="flex h-48 items-center justify-center rounded-xl border border-dashed border-gray-200 text-sm text-gray-500 dark:border-gray-700 dark:text-gray-400"
        >
          {{ t('admin.dashboard.noDataAvailable') }}
        </div>
      </div>

      <div class="min-w-0 overflow-hidden rounded-xl border border-gray-100 dark:border-gray-700">
        <div class="overflow-auto">
          <table class="w-full text-xs">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr class="text-gray-500 dark:text-gray-400">
                <th class="px-3 py-3 text-left">{{ t('usage.requestedModel') }}</th>
                <th class="px-3 py-3 text-left">{{ t('usage.reasoningEffort') }}</th>
                <th class="px-3 py-3 text-right">{{ t('usage.requestCount') }}</th>
                <th class="px-3 py-3 text-right">{{ t('usage.matchCount') }}</th>
                <th class="px-3 py-3 text-right">{{ t('usage.matchRatio') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading">
                <td colspan="5" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                  {{ t('common.loading') }}
                </td>
              </tr>
              <tr v-else-if="rows.length === 0">
                <td colspan="5" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.noDataAvailable') }}
                </td>
              </tr>
              <tr
                v-for="item in rows"
                :key="item.key"
                class="border-t border-gray-100 dark:border-gray-700"
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Line } from 'vue-chartjs'
import { CategoryScale, Chart as ChartJS, Filler, Legend, LinearScale, LineElement, PointElement, Tooltip } from 'chart.js'
import type {
  AdminOpenAIReasoningGuardModelEffortItem,
  AdminOpenAIReasoningGuardModelEffortTrendPoint,
  AdminOpenAIReasoningGuardRuntime,
} from '@/api/admin/usage'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Legend, Filler)

interface Props {
  rows: AdminOpenAIReasoningGuardModelEffortItem[]
  trend: AdminOpenAIReasoningGuardModelEffortTrendPoint[]
  runtime: AdminOpenAIReasoningGuardRuntime | null
  loading: boolean
}

const props = defineProps<Props>()
const { t } = useI18n()

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))
const colors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  muted: isDarkMode.value ? '#9ca3af' : '#6b7280',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb',
  intercept: '#3b82f6',
  matchRatio: '#8b5cf6'
}))

const aggregateTrendPoints = computed(() => {
  const pointMap = new Map<
    string,
    { date: string; request_count: number; match_count: number; match_ratio: number; intercept_count: number }
  >()
  for (const point of props.trend) {
    const current = pointMap.get(point.date) || {
      date: point.date,
      request_count: 0,
      match_count: 0,
      match_ratio: 0,
      intercept_count: 0,
    }
    current.request_count += point.request_count
    current.match_count += point.match_count
    current.intercept_count += point.intercept_count
    pointMap.set(point.date, current)
  }
  return Array.from(pointMap.values()).map((point) => ({
    ...point,
    match_ratio: point.request_count > 0 ? point.match_count / point.request_count : 0,
  }))
})

const chartData = computed(() => {
  if (aggregateTrendPoints.value.length === 0) return null
  const c = colors.value
  return {
    labels: aggregateTrendPoints.value.map((point) => point.date),
    datasets: [
      {
        label: t('usage.interceptCount'),
        data: aggregateTrendPoints.value.map((point) => point.intercept_count),
        borderColor: c.intercept,
        backgroundColor: `${c.intercept}20`,
        fill: true,
        tension: 0.3,
        pointRadius: 2,
        pointHoverRadius: 4,
        yAxisID: 'y',
      },
      {
        label: t('usage.matchRatio'),
        data: aggregateTrendPoints.value.map((point) => point.match_ratio * 100),
        borderColor: c.matchRatio,
        backgroundColor: `${c.matchRatio}20`,
        borderDash: [5, 5],
        fill: false,
        tension: 0.3,
        pointRadius: 2,
        pointHoverRadius: 4,
        yAxisID: 'yPercent',
      },
    ],
  }
})

const chartOptions = computed(() => {
  const c = colors.value
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
      intersect: false,
      mode: 'index' as const
    },
    animation: { duration: 0 },
    plugins: {
      legend: {
        position: 'top' as const,
        labels: {
          color: c.text,
          usePointStyle: true,
          pointStyle: 'circle',
          padding: 15,
          font: { size: 11 }
        }
      },
      tooltip: {
        callbacks: {
          label: (context: any) => {
            if (context.dataset.yAxisID === 'yPercent') {
              return `${context.dataset.label}: ${Number(context.raw || 0).toFixed(1)}%`
            }
            return `${context.dataset.label}: ${formatNumber(Number(context.raw || 0))}`
          },
          footer: (tooltipItems: any) => {
            const dataIndex = tooltipItems[0]?.dataIndex
            const point = dataIndex !== undefined ? aggregateTrendPoints.value[dataIndex] : null
            if (!point) return ''
            return `${t('usage.requestCount')}: ${formatNumber(point.request_count)} | ${t('usage.matchCount')}: ${formatNumber(point.match_count)}`
          }
        }
      }
    },
    scales: {
      x: {
        grid: { color: c.grid, borderDash: [4, 4] },
        ticks: { color: c.text, font: { size: 10 } }
      },
      y: {
        beginAtZero: true,
        grid: { color: c.grid, borderDash: [4, 4] },
        ticks: {
          color: c.text,
          precision: 0
        }
      },
      yPercent: {
        position: 'right' as const,
        min: 0,
        max: 100,
        grid: {
          drawOnChartArea: false
        },
        ticks: {
          color: c.matchRatio,
          callback: (value: string | number) => `${value}%`,
          font: { size: 10 }
        }
      }
    }
  }
})

const formatNumber = (value: number): string => value.toLocaleString()
const formatRatio = (value: number): string => `${(value * 100).toFixed(1)}%`
</script>
