<template>
  <div class="bg-white dark:bg-gray-800 shadow rounded-lg p-4">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
        Connection Monitor
      </h3>
      <div class="flex items-center space-x-2">
        <button
          @click="startMonitoring"
          :disabled="isMonitoring"
          :class="[
            'px-3 py-1 rounded-md text-sm font-medium',
            isMonitoring
              ? 'bg-gray-100 text-gray-400 dark:bg-gray-700 dark:text-gray-500'
              : 'bg-green-600 text-white hover:bg-green-700 dark:hover:bg-green-500'
          ]"
        >
          Start
        </button>
        <button
          @click="stopMonitoring"
          :disabled="!isMonitoring"
          :class="[
            'px-3 py-1 rounded-md text-sm font-medium',
            !isMonitoring
              ? 'bg-gray-100 text-gray-400 dark:bg-gray-700 dark:text-gray-500'
              : 'bg-red-600 text-white hover:bg-red-700 dark:hover:bg-red-500'
          ]"
        >
          Stop
        </button>
      </div>
    </div>

    <!-- Metrics Overview -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
        <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
          Active Connections
        </div>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">
          {{ metrics.activeConnections }}
        </div>
      </div>
      <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
        <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
          Average Response Time
        </div>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">
          {{ formatTime(metrics.avgResponseTime) }}
        </div>
      </div>
      <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
        <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
          Total Queries
        </div>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-gray-100">
          {{ metrics.totalQueries }}
        </div>
      </div>
    </div>

    <!-- Connection List with Metrics -->
    <div class="space-y-4">
      <div v-for="conn in connections" :key="conn.id" class="border dark:border-gray-700 rounded-lg p-4">
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center space-x-2">
            <div class="w-2 h-2 rounded-full" :class="getStatusColor(conn.status)"></div>
            <h4 class="font-medium text-gray-900 dark:text-gray-100">{{ conn.name }}</h4>
          </div>
          <span class="text-sm text-gray-500 dark:text-gray-400">
            {{ formatUptime(conn.uptime) }}
          </span>
        </div>
        
        <!-- Connection Metrics -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mt-3">
          <div>
            <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
              Queries/sec
            </div>
            <div class="mt-1 text-lg font-medium text-gray-900 dark:text-gray-100">
              {{ formatNumber(conn.metrics?.qps || 0) }}
            </div>
          </div>
          <div>
            <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
              Response Time
            </div>
            <div class="mt-1 text-lg font-medium text-gray-900 dark:text-gray-100">
              {{ formatTime(conn.metrics?.responseTime || 0) }}
            </div>
          </div>
          <div>
            <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
              Active Sessions
            </div>
            <div class="mt-1 text-lg font-medium text-gray-900 dark:text-gray-100">
              {{ conn.metrics?.activeSessions || 0 }}
            </div>
          </div>
          <div>
            <div class="text-sm font-medium text-gray-500 dark:text-gray-400">
              Memory Usage
            </div>
            <div class="mt-1 text-lg font-medium text-gray-900 dark:text-gray-100">
              {{ formatBytes(conn.metrics?.memoryUsage || 0) }}
            </div>
          </div>
        </div>

        <!-- Mini Chart -->
        <div class="h-20 mt-4">
          <line-chart
            v-if="conn.metrics?.history"
            :data="conn.metrics.history"
            :options="chartOptions"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Line as LineChart } from 'vue-chartjs'
import { 
  Chart as ChartJS, 
  CategoryScale, 
  LinearScale, 
  PointElement, 
  LineElement, 
  Title, 
  Tooltip, 
  Legend, 
  Filler 
} from 'chart.js'
import { useDatabaseStore } from '@/stores/database'

// Only register Chart.js in non-test environment
if (process.env.NODE_ENV !== 'test') {
  ChartJS.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend,
    Filler
  )
}

const props = defineProps({
  connections: {
    type: Array,
    required: true
  }
})

const databaseStore = useDatabaseStore()
const isMonitoring = ref(false)
const monitoringInterval = ref(null)
const metrics = ref({
  activeConnections: 0,
  avgResponseTime: 0,
  totalQueries: 0
})

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
      display: false
    }
  }
}

// Start monitoring
function startMonitoring() {
  if (isMonitoring.value) return
  
  isMonitoring.value = true
  monitoringInterval.value = setInterval(async () => {
    try {
      await databaseStore.updateConnectionMetrics()
      updateMetrics()
    } catch (error) {
      console.error('Failed to update metrics:', error)
    }
  }, 5000) // Update every 5 seconds
}

// Stop monitoring
function stopMonitoring() {
  if (!isMonitoring.value) return
  
  isMonitoring.value = false
  if (monitoringInterval.value) {
    clearInterval(monitoringInterval.value)
    monitoringInterval.value = null
  }
}

// Update overall metrics
function updateMetrics() {
  const activeConns = props.connections.filter(conn => conn.status === 'connected').length
  const totalQueries = props.connections.reduce((sum, conn) => sum + (conn.metrics?.totalQueries || 0), 0)
  const avgResponse = props.connections.reduce((sum, conn) => {
    return conn.metrics?.responseTime ? sum + conn.metrics.responseTime : sum
  }, 0) / activeConns || 0

  metrics.value = {
    activeConnections: activeConns,
    avgResponseTime: avgResponse,
    totalQueries: totalQueries
  }
}

// Utility functions
function formatTime(ms) {
  if (ms < 1) return '0ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

function formatNumber(num) {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2
  }).format(num)
}

function formatBytes(bytes) {
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes
  let unitIndex = 0
  
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }
  
  return `${size.toFixed(1)}${units[unitIndex]}`
}

function formatUptime(seconds) {
  if (!seconds) return 'N/A'
  
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
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

// Lifecycle hooks
onMounted(() => {
  updateMetrics()
})

onUnmounted(() => {
  stopMonitoring()
})
</script>
