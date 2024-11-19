/**
 * 分页工具类
 */
export class Pagination {
  constructor(pageSize = 20) {
    this.pageSize = pageSize
    this.currentPage = 1
    this.total = 0
  }

  /**
   * 获取分页参数
   */
  getParams() {
    return {
      page: this.currentPage,
      pageSize: this.pageSize,
    }
  }

  /**
   * 更新总数
   */
  setTotal(total) {
    this.total = total
  }

  /**
   * 获取总页数
   */
  getTotalPages() {
    return Math.ceil(this.total / this.pageSize)
  }

  /**
   * 是否有下一页
   */
  hasNextPage() {
    return this.currentPage < this.getTotalPages()
  }

  /**
   * 是否有上一页
   */
  hasPrevPage() {
    return this.currentPage > 1
  }

  /**
   * 下一页
   */
  nextPage() {
    if (this.hasNextPage()) {
      this.currentPage++
      return true
    }
    return false
  }

  /**
   * 上一页
   */
  prevPage() {
    if (this.hasPrevPage()) {
      this.currentPage--
      return true
    }
    return false
  }

  /**
   * 重置分页
   */
  reset() {
    this.currentPage = 1
    this.total = 0
  }
}

/**
 * 虚拟滚动工具类
 */
export class VirtualScroll {
  constructor(options = {}) {
    this.itemHeight = options.itemHeight || 50
    this.bufferSize = options.bufferSize || 5
    this.viewportHeight = options.viewportHeight || 500
    this.items = []
    this.startIndex = 0
    this.visibleCount = Math.ceil(this.viewportHeight / this.itemHeight)
  }

  /**
   * 更新数据源
   */
  setItems(items) {
    this.items = items
    this.updateVisibleItems()
  }

  /**
   * 获取可见项
   */
  getVisibleItems() {
    const start = Math.max(0, this.startIndex - this.bufferSize)
    const end = Math.min(
      this.items.length,
      this.startIndex + this.visibleCount + this.bufferSize
    )
    return this.items.slice(start, end)
  }

  /**
   * 更新滚动位置
   */
  updateScrollPosition(scrollTop) {
    this.startIndex = Math.floor(scrollTop / this.itemHeight)
    return this.getVisibleItems()
  }

  /**
   * 获取总高度
   */
  getTotalHeight() {
    return this.items.length * this.itemHeight
  }

  /**
   * 获取偏移量
   */
  getOffset() {
    const start = Math.max(0, this.startIndex - this.bufferSize)
    return start * this.itemHeight
  }
}

/**
 * 缓存管理工具类
 */
export class CacheManager {
  constructor(options = {}) {
    this.maxAge = options.maxAge || 5 * 60 * 1000 // 默认5分钟
    this.maxSize = options.maxSize || 100 // 默认缓存100条
    this.cache = new Map()
    this.expirations = new Map()
  }

  /**
   * 获取缓存
   */
  get(key) {
    const item = this.cache.get(key)
    if (!item) return null
    if (this.isExpired(key)) {
      this.delete(key)
      return null
    }
    return item
  }

  /**
   * 设置缓存
   */
  set(key, value) {
    if (this.cache.size >= this.maxSize) {
      // 删除最旧的缓存
      const oldestKey = this.cache.keys().next().value
      this.delete(oldestKey)
    }
    this.cache.set(key, value)
    this.expirations.set(key, Date.now() + this.maxAge)
  }

  /**
   * 删除缓存
   */
  delete(key) {
    this.cache.delete(key)
    this.expirations.delete(key)
  }

  /**
   * 清除缓存
   */
  clear() {
    this.cache.clear()
    this.expirations.clear()
  }

  /**
   * 检查缓存是否过期
   */
  isExpired(key) {
    const expiration = this.expirations.get(key)
    return !expiration || expiration <= Date.now()
  }

  /**
   * 获取即将过期的缓存键
   */
  getExpiringKeys(timeWindow) {
    const now = Date.now()
    const expiringKeys = []
    this.expirations.forEach((expiration, key) => {
      if (expiration - now <= timeWindow) {
        expiringKeys.push(key)
      }
    })
    return expiringKeys
  }
}
