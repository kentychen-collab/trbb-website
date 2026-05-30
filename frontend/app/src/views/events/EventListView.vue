<template>
  <div class="events-page">

    <div class="container" style="padding-top:calc(64px + 2rem); padding-bottom:4rem">
      <!-- Header -->
      <div class="page-hero">
        <div class="page-hero-tag">RACE EVENTS</div>
        <h1 class="page-hero-title">賽事報名</h1>
        <p class="page-hero-desc">鐵人三項、路跑、游泳、單車，每一場賽事都是突破自己的機會。</p>
      </div>

      <!-- Filters -->
      <div class="events-filter">
        <input v-model="keyword" placeholder="搜尋賽事名稱 / 地點..." @keyup.enter="fetchEvents" class="filter-input" />
        <button class="btn btn-primary btn-sm" @click="fetchEvents">搜尋</button>
        <button v-if="keyword" class="btn btn-ghost btn-sm" @click="keyword=''; fetchEvents()">清除</button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="loading-state">
        <div class="loading-spinner"></div>
        <span>載入中...</span>
      </div>

      <!-- Empty -->
      <div v-else-if="!events.length" class="empty-state">
        <div class="empty-icon">🏁</div>
        <p>目前沒有開放報名的賽事</p>
      </div>

      <!-- List -->
      <div v-else class="events-grid">
        <RouterLink v-for="ev in events" :key="ev.id"
          :to="`/events/${ev.id}`" class="event-card">
          <div class="event-cover">
            <img v-if="ev.cover_url" :src="ev.cover_url" :alt="ev.title" />
            <div v-else class="event-cover-placeholder">
              <span>{{ eventTypeIcon(ev.event_type) }}</span>
            </div>
            <div class="event-type-tag">{{ eventTypeLabel(ev.event_type) }}</div>
            <div class="event-status-tag" :class="eventStatusClass(ev)">{{ eventStatusLabel(ev) }}</div>
          </div>
          <div class="event-body">
            <h3 class="event-title">{{ ev.title }}</h3>
            <div class="event-meta">
              <span class="meta-item">📅 {{ formatDate(ev.start_at) }}</span>
              <span class="meta-item">📍 {{ ev.location }}</span>
            </div>
            <div class="event-footer">
              <span class="event-fee" v-if="ev.fee > 0">NT$ {{ ev.fee.toLocaleString() }}</span>
              <span class="event-fee free" v-else>免費</span>
              <span class="event-spots" v-if="ev.max_participants">
                {{ ev.registered_count }}/{{ ev.max_participants }} 人
              </span>
              <span class="event-reg-info">{{ regStatusText(ev) }}</span>
            </div>
          </div>
        </RouterLink>
      </div>

      <!-- Pagination -->
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

const auth = useAuthStore()
const events     = ref([])
const loading    = ref(false)
const keyword    = ref('')
const page       = ref(1)
const totalPages = ref(1)

async function fetchEvents() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: 12 }
    if (keyword.value) params.keyword = keyword.value
    const { data } = await api.get('/events', { params })
    events.value      = data.events || []
    totalPages.value  = data.pages  || 1
  } catch {}
  finally { loading.value = false }
}

function goPage(p) { page.value = p; fetchEvents() }

const typeMap  = { 1:'鐵人三項', 2:'路跑', 3:'游泳', 4:'單車', 5:'訓練', 6:'其他' }
const typeIcon = { 1:'🏊', 2:'🏃', 3:'🏊', 4:'🚴', 5:'💪', 6:'🏅' }
function eventTypeLabel(t) { return typeMap[t] || '其他' }
function eventTypeIcon(t)  { return typeIcon[t] || '🏅' }

function formatDate(d) {
  return new Date(d).toLocaleDateString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit' })
}

function eventStatusClass(ev) {
  const now = new Date()
  const regEnd = new Date(ev.reg_end_at)
  const regStart = new Date(ev.reg_start_at)
  if (now < regStart) return 'tag-upcoming'
  if (now > regEnd)   return 'tag-closed'
  if (ev.max_participants && ev.registered_count >= ev.max_participants) return 'tag-full'
  return 'tag-open'
}
function eventStatusLabel(ev) {
  const cls = eventStatusClass(ev)
  return { 'tag-upcoming':'即將開放', 'tag-closed':'已截止', 'tag-full':'額滿', 'tag-open':'報名中' }[cls]
}
function regStatusText(ev) {
  const now = new Date()
  const end = new Date(ev.reg_end_at)
  const start = new Date(ev.reg_start_at)
  if (now < start) return `${formatDate(ev.reg_start_at)} 開放`
  if (now > end)   return `截止 ${formatDate(ev.reg_end_at)}`
  return `截止 ${formatDate(ev.reg_end_at)}`
}

onMounted(fetchEvents)
</script>

<style scoped>
.nav-link:hover { color:var(--color-primary); }

.page-hero { padding:3rem 0 2rem;text-align:center; }
.page-hero-tag { font-family:var(--font-cond);font-size:.75rem;font-weight:700;letter-spacing:.3em;color:var(--color-primary);margin-bottom:.75rem; }
.page-hero-title { font-family:var(--font-display);font-size:clamp(3rem,8vw,6rem);margin-bottom:.75rem; }
.page-hero-desc { color:var(--color-gray-1);font-size:1rem;max-width:500px;margin:0 auto; }

.events-filter { display:flex;gap:.75rem;align-items:center;margin-bottom:2rem;flex-wrap:wrap; }
.filter-input { flex:1;min-width:200px;max-width:400px;height:40px; }

.loading-state { display:flex;align-items:center;justify-content:center;gap:1rem;padding:4rem;color:var(--color-gray-2); }
.loading-spinner { width:24px;height:24px;border:2px solid var(--color-border);border-top-color:var(--color-primary);border-radius:50%;animation:spin .7s linear infinite; }
@keyframes spin { to { transform:rotate(360deg) } }
.empty-state { text-align:center;padding:4rem;color:var(--color-gray-2); }
.empty-icon { font-size:3rem;margin-bottom:1rem; }

.events-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(320px,1fr));gap:1.5rem;margin-bottom:2rem; }

.event-card { background:var(--color-bg-card);border:1px solid var(--color-border);border-radius:8px;overflow:hidden;transition:all .2s;display:block; }
.event-card:hover { border-color:var(--color-primary);transform:translateY(-3px);box-shadow:0 8px 30px rgba(229,25,26,.15); }

.event-cover { position:relative;height:180px;overflow:hidden;background:var(--color-bg-2); }
.event-cover img { width:100%;height:100%;object-fit:cover; }
.event-cover-placeholder { width:100%;height:100%;display:flex;align-items:center;justify-content:center;font-size:4rem;background:linear-gradient(135deg,#1a1a2e,#16213e); }
.event-type-tag { position:absolute;top:.75rem;left:.75rem;font-family:var(--font-cond);font-size:.72rem;font-weight:700;letter-spacing:.08em;text-transform:uppercase;padding:.2rem .65rem;background:rgba(0,0,0,.7);border:1px solid rgba(255,255,255,.15);border-radius:3px; }
.event-status-tag { position:absolute;top:.75rem;right:.75rem;font-family:var(--font-cond);font-size:.7rem;font-weight:700;letter-spacing:.06em;padding:.2rem .6rem;border-radius:3px; }
.tag-open     { background:rgba(34,197,94,.2);color:#22c55e;border:1px solid rgba(34,197,94,.3); }
.tag-upcoming { background:rgba(245,158,11,.2);color:#f59e0b;border:1px solid rgba(245,158,11,.3); }
.tag-closed   { background:rgba(107,114,128,.2);color:#9ca3af;border:1px solid rgba(107,114,128,.3); }
.tag-full     { background:rgba(239,68,68,.2);color:#ef4444;border:1px solid rgba(239,68,68,.3); }

.event-body { padding:1.25rem; }
.event-title { font-size:1rem;font-weight:700;line-height:1.4;margin-bottom:.6rem; }
.event-meta { display:flex;flex-direction:column;gap:.25rem;margin-bottom:.75rem; }
.meta-item { font-size:.82rem;color:var(--color-gray-2); }
.event-footer { display:flex;align-items:center;gap:.75rem;flex-wrap:wrap;padding-top:.75rem;border-top:1px solid var(--color-border); }
.event-fee { font-family:var(--font-cond);font-size:1.1rem;font-weight:700;color:var(--color-primary); }
.event-fee.free { color:#22c55e; }
.event-spots { font-size:.78rem;color:var(--color-gray-2); }
.event-reg-info { font-size:.75rem;color:var(--color-gray-2);margin-left:auto; }

.pagination { display:flex;align-items:center;justify-content:center;gap:1rem;margin-top:2rem; }
</style>
