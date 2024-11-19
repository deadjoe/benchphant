<template>
  <div class="min-h-screen bg-gray-100 dark:bg-gray-900">
    <nav v-if="isAuthenticated" class="bg-white dark:bg-gray-800 shadow">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <div class="flex-shrink-0 flex items-center">
              <img class="h-8 w-auto" src="@/assets/logo.svg" alt="Benchphant" />
            </div>
            <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
              <router-link
                v-for="item in navigation"
                :key="item.name"
                :to="item.href"
                :class="[
                  item.current
                    ? 'border-indigo-500 text-gray-900 dark:text-white'
                    : 'border-transparent text-gray-500 dark:text-gray-400 hover:border-gray-300 hover:text-gray-700 dark:hover:text-gray-300',
                  'inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium'
                ]"
              >
                {{ item.name }}
              </router-link>
            </div>
          </div>
          <div class="flex items-center">
            <button
              @click="toggleTheme"
              class="p-2 rounded-md text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
            >
              <sun-icon v-if="isDark" class="h-6 w-6" />
              <moon-icon v-else class="h-6 w-6" />
            </button>
            <div class="ml-3 relative">
              <button
                @click="logout"
                class="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-md text-sm font-medium"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </div>
    </nav>

    <main>
      <div class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <router-view></router-view>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { SunIcon, MoonIcon } from '@heroicons/vue/24/outline'

const router = useRouter()
const isAuthenticated = ref(false)
const isDark = ref(false)

const navigation = [
  { name: 'Dashboard', href: '/', current: true },
  { name: 'Databases', href: '/databases', current: false },
  { name: 'Benchmarks', href: '/benchmarks', current: false },
  { name: 'Reports', href: '/reports', current: false },
]

onMounted(() => {
  // Check authentication status
  const token = localStorage.getItem('token')
  isAuthenticated.value = !!token

  // Check theme preference
  isDark.value = localStorage.getItem('theme') === 'dark'
  applyTheme()
})

function toggleTheme() {
  isDark.value = !isDark.value
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  applyTheme()
}

function applyTheme() {
  document.documentElement.classList.toggle('dark', isDark.value)
}

async function logout() {
  try {
    // Call logout API
    await fetch('/api/logout', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      }
    })
    
    // Clear local storage and redirect to login
    localStorage.removeItem('token')
    router.push('/login')
  } catch (error) {
    console.error('Logout failed:', error)
  }
}
</script>

<style>
@import 'tailwindcss/base';
@import 'tailwindcss/components';
@import 'tailwindcss/utilities';
</style>
