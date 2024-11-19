<template>
  <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg">
    <div class="px-4 py-5 sm:px-6">
      <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">Connection History</h3>
      <p class="mt-1 max-w-2xl text-sm text-gray-500 dark:text-gray-400">
        Recent connection attempts and status changes
      </p>
    </div>

    <div class="border-t border-gray-200 dark:border-gray-700">
      <div v-if="loading" class="p-4 text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500 mx-auto"></div>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">Loading history...</p>
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

      <div v-else-if="history.length === 0" class="p-4 text-center">
        <ClockIcon class="mx-auto h-12 w-12 text-gray-400" aria-hidden="true" />
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No history</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          No connection history available yet.
        </p>
      </div>

      <ul v-else role="list" class="divide-y divide-gray-200 dark:divide-gray-700">
        <li v-for="(event, index) in history" :key="index" class="px-4 py-4 sm:px-6">
          <div class="flex items-center justify-between">
            <div class="flex flex-col min-w-0 flex-1">
              <div class="flex items-center">
                <div class="flex-shrink-0">
                  <span
                    :class="[
                      event.success
                        ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                        : 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
                      'h-8 w-8 rounded-full flex items-center justify-center',
                    ]"
                  >
                    <CheckCircleIcon
                      v-if="event.success"
                      class="h-5 w-5"
                      aria-hidden="true"
                    />
                    <XCircleIcon
                      v-else
                      class="h-5 w-5"
                      aria-hidden="true"
                    />
                  </span>
                </div>
                <div class="ml-4">
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ event.action }}
                  </p>
                  <p class="text-sm text-gray-500 dark:text-gray-400">
                    {{ event.timestamp | formatDate }}
                  </p>
                </div>
              </div>
              <div class="mt-2 text-sm text-gray-500 dark:text-gray-400">
                <div v-if="event.details" class="mt-1">
                  <dl class="grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-2">
                    <div v-if="event.details.responseTime" class="sm:col-span-1">
                      <dt class="font-medium">Response Time</dt>
                      <dd>{{ event.details.responseTime }}ms</dd>
                    </div>
                    <div v-if="event.details.serverVersion" class="sm:col-span-1">
                      <dt class="font-medium">Server Version</dt>
                      <dd>{{ event.details.serverVersion }}</dd>
                    </div>
                    <div v-if="event.details.error" class="sm:col-span-2">
                      <dt class="font-medium">Error</dt>
                      <dd class="text-red-600 dark:text-red-400">{{ event.details.error }}</dd>
                    </div>
                  </dl>
                </div>
              </div>
            </div>
            <div class="ml-4 flex-shrink-0">
              <span
                :class="[
                  event.success
                    ? 'text-green-600 dark:text-green-400'
                    : 'text-red-600 dark:text-red-400',
                  'text-sm font-medium',
                ]"
              >
                {{ event.success ? 'Success' : 'Failed' }}
              </span>
            </div>
          </div>
        </li>
      </ul>

      <div v-if="hasMore" class="px-4 py-4 sm:px-6 border-t border-gray-200 dark:border-gray-700">
        <button
          type="button"
          @click="$emit('load-more')"
          :disabled="loadingMore"
          class="w-full flex justify-center items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-indigo-700 bg-indigo-100 hover:bg-indigo-200 dark:text-indigo-300 dark:bg-indigo-900 dark:hover:bg-indigo-800 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <span v-if="loadingMore" class="mr-2">
            <svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
            </svg>
          </span>
          Load More
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { CheckCircleIcon, XCircleIcon, ClockIcon } from '@heroicons/vue/24/outline'

const props = defineProps({
  history: {
    type: Array,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  },
  loadingMore: {
    type: Boolean,
    default: false
  },
  hasMore: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: null
  }
})

defineEmits(['load-more'])
</script>
