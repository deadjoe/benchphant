<template>
  <div class="space-y-4">
    <!-- Import Section -->
    <div class="bg-white dark:bg-gray-800 shadow rounded-lg p-4">
      <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
        Import Connections
      </h3>
      
      <div class="flex items-center justify-center w-full">
        <label
          class="flex flex-col items-center justify-center w-full h-32 border-2 border-gray-300 border-dashed rounded-lg cursor-pointer bg-gray-50 dark:hover:bg-gray-700 dark:bg-gray-700 hover:bg-gray-100 dark:border-gray-600 dark:hover:border-gray-500"
          :class="{ 'border-indigo-500 dark:border-indigo-400': isDragging }"
          @dragenter.prevent="isDragging = true"
          @dragleave.prevent="isDragging = false"
          @dragover.prevent
          @drop.prevent="handleFileDrop"
        >
          <div class="flex flex-col items-center justify-center pt-5 pb-6">
            <svg
              class="w-8 h-8 mb-4 text-gray-500 dark:text-gray-400"
              aria-hidden="true"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 20 16"
            >
              <path
                stroke="currentColor"
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13 13h3a3 3 0 0 0 0-6h-.025A5.56 5.56 0 0 0 16 6.5 5.5 5.5 0 0 0 5.207 5.021C5.137 5.017 5.071 5 5 5a4 4 0 0 0 0 8h2.167M10 15V6m0 0L8 8m2-2 2 2"
              />
            </svg>
            <p class="mb-2 text-sm text-gray-500 dark:text-gray-400">
              <span class="font-semibold">Click to upload</span> or drag and drop
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">JSON files only</p>
          </div>
          <input
            type="file"
            class="hidden"
            accept=".json"
            @change="handleFileSelect"
          />
        </label>
      </div>

      <!-- Import Progress -->
      <div v-if="importProgress" class="mt-4">
        <div class="flex justify-between mb-1">
          <span class="text-sm font-medium text-indigo-700 dark:text-indigo-300">
            Importing...
          </span>
          <span class="text-sm font-medium text-indigo-700 dark:text-indigo-300">
            {{ importProgress }}%
          </span>
        </div>
        <div class="w-full bg-gray-200 rounded-full h-2 dark:bg-gray-700">
          <div
            class="bg-indigo-600 h-2 rounded-full"
            :style="{ width: `${importProgress}%` }"
          ></div>
        </div>
      </div>

      <!-- Import Results -->
      <div v-if="importResults" class="mt-4">
        <div
          class="p-4 rounded-lg"
          :class="{
            'bg-green-50 dark:bg-green-900': importResults.success,
            'bg-red-50 dark:bg-red-900': !importResults.success
          }"
        >
          <div class="flex">
            <div class="flex-shrink-0">
              <CheckCircleIcon
                v-if="importResults.success"
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
              <h3
                class="text-sm font-medium"
                :class="{
                  'text-green-800 dark:text-green-200': importResults.success,
                  'text-red-800 dark:text-red-200': !importResults.success
                }"
              >
                {{ importResults.message }}
              </h3>
              <div
                v-if="importResults.details"
                class="mt-2 text-sm"
                :class="{
                  'text-green-700 dark:text-green-300': importResults.success,
                  'text-red-700 dark:text-red-300': !importResults.success
                }"
              >
                <ul class="list-disc pl-5 space-y-1">
                  <li v-for="(detail, index) in importResults.details" :key="index">
                    {{ detail }}
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Export Section -->
    <div class="bg-white dark:bg-gray-800 shadow rounded-lg p-4">
      <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4">
        Export Connections
      </h3>
      
      <div class="space-y-4">
        <!-- Connection Selection -->
        <div v-if="connections.length > 0">
          <label class="flex items-center space-x-2">
            <input
              type="checkbox"
              v-model="selectAll"
              @change="toggleSelectAll"
              class="rounded border-gray-300 text-indigo-600 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50 dark:border-gray-600 dark:bg-gray-700"
            />
            <span class="text-sm text-gray-700 dark:text-gray-300">Select All</span>
          </label>

          <div class="mt-2 space-y-2">
            <label
              v-for="conn in connections"
              :key="conn.id"
              class="flex items-center space-x-2"
            >
              <input
                type="checkbox"
                v-model="selectedConnections"
                :value="conn.id"
                class="rounded border-gray-300 text-indigo-600 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50 dark:border-gray-600 dark:bg-gray-700"
              />
              <span class="text-sm text-gray-700 dark:text-gray-300">
                {{ conn.name }}
              </span>
            </label>
          </div>
        </div>
        
        <div v-else class="text-sm text-gray-500 dark:text-gray-400">
          No connections available to export.
        </div>

        <!-- Export Button -->
        <button
          @click="exportConnections"
          :disabled="selectedConnections.length === 0 || exporting"
          class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed dark:focus:ring-offset-gray-800"
        >
          <ArrowDownTrayIcon
            v-if="!exporting"
            class="-ml-1 mr-2 h-5 w-5"
            aria-hidden="true"
          />
          <SpinnerIcon
            v-else
            class="-ml-1 mr-2 h-5 w-5 animate-spin"
            aria-hidden="true"
          />
          {{ exporting ? 'Exporting...' : 'Export Selected' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import {
  ArrowDownTrayIcon,
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

const databaseStore = useDatabaseStore()
const isDragging = ref(false)
const importProgress = ref(null)
const importResults = ref(null)
const selectedConnections = ref([])
const exporting = ref(false)

// Computed
const selectAll = computed({
  get: () => selectedConnections.value.length === props.connections.length,
  set: (value) => {
    selectedConnections.value = value
      ? props.connections.map(conn => conn.id)
      : []
  }
})

// Methods
function toggleSelectAll() {
  if (selectAll.value) {
    selectedConnections.value = props.connections.map(conn => conn.id)
  } else {
    selectedConnections.value = []
  }
}

function handleFileDrop(event) {
  isDragging.value = false
  const file = event.dataTransfer.files[0]
  if (file && file.type === 'application/json') {
    importFile(file)
  }
}

function handleFileSelect(event) {
  const file = event.target.files[0]
  if (file) {
    importFile(file)
  }
}

async function importFile(file) {
  try {
    importProgress.value = 0
    importResults.value = null
    
    const response = await databaseStore.importConnections(file, (progress) => {
      importProgress.value = Math.round(progress)
    })
    
    importResults.value = {
      success: true,
      message: 'Import completed successfully',
      details: [
        `${response.imported} connections imported`,
        `${response.updated} connections updated`,
        `${response.skipped} connections skipped`
      ]
    }
  } catch (error) {
    importResults.value = {
      success: false,
      message: 'Import failed',
      details: [error.message]
    }
  } finally {
    importProgress.value = null
  }
}

async function exportConnections() {
  if (selectedConnections.value.length === 0) return
  
  try {
    exporting.value = true
    const blob = await databaseStore.exportConnections(selectedConnections.value)
    
    // Create download link
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `connections_${new Date().toISOString().slice(0, 10)}.json`
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (error) {
    console.error('Export failed:', error)
  } finally {
    exporting.value = false
  }
}
</script>
