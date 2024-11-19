export function formatDate(date) {
  if (!date) return ''
  
  const d = new Date(date)
  if (isNaN(d.getTime())) return ''

  const now = new Date()
  const diff = now - d
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  // 如果是今天
  if (d.toDateString() === now.toDateString()) {
    if (seconds < 60) {
      return 'just now'
    }
    if (minutes < 60) {
      return `${minutes} minute${minutes === 1 ? '' : 's'} ago`
    }
    if (hours < 24) {
      return `${hours} hour${hours === 1 ? '' : 's'} ago`
    }
  }

  // 如果是昨天
  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  if (d.toDateString() === yesterday.toDateString()) {
    return 'Yesterday'
  }

  // 如果是本周
  if (days < 7) {
    return `${days} day${days === 1 ? '' : 's'} ago`
  }

  // 如果是本年
  if (d.getFullYear() === now.getFullYear()) {
    return d.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric'
    })
  }

  // 其他情况
  return d.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

export function formatDateTime(date) {
  if (!date) return ''
  
  const d = new Date(date)
  if (isNaN(d.getTime())) return ''

  return d.toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

export function formatTime(date) {
  if (!date) return ''
  
  const d = new Date(date)
  if (isNaN(d.getTime())) return ''

  return d.toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit'
  })
}

export function formatDuration(ms) {
  if (!ms || isNaN(ms)) return '0ms'

  const seconds = Math.floor(ms / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`
  }
  if (seconds > 0) {
    return `${seconds}s`
  }
  return `${ms}ms`
}
