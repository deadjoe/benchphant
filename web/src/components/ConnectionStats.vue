<template>
  <div class="bg-white dark:bg-gray-800 shadow rounded-lg p-4">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
        Connection Statistics
      </h3>
      <div class="flex items-center space-x-2">
        <select
          v-model="timeRange"
          class="block pl-3 pr-10 py-2 text-sm border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white"
        >
          <option value="day">Last 24 Hours</option>
          <option value="week">Last 7 Days</option>
          <option value="month">Last 30 Days</option>
          <option value="year">Last Year</option>
        </select>
        <button
          @click="refreshStats"
          :disabled="loading"
          class="inline-flex items-center p-2 border border-transparent rounded-md shadow-sm text-indigo-600 hover:bg-indigo-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:text-indigo-400 dark:hover:bg-gray-700 dark:focus:ring-offset-gray-800"
        >
          <ArrowPathIcon
            class="h-5 w-5"
            :class="{ 'animate-spin': loading }"
            aria-hidden="true"
          />
        </button>
      </div>
    </div>

    <!-- Stats Overview -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
      <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
        <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
          Total Queries
        </div>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">
          {{ formatNumber(stats.totalQueries) }}
        </div>
      </div>
      <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
        <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
          Average Response Time
        </div>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">
          {{ formatTime(stats.avgResponseTime) }}
        </div>
      </div>
      <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
        <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
          Peak Connections
        </div>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">
          {{ stats.peakConnections }}
        </div>
      </div>
      <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
        <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
          Success Rate
        </div>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">
          {{ formatPercentage(stats.successRate) }}
        </div>
      </div>
    </div>

    <!-- Usage Trends Chart -->
    <div class="mb-8">
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-4">
        Usage Trends
      </h4>
      <div class="h-64">
        <line-chart
          v-if="chartData"
          :data="chartData"
          :options="chartOptions"
        />
      </div>
    </div>

    <!-- Connection Details -->
    <div class="space-y-4">
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">
        Connection Details
      </h4>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
              >
                Connection
              </th>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
              >
                Queries
              </th>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
              >
                Avg Response
              </th>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
              >
                Success Rate
              </th>
              <th
                scope="col"
                class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider"
              >
                Last Used
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="conn in connectionStats" :key="conn.id">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <div class="flex-shrink-0 h-2 w-2 rounded-full" :class="getStatusColor(conn.status)"></div>
                  <div class="ml-4">
                    <div class="text-sm font-medium text-gray-900 dark:text-gray-100">
                      {{ conn.name }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400">
                      {{ conn.type }}
                    </div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm text-gray-900 dark:text-gray-100">
                  {{ formatNumber(conn.queries) }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm text-gray-900 dark:text-gray-100">
                  {{ formatTime(conn.avgResponse) }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="text-sm text-gray-900 dark:text-gray-100">
                  {{ formatPercentage(conn.successRate) }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                {{ formatDate(conn.lastUsed) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Line as LineChart } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend } from 'chart.js'
import { ArrowPathIcon } from '@heroicons/vue/24/outline'
import { useDatabaseStore } from '@/stores/database'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

const props = defineProps({
  connections: {
    type: Array,
    required: true
  }
})

const databaseStore = useDatabaseStore()
const timeRange = ref('day')
const loading = ref(false)
const stats = ref({
  totalQueries: 0,
  avgResponseTime: 0,
  peakConnections: 0,
  successRate: 0
})
const connectionStats = ref([])
const chartData = ref(null)

// Chart options
const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    y: {
      beginAtZero: true
    }
  },
  plugins: {
    legend: {
      position: 'top'
    }
  }
}

// Methods
async function refreshStats() {
  if (loading.value) return
  
  try {
    loading.value = true
    
    // Get connection stats
    const response = await databaseStore.getBatchConnectionStats(
      props.connections.map(conn => conn.id),
      timeRange.value
    )
    
    // Update stats
    stats.value = response.overall
    connectionStats.value = response.connections
    
    // Update chart data
    chartData.value = {
      labels: response.timeline.map(t => formatChartDate(t.timestamp)),
      datasets: [
        {
          label: 'Queries',
          data: response.timeline.map(t => t.queries),
          borderColor: 'rgb(99, 102, 241)',
          tension: 0.1
        },
        {
          label: 'Response Time (ms)',
          data: response.timeline.map(t => t.responseTime),
          borderColor: 'rgb(251, 146, 60)',
          tension: 0.1
        }
      ]
    }
  } catch (error) {
    console.error('Failed to refresh stats:', error)
  } finally {
    loading.value = false
  }
}

// Utility functions
function formatNumber(num) {
  return new Intl.NumberFormat().format(num)
}

function formatTime(ms) {
  if (ms < 1) return '0ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatPercentage(value) {
  return `${(value * 100).toFixed(1)}%`
}

function formatDate(date) {
  if (!date) return 'Never'
  return new Date(date).toLocaleString()
}

function formatChartDate(date) {
  const d = new Date(date)
  switch (timeRange.value) {
    case 'day':
      return d.toLocaleTimeString()
    case 'week':
      return d.toLocaleDateString()
    default:
      return d.toLocaleDateString()
  }
}

function getStatusColor(status) {
  switch (status) {
    case 'connected':
      return 'bg-green-500'
    case 'disconnected':
      return 'bg-red-500'
    default:
      return 'bg-gray-500'
  }
}

// Watch for time range changes
watch(timeRange, () => {
  refreshStats()
})

// Initial load
onMounted(() => {
  refreshStats()
})
</script>
