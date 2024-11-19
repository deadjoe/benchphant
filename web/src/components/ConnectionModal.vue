<template>
  <TransitionRoot as="template" :show="show">
    <Dialog as="div" class="relative z-10" @close="closeModal">
      <TransitionChild
        as="template"
        enter="ease-out duration-300"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="ease-in duration-200"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
      </TransitionChild>

      <div class="fixed inset-0 z-10 overflow-y-auto">
        <div class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0">
          <TransitionChild
            as="template"
            enter="ease-out duration-300"
            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            enter-to="opacity-100 translate-y-0 sm:scale-100"
            leave="ease-in duration-200"
            leave-from="opacity-100 translate-y-0 sm:scale-100"
            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
          >
            <DialogPanel
              class="relative transform overflow-hidden rounded-lg bg-white dark:bg-gray-800 px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
            >
              <div>
                <div class="mt-3 sm:mt-5">
                  <DialogTitle as="h3" class="text-lg font-medium leading-6 text-gray-900 dark:text-white">
                    {{ connection ? 'Edit Connection' : 'New Connection' }}
                  </DialogTitle>
                  <div class="mt-6">
                    <form @submit.prevent="handleSubmit" class="space-y-6">
                      <!-- Connection Name -->
                      <div>
                        <label for="name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                          Connection Name
                        </label>
                        <div class="mt-1">
                          <input
                            type="text"
                            name="name"
                            id="name"
                            v-model="form.name"
                            class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white sm:text-sm"
                            required
                          />
                        </div>
                      </div>

                      <!-- Database Type -->
                      <div>
                        <label for="type" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                          Database Type
                        </label>
                        <div class="mt-1">
                          <select
                            id="type"
                            name="type"
                            v-model="form.type"
                            class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white sm:text-sm"
                            required
                          >
                            <option value="mysql">MySQL</option>
                            <option value="postgresql">PostgreSQL</option>
                          </select>
                        </div>
                      </div>

                      <!-- Host -->
                      <div>
                        <label for="host" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                          Host
                        </label>
                        <div class="mt-1">
                          <input
                            type="text"
                            name="host"
                            id="host"
                            v-model="form.host"
                            class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white sm:text-sm"
                            required
                          />
                        </div>
                      </div>

                      <!-- Port -->
                      <div>
                        <label for="port" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                          Port
                        </label>
                        <div class="mt-1">
                          <input
                            type="number"
                            name="port"
                            id="port"
                            v-model="form.port"
                            class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white sm:text-sm"
                            required
                          />
                        </div>
                      </div>

                      <!-- Database Name -->
                      <div>
                        <label for="database" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                          Database Name
                        </label>
                        <div class="mt-1">
                          <input
                            type="text"
                            name="database"
                            id="database"
                            v-model="form.database"
                            class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white sm:text-sm"
                            required
                          />
                        </div>
                      </div>

                      <!-- Username -->
                      <div>
                        <label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                          Username
                        </label>
                        <div class="mt-1">
                          <input
                            type="text"
                            name="username"
                            id="username"
                            v-model="form.username"
                            class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white sm:text-sm"
                            required
                          />
                        </div>
                      </div>

                      <!-- Password -->
                      <div>
                        <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                          Password
                        </label>
                        <div class="mt-1">
                          <input
                            type="password"
                            name="password"
                            id="password"
                            v-model="form.password"
                            class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white sm:text-sm"
                            :required="!connection"
                          />
                        </div>
                        <p v-if="connection" class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                          Leave blank to keep the current password
                        </p>
                      </div>

                      <!-- Connection Limits -->
                      <div class="space-y-4">
                        <div>
                          <label for="maxIdleConn" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                            Max Idle Connections
                          </label>
                          <div class="mt-1">
                            <input
                              type="number"
                              name="maxIdleConn"
                              id="maxIdleConn"
                              v-model="form.maxIdleConn"
                              class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white sm:text-sm"
                              required
                            />
                          </div>
                        </div>

                        <div>
                          <label for="maxOpenConn" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                            Max Open Connections
                          </label>
                          <div class="mt-1">
                            <input
                              type="number"
                              name="maxOpenConn"
                              id="maxOpenConn"
                              v-model="form.maxOpenConn"
                              class="block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:border-gray-600 dark:bg-gray-700 dark:text-white sm:text-sm"
                              required
                            />
                          </div>
                        </div>
                      </div>
                    </form>
                  </div>
                </div>
              </div>

              <div class="mt-6 sm:mt-6 sm:grid sm:grid-flow-row-dense sm:grid-cols-2 sm:gap-3">
                <button
                  type="submit"
                  @click="handleSubmit"
                  class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:col-start-2"
                >
                  {{ connection ? 'Save Changes' : 'Create Connection' }}
                </button>
                <button
                  type="button"
                  @click="closeModal"
                  class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:col-start-1 sm:mt-0"
                >
                  Cancel
                </button>
              </div>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup>
import { ref, watchEffect } from 'vue'
import {
  Dialog,
  DialogPanel,
  DialogTitle,
  TransitionChild,
  TransitionRoot,
} from '@headlessui/vue'

const props = defineProps({
  show: {
    type: Boolean,
    required: true
  },
  connection: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close', 'save'])

const form = ref({
  name: '',
  type: 'mysql',
  host: 'localhost',
  port: 3306,
  database: '',
  username: '',
  password: '',
  maxIdleConn: 10,
  maxOpenConn: 100
})

// Watch for connection changes and update form
watchEffect(() => {
  if (props.connection) {
    form.value = {
      ...props.connection,
      password: '' // Don't show the existing password
    }
  } else {
    // Reset form to defaults
    form.value = {
      name: '',
      type: 'mysql',
      host: 'localhost',
      port: 3306,
      database: '',
      username: '',
      password: '',
      maxIdleConn: 10,
      maxOpenConn: 100
    }
  }
})

function closeModal() {
  emit('close')
}

function handleSubmit() {
  // If editing and password is empty, don't include it in the update
  const data = { ...form.value }
  if (props.connection && !data.password) {
    delete data.password
  }
  emit('save', data)
}
</script>
