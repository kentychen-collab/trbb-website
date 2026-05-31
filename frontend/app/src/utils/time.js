const TZ = 'Asia/Taipei'

export function fmtDate(d) {
  if (!d) return ''
  try {
    const s = String(d)
    if (/^\d{4}-\d{2}-\d{2}$/.test(s)) {
      return s.replace(/-/g, '/')
    }
    const dt = new Date(s)
    const y  = dt.toLocaleString('zh-TW', { timeZone: TZ, year:  'numeric' })
    const mo = dt.toLocaleString('zh-TW', { timeZone: TZ, month: '2-digit' })
    const dy = dt.toLocaleString('zh-TW', { timeZone: TZ, day:   '2-digit' })
    return `${y}/${mo}/${dy}`
  } catch { return String(d).substring(0, 10) }
}

export function fmtTime(d) {
  if (!d) return ''
  try {
    return new Date(d).toLocaleTimeString('zh-TW', {
      timeZone: TZ,
      hour: '2-digit', minute: '2-digit',
      hour12: true   // 上午/下午
    })
  } catch { return '' }
}

export function fmtDateTime(d) {
  if (!d) return ''
  try {
    const dt = new Date(d)
    const date = fmtDate(d)
    const time = fmtTime(d)
    return `${date} ${time}`
  } catch { return '' }
}

export function fmtRelative(d) {
  if (!d) return ''
  const diff = (Date.now() - new Date(d).getTime()) / 1000
  if (diff < 60)      return '剛剛'
  if (diff < 3600)    return `${Math.floor(diff/60)} 分鐘前`
  if (diff < 86400)   return `${Math.floor(diff/3600)} 小時前`
  if (diff < 86400*7) return `${Math.floor(diff/86400)} 天前`
  return fmtDate(d)
}
