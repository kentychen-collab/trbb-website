import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { useSiteSettingsStore } from '@/stores/siteSettings'
import App from './App.vue'
import router from './router'
import './assets/css/main.css'

async function bootstrap() {
  const app = createApp(App)
  const pinia = createPinia()
  app.use(pinia)
  app.use(router)

  // 載入網站設定，套用 CSS 變數
  try {
    const siteStore = useSiteSettingsStore()
    await siteStore.load()
  } catch(e) {
    console.warn('site settings unavailable', e)
  }

  app.mount('#app')
}

bootstrap()
