import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const databaseAPI = {
  // Get all database connections
  getConnections: () => api.get('/connections'),
  
  // Get a single database connection
  getConnection: (id) => api.get(`/connections/${id}`),
  
  // Create a new database connection
  createConnection: (data) => api.post('/connections', data),
  
  // Update an existing database connection
  updateConnection: (id, data) => api.put(`/connections/${id}`, data),
  
  // Delete a database connection
  deleteConnection: (id) => api.delete(`/connections/${id}`),
  
  // Get connection metrics
  getConnectionMetrics: (id) => api.get(`/connections/${id}/metrics`),
  
  // Get metrics history
  getMetricsHistory: (id, params) => api.get(`/connections/${id}/metrics/history`, { params }),
  
  // Start monitoring
  startMonitoring: (id) => api.post(`/connections/${id}/monitor/start`),
  
  // Stop monitoring
  stopMonitoring: (id) => api.post(`/connections/${id}/monitor/stop`),
  
  // Get connection stats
  getConnectionStats: (id, timeRange) => api.get(`/connections/${id}/stats`, { params: { timeRange } }),
  
  // Import connections
  importConnections: (data) => api.post('/connections/import', data),
  
  // Export connections
  exportConnections: (connectionIds) => api.post('/connections/export', { connectionIds }),
  
  // Clone connection
  cloneConnection: (id, options) => api.post(`/connections/${id}/clone`, options),
  
  // Update connection tags
  updateConnectionTags: (id, tags) => api.put(`/connections/${id}/tags`, { tags }),
  
  // Batch update tags
  batchUpdateTags: (connectionIds, tags, operation) => 
    api.post('/connections/batch/tags', { connectionIds, tags, operation }),
  
  // Batch test connections
  batchTest: (connectionIds) => api.post('/connections/batch/test', { connectionIds }),
  
  // Batch delete connections
  batchDelete: (connectionIds) => api.post('/connections/batch/delete', { connectionIds }),
  
  // Test a database connection
  testConnection: (data) => api.post('/connections/test', data, {
    timeout: 30000, 
    validateStatus: function (status) {
      return status >= 200 && status < 500
    }
  }),
  
  // Get connection history
  getConnectionHistory: (id) => api.get(`/connections/${id}/history`),
  
  // Get connection metrics history
  getConnectionMetricsHistory: (connectionId) => api.get(`/connections/${connectionId}/metrics/history`),
  
  // Get batch connection stats
  getBatchConnectionStats: (connectionIds) => api.post('/connections/batch/stats', { connectionIds }),
  
  // Batch test connections
  batchTestConnections: (ids) => api.post('/connections/batch-test', { ids }, {
    timeout: 60000
  }),
  
  // Batch delete connections
  batchDeleteConnections: (ids) => api.post('/connections/batch-delete', { ids }),
}

export const benchmarkAPI = {
  // Get all benchmarks
  getBenchmarks: () => api.get('/api/benchmarks'),
  
  // Get a single benchmark
  getBenchmark: (id) => api.get(`/api/benchmarks/${id}`),
  
  // Create a new benchmark
  createBenchmark: (data) => api.post('/api/benchmarks', data),
  
  // Start a benchmark
  startBenchmark: (id) => api.post(`/api/benchmarks/${id}/start`),
  
  // Stop a benchmark
  stopBenchmark: (id) => api.post(`/api/benchmarks/${id}/stop`),
  
  // Get benchmark results
  getBenchmarkResults: (id) => api.get(`/api/benchmarks/${id}/results`)
}

export const authAPI = {
  // Login
  login: (credentials) => api.post('/api/auth/login', credentials),
  
  // Logout
  logout: () => api.post('/api/auth/logout'),
  
  // Get current user
  getCurrentUser: () => api.get('/api/auth/user')
}

export default api
