<template>
  <TransitionRoot as="template" :show="show">
    <Dialog as="div" class="relative z-10" @close="$emit('close')">
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
            <DialogPanel class="relative transform overflow-hidden rounded-lg bg-white dark:bg-gray-800 px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6">
              <div>
                <div :class="[
                  result.success ? 'bg-green-100 dark:bg-green-900' : 'bg-red-100 dark:bg-red-900',
                  'mx-auto flex h-12 w-12 items-center justify-center rounded-full'
                ]">
                  <CheckCircleIcon
                    v-if="result.success"
                    class="h-6 w-6 text-green-600 dark:text-green-400"
                    aria-hidden="true"
                  />
                  <XCircleIcon
                    v-else
                    class="h-6 w-6 text-red-600 dark:text-red-400"
                    aria-hidden="true"
                  />
                </div>
                <div class="mt-3 text-center sm:mt-5">
                  <DialogTitle
                    as="h3"
                    class="text-lg font-medium leading-6 text-gray-900 dark:text-white"
                  >
                    Connection Test Result
                  </DialogTitle>
                  <div class="mt-2">
                    <p :class="[
                      result.success ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400',
                      'text-sm'
                    ]">
                      {{ result.success ? 'Connection successful!' : 'Connection failed' }}
                    </p>
                    
                    <!-- Connection Details -->
                    <div class="mt-4 text-left">
                      <div class="space-y-4">
                        <!-- Basic Info -->
                        <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                          <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-2">Connection Info</h4>
                          <dl class="grid grid-cols-1 gap-x-4 gap-y-4 sm:grid-cols-2">
                            <div class="sm:col-span-1">
                              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Host</dt>
                              <dd class="mt-1 text-sm text-gray-900 dark:text-white">{{ connection.host }}:{{ connection.port }}</dd>
                            </div>
                            <div class="sm:col-span-1">
                              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Database</dt>
                              <dd class="mt-1 text-sm text-gray-900 dark:text-white">{{ connection.database }}</dd>
                            </div>
                            <div class="sm:col-span-1">
                              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Type</dt>
                              <dd class="mt-1 text-sm text-gray-900 dark:text-white capitalize">{{ connection.type }}</dd>
                            </div>
                            <div class="sm:col-span-1">
                              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Username</dt>
                              <dd class="mt-1 text-sm text-gray-900 dark:text-white">{{ connection.username }}</dd>
                            </div>
                          </dl>
                        </div>

                        <!-- Test Details -->
                        <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
                          <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-2">Test Details</h4>
                          <dl class="space-y-3">
                            <div>
                              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Response Time</dt>
                              <dd class="mt-1 text-sm text-gray-900 dark:text-white">{{ result.responseTime }}ms</dd>
                            </div>
                            <div>
                              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Server Version</dt>
                              <dd class="mt-1 text-sm text-gray-900 dark:text-white">{{ result.serverVersion || 'N/A' }}</dd>
                            </div>
                            <div v-if="!result.success">
                              <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Error Message</dt>
                              <dd class="mt-1 text-sm text-red-600 dark:text-red-400">{{ result.error }}</dd>
                            </div>
                          </dl>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
              <div class="mt-5 sm:mt-6">
                <button
                  type="button"
                  class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                  @click="$emit('close')"
                >
                  Close
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
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { CheckCircleIcon, XCircleIcon } from '@heroicons/vue/24/outline'

defineProps({
  show: {
    type: Boolean,
    required: true
  },
  connection: {
    type: Object,
    required: true
  },
  result: {
    type: Object,
    required: true,
    validator: (value) => {
      return typeof value.success === 'boolean' &&
        (value.success ? true : typeof value.error === 'string')
    }
  }
})

defineEmits(['close'])
</script>
