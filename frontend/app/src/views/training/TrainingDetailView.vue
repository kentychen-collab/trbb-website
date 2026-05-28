<template>
  <div class="training-detail-page">
    <nav class="navbar">
      <div class="container nav-inner">
        <RouterLink to="/" class="trbb-logo nav-logo"><span class="tr">TR</span><span class="bb">BB</span></RouterLink>
        <div class="flex items-center gap-2">
          <RouterLink to="/training" class="nav-link">← 訓練動態</RouterLink>
          <RouterLink v-if="auth.isLoggedIn" to="/me/training" class="btn btn-ghost btn-sm">我的日記</RouterLink>
          <RouterLink v-else to="/login" class="btn btn-primary btn-sm">登入</RouterLink>
        </div>
      </div>
    </nav>

    <div v-if="loading" class="loading-full"><div class="loading-spinner"></div></div>
    <div v-else-if="!log" class="not-found container">
      <p>訓練記錄不存在或尚未公開</p>
      <RouterLink to="/training" class="btn btn-ghost" style="margin-top:1rem">返回訓練動態</RouterLink>
    </div>

    <template v-else>
      <!-- Hero -->
      <div class="detail-hero" :style="log.photos?.length ? `background-image:url(${imgUrl(log.photos[0])})` : ''">
        <div class="hero-overlay"></div>
        <div class="container hero-content">
          <div class="hero-sport">{{ sportIcon(log.sport_type) }} {{ sportLabel(log.sport_type) }}</div>
          <h1 class="hero-title">{{ log.title }}</h1>
          <div class="hero-meta">
            <span>👤 {{ log.display_name || log.username }}</span>
            <span>📅 {{ log.date }}</span>
            <span v-if="log.source !== 'manual'">📡 {{ sourceLabel(log.source) }}</span>
          </div>
        </div>
      </div>

      <div class="container detail-main">
        <!-- Stats row -->
        <div class="stats-card card">
          <div class="stat-item" v-if="log.distance_km">
            <span class="stat-val">{{ Number(log.distance_km).toFixed(2) }}</span>
            <span class="stat-unit">公里</span>
          </div>
          <div class="stat-div" v-if="log.distance_km && log.duration_min"></div>
          <div class="stat-item" v-if="log.duration_min">
            <span class="stat-val">{{ fmtDuration(log.duration_min) }}</span>
            <span class="stat-unit">時間</span>
          </div>
          <div class="stat-div" v-if="log.avg_pace"></div>
          <div class="stat-item" v-if="log.avg_pace">
            <span class="stat-val">{{ log.avg_pace }}</span>
            <span class="stat-unit">/km 配速</span>
          </div>
          <div class="stat-div" v-if="log.avg_heart_rate"></div>
          <div class="stat-item" v-if="log.avg_heart_rate">
            <span class="stat-val">{{ log.avg_heart_rate }}</span>
            <span class="stat-unit">bpm 均心率</span>
          </div>
          <div class="stat-div" v-if="log.elevation_m"></div>
          <div class="stat-item" v-if="log.elevation_m">
            <span class="stat-val">{{ log.elevation_m }}</span>
            <span class="stat-unit">m 爬升</span>
          </div>
          <div class="stat-div" v-if="log.calories"></div>
          <div class="stat-item" v-if="log.calories">
            <span class="stat-val">{{ log.calories }}</span>
            <span class="stat-unit">kcal</span>
          </div>
        </div>

        <!-- Map -->
        <div v-if="log.route_points?.length" class="map-section card">
          <h3 class="section-label">GPS 路線</h3>
          <TrainingMap :route-points="log.route_points"
            :start-lat="log.start_lat" :start-lng="log.start_lng"
            height="320px" />
        </div>

        <!-- Photos -->
        <div v-if="log.photos?.length > 1" class="photos-section card">
          <h3 class="section-label">照片</h3>
          <div class="photos-grid">
            <img v-for="(p, i) in log.photos" :key="i"
              :src="imgUrl(p)" @click="activePhoto=p"
              class="photo-thumb" />
          </div>
        </div>

        <!-- Note -->
        <div v-if="log.note" class="note-section card">
          <h3 class="section-label">訓練備註</h3>
          <p class="note-text">{{ log.note }}</p>
        </div>

        <!-- Share -->
        <div class="share-section card">
          <h3 class="section-label">分享</h3>
          <div class="share-btns">
            <button class="share-btn" @click="copyLink">🔗 複製連結</button>
            <a :href="lineUrl" target="_blank" class="share-btn line">LINE</a>
            <a :href="fbUrl"   target="_blank" class="share-btn fb">Facebook</a>
            <button class="share-btn ig" @click="copyForIG">📷 Instagram</button>
          </div>
          <div v-if="copied" class="copied-hint">✓ 連結已複製</div>
        </div>
      </div>
    </template>

    <!-- Photo lightbox -->
    <div v-if="activePhoto" class="lightbox" @click="activePhoto=null">
      <img :src="imgUrl(activePhoto)" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'
import TrainingMap from '@/components/TrainingMap.vue'

const route  = useRoute()
const auth   = useAuthStore()
const log    = ref(null)
const loading = ref(true)
const activePhoto = ref(null)
const copied = ref(false)

const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || ''
function imgUrl(p) { return p?.startsWith('http') ? p : `${IMAGE_BASE}/images/${p}` }

const shareUrl = computed(() => `${window.location.origin}/training/share/${log.value?.uuid}`)
const lineUrl  = computed(() => `https://social-plugins.line.me/lineit/share?url=${encodeURIComponent(shareUrl.value)}`)
const fbUrl    = computed(() => `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(shareUrl.value)}`)

function copyLink() {
  navigator.clipboard.writeText(shareUrl.value).then(() => {
    copied.value = true
    setTimeout(() => { copied.value = false }, 2500)
  })
}
function copyForIG() {
  navigator.clipboard.writeText(shareUrl.value).then(() => alert('連結已複製，請貼到 Instagram 限時動態！'))
}

function sportIcon(t) { return { 1:'🏃', 2:'🏊', 3:'🚴', 4:'🏅', 5:'💪', 6:'🏋️' }[t] || '🏅' }
function sportLabel(t) { return { 1:'路跑', 2:'游泳', 3:'單車', 4:'鐵人三項', 5:'重訓', 6:'其他' }[t] || '其他' }
function sourceLabel(s) { return { gpx:'GPX 匯入', fit:'FIT 匯入', garmin:'Garmin 同步' }[s] || '' }
function fmtDuration(min) {
  if (!min) return ''
  const h = Math.floor(min / 60), m = min % 60
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

onMounted(async () => {
  try {
    // route.params.id 可能是數字 ID 或 uuid (share link)
    const id = route.params.id
    const isUUID = id.includes('-')
    const url = isUUID ? `/training/share/${id}` : `/training/${id}`
    const { data } = await api.get(url)
    log.value = data
  } catch { log.value = null }
  finally { loading.value = false }
})
</script>

<style scoped>
.training-detail-page { background:var(--color-bg); min-height:100vh; }
.navbar { position:fixed;top:0;left:0;right:0;z-index:100;background:rgba(0,0,0,.9);backdrop-filter:blur(12px);border-bottom:1px solid var(--color-border); }
.nav-inner { display:flex;align-items:center;justify-content:space-between;height:64px; }
.nav-logo { font-size:2rem; }
.nav-link { font-family:var(--font-cond);font-weight:600;letter-spacing:.05em;font-size:.85rem;color:var(--color-gray-1); }
.nav-link:hover { color:var(--color-primary); }
.btn-sm { padding:.4rem 1rem;font-size:.82rem; }
.loading-full { display:flex;align-items:center;justify-content:center;height:100vh; }
.loading-spinner { width:32px;height:32px;border:3px solid var(--color-border);border-top-color:var(--color-primary);border-radius:50%;animation:spin .7s linear infinite; }
@keyframes spin { to { transform:rotate(360deg) } }
.not-found { padding-top:calc(64px+4rem); text-align:center; color:var(--color-gray-2); }

/* Hero */
.detail-hero { height:320px; margin-top:64px; background:#0a0a1a; background-size:cover; background-position:center; position:relative; }
.hero-overlay { position:absolute;inset:0;background:linear-gradient(to bottom,rgba(0,0,0,.2),rgba(0,0,0,.8)); }
.hero-content { position:relative;z-index:1;height:100%;display:flex;flex-direction:column;justify-content:flex-end;padding-bottom:2rem; }
.hero-sport { font-family:var(--font-cond);font-size:.8rem;letter-spacing:.15em;text-transform:uppercase;color:var(--color-primary);margin-bottom:.4rem; }
.hero-title { font-family:var(--font-display);font-size:clamp(1.8rem,5vw,3rem);line-height:1.2;margin-bottom:.75rem; }
.hero-meta { display:flex;gap:1.5rem;font-size:.85rem;color:var(--color-gray-1);flex-wrap:wrap; }

/* Main content */
.detail-main { padding:2rem 0 4rem; display:flex; flex-direction:column; gap:1.25rem; max-width:720px; }

/* Stats */
.stats-card { padding:1.5rem; display:flex; align-items:center; flex-wrap:wrap; gap:1rem; }
.stat-item { display:flex; flex-direction:column; align-items:center; flex:1; min-width:80px; }
.stat-val { font-family:var(--font-display); font-size:2rem; color:var(--color-primary); }
.stat-unit { font-size:.7rem; color:var(--color-gray-2); text-transform:uppercase; letter-spacing:.08em; }
.stat-div { width:1px; height:40px; background:var(--color-border); }

/* Sections */
.section-label { font-family:var(--font-cond); font-size:.78rem; font-weight:700; letter-spacing:.12em; text-transform:uppercase; color:var(--color-gray-2); margin-bottom:1rem; }
.map-section { padding:1.25rem; }
.photos-section { padding:1.25rem; }
.photos-grid { display:flex; gap:.5rem; flex-wrap:wrap; }
.photo-thumb { width:100px; height:100px; border-radius:6px; object-fit:cover; cursor:pointer; transition:opacity .15s; }
.photo-thumb:hover { opacity:.85; }
.note-section { padding:1.25rem; }
.note-text { color:var(--color-gray-1); line-height:1.9; white-space:pre-wrap; }

/* Share */
.share-section { padding:1.25rem; }
.share-btns { display:flex; gap:.5rem; flex-wrap:wrap; }
.share-btn { padding:.5rem 1.25rem; border-radius:4px; border:1px solid var(--color-border); font-size:.85rem; font-weight:600; cursor:pointer; background:none; color:var(--color-gray-1); transition:all .15s; text-decoration:none; display:inline-block; }
.share-btn:hover { border-color:var(--color-primary); color:var(--color-primary); }
.share-btn.line { border-color:#06c755; color:#06c755; }
.share-btn.fb   { border-color:#1877f2; color:#1877f2; }
.share-btn.ig   { border-color:#e1306c; color:#e1306c; }
.copied-hint { font-size:.78rem; color:#22c55e; margin-top:.5rem; }

/* Lightbox */
.lightbox { position:fixed; inset:0; background:rgba(0,0,0,.9); z-index:300; display:flex; align-items:center; justify-content:center; cursor:pointer; }
.lightbox img { max-width:90vw; max-height:90vh; border-radius:4px; }
</style>
