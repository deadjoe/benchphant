import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { databaseAPI } from '@/services/api'
import { CacheManager } from '@/utils/pagination'

export const useDatabaseStore = defineStore('database', () => {
  // 状态
  const connections = ref([])
  const connectionMetrics = ref({})
  const connectionStats = ref({})
  const loading = ref(false)
  const error = ref(null)
  const pendingRequests = ref(0)
  const monitoring = ref(false)
  const selectedConnections = ref([])
  const connectionHistory = ref([])
  const metricsHistory = ref({})
  const tags = ref([])
  const currentConnection = ref(null)
  const historyLoading = ref(false)
  const historyError = ref(null)
  const hasMoreHistory = ref(false)
  const historyPage = ref(1)
  const historyPageSize = 10

  // 缓存管理
  const connectionCache = new CacheManager({ maxAge: 5 * 60 * 1000 }) // 5分钟缓存
  const metricsCache = new CacheManager({ maxAge: 30 * 1000 }) // 30秒缓存
  const statsCache = new CacheManager({ maxAge: 5 * 60 * 1000 }) // 5分钟缓存

  // 分页
  const pagination = ref({
    currentPage: 1,
    pageSize: 20,
    total: 0
  })

  // 处理API错误
  function handleApiError(err, operation) {
    const errorMessage = err.response?.data?.message || err.message
    error.value = `Failed to ${operation}: ${errorMessage}`
    console.error(`API Error during ${operation}:`, err)
    throw error.value
  }

  // 验证响应格式
  function validateResponse(response, expectedKeys) {
    if (!response || !response.data || typeof response.data !== 'object') {
      throw new Error('Invalid response format: Response data is missing or invalid')
    }
    
    for (const key of expectedKeys) {
      if (!(key in response.data)) {
        throw new Error(`Invalid response format: Missing required field '${key}'`)
      }
    }
    
    return response.data
  }

  // 获取或创建请求Promise
  function getOrCreateRequest(key, requestFn) {
    if (pendingRequests.value.has(key)) {
      return pendingRequests.value.get(key)
    }

    const promise = requestFn().finally(() => {
      pendingRequests.value.delete(key)
    })

    pendingRequests.value.set(key, promise)
    return promise
  }

  // 获取所有连接（支持分页）
  async function fetchConnections(params = {}) {
    const cacheKey = JSON.stringify({ ...pagination.getParams(), ...params })
    const cachedData = connectionCache.get(cacheKey)
    
    if (cachedData && !connectionCache.isExpired(cacheKey)) {
      connections.value = cachedData.connections
      pagination.setTotal(cachedData.total)
      return cachedData
    }

    loading.value = true
    error.value = null

    try {
      const requestKey = `connections_${cacheKey}`
      const response = await getOrCreateRequest(requestKey, async () => {
        const resp = await databaseAPI.getConnections({
          ...pagination.getParams(),
          ...params
        })
        return validateResponse(resp, ['connections', 'total'])
      })

      connections.value = response.connections
      pagination.setTotal(response.total)
      
      connectionCache.set(cacheKey, {
        connections: response.connections,
        total: response.total
      })
      
      return response
    } catch (err) {
      handleApiError(err, 'fetch connections')
    } finally {
      loading.value = false
    }
  }

  // 获取连接指标（带缓存）
  async function fetchConnectionMetrics(connectionId) {
    const cacheKey = `metrics_${connectionId}`
    const cachedData = metricsCache.get(cacheKey)
    
    if (cachedData && !metricsCache.isExpired(cacheKey)) {
      connectionMetrics.value[connectionId] = cachedData
      return cachedData
    }

    try {
      const requestKey = `metrics_${connectionId}`
      const response = await getOrCreateRequest(requestKey, async () => {
        const resp = await databaseAPI.getConnectionMetrics(connectionId)
        return validateResponse(resp, ['activeConnections', 'avgResponseTime'])
      })

      connectionMetrics.value[connectionId] = response
      metricsCache.set(cacheKey, response)
      return response
    } catch (err) {
      handleApiError(err, `fetch metrics for connection ${connectionId}`)
    }
  }

  // 获取连接统计信息（带缓存）
  async function fetchConnectionStats(connectionId, timeRange) {
    const cacheKey = `stats_${connectionId}_${timeRange}`
    const cachedData = statsCache.get(cacheKey)
    
    if (cachedData && !statsCache.isExpired(cacheKey)) {
      connectionStats.value[connectionId] = cachedData
      return cachedData
    }

    try {
      const response = await databaseAPI.getConnectionStats(connectionId, timeRange)
      connectionStats.value[connectionId] = response.data
      statsCache.set(cacheKey, response.data)
      return response.data
    } catch (err) {
      console.error(`Error fetching stats for connection ${connectionId}:`, err)
      throw err
    }
  }

  // 清除缓存
  function clearCache() {
    connectionCache.clear()
    metricsCache.clear()
    statsCache.clear()
  }

  // 清除特定连接的缓存
  function clearConnectionCache(connectionId) {
    metricsCache.delete(`metrics_${connectionId}`)
    statsCache.delete(new RegExp(`stats_${connectionId}_.*`))
  }

  // Getters
  const activeConnections = computed(() => 
    connections.value.filter(conn => conn.status === 'connected')
  )

  const connectionsByType = computed(() => {
    const grouped = {}
    connections.value.forEach(conn => {
      if (!grouped[conn.type]) {
        grouped[conn.type] = []
      }
      grouped[conn.type].push(conn)
    })
    return grouped
  })

  const getConnectionsByTags = computed(() => (tags) => {
    if (!tags || tags.length === 0) return connections.value
    return connections.value.filter(conn => 
      tags.every(tag => conn.tags && conn.tags.includes(tag))
    )
  })

  const tagList = computed(() => {
    const tags = new Set()
    connections.value.forEach(conn => {
      if (conn.tags) {
        conn.tags.forEach(tag => tags.add(tag))
      }
    })
    return Array.from(tags)
  })

  // Actions
  async function createConnection(connectionData) {
    try {
      loading.value = true
      error.value = null
      const response = await databaseAPI.createConnection(connectionData)
      connections.value.push(response.data)
      return response.data
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to create connection'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  async function updateConnection(id, connectionData) {
    try {
      loading.value = true
      error.value = null
      const response = await databaseAPI.updateConnection(id, connectionData)
      const index = connections.value.findIndex(conn => conn.id === id)
      if (index !== -1) {
        connections.value[index] = response.data
      }
      return response.data
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to update connection'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  async function deleteConnection(id) {
    try {
      loading.value = true
      error.value = null
      await databaseAPI.deleteConnection(id)
      connections.value = connections.value.filter(conn => conn.id !== id)
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to delete connection'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  async function testConnection(connectionData) {
    try {
      // 设置连接状态为testing
      if (connectionData.id) {
        const index = connections.value.findIndex(conn => conn.id === connectionData.id)
        if (index !== -1) {
          connections.value[index] = { ...connections.value[index], status: 'testing' }
        }
      }

      error.value = null
      const response = await databaseAPI.testConnection(connectionData)

      // 更新连接状态
      if (connectionData.id) {
        const index = connections.value.findIndex(conn => conn.id === connectionData.id)
        if (index !== -1) {
          connections.value[index] = { 
            ...connections.value[index], 
            status: response.data.success ? 'connected' : 'disconnected'
          }
        }
      }

      return response.data
    } catch (err) {
      // 如果测试失败，更新状态为disconnected
      if (connectionData.id) {
        const index = connections.value.findIndex(conn => conn.id === connectionData.id)
        if (index !== -1) {
          connections.value[index] = { ...connections.value[index], status: 'disconnected' }
        }
      }

      error.value = err.response?.data?.message || 'Connection test failed'
      throw error.value
    }
  }

  async function fetchConnection(id) {
    try {
      loading.value = true
      error.value = null
      const response = await databaseAPI.getConnection(id)
      currentConnection.value = response.data
      return response.data
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to fetch connection'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  async function fetchConnectionHistory(id, page = 1) {
    try {
      historyLoading.value = true
      historyError.value = null
      const response = await databaseAPI.getConnectionHistory(id)
      
      if (page === 1) {
        connectionHistory.value = response.data.items
      } else {
        connectionHistory.value = [...connectionHistory.value, ...response.data.items]
      }
      
      hasMoreHistory.value = response.data.hasMore
      historyPage.value = page
      return response.data
    } catch (err) {
      historyError.value = err.response?.data?.message || 'Failed to fetch connection history'
      throw historyError.value
    } finally {
      historyLoading.value = false
    }
  }

  async function loadMoreHistory(id) {
    if (!hasMoreHistory.value || historyLoading.value) return
    return fetchConnectionHistory(id, historyPage.value + 1)
  }

  async function batchTestConnections(ids) {
    try {
      loading.value = true
      error.value = null
      
      // 更新选中连接的状态为testing
      ids.forEach(id => {
        const index = connections.value.findIndex(conn => conn.id === id)
        if (index !== -1) {
          connections.value[index] = { ...connections.value[index], status: 'testing' }
        }
      })

      const response = await databaseAPI.batchTestConnections(ids)
      
      // 更新测试结果
      response.data.forEach(result => {
        const index = connections.value.findIndex(conn => conn.id === result.id)
        if (index !== -1) {
          connections.value[index] = { 
            ...connections.value[index], 
            status: result.success ? 'connected' : 'disconnected'
          }
        }
      })

      return response.data
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to test connections'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  async function batchDeleteConnections(ids) {
    try {
      loading.value = true
      error.value = null
      await databaseAPI.batchDeleteConnections(ids)
      connections.value = connections.value.filter(conn => !ids.includes(conn.id))
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to delete connections'
      throw error.value
    } finally {
      loading.value = false
    }
  }

  async function updateConnectionTags(connectionId, tags) {
    try {
      loading.value = true
      error.value = null
      const response = await databaseAPI.put(`/connections/${connectionId}/tags`, { tags })
      
      // 更新本地状态
      const index = connections.value.findIndex(c => c.id === connectionId)
      if (index !== -1) {
        connections.value[index] = {
          ...connections.value[index],
          tags: response.data.tags
        }
      }
      
      return response.data
    } catch (error) {
      error.value = error.message
      throw error
    } finally {
      loading.value = false
    }
  }

  async function updateBatchConnectionTags(connectionIds, tags) {
    try {
      loading.value = true
      error.value = null
      const response = await databaseAPI.put('/connections/batch/tags', { connectionIds, tags })
      
      // 更新本地状态
      connectionIds.forEach(id => {
        const index = connections.value.findIndex(c => c.id === id)
        if (index !== -1) {
          connections.value[index] = {
            ...connections.value[index],
            tags: tags
          }
        }
      })
      
      return response.data
    } catch (error) {
      error.value = error.message
      throw error
    } finally {
      loading.value = false
    }
  }

  // 更新连接指标
  async function updateConnectionMetrics() {
    try {
      const response = await databaseAPI.getConnectionMetrics()
      const metrics = response.data

      // 更新当前指标
      connectionMetrics.value = metrics.current

      // 更新历史数据
      Object.keys(metrics.history || {}).forEach(connId => {
        if (!metricsHistory.value[connId]) {
          metricsHistory.value[connId] = []
        }
        metricsHistory.value[connId].push({
          timestamp: new Date(),
          ...metrics.history[connId]
        })
        // 保留最近100个数据点
        if (metricsHistory.value[connId].length > 100) {
          metricsHistory.value[connId].shift()
        }
      })
    } catch (error) {
      console.error('Failed to update metrics:', error)
      throw error
    }
  }

  // 开始定期更新指标
  function startMetricsUpdate(interval = 5000) {
    if (monitoring.value) return
    
    monitoring.value = true
    setInterval(async () => {
      try {
        await updateConnectionMetrics()
      } catch (error) {
        console.error('Failed to update metrics:', error)
      }
    }, interval)
  }

  // 停止更新指标
  function stopMetricsUpdate() {
    monitoring.value = false
  }

  // 获取连接的指标历史
  function getConnectionMetricsHistory(connectionId) {
    return metricsHistory.value[connectionId] || []
  }

  // 获取连接的当前指标
  function getConnectionMetrics(connectionId) {
    return connectionMetrics.value[connectionId] || null
  }

  // 清除指标数据
  function clearMetrics() {
    connectionMetrics.value = {}
    metricsHistory.value = {}
  }

  // 克隆连接
  async function cloneConnection(connectionId, options = {}) {
    try {
      const response = await databaseAPI.post(`/connections/${connectionId}/clone`, options)
      connections.value.push(response.data)
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  // 导入连接
  async function importConnections(connectionsData) {
    try {
      const response = await databaseAPI.post('/connections/import', connectionsData)
      await fetchConnections() // 刷新连接列表
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  // 导出连接
  async function exportConnections(connectionIds) {
    try {
      const response = await databaseAPI.post('/connections/export', { connectionIds })
      return response.data
    } catch (err) {
      error.value = err.message
      throw err
    }
  }

  // 按标签过滤连接
  const filterConnectionsByTags = (selectedTags) => {
    if (!selectedTags || selectedTags.length === 0) return connections.value
    return connections.value.filter(conn =>
      selectedTags.every(tag => conn.tags && conn.tags.includes(tag))
    )
  }

  return {
    // State
    connections,
    connectionMetrics,
    connectionStats,
    loading,
    error,
    pendingRequests,
    monitoring,
    selectedConnections,
    connectionHistory,
    metricsHistory,
    tags,
    currentConnection,
    historyLoading,
    historyError,
    hasMoreHistory,
    historyPage,
    pagination,

    // Cache managers
    connectionCache,
    metricsCache,
    statsCache,

    // Computed
    activeConnections,
    connectionsByType,
    getConnectionsByTags,
    tagList,

    // Methods
    setLoading: (value) => {
      loading.value = value
    },
    incrementPendingRequests: () => {
      pendingRequests.value++
      loading.value = true
    },
    decrementPendingRequests: () => {
      pendingRequests.value--
      if (pendingRequests.value === 0) {
        loading.value = false
      }
    },
    handleError: (err) => {
      error.value = err.message
      loading.value = false
    },
    fetchConnections,
    fetchConnectionMetrics,
    getConnectionsByTags,
    clearCache,
    clearConnectionCache,
    fetchConnectionStats,
    createConnection,
    updateConnection,
    deleteConnection,
    testConnection,
    fetchConnection,
    fetchConnectionHistory,
    loadMoreHistory,
    batchTestConnections,
    batchDeleteConnections,
    updateConnectionTags,
    updateBatchConnectionTags,
    updateConnectionMetrics,
    startMetricsUpdate,
    stopMetricsUpdate,
    getConnectionMetricsHistory,
    getConnectionMetrics,
    clearMetrics,
    cloneConnection,
    importConnections,
    exportConnections,
    filterConnectionsByTags
  }
})
