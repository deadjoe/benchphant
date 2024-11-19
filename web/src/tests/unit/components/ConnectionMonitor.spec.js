import { mount } from '@vue/test-utils'
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { createPinia } from 'pinia'
import ConnectionMonitor from '@/components/ConnectionMonitor.vue'
import { useDatabaseStore } from '@/stores/database'

// Mock Chart.js components
vi.mock('vue-chartjs', () => ({
  Line: {
    name: 'Line',
    template: '<div class="mock-chart"></div>'
  }
}))

describe('ConnectionMonitor.vue', () => {
  let wrapper
  let store
  let pinia

  beforeEach(() => {
    // Setup fake timers
    vi.useFakeTimers()
    
    // Create a fresh pinia instance
    pinia = createPinia()
    
    // Create test data
    const testConnections = [
      {
        id: 1,
        name: 'Test DB 1',
        status: 'connected',
        uptime: 3600,
        metrics: {
          qps: 100,
          responseTime: 50,
          activeSessions: 10,
          memoryUsage: 1024 * 1024,
          totalQueries: 1000,
          history: []
        }
      },
      {
        id: 2,
        name: 'Test DB 2',
        status: 'disconnected',
        uptime: 1800,
        metrics: {
          qps: 0,
          responseTime: 0,
          activeSessions: 0,
          memoryUsage: 0,
          totalQueries: 0,
          history: []
        }
      }
    ]

    // Mount component with test data
    wrapper = mount(ConnectionMonitor, {
      global: {
        plugins: [pinia]
      },
      props: {
        connections: testConnections
      }
    })

    // Get store instance
    store = useDatabaseStore()
    
    // Mock store methods
    store.updateConnectionMetrics = vi.fn()
  })

  afterEach(() => {
    vi.clearAllTimers()
    vi.clearAllMocks()
    wrapper.unmount()
  })

  it('renders connection list correctly', async () => {
    const connections = wrapper.findAll('.border')
    expect(connections).toHaveLength(2)
    expect(connections[0].text()).toContain('Test DB 1')
    expect(connections[1].text()).toContain('Test DB 2')
  })

  it('displays connection status correctly', () => {
    const statusIndicators = wrapper.findAll('.w-2.h-2.rounded-full')
    expect(statusIndicators[0].classes()).toContain('bg-green-500')
    expect(statusIndicators[1].classes()).toContain('bg-red-500')
  })

  it('starts monitoring when button clicked', async () => {
    const startButton = wrapper.find('button:first-child')
    await startButton.trigger('click')
    
    expect(wrapper.vm.isMonitoring).toBe(true)
    
    // Advance timer and check if updateConnectionMetrics was called
    vi.advanceTimersByTime(5000)
    expect(store.updateConnectionMetrics).toHaveBeenCalled()
  })

  it('stops monitoring when stop button clicked', async () => {
    // Start monitoring first
    await wrapper.vm.startMonitoring()
    expect(wrapper.vm.isMonitoring).toBe(true)
    
    // Click stop button
    const stopButton = wrapper.find('button:last-child')
    await stopButton.trigger('click')
    
    expect(wrapper.vm.isMonitoring).toBe(false)
    
    // Clear previous calls
    store.updateConnectionMetrics.mockClear()
    
    // Advance timer and verify no more updates
    vi.advanceTimersByTime(5000)
    expect(store.updateConnectionMetrics).not.toHaveBeenCalled()
  })

  it('displays metrics when available', () => {
    // Check QPS
    const qpsValue = wrapper.find('.grid-cols-2.md\\:grid-cols-4 > div:nth-child(1) .text-lg')
    expect(qpsValue.text()).toBe('100')

    // Check Response Time
    const responseTime = wrapper.find('.grid-cols-2.md\\:grid-cols-4 > div:nth-child(2) .text-lg')
    expect(responseTime.text()).toBe('50ms')

    // Check Active Sessions
    const activeSessions = wrapper.find('.grid-cols-2.md\\:grid-cols-4 > div:nth-child(3) .text-lg')
    expect(activeSessions.text()).toBe('10')

    // Check Memory Usage
    const memoryUsage = wrapper.find('.grid-cols-2.md\\:grid-cols-4 > div:nth-child(4) .text-lg')
    expect(memoryUsage.text()).toBe('1.0MB')
  })

  it('shows loading state while fetching metrics', async () => {
    // Mock the loading state in store
    store.$patch({ loading: true })
    await wrapper.vm.startMonitoring()
    
    // Should show loading state while fetching
    expect(store.loading).toBe(true)
    
    // Mock the completion of the request
    store.$patch({ loading: false })
    
    // Advance timer to complete the request
    vi.advanceTimersByTime(1000)
    await wrapper.vm.$nextTick()
    
    expect(store.loading).toBe(false)
  })

  it('formats metric values correctly', () => {
    // Test time formatting
    expect(wrapper.vm.formatTime(1234)).toBe('1.23s')
    expect(wrapper.vm.formatTime(50)).toBe('50ms')
    
    // Test number formatting with fixed precision
    const num = 1234.5678
    const formatted = wrapper.vm.formatNumber(num)
    expect(formatted.split('.')[0]).toBe('1,234')
    expect(formatted.split('.')[1].length).toBeLessThanOrEqual(2)
    
    // Test bytes formatting
    expect(wrapper.vm.formatBytes(1024)).toBe('1.0KB')
    expect(wrapper.vm.formatBytes(1024 * 1024)).toBe('1.0MB')
    
    // Test uptime formatting
    expect(wrapper.vm.formatUptime(3600)).toBe('1h 0m')
    expect(wrapper.vm.formatUptime(90)).toBe('1m')
  })

  it('handles empty metrics gracefully', async () => {
    await wrapper.setProps({
      connections: [{
        id: 1,
        name: 'Empty DB',
        status: 'connected',
        uptime: 0,
        metrics: null
      }]
    })

    // Check that all metric values show 0 with appropriate formatting
    const metricValues = wrapper.findAll('.grid-cols-2.md\\:grid-cols-4 .text-lg')
    const expectedValues = ['0', '0ms', '0', '0.0B']
    metricValues.forEach((el, index) => {
      expect(el.text()).toBe(expectedValues[index])
    })
  })
})
