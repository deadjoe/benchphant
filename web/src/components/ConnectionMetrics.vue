<template>
  <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg">
    <div class="px-4 py-5 sm:px-6">
      <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">Connection Metrics</h3>
      <p class="mt-1 max-w-2xl text-sm text-gray-500 dark:text-gray-400">
        Real-time monitoring and performance statistics
      </p>
    </div>

    <div class="border-t border-gray-200 dark:border-gray-700">
      <div v-if="loading" class="p-4 text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500 mx-auto"></div>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">Loading metrics...</p>
      </div>

      <div v-else-if="error" class="p-4">
        <div class="rounded-md bg-red-50 dark:bg-red-900 p-4">
          <div class="flex">
            <XCircleIcon class="h-5 w-5 text-red-400" aria-hidden="true" />
            <div class="ml-3">
              <h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
              <div class="mt-2 text-sm text-red-700 dark:text-red-300">
                {{ error }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="px-4 py-5 sm:p-6">
        <!-- Performance Metrics -->
        <div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
          <div class="relative overflow-hidden rounded-lg bg-white dark:bg-gray-900 px-4 py-5 shadow sm:px-6">
            <dt>
              <div class="absolute rounded-md bg-indigo-500 p-3">
                <ClockIcon class="h-6 w-6 text-white" aria-hidden="true" />
              </div>
              <p class="ml-16 truncate text-sm font-medium text-gray-500 dark:text-gray-400">Average Response Time</p>
            </dt>
            <dd class="ml-16 flex items-baseline">
              <p class="text-2xl font-semibold text-gray-900 dark:text-white">{{ formatDuration(metrics.avgResponseTime) }}</p>
              <p
                :class="[
                  metrics.responseTimeTrend > 0 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400',
                  'ml-2 flex items-baseline text-sm font-semibold',
                ]"
              >
                <ArrowUpIcon
                  v-if="metrics.responseTimeTrend > 0"
                  class="h-5 w-5 flex-shrink-0 self-center text-red-500"
                  aria-hidden="true"
                />
                <ArrowDownIcon
                  v-else
                  class="h-5 w-5 flex-shrink-0 self-center text-green-500"
                  aria-hidden="true"
                />
                <span class="sr-only">
                  {{ metrics.responseTimeTrend > 0 ? 'Increased' : 'Decreased' }} by
                </span>
                {{ Math.abs(metrics.responseTimeTrend) }}%
              </p>
            </dd>
          </div>

          <div class="relative overflow-hidden rounded-lg bg-white dark:bg-gray-900 px-4 py-5 shadow sm:px-6">
            <dt>
              <div class="absolute rounded-md bg-indigo-500 p-3">
                <BoltIcon class="h-6 w-6 text-white" aria-hidden="true" />
              </div>
              <p class="ml-16 truncate text-sm font-medium text-gray-500 dark:text-gray-400">Active Queries</p>
            </dt>
            <dd class="ml-16 flex items-baseline">
              <p class="text-2xl font-semibold text-gray-900 dark:text-white">{{ metrics.activeQueries }}</p>
              <p
                :class="[
                  metrics.activeQueriesTrend > 0 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400',
                  'ml-2 flex items-baseline text-sm font-semibold',
                ]"
              >
                <ArrowUpIcon
                  v-if="metrics.activeQueriesTrend > 0"
                  class="h-5 w-5 flex-shrink-0 self-center text-red-500"
                  aria-hidden="true"
                />
                <ArrowDownIcon
                  v-else
                  class="h-5 w-5 flex-shrink-0 self-center text-green-500"
                  aria-hidden="true"
                />
                <span class="sr-only">
                  {{ metrics.activeQueriesTrend > 0 ? 'Increased' : 'Decreased' }} by
                </span>
                {{ Math.abs(metrics.activeQueriesTrend) }}%
              </p>
            </dd>
          </div>

          <div class="relative overflow-hidden rounded-lg bg-white dark:bg-gray-900 px-4 py-5 shadow sm:px-6">
            <dt>
              <div class="absolute rounded-md bg-indigo-500 p-3">
                <CircleStackIcon class="h-6 w-6 text-white" aria-hidden="true" />
              </div>
              <p class="ml-16 truncate text-sm font-medium text-gray-500 dark:text-gray-400">Memory Usage</p>
            </dt>
            <dd class="ml-16 flex items-baseline">
              <p class="text-2xl font-semibold text-gray-900 dark:text-white">{{ formatBytes(metrics.memoryUsage) }}</p>
              <p
                :class="[
                  metrics.memoryUsageTrend > 0 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400',
                  'ml-2 flex items-baseline text-sm font-semibold',
                ]"
              >
                <ArrowUpIcon
                  v-if="metrics.memoryUsageTrend > 0"
                  class="h-5 w-5 flex-shrink-0 self-center text-red-500"
                  aria-hidden="true"
                />
                <ArrowDownIcon
                  v-else
                  class="h-5 w-5 flex-shrink-0 self-center text-green-500"
                  aria-hidden="true"
                />
                <span class="sr-only">
                  {{ metrics.memoryUsageTrend > 0 ? 'Increased' : 'Decreased' }} by
                </span>
                {{ Math.abs(metrics.memoryUsageTrend) }}%
              </p>
            </dd>
          </div>

          <div class="relative overflow-hidden rounded-lg bg-white dark:bg-gray-900 px-4 py-5 shadow sm:px-6">
            <dt>
              <div class="absolute rounded-md bg-indigo-500 p-3">
                <CpuChipIcon class="h-6 w-6 text-white" aria-hidden="true" />
              </div>
              <p class="ml-16 truncate text-sm font-medium text-gray-500 dark:text-gray-400">CPU Usage</p>
            </dt>
            <dd class="ml-16 flex items-baseline">
              <p class="text-2xl font-semibold text-gray-900 dark:text-white">{{ metrics.cpuUsage }}%</p>
              <p
                :class="[
                  metrics.cpuUsageTrend > 0 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400',
                  'ml-2 flex items-baseline text-sm font-semibold',
                ]"
              >
                <ArrowUpIcon
                  v-if="metrics.cpuUsageTrend > 0"
                  class="h-5 w-5 flex-shrink-0 self-center text-red-500"
                  aria-hidden="true"
                />
                <ArrowDownIcon
                  v-else
                  class="h-5 w-5 flex-shrink-0 self-center text-green-500"
                  aria-hidden="true"
                />
                <span class="sr-only">
                  {{ metrics.cpuUsageTrend > 0 ? 'Increased' : 'Decreased' }} by
                </span>
                {{ Math.abs(metrics.cpuUsageTrend) }}%
              </p>
            </dd>
          </div>
        </div>

        <!-- Query Statistics -->
        <div class="mt-8">
          <h4 class="text-base font-medium text-gray-900 dark:text-white">Query Statistics</h4>
          <dl class="mt-5 grid grid-cols-1 gap-5 sm:grid-cols-2">
            <div class="relative overflow-hidden rounded-lg bg-white dark:bg-gray-900 px-4 py-5 shadow sm:px-6">
              <dt class="truncate text-sm font-medium text-gray-500 dark:text-gray-400">Queries Per Second</dt>
              <dd class="mt-1 text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ metrics.queriesPerSecond }}
              </dd>
            </div>
            <div class="relative overflow-hidden rounded-lg bg-white dark:bg-gray-900 px-4 py-5 shadow sm:px-6">
              <dt class="truncate text-sm font-medium text-gray-500 dark:text-gray-400">Slow Queries</dt>
              <dd class="mt-1 text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ metrics.slowQueries }}
              </dd>
            </div>
          </dl>
        </div>

        <!-- Connection Pool -->
        <div class="mt-8">
          <h4 class="text-base font-medium text-gray-900 dark:text-white">Connection Pool</h4>
          <dl class="mt-5 grid grid-cols-1 gap-5 sm:grid-cols-3">
            <div class="relative overflow-hidden rounded-lg bg-white dark:bg-gray-900 px-4 py-5 shadow sm:px-6">
              <dt class="truncate text-sm font-medium text-gray-500 dark:text-gray-400">Active Connections</dt>
              <dd class="mt-1 text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ metrics.activeConnections }}
              </dd>
            </div>
            <div class="relative overflow-hidden rounded-lg bg-white dark:bg-gray-900 px-4 py-5 shadow sm:px-6">
              <dt class="truncate text-sm font-medium text-gray-500 dark:text-gray-400">Idle Connections</dt>
              <dd class="mt-1 text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ metrics.idleConnections }}
              </dd>
            </div>
            <div class="relative overflow-hidden rounded-lg bg-white dark:bg-gray-900 px-4 py-5 shadow sm:px-6">
              <dt class="truncate text-sm font-medium text-gray-500 dark:text-gray-400">Max Connections</dt>
              <dd class="mt-1 text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">
                {{ metrics.maxConnections }}
              </dd>
            </div>
          </dl>
        </div>

        <!-- Refresh Button -->
        <div class="mt-8 flex justify-end">
          <button
            type="button"
            @click="$emit('refresh')"
            :disabled="loading"
            class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 dark:bg-gray-700 dark:text-white dark:ring-gray-600 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <ArrowPathIcon
              :class="[loading ? 'animate-spin' : '', '-ml-0.5 mr-1.5 h-5 w-5']"
              aria-hidden="true"
            />
            Refresh
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {
  ArrowUpIcon,
  ArrowDownIcon,
  ClockIcon,
  BoltIcon,
  CircleStackIcon,
  CpuChipIcon,
  ArrowPathIcon,
  XCircleIcon
} from '@heroicons/vue/24/outline'
import { formatDuration } from '@/utils/date'

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}

defineProps({
  metrics: {
    type: Object,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: null
  }
})

defineEmits(['refresh'])
</script>
