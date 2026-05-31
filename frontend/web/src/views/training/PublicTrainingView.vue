<template>
  <div class="training-pub-page">

    <div class="container" style="padding-top:calc(64px + 2rem);padding-bottom:4rem">
      <div class="page-hero">
        <div class="page-hero-tag">TRAINING DIARY</div>
        <h1 class="page-hero-title">訓練動態</h1>
        <p class="page-hero-desc">TRBB 會員的訓練足跡，每一步都是進步的累積。</p>
      </div>

      <!-- Sport filter -->
      <div class="sport-filter">
        <button v-for="s in sports" :key="s.value"
          class="sport-btn" :class="{ active: selectedSport === s.value }"
          @click="selectSport(s.value)">
          {{ s.icon }} {{ s.label }}
        </button>
      </div>

      <div v-if="loading" class="loading-state"><div class="loading-spinner"></div></div>
      <div v-else-if="!logs.length" class="empty-state"><p>目前沒有公開的訓練記錄</p></div>
      <div v-else class="logs-grid">
        <RouterLink v-for="log in logs" :key="log.id"
          :to="`/training/${log.id}`" class="log-card card">
          <div class="log-cover" :class="`sport-${log.sport_type}`">
            <img v-if="log.photos?.length" :src="imgUrl(log.photos[0])" class="cover-img" />
            <span v-else class="cover-icon">{{ sportIcon(log.sport_type) }}</span>
            <span class="sport-badge">{{ sportLabel(log.sport_type) }}</span>
          </div>
          <div class="log-body">
            <div class="log-meta">
              <span class="log-author">{{ log.display_name || log.username }}</span>
              <span class="log-date text-gray">{{ log.created_at ? fmtDateTime(log.created_at) : fmtDate(log.date) }}</span>
            </div>
            <h3 class="log-title">{{ log.title }}</h3>
            <div class="log-stats">
              <span v-if="log.distance_km">📏 {{ Number(log.distance_km).toFixed(2) }} km</span>
              <span v-if="log.duration_min">⏱ {{ fmtDuration(log.duration_min) }}</span>
              <span v-if="log.avg_heart_rate">❤️ {{ log.avg_heart_rate }} bpm</span>
              <span v-if="log.elevation_m">⛰ {{ log.elevation_m }} m</span>
            </div>
          </div>
        </RouterLink>
      </div>

      <div class="pagination" v-if="totalPages > 1">
        <button :disabled="page===1" @click="goPage(page-1)" class="btn btn-ghost">‹</button>
        <span class="text-gray">{{ page }} / {{ totalPages }}</span>
        <button :disabled="page===totalPages" @click="goPage(page+1)" class="btn btn-ghost">›</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'
import { fmtDate, fmtDateTime } from '@/utils/time'

const auth = useAuthStore()
const logs = ref([])
const loading = ref(false)
const page = ref(1)
const totalPages = ref(1)
const selectedSport = ref(0)

const sports = [
  { value:0, icon:'🏅', label:'全部' }, { value:1, icon:'🏃', label:'路跑' },
  { value:2, icon:'🏊', label:'游泳' }, { value:3, icon:'🚴', label:'單車' },
  { value:4, icon:'🏅', label:'鐵人' }, { value:5, icon:'💪', label:'重訓' },
]
const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || ''
function imgUrl(p) { return p?.startsWith('http') ? p : `${IMAGE_BASE}/images/${p}` }
function sportIcon(t) { return { 1:'🏃', 2:'🏊', 3:'🚴', 4:'🏅', 5:'💪', 6:'🏋️' }[t] || '🏅' }
function sportLabel(t) { return { 1:'路跑', 2:'游泳', 3:'單車', 4:'鐵人', 5:'重訓', 6:'其他' }[t] || '其他' }
function fmtDuration(min) {
  if (!min) return ''
  const h = Math.floor(min/60), m = min%60
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}
async function fetchLogs() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: 12 }
    if (selectedSport.value) params.sport_type = selectedSport.value
    const { data } = await api.get('/training', { params })
    logs.value = data.logs || []
    totalPages.value = data.pages || 1
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}
function selectSport(v) { selectedSport.value = v; page.value = 1; fetchLogs() }
function goPage(p) { page.value = p; fetchLogs() }
onMounted(fetchLogs)
</script>

<style scoped>
.nav-link:hover { color:var(--color-primary); }
.page-hero { padding:3rem 0 2rem;text-align:center; }
.page-hero-tag { font-family:var(--font-cond);font-size:.75rem;font-weight:700;letter-spacing:.3em;color:var(--color-primary);margin-bottom:.75rem; }
.page-hero-title { font-family:var(--font-display);font-size:clamp(3rem,8vw,5rem);margin-bottom:.75rem; }
.page-hero-desc { color:var(--color-gray-1);max-width:400px;margin:0 auto; }
.sport-filter { display:flex;gap:.5rem;flex-wrap:wrap;margin-bottom:2rem; }
.sport-btn { padding:.4rem 1rem;border-radius:4px;border:1px solid var(--color-border);font-size:.82rem;font-weight:600;cursor:pointer;background:none;color:var(--color-gray-2);transition:all .15s; }
.sport-btn.active { border-color:var(--color-primary);color:var(--color-primary); }
.loading-state,.empty-state { display:flex;align-items:center;justify-content:center;padding:4rem;color:var(--color-gray-2); }
.loading-spinner { width:24px;height:24px;border:2px solid var(--color-border);border-top-color:var(--color-primary);border-radius:50%;animation:spin .7s linear infinite; }
@keyframes spin { to { transform:rotate(360deg) } }
.logs-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:1.5rem;margin-bottom:2rem; }
.log-card { display:block;background:var(--color-bg-card);border:1px solid var(--color-border);border-radius:8px;overflow:hidden;transition:all .2s; }
.log-card:hover { border-color:var(--color-primary);transform:translateY(-3px); }
.log-cover { height:150px;display:flex;align-items:center;justify-content:center;position:relative; }
.sport-1 { background:linear-gradient(135deg,#1a1a2e,#2d1b4e); }
.sport-2 { background:linear-gradient(135deg,#0a1628,#0e3460); }
.sport-3 { background:linear-gradient(135deg,#1a1a1a,#2d2000); }
.sport-4 { background:linear-gradient(135deg,#1a0d1a,#3a1a3a); }
.sport-5, .sport-6 { background:linear-gradient(135deg,#1a0d0d,#3a1a1a); }
.cover-img { width:100%;height:100%;object-fit:cover;position:absolute;inset:0; }
.cover-icon { font-size:3rem;position:relative;z-index:1; }
.sport-badge { position:absolute;bottom:.5rem;left:.5rem;font-family:var(--font-cond);font-size:.65rem;font-weight:700;letter-spacing:.1em;text-transform:uppercase;padding:.15rem .5rem;background:rgba(0,0,0,.6);border-radius:3px;color:#fff; }
.log-body { padding:1rem; }
.log-meta { display:flex;justify-content:space-between;margin-bottom:.3rem;font-size:.78rem; }
.log-author { font-weight:600;color:var(--color-gray-1); }
.log-title { font-size:.95rem;font-weight:700;line-height:1.4;margin-bottom:.4rem; }
.log-stats { display:flex;gap:.6rem;flex-wrap:wrap;font-size:.75rem;color:var(--color-gray-2); }
.pagination { display:flex;align-items:center;justify-content:center;gap:1rem; }
</style>
