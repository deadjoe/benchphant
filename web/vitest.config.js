import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/tests/setup.js'],
    include: ['./src/tests/unit/**/*.spec.js'],
    exclude: ['**/node_modules/**'],
    coverage: {
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/tests/setup.js',
      ],
    },
    testTimeout: 5000,
    hookTimeout: 5000,
    pool: 'threads',
    maxThreads: 1,
    minThreads: 1,
    isolate: true,
    watch: false,
    reporters: ['verbose'],
    onConsoleLog: (log, type) => {
      console.log(`[${type}] ${log}`)
      return false
    }
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  }
})
