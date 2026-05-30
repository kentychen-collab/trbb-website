import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import api from '@/services/api'

export const useSiteSettingsStore = defineStore('siteSettings', () => {
  const settings = reactive({})
  const loaded = ref(false)

  async function load() {
    try {
      const { data } = await api.get('/settings')
      Object.keys(settings).forEach(k => delete settings[k])
      Object.assign(settings, data)
      applyCSS(data)
      loaded.value = true
    } catch(e) {
      console.warn('site settings load failed', e)
    }
  }

  function get(key, fallback = '') {
    const v = settings[key]
    return (v === null || v === undefined || v === '') ? fallback : v
  }

  function applyCSS(s) {
    const root = document.documentElement

    // Helper: gradient value → CSS
    const grad = (val, fallback) => {
      if (!val) return fallback
      try {
        const g = JSON.parse(val)
        if (g.type === 'solid') return g.colors[0]
        if (g.type === 'linear') {
          const deg = g.angle || '135deg'
          const stops = g.colors.map((c, i) => {
            const pct = g.stops?.[i] ?? Math.round(i * 100 / (g.colors.length - 1))
            return `${c} ${pct}%`
          }).join(', ')
          return `linear-gradient(${deg}, ${stops})`
        }
        if (g.type === 'radial') {
          const stops = g.colors.map((c, i) => {
            const pct = g.stops?.[i] ?? Math.round(i * 100 / (g.colors.length - 1))
            return `${c} ${pct}%`
          }).join(', ')
          return `radial-gradient(circle, ${stops})`
        }
      } catch {}
      return val || fallback
    }

    // Background
    const bgVal = grad(s.bg_color, '#FFFFF3')
    if (bgVal.includes('gradient')) {
      root.style.setProperty('--color-bg', 'transparent')
      document.body.style.background = bgVal
    } else {
      root.style.setProperty('--color-bg', bgVal)
      document.body.style.background = ''
    }
    root.style.setProperty('--color-bg-2',    grad(s.bg2_color,   '#F5F5E8'))
    root.style.setProperty('--color-bg-card', grad(s.card_color,  '#FFFFFF'))
    root.style.setProperty('--color-border',  s.border_color  || '#E0E3DA')

    // Brand
    root.style.setProperty('--color-primary', s.primary_color || '#CF2027')
    root.style.setProperty('--color-navy',    s.navy_color    || '#1A3A7A')
    root.style.setProperty('--color-accent',  s.accent_color  || '#A593E0')

    // Typography — body
    if (s.font_body)        root.style.setProperty('--font-body',           `'${s.font_body}', sans-serif`)
    if (s.font_body_size)   root.style.setProperty('--font-body-size',      s.font_body_size)
    if (s.font_body_color)  root.style.setProperty('--color-text',          s.font_body_color)
    if (s.font_body_color)  root.style.setProperty('--color-gray-1',        s.font_body_color)
    if (s.font_body_weight) root.style.setProperty('--font-body-weight',    s.font_body_weight)
    document.body.style.fontSize   = s.font_body_size   || ''
    document.body.style.color      = s.font_body_color  || ''
    document.body.style.fontWeight = s.font_body_weight || ''

    // Typography — heading
    if (s.font_heading)        root.style.setProperty('--font-cond',          `'${s.font_heading}', sans-serif`)
    if (s.font_heading_color)  root.style.setProperty('--color-navy',         s.font_heading_color)
    if (s.font_heading_weight) root.style.setProperty('--font-heading-weight', s.font_heading_weight)

    // Typography — display
    if (s.font_display)       root.style.setProperty('--font-display',    `'${s.font_display}', sans-serif`)
    if (s.font_display_color) root.style.setProperty('--color-display',   s.font_display_color)

    // Navbar
    const navBg = grad(s.navbar_bg, 'rgba(255,255,243,0.95)')
    root.style.setProperty('--navbar-bg',   navBg)
    root.style.setProperty('--navbar-text', s.navbar_text || '#1A3A7A')

    // Inject google font if custom font
    injectGoogleFont(s.font_body)
    injectGoogleFont(s.font_heading)
    injectGoogleFont(s.font_display)

    // Favicon
    if (s.site_icon) applyFavicon(s.site_icon)
  }

  function applyFavicon(iconUrl) {
    if (!iconUrl) return
    // 更新 <link rel="icon">
    let link = document.querySelector("link[rel~='icon']")
    if (!link) {
      link = document.createElement('link')
      link.rel = 'icon'
      document.head.appendChild(link)
    }
    link.href = iconUrl

    // 更新 Apple Touch Icon
    let appleLink = document.querySelector("link[rel='apple-touch-icon']")
    if (!appleLink) {
      appleLink = document.createElement('link')
      appleLink.rel = 'apple-touch-icon'
      document.head.appendChild(appleLink)
    }
    appleLink.href = s.site_icon_lg || iconUrl
  }

  function injectGoogleFont(fontName) {
    if (!fontName || fontName.includes('sans-serif') || fontName.includes('serif')) return
    const id = `gf-${fontName.replace(/\s+/g, '-')}`
    if (document.getElementById(id)) return
    const link = document.createElement('link')
    link.id = id
    link.rel = 'stylesheet'
    link.href = `https://fonts.googleapis.com/css2?family=${encodeURIComponent(fontName)}:wght@300;400;500;600;700;900&display=swap`
    document.head.appendChild(link)
  }

  return { settings, loaded, load, get }
})
