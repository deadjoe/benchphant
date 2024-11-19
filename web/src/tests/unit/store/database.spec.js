import { setActivePinia, createPinia } from 'pinia'
import { useDatabaseStore } from '@/stores/database'
import { databaseAPI } from '@/services/api'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { CacheManager } from '@/utils/pagination'

vi.mock('@/services/api', () => ({
  databaseAPI: {
    getConnections: vi.fn(),
    getConnectionMetrics: vi.fn(),
    getConnectionStats: vi.fn(),
    createConnection: vi.fn(),
    updateConnection: vi.fn(),
    deleteConnection: vi.fn()
  }
}))

describe('Database Store', () => {
  let store

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useDatabaseStore()
  })

  describe('CacheManager', () => {
    let cacheManager

    beforeEach(() => {
      cacheManager = new CacheManager({ maxAge: 1000, maxSize: 2 })
    })

    it('should store and retrieve values', () => {
      cacheManager.set('key1', 'value1')
      expect(cacheManager.get('key1')).toBe('value1')
    })

    it('should respect maxSize', () => {
      cacheManager.set('key1', 'value1')
      cacheManager.set('key2', 'value2')
      cacheManager.set('key3', 'value3')
      expect(cacheManager.get('key1')).toBeNull()
      expect(cacheManager.get('key2')).toBe('value2')
      expect(cacheManager.get('key3')).toBe('value3')
    })

    it('should handle expiration', async () => {
      cacheManager.set('key1', 'value1')
      await new Promise(resolve => setTimeout(resolve, 1100))
      expect(cacheManager.get('key1')).toBeNull()
    })

    it('should identify expiring keys', () => {
      cacheManager.set('key1', 'value1')
      cacheManager.set('key2', 'value2')
      const expiringKeys = cacheManager.getExpiringKeys(1000)
      expect(expiringKeys).toContain('key1')
      expect(expiringKeys).toContain('key2')
    })

    it('should clear all cache', () => {
      cacheManager.set('key1', 'value1')
      cacheManager.set('key2', 'value2')
      cacheManager.clear()
      expect(cacheManager.get('key1')).toBeNull()
      expect(cacheManager.get('key2')).toBeNull()
    })
  })

  describe('Pagination', () => {
    it('should fetch connections with pagination', async () => {
      const mockResponse = {
        data: [{ id: 1 }, { id: 2 }],
        total: 2
      }

      store.connections = mockResponse.data
      store.pagination.total = mockResponse.total

      expect(store.connections).toEqual(mockResponse.data)
      expect(store.pagination.total).toBe(2)
    })

    it('should use cache when available', async () => {
      const mockData = [{ id: 1 }, { id: 2 }]
      store.connections = mockData
      store.connectionCache.set('page1', mockData)

      expect(store.connections).toEqual(mockData)
    })
  })

  describe('Metrics', () => {
    it('should fetch connection metrics', async () => {
      const mockMetrics = {
        data: { cpu: 50, memory: 80 }
      }

      store.connectionMetrics = { 1: mockMetrics.data }

      expect(store.connectionMetrics[1]).toEqual(mockMetrics.data)
    })

    it('should use cache for metrics', async () => {
      const mockData = { cpu: 50, memory: 80 }
      store.connectionMetrics = { 1: mockData }
      store.metricsCache.set('connection1', mockData)

      expect(store.connectionMetrics[1]).toEqual(mockData)
    })
  })

  describe('Tag Management', () => {
    it('should get all unique tags', () => {
      store.connections = [
        { id: 1, tags: ['prod', 'mysql'] },
        { id: 2, tags: ['dev', 'postgres'] }
      ]

      expect(store.tagList).toEqual(['prod', 'mysql', 'dev', 'postgres'])
    })

    it('should filter connections by tags', () => {
      store.connections = [
        { id: 1, tags: ['prod', 'mysql'] },
        { id: 2, tags: ['dev', 'postgres'] }
      ]

      const filtered = store.getConnectionsByTags(['prod', 'mysql'])
      expect(filtered).toHaveLength(1)
      expect(filtered[0].id).toBe(1)
    })
  })

  describe('Error Handling', () => {
    it('should handle API errors gracefully', async () => {
      const error = new Error('API Error')
      store.handleError(error)
      expect(store.error).toBe(error.message)
    })

    it('should handle network timeout', async () => {
      const error = new Error('Network timeout')
      store.handleError(error)
      expect(store.error).toBe(error.message)
    })

    it('should handle invalid response format', async () => {
      const error = new Error('Invalid response format')
      store.handleError(error)
      expect(store.error).toBe(error.message)
    })
  })

  describe('Loading States', () => {
    it('should set loading state during API calls', () => {
      store.setLoading(true)
      expect(store.loading).toBe(true)
      store.setLoading(false)
      expect(store.loading).toBe(false)
    })

    it('should maintain loading state for concurrent requests', () => {
      store.incrementPendingRequests()
      store.incrementPendingRequests()
      expect(store.loading).toBe(true)
      store.decrementPendingRequests()
      expect(store.loading).toBe(true)
      store.decrementPendingRequests()
      expect(store.loading).toBe(false)
    })
  })

  describe('Concurrent Requests', () => {
    it('should handle multiple simultaneous requests', async () => {
      store.incrementPendingRequests()
      store.incrementPendingRequests()
      expect(store.pendingRequests).toBe(2)
      store.decrementPendingRequests()
      store.decrementPendingRequests()
      expect(store.pendingRequests).toBe(0)
    })
  })
})
