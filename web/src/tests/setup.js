import { vi } from 'vitest'
import { config } from '@vue/test-utils'

// 设置测试环境
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    go: vi.fn(),
  }),
  useRoute: () => ({
    params: {},
    query: {},
    path: '/',
  }),
}))

// Mock Chart.js
vi.mock('chart.js', () => ({
  Chart: {
    register: vi.fn()
  },
  CategoryScale: {},
  LinearScale: {},
  PointElement: {},
  LineElement: {},
  Title: {},
  Tooltip: {},
  Legend: {},
  Filler: {}
}))

// Mock vue-chartjs
vi.mock('vue-chartjs', () => ({
  Line: vi.fn()
}))

// 全局 mocks
config.global.mocks = {
  $t: (key) => key,
  $router: {
    push: vi.fn(),
    replace: vi.fn(),
  },
  $route: {
    params: {},
    query: {},
  },
}

// Store original console methods
const originalConsoleError = console.error
const originalConsoleWarn = console.warn

// Override console.error to throw on errors during tests
console.error = (...args) => {
  originalConsoleError(...args)
  throw new Error(args.join(' '))
}

// Silence Vue warnings
console.warn = (...args) => {
  if (args[0]?.includes?.('[Vue warn]')) return
  originalConsoleWarn(...args)
}

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Mock IntersectionObserver
const mockIntersectionObserver = vi.fn()
mockIntersectionObserver.mockReturnValue({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
})
window.IntersectionObserver = mockIntersectionObserver

// Mock ResizeObserver
const mockResizeObserver = vi.fn()
mockResizeObserver.mockReturnValue({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
})
window.ResizeObserver = mockResizeObserver

// Mock requestAnimationFrame
window.requestAnimationFrame = vi.fn(cb => setTimeout(cb, 0))
window.cancelAnimationFrame = vi.fn(id => clearTimeout(id))

// 清理函数
afterEach(() => {
  vi.clearAllMocks()
  vi.clearAllTimers()
  document.body.innerHTML = ''
})

// 设置测试环境变量
process.env.VITE_API_BASE_URL = 'http://localhost:3000'
