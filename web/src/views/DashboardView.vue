<template>
  <div class="space-y-6">
    <div class="bg-white dark:bg-gray-800 shadow px-4 py-5 sm:rounded-lg sm:p-6">
      <div class="md:grid md:grid-cols-3 md:gap-6">
        <div class="md:col-span-1">
          <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">System Status</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Overview of your database connections and benchmark status.
          </p>
        </div>
        <div class="mt-5 md:mt-0 md:col-span-2">
          <div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
            <div class="bg-gray-50 dark:bg-gray-700 overflow-hidden shadow rounded-lg">
              <div class="px-4 py-5 sm:p-6">
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                  Active Connections
                </dt>
                <dd class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">
                  {{ stats.activeConnections }}
                </dd>
              </div>
            </div>

            <div class="bg-gray-50 dark:bg-gray-700 overflow-hidden shadow rounded-lg">
              <div class="px-4 py-5 sm:p-6">
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                  Running Benchmarks
                </dt>
                <dd class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">
                  {{ stats.runningBenchmarks }}
                </dd>
              </div>
            </div>

            <div class="bg-gray-50 dark:bg-gray-700 overflow-hidden shadow rounded-lg">
              <div class="px-4 py-5 sm:p-6">
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400 truncate">
                  Total Reports
                </dt>
                <dd class="mt-1 text-3xl font-semibold text-gray-900 dark:text-white">
                  {{ stats.totalReports }}
                </dd>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="bg-white dark:bg-gray-800 shadow px-4 py-5 sm:rounded-lg sm:p-6">
      <div class="md:grid md:grid-cols-3 md:gap-6">
        <div class="md:col-span-1">
          <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">Recent Activity</h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Latest benchmark runs and database connections.
          </p>
        </div>
        <div class="mt-5 md:mt-0 md:col-span-2">
          <div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
            <table class="min-w-full divide-y divide-gray-300 dark:divide-gray-600">
              <thead class="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 dark:text-white sm:pl-6">Type</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Description</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Status</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold text-gray-900 dark:text-white">Time</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 dark:divide-gray-600 bg-white dark:bg-gray-800">
                <tr v-for="activity in recentActivity" :key="activity.id">
                  <td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 dark:text-white sm:pl-6">
                    {{ activity.type }}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-400">
                    {{ activity.description }}
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm">
                    <span :class="[
                      activity.status === 'completed' ? 'text-green-800 bg-green-100 dark:text-green-400' :
                      activity.status === 'running' ? 'text-yellow-800 bg-yellow-100 dark:text-yellow-400' :
                      'text-red-800 bg-red-100 dark:text-red-400',
                      'inline-flex rounded-full px-2 text-xs font-semibold leading-5'
                    ]">
                      {{ activity.status }}
                    </span>
                  </td>
                  <td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500 dark:text-gray-400">
                    {{ activity.time }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const stats = ref({
  activeConnections: 0,
  runningBenchmarks: 0,
  totalReports: 0
})

const recentActivity = ref([])

onMounted(async () => {
  try {
    // Fetch dashboard stats
    const statsResponse = await fetch('/api/stats', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    if (statsResponse.ok) {
      stats.value = await statsResponse.json()
    }

    // Fetch recent activity
    const activityResponse = await fetch('/api/activity', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    if (activityResponse.ok) {
      recentActivity.value = await activityResponse.json()
    }
  } catch (error) {
    console.error('Failed to fetch dashboard data:', error)
  }
})
</script>
