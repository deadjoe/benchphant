<template>
  <div class="bg-white dark:bg-gray-800 shadow rounded-lg p-4">
    <div class="mb-4">
      <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100">
        Clone Connection
      </h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        Create a new connection based on an existing one
      </p>
    </div>

    <form @submit.prevent="handleSubmit" class="space-y-4">
      <!-- Source Connection -->
      <div>
        <label
          for="sourceConnection"
          class="block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          Source Connection
        </label>
        <select
          id="sourceConnection"
          v-model="selectedConnectionId"
          class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white sm:text-sm"
          required
        >
          <option value="">Select a connection</option>
          <option
            v-for="conn in connections"
            :key="conn.id"
            :value="conn.id"
          >
            {{ conn.name }}
          </option>
        </select>
      </div>

      <!-- New Connection Name -->
      <div>
        <label
          for="newName"
          class="block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          New Connection Name
        </label>
        <div class="mt-1">
          <input
            type="text"
            id="newName"
            v-model="newName"
            class="shadow-sm focus:ring-indigo-500 focus:border-indigo-500 block w-full sm:text-sm border-gray-300 rounded-md dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            required
            :placeholder="selectedConnection ? `Copy of ${selectedConnection.name}` : 'Enter new name'"
          />
        </div>
      </div>

      <!-- Clone Options -->
      <div class="space-y-2">
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
          Clone Options
        </label>
        <div class="space-y-2">
          <label class="flex items-center">
            <input
              type="checkbox"
              v-model="cloneOptions.includeTags"
              class="rounded border-gray-300 text-indigo-600 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50 dark:border-gray-600 dark:bg-gray-700"
            />
            <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">
              Include tags
            </span>
          </label>
          <label class="flex items-center">
            <input
              type="checkbox"
              v-model="cloneOptions.includeMetadata"
              class="rounded border-gray-300 text-indigo-600 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50 dark:border-gray-600 dark:bg-gray-700"
            />
            <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">
              Include metadata
            </span>
          </label>
        </div>
      </div>

      <!-- Submit Button -->
      <div class="flex justify-end">
        <button
          type="submit"
          :disabled="!canSubmit || loading"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed dark:focus:ring-offset-gray-800"
        >
          <SpinnerIcon
            v-if="loading"
            class="-ml-1 mr-2 h-5 w-5 animate-spin"
            aria-hidden="true"
          />
          {{ loading ? 'Cloning...' : 'Clone Connection' }}
        </button>
      </div>
    </form>

    <!-- Result Message -->
    <div
      v-if="resultMessage"
      class="mt-4 p-4 rounded-md"
      :class="{
        'bg-green-50 dark:bg-green-900': !error,
        'bg-red-50 dark:bg-red-900': error
      }"
    >
      <div class="flex">
        <div class="flex-shrink-0">
          <CheckCircleIcon
            v-if="!error"
            class="h-5 w-5 text-green-400"
            aria-hidden="true"
          />
          <XCircleIcon
            v-else
            class="h-5 w-5 text-red-400"
            aria-hidden="true"
          />
        </div>
        <div class="ml-3">
          <p
            class="text-sm font-medium"
            :class="{
              'text-green-800 dark:text-green-200': !error,
              'text-red-800 dark:text-red-200': error
            }"
          >
            {{ resultMessage }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import {
  CheckCircleIcon,
  XCircleIcon
} from '@heroicons/vue/24/outline'
import SpinnerIcon from '@heroicons/vue/24/outline/ArrowPathIcon'
import { useDatabaseStore } from '@/stores/database'

const props = defineProps({
  connections: {
    type: Array,
    required: true
  }
})

const emit = defineEmits(['cloned'])

const databaseStore = useDatabaseStore()
const selectedConnectionId = ref('')
const newName = ref('')
const loading = ref(false)
const resultMessage = ref('')
const error = ref(false)
const cloneOptions = ref({
  includeTags: true,
  includeMetadata: false
})

// Computed
const selectedConnection = computed(() => {
  return props.connections.find(conn => conn.id === selectedConnectionId.value)
})

const canSubmit = computed(() => {
  return selectedConnectionId.value && newName.value.trim()
})

// Methods
async function handleSubmit() {
  if (!canSubmit.value || loading.value) return
  
  try {
    loading.value = true
    error.value = false
    resultMessage.value = ''
    
    const clonedConnection = await databaseStore.cloneConnection(
      selectedConnectionId.value,
      newName.value,
      cloneOptions.value
    )
    
    resultMessage.value = 'Connection cloned successfully'
    emit('cloned', clonedConnection)
    
    // Reset form
    selectedConnectionId.value = ''
    newName.value = ''
  } catch (err) {
    error.value = true
    resultMessage.value = err.message || 'Failed to clone connection'
  } finally {
    loading.value = false
  }
}

// Watch for selected connection changes to update default name
watch(selectedConnection, (newConn) => {
  if (newConn && !newName.value) {
    newName.value = `Copy of ${newConn.name}`
  }
})
</script>
