<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="md:flex md:items-center md:justify-between">
      <div class="min-w-0 flex-1">
        <h2 class="text-2xl font-bold leading-7 text-gray-900 dark:text-white sm:truncate sm:text-3xl sm:tracking-tight">
          Database Connections
        </h2>
      </div>
      <div class="mt-4 flex md:ml-4 md:mt-0 space-x-2">
        <button
          v-if="selectedConnections.length > 0"
          type="button"
          @click="batchTest"
          :disabled="batchLoading"
          class="inline-flex items-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 dark:bg-gray-700 dark:text-white dark:ring-gray-600 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <ArrowPathIcon class="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
          Test Selected
        </button>
        <button
          v-if="selectedConnections.length > 0"
          type="button"
          @click="confirmBatchDelete"
          class="inline-flex items-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600"
        >
          <TrashIcon class="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
          Delete Selected
        </button>
        <button
          type="button"
          @click="showNewConnectionModal = true"
          class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
          <PlusIcon class="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
          New Connection
        </button>
      </div>
    </div>

    <!-- Connection List -->
    <div class="bg-white dark:bg-gray-800 shadow sm:rounded-lg">
      <div v-if="loading" class="p-4 text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-500 mx-auto"></div>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">Loading connections...</p>
      </div>

      <div v-else-if="error" class="p-4">
        <div class="rounded-md bg-red-50 dark:bg-red-900 p-4">
          <div class="flex">
            <ExclamationCircleIcon class="h-5 w-5 text-red-400" aria-hidden="true" />
            <div class="ml-3">
              <h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
              <div class="mt-2 text-sm text-red-700 dark:text-red-300">
                {{ error }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-else-if="connections.length === 0" class="p-4 text-center">
        <DatabaseIcon class="mx-auto h-12 w-12 text-gray-400" aria-hidden="true" />
        <h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No connections</h3>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Get started by creating a new database connection.
        </p>
        <div class="mt-6">
          <button
            type="button"
            @click="showNewConnectionModal = true"
            class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
          >
            <PlusIcon class="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
            New Connection
          </button>
        </div>
      </div>

      <div v-else>
        <!-- Batch Selection Header -->
        <div class="px-4 py-3 sm:px-6 border-b border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div class="flex items-center">
              <input
                type="checkbox"
                :checked="isAllSelected"
                @change="toggleSelectAll"
                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 dark:border-gray-600 dark:bg-gray-700"
              />
              <span class="ml-3 text-sm text-gray-500 dark:text-gray-400">
                {{ selectedConnections.length }} selected
              </span>
            </div>
            <div class="flex items-center space-x-2">
              <select
                v-model="filterType"
                class="block w-full rounded-md border-0 py-1.5 pl-3 pr-10 text-gray-900 ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6 dark:bg-gray-700 dark:text-white dark:ring-gray-600"
              >
                <option value="">All Types</option>
                <option value="mysql">MySQL</option>
                <option value="postgresql">PostgreSQL</option>
                <option value="mongodb">MongoDB</option>
                <option value="redis">Redis</option>
              </select>
              <select
                v-model="filterStatus"
                class="block w-full rounded-md border-0 py-1.5 pl-3 pr-10 text-gray-900 ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6 dark:bg-gray-700 dark:text-white dark:ring-gray-600"
              >
                <option value="">All Status</option>
                <option value="connected">Connected</option>
                <option value="disconnected">Disconnected</option>
              </select>
              <ConnectionTags
                v-model="selectedTags"
                :suggestions="databaseStore.getAllTags"
                placeholder="Filter by tag..."
              />
            </div>
          </div>
        </div>

        <!-- Connection List -->
        <ul role="list" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li v-for="connection in filteredConnections" :key="connection.id" class="px-4 py-4 sm:px-6">
            <div class="flex items-center">
              <div class="flex-shrink-0">
                <input
                  type="checkbox"
                  v-model="selectedConnections"
                  :value="connection.id"
                  class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 dark:border-gray-600 dark:bg-gray-700"
                />
              </div>
              <div class="flex min-w-0 flex-1 items-center">
                <div class="flex-shrink-0">
                  <img
                    v-if="connection.type === 'mysql'"
                    class="h-12 w-12"
                    src="@/assets/mysql-logo.png"
                    alt="MySQL"
                  />
                  <img
                    v-else-if="connection.type === 'postgresql'"
                    class="h-12 w-12"
                    src="@/assets/postgresql-logo.png"
                    alt="PostgreSQL"
                  />
                  <img
                    v-else-if="connection.type === 'mongodb'"
                    class="h-12 w-12"
                    src="@/assets/mongodb-logo.png"
                    alt="MongoDB"
                  />
                  <img
                    v-else-if="connection.type === 'redis'"
                    class="h-12 w-12"
                    src="@/assets/redis-logo.png"
                    alt="Redis"
                  />
                  <DatabaseIcon
                    v-else
                    class="h-12 w-12 text-gray-400"
                    aria-hidden="true"
                  />
                </div>
                <div class="min-w-0 flex-1 px-4">
                  <div class="flex items-center justify-between">
                    <p class="truncate text-sm font-medium text-indigo-600 dark:text-indigo-400">
                      {{ connection.name }}
                    </p>
                    <div class="ml-2 flex flex-shrink-0">
                      <span
                        :class="[
                          connection.status === 'connected'
                            ? 'bg-green-50 text-green-700 ring-green-600/20 dark:bg-green-900 dark:text-green-300'
                            : connection.status === 'testing'
                            ? 'bg-yellow-50 text-yellow-700 ring-yellow-600/20 dark:bg-yellow-900 dark:text-yellow-300'
                            : 'bg-red-50 text-red-700 ring-red-600/20 dark:bg-red-900 dark:text-red-300',
                          'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset',
                        ]"
                      >
                        <span v-if="connection.status === 'testing'" class="mr-1">
                          <svg class="animate-spin h-3 w-3" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
                            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                          </svg>
                        </span>
                        {{ connection.status }}
                      </span>
                    </div>
                  </div>
                  <div class="mt-2 flex flex-wrap gap-4">
                    <div class="flex items-center text-sm text-gray-500 dark:text-gray-400">
                      <ServerIcon class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400" aria-hidden="true" />
                      <span>{{ connection.host }}:{{ connection.port }}</span>
                    </div>
                    <div class="flex items-center text-sm text-gray-500 dark:text-gray-400">
                      <DatabaseIcon class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400" aria-hidden="true" />
                      <span>{{ connection.database }}</span>
                    </div>
                    <div class="flex items-center text-sm text-gray-500 dark:text-gray-400">
                      <UserIcon class="mr-1.5 h-5 w-5 flex-shrink-0 text-gray-400" aria-hidden="true" />
                      <span>{{ connection.username }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div class="flex flex-shrink-0 space-x-2">
                <button
                  type="button"
                  @click="editConnection(connection)"
                  class="inline-flex items-center rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 dark:bg-gray-700 dark:text-white dark:ring-gray-600 dark:hover:bg-gray-600"
                >
                  <PencilIcon class="-ml-0.5 mr-1.5 h-4 w-4" aria-hidden="true" />
                  Edit
                </button>
                <button
                  type="button"
                  @click="testConnection(connection)"
                  :disabled="connection.status === 'testing'"
                  class="inline-flex items-center rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed dark:bg-gray-700 dark:text-white dark:ring-gray-600 dark:hover:bg-gray-600"
                >
                  <ArrowPathIcon class="-ml-0.5 mr-1.5 h-4 w-4" aria-hidden="true" />
                  Test
                </button>
                <button
                  type="button"
                  @click="confirmDelete(connection)"
                  class="inline-flex items-center rounded-md bg-red-600 px-2.5 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-red-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600"
                >
                  <TrashIcon class="-ml-0.5 mr-1.5 h-4 w-4" aria-hidden="true" />
                  Delete
                </button>
              </div>
            </div>
          </li>
        </ul>
      </div>
    </div>

    <!-- New/Edit Connection Modal -->
    <ConnectionModal
      v-if="showNewConnectionModal || showEditConnectionModal"
      :connection="editingConnection"
      :show="showNewConnectionModal || showEditConnectionModal"
      @close="closeModal"
      @save="saveConnection"
    />

    <!-- Delete Confirmation Modal -->
    <ConfirmationModal
      v-if="showDeleteModal"
      :show="showDeleteModal"
      title="Delete Connection"
      message="Are you sure you want to delete this connection? This action cannot be undone."
      @confirm="deleteConnection"
      @cancel="showDeleteModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useDatabaseStore } from '@/stores/database'
import {
  PlusIcon,
  DatabaseIcon,
  ServerIcon,
  UserIcon,
  ExclamationCircleIcon,
  PencilIcon,
  ArrowPathIcon,
  TrashIcon
} from '@heroicons/vue/24/outline'
import ConnectionModal from '@/components/ConnectionModal.vue'
import ConfirmationModal from '@/components/ConfirmationModal.vue'
import ConnectionTags from '@/components/ConnectionTags.vue'
import ConnectionTestResult from '@/components/ConnectionTestResult.vue'
import { useToast } from '@/composables/useToast'

const databaseStore = useDatabaseStore()
const { showToast } = useToast()

// State
const showNewConnectionModal = ref(false)
const showEditConnectionModal = ref(false)
const showDeleteModal = ref(false)
const showTestResultModal = ref(false)
const editingConnection = ref(null)
const deletingConnection = ref(null)
const testingConnection = ref(null)
const testResult = ref(null)
const selectedConnections = ref([])
const filterType = ref('')
const filterStatus = ref('')
const batchLoading = ref(false)
const selectedTags = ref([])

// Computed
const connections = computed(() => databaseStore.connections)
const loading = computed(() => databaseStore.loading)
const error = computed(() => databaseStore.error)

const filteredConnections = computed(() => {
  let connections = databaseStore.connections
  if (filterType.value) {
    connections = connections.filter(conn => conn.type === filterType.value)
  }
  if (filterStatus.value) {
    connections = connections.filter(conn => conn.status === filterStatus.value)
  }
  if (selectedTags.value.length > 0) {
    connections = connections.filter(conn => selectedTags.value.every(tag => conn.tags.includes(tag)))
  }
  return connections
})

const isAllSelected = computed(() => {
  return filteredConnections.value.length > 0 && 
    filteredConnections.value.every(conn => selectedConnections.value.includes(conn.id))
})

// Methods
function toggleSelectAll() {
  if (isAllSelected.value) {
    selectedConnections.value = []
  } else {
    selectedConnections.value = filteredConnections.value.map(conn => conn.id)
  }
}

async function batchTest() {
  if (selectedConnections.value.length === 0) return
  
  try {
    batchLoading.value = true
    await databaseStore.batchTestConnections(selectedConnections.value)
    showToast('Connection tests completed', 'success')
  } catch (error) {
    showToast(error.message || 'Failed to test connections', 'error')
  } finally {
    batchLoading.value = false
  }
}

function confirmBatchDelete() {
  if (selectedConnections.value.length === 0) return
  showDeleteModal.value = true
}

async function batchDelete() {
  if (selectedConnections.value.length === 0) return
  
  try {
    await databaseStore.batchDeleteConnections(selectedConnections.value)
    selectedConnections.value = []
    showToast('Connections deleted successfully', 'success')
  } catch (error) {
    showToast(error.message || 'Failed to delete connections', 'error')
  }
  showDeleteModal.value = false
}

function editConnection(connection) {
  editingConnection.value = { ...connection }
  showEditConnectionModal.value = true
}

async function saveConnection(connectionData) {
  try {
    if (connectionData.id) {
      await databaseStore.updateConnection(connectionData.id, connectionData)
      showToast('Connection updated successfully', 'success')
    } else {
      await databaseStore.createConnection(connectionData)
      showToast('Connection created successfully', 'success')
    }
    closeModal()
    await fetchConnections()
  } catch (error) {
    showToast(error, 'error')
  }
}

async function testConnection(connection) {
  try {
    const result = await databaseStore.testConnection(connection)
    showToast(result.message || 'Connection test successful', 'success')
  } catch (error) {
    showToast(error, 'error')
  }
}

function confirmDelete(connection) {
  deletingConnection.value = connection
  showDeleteModal.value = true
}

async function deleteConnection() {
  try {
    await databaseStore.deleteConnection(deletingConnection.value.id)
    showToast('Connection deleted successfully', 'success')
    showDeleteModal.value = false
    deletingConnection.value = null
  } catch (error) {
    showToast(error, 'error')
  }
}

function closeModal() {
  showNewConnectionModal.value = false
  showEditConnectionModal.value = false
  editingConnection.value = null
}

// Lifecycle
onMounted(async () => {
  await fetchConnections()
})

async function fetchConnections() {
  try {
    await databaseStore.fetchConnections()
  } catch (error) {
    showToast('Failed to fetch connections', 'error')
  }
}
</script>
