<template>
  <div class="home">
    <!-- ── Navbar ─────────────────────────────────────── -->

    <!-- ── Hero ──────────────────────────────────────── -->
    <section class="hero">
      <div class="hero-bg">
        <div class="hero-grid"></div>
        <div class="hero-glow"></div>
      </div>
      <div class="container hero-content">
        <!-- Dynamic Banner from Site Settings -->
        <div v-if="site.get('banner_visible')=='1' && site.get('banner_image')" class="anniversary-banner">
          <a :href="site.get('banner_link') || undefined" :target="site.get('banner_link') ? '_blank' : undefined">
            <img :src="site.get('banner_image')" :alt="site.get('banner_text') || 'Banner'" class="anniversary-img" />
          </a>
          <p v-if="site.get('banner_text')" class="banner-caption">{{ site.get('banner_text') }}</p>
        </div>
        <div class="hero-tag">TRIATHLON · RUNNING · BIKING · SWIMMING</div>
        <h1 class="hero-title">
          <span class="tr">TR</span><span class="bb">BB</span>
          <br />
          <span class="hero-subtitle">鐵人三項運動社團</span>
        </h1>
        <p class="hero-desc">挑戰極限，超越自我。每一滴汗水，都是向終點線的宣言。</p>
        <div class="hero-cta">
          <RouterLink to="/events" class="btn btn-primary btn-lg">查看賽事報名</RouterLink>
          <RouterLink to="/register" class="btn btn-ghost btn-lg">加入我們</RouterLink>
        </div>
        <div class="hero-stats">
          <div class="stat"><span class="stat-num">500+</span><span class="stat-label">活躍會員</span></div>
          <div class="stat-div"></div>
          <div class="stat"><span class="stat-num">120+</span><span class="stat-label">年度賽事</span></div>
          <div class="stat-div"></div>
          <div class="stat"><span class="stat-num">10+</span><span class="stat-label">年資歷</span></div>
        </div>
      </div>
      <div class="hero-scroll-hint">
        <span></span>SCROLL
      </div>
    </section>

    <!-- ── Latest Events（API 載入） ──────────────────── -->
    <section class="latest-events">
      <div class="container">
        <div class="flex justify-between items-center mb-2">
          <h2 class="section-title">最新賽事</h2>
          <RouterLink to="/events" class="btn btn-ghost">查看全部</RouterLink>
        </div>

        <!-- Loading -->
        <div v-if="eventsLoading" class="events-loading">
          <div class="loading-spinner"></div>
          <span>載入賽事中...</span>
        </div>

        <!-- Empty -->
        <div v-else-if="!latestEvents.length" class="events-empty">
          <p>目前沒有公開賽事，敬請期待。</p>
        </div>

        <!-- Cards -->
        <div v-else class="events-grid">
          <RouterLink v-for="ev in latestEvents" :key="ev.id"
            :to="`/events/${ev.id}`" class="event-card card">
            <div class="event-cover" :class="!ev.cover_url ? 'event-cover-placeholder' : ''"
              :style="ev.cover_url ? `background-image:url(${ev.cover_url})` : ''">
              <span v-if="!ev.cover_url" class="event-cover-icon">{{ eventTypeIcon(ev.event_type) }}</span>
              <span class="event-type-badge">{{ eventTypeLabel(ev.event_type) }}</span>
            </div>
            <div class="event-body">
              <div class="event-date text-red text-sm uppercase">
                {{ formatDate(ev.start_at) }}
              </div>
              <h3 class="event-title">{{ ev.title }}</h3>
              <div class="event-meta text-gray text-sm">
                <span>📍 {{ ev.location }}</span>
                <span v-if="ev.fee > 0">💰 NT$ {{ Number(ev.fee).toLocaleString() }}</span>
                <span v-else>💰 免費</span>
              </div>
              <div class="event-reg-status" :class="regStatusClass(ev)">
                {{ regStatusText(ev) }}
              </div>
              <div class="btn btn-primary mt-2" style="width:100%;text-align:center">
                {{ isRegOpen(ev) ? '立即報名' : '查看詳情' }}
              </div>
            </div>
          </RouterLink>
        </div>
      </div>
    </section>

    <!-- ── Latest Products ────────────────────────────── -->
    <section class="latest-products">
      <div class="container">
        <div class="flex justify-between items-center mb-2">
          <h2 class="section-title">近期商品</h2>
          <RouterLink to="/shop" class="btn btn-ghost">查看全部</RouterLink>
        </div>
        <div v-if="productsLoading" class="loading-box"><div class="loading-spinner"></div></div>
        <div v-else-if="!latestProducts.length" class="empty-box"><p>目前無商品</p></div>
        <div v-else class="products-mini-grid">
          <RouterLink v-for="p in latestProducts" :key="p.id"
            :to="`/shop/${p.id}`" class="product-mini-card card">
            <div class="pmc-img">
              <img v-if="p.images && p.images.length" :src="imgUrl(p.images[0])" :alt="p.title" />
              <span v-else class="pmc-placeholder">🛍</span>
            </div>
            <div class="pmc-body">
              <div class="pmc-title">{{ p.title }}</div>
              <div class="pmc-price">NT$ {{ Number(p.price).toLocaleString() }}</div>
            </div>
          </RouterLink>
        </div>
      </div>
    </section>
    <!-- ── Public Training Feed ─────────────────────── -->
    <section class="training-feed">
      <div class="container">
        <div class="flex justify-between items-center mb-2">
          <h2 class="section-title">會員訓練動態</h2>
          <RouterLink to="/training" class="btn btn-ghost">查看全部</RouterLink>
        </div>

        <div v-if="trainingLoading" class="events-loading">
          <div class="loading-spinner"></div>
        </div>
        <div v-else-if="!trainingLogs.length" class="events-empty">
          <p>目前沒有公開的訓練記錄</p>
        </div>
        <div v-else class="training-grid">
          <RouterLink v-for="log in trainingLogs" :key="log.id"
            :to="`/training/${log.id}`" class="training-card card">
            <!-- 封面照片 or 地圖佔位 -->
            <div class="training-cover" :class="`sport-${log.sport_type}`">
              <img v-if="log.photos && log.photos.length"
                :src="imgUrl(log.photos[0])" :alt="log.title" class="cover-img" />
              <div v-else class="cover-icon">{{ sportIcon(log.sport_type) }}</div>
              <span class="sport-badge">{{ sportLabel(log.sport_type) }}</span>
            </div>
            <div class="training-body">
              <div class="training-meta text-gray text-sm">
                <span>{{ log.display_name || log.username }}</span>
                <span>{{ log.date }}</span>
              </div>
              <h3 class="training-title">{{ log.title }}</h3>
              <div class="training-stats">
                <span v-if="log.distance_km">📏 {{ Number(log.distance_km).toFixed(2) }} km</span>
                <span v-if="log.duration_min">⏱ {{ fmtDuration(log.duration_min) }}</span>
                <span v-if="log.avg_heart_rate">❤️ {{ log.avg_heart_rate }} bpm</span>
              </div>
            </div>
          </RouterLink>
        </div>
      </div>
    </section>
    <!-- ── Features ──────────────────────────────────── -->
    <section class="features">
      <div class="container">
        <h2 class="section-title text-center">社團服務</h2>
        <div class="features-grid">
          <div class="feature-card card" v-for="f in features" :key="f.title">
            <div class="feature-icon">{{ f.icon }}</div>
            <h3>{{ f.title }}</h3>
            <p>{{ f.desc }}</p>
          </div>
        </div>
      </div>
    </section>
    <!-- ── Garmin Banner ─────────────────────────────── -->
    <section class="garmin-banner">
      <div class="container garmin-inner">
        <div class="garmin-text">
          <h2>串接 Garmin Connect</h2>
          <p>自動同步訓練數據，記錄每一次游泳、騎車、跑步的進步軌跡。</p>
          <RouterLink to="/me/garmin" class="btn btn-primary">立即連結</RouterLink>
        </div>
        <div class="garmin-visual">
          <div class="garmin-ring">
            <div class="ring swim">游</div>
            <div class="ring bike">騎</div>
            <div class="ring run">跑</div>
          </div>
        </div>
      </div>
    </section>

    <!-- ── Footer ────────────────────────────────────── -->
    <footer class="footer">
      <div class="container footer-inner">
        <div class="footer-brand">
          <span class="trbb-logo footer-logo"><span class="tr">TR</span><span class="bb">BB</span></span>
          <p>鐵人三項運動社團</p>
          <p class="text-gray text-sm">© 2025 TRBB. All rights reserved.</p>
        </div>
        <div class="footer-links">
          <div class="footer-col">
            <h4>社團</h4>
            <ul>
              <li><RouterLink to="/events">賽事報名</RouterLink></li>
              <li><RouterLink to="/shop">商品販售</RouterLink></li>
              <li><RouterLink to="/secondhand">二手交換</RouterLink></li>
            </ul>
          </div>
          <div class="footer-col">
            <h4>會員</h4>
            <ul>
              <li><RouterLink to="/register">加入會員</RouterLink></li>
              <li><RouterLink to="/me/training">訓練日記</RouterLink></li>
              <li><RouterLink to="/me/garmin">Garmin 串接</RouterLink></li>
            </ul>
          </div>
          <div class="footer-col">
            <h4>聯絡</h4>
            <ul>
              <li><a href="mailto:info@trbbtw.com">info@trbbtw.com</a></li>
              <li><a href="#">Facebook</a></li>
              <li><a href="#">Instagram</a></li>
            </ul>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSiteSettingsStore } from '@/stores/siteSettings'
import api from '@/services/api'

const auth = useAuthStore()
const site = useSiteSettingsStore()
const latestEvents  = ref([])
const eventsLoading  = ref(true)
const trainingLogs    = ref([])
const trainingLoading  = ref(true)
const latestProducts  = ref([])
const productsLoading = ref(true)

// 載入最新 3 筆已發布賽事
async function fetchLatestEvents() {
  eventsLoading.value = true
  try {
    const { data } = await api.get('/events', { params: { page: 1, page_size: 3 } })
    latestEvents.value = data.events || []
  } catch(e) {
    console.error('fetchLatestEvents error', e)
    latestEvents.value = []
  } finally {
    eventsLoading.value = false
  }
}

async function fetchProducts() {
  productsLoading.value = true
  try {
    const { data } = await api.get('/products', { params: { page: 1, page_size: 4 } })
    latestProducts.value = data.products || []
  } catch(e) {
    console.error('fetchProducts error', e)
  } finally {
    productsLoading.value = false
  }
}

async function fetchTrainingLogs() {
  trainingLoading.value = true
  try {
    const { data } = await api.get('/training', { params: { page: 1, page_size: 6 } })
    trainingLogs.value = data.logs || []
  } catch(e) {
    console.error('fetchTrainingLogs error', e)
    trainingLogs.value = []
  } finally {
    trainingLoading.value = false
  }
}

// ── Helpers ───────────────────────────────────────────────
const typeMap  = { 1:'鐵人三項', 2:'路跑', 3:'游泳', 4:'單車', 5:'訓練', 6:'其他' }
const typeIcon = { 1:'🏊', 2:'🏃', 3:'🏊', 4:'🚴', 5:'💪', 6:'🏅' }
function eventTypeLabel(t) { return typeMap[t] || '其他' }
function eventTypeIcon(t)  { return typeIcon[t] || '🏅' }

function formatDate(d) {
  return new Date(d).toLocaleDateString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit' })
}

function imgUrl(path) {
  const base = import.meta.env.VITE_IMAGE_BASE_URL || ''
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `${base}/images/${path}`
}
function sportIcon(t) { return { 1:'🏃', 2:'🏊', 3:'🚴', 4:'🏅', 5:'💪', 6:'🏋️' }[t] || '🏅' }
function sportLabel(t) { return { 1:'路跑', 2:'游泳', 3:'單車', 4:'鐵人', 5:'重訓', 6:'其他' }[t] || '其他' }
function fmtDuration(min) {
  if (!min) return ''
  const h = Math.floor(min / 60), m = min % 60
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

function isRegOpen(ev) {
  const now = new Date()
  return now >= new Date(ev.reg_start_at) && now <= new Date(ev.reg_end_at)
}

function regStatusClass(ev) {
  const now = new Date()
  if (now < new Date(ev.reg_start_at)) return 'reg-upcoming'
  if (now > new Date(ev.reg_end_at))   return 'reg-closed'
  if (ev.max_participants && ev.registered_count >= ev.max_participants) return 'reg-full'
  return 'reg-open'
}

function regStatusText(ev) {
  const cls = regStatusClass(ev)
  return { 'reg-open':'🟢 報名中', 'reg-upcoming':'🟡 即將開放', 'reg-closed':'🔴 已截止', 'reg-full':'🔴 額滿' }[cls]
}

const features = [
  { icon: '🏊', title: '賽事報名', desc: '鐵人三項、路跑、游泳、單車全系列賽事一站式報名，支援個人及團體。' },
  { icon: '🛒', title: '裝備商城', desc: '官方認證裝備、訓練補給，會員享專屬折扣，到貨快速安心。' },
  { icon: '🔄', title: '二手交換', desc: '讓好裝備流轉，以合理價格找到需要的人，減少浪費。' },
  { icon: '📔', title: '訓練日記', desc: '記錄每日訓練，串接 Garmin 自動同步，分析進步曲線。' },
  { icon: '🚌', title: '賽事交通', desc: '團體包車服務，從集合點直送賽場，輕鬆無壓力。' },
  { icon: '🏅', title: '會員制度', desc: '分級會員系統，解鎖專屬訓練計畫、折扣及優先報名資格。' },
]

onMounted(() => { fetchLatestEvents(); fetchTrainingLogs(); fetchProducts() })
</script>

<style scoped>
/* ── Navbar ──────────────────────────────────────────────── */

/* ── Hero ────────────────────────────────────────────────── */
.hero { position: relative; min-height: 100vh; display: flex; align-items: center; padding-top: 64px; }
.hero-bg { position: absolute; inset: 0; background: var(--color-navy); }
.hero-grid { position: absolute; inset: 0; background-image: linear-gradient(rgba(229,25,26,0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(229,25,26,0.05) 1px, transparent 1px); background-size: 60px 60px; }
.hero-glow { position: absolute; top: 30%; left: 50%; transform: translate(-50%,-50%); width: 600px; height: 600px; border-radius: 50%; background: radial-gradient(circle, rgba(229,25,26,0.12) 0%, transparent 70%); }
.hero-content { position: relative; z-index: 1; padding: 4rem 0; }
.hero-tag { font-family: var(--font-cond); font-size: 0.75rem; font-weight: 600; letter-spacing: 0.25em; color: var(--color-primary); margin-bottom: 1.5rem; text-transform: uppercase; }
.hero-title { font-family: var(--font-display); font-size: clamp(5rem, 15vw, 12rem); line-height: 0.9; margin-bottom: 0.5rem; }
.hero-title .tr { color: #E0E3DA; }
.hero-title .bb { color: #C9A84C; }
.hero-subtitle { font-family: var(--font-cond); font-size: clamp(1.2rem, 3vw, 2rem); font-weight: 300; letter-spacing: 0.3em; color: var(--color-gray-1); display: block; }
.hero-desc { max-width: 520px; color: var(--color-gray-1); font-size: 1.05rem; margin: 1.5rem 0 2rem; line-height: 1.8; }
.hero-cta { display: flex; gap: 1rem; flex-wrap: wrap; margin-bottom: 3rem; }
.btn-lg { padding: 0.9rem 2.4rem; font-size: 1rem; }
.hero-stats { display: flex; align-items: center; gap: 2rem; }
.stat { display: flex; flex-direction: column; }
.stat-num { font-family: var(--font-display); font-size: 2.5rem; color: var(--color-primary); }
.stat-label { font-family: var(--font-cond); font-size: 0.75rem; letter-spacing: 0.15em; color: var(--color-gray-2); text-transform: uppercase; }
.stat-div { width: 1px; height: 40px; background: var(--color-border); }
.hero-scroll-hint { position: absolute; bottom: 2rem; left: 50%; transform: translateX(-50%); display: flex; flex-direction: column; align-items: center; gap: 0.5rem; font-family: var(--font-cond); font-size: 0.7rem; letter-spacing: 0.2em; color: var(--color-gray-3); }
.hero-scroll-hint span { width: 1px; height: 40px; background: linear-gradient(to bottom, transparent, var(--color-primary)); animation: scrollLine 2s ease-in-out infinite; }
@keyframes scrollLine { 0%,100% { opacity: 0.3; } 50% { opacity: 1; } }

/* Anniversary banner */
.anniversary-banner { margin-bottom: 1.5rem; }
.anniversary-img {
  max-height: 120px; width: auto;
  filter: drop-shadow(0 4px 16px rgba(0,0,0,0.5));
}
.banner-caption { font-size:.85rem; color:rgba(255,255,255,.8); margin-top:.35rem; }

/* ── Features ────────────────────────────────────────────── */
.features { padding: 6rem 0; background: var(--color-bg-2); }
.text-center { text-align: center; }
.section-title.text-center::after { left: 50%; transform: translateX(-50%); }
.features-grid { display: grid; gap: 1.5rem; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); }
.feature-card { padding: 2rem; }
.feature-icon { font-size: 2.5rem; margin-bottom: 1rem; }
.feature-card h3 { font-size: 1.1rem; margin-bottom: 0.5rem; }
.feature-card p { color: var(--color-gray-2); font-size: 0.9rem; line-height: 1.7; }

/* ── Latest Events ───────────────────────────────────────── */
.latest-events { padding: 6rem 0; }
.events-loading { display: flex; align-items: center; justify-content: center; gap: 1rem; padding: 3rem; color: var(--color-gray-2); }
.loading-spinner { width: 24px; height: 24px; border: 2px solid var(--color-border); border-top-color: var(--color-primary); border-radius: 50%; animation: spin .7s linear infinite; }
@keyframes spin { to { transform: rotate(360deg) } }
.events-empty { text-align: center; padding: 3rem; color: var(--color-gray-2); }
.events-grid { display: grid; gap: 1.5rem; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); }

.event-card { background: var(--color-bg-card); border: 1px solid var(--color-border); border-radius: 8px; overflow: hidden; transition: all .2s; display: block; }
.event-card:hover { border-color: var(--color-primary); transform: translateY(-3px); box-shadow: 0 8px 30px rgba(229,25,26,.15); }

.event-cover { height: 160px; display: flex; align-items: flex-end; padding: 0.75rem; position: relative; background-size: cover; background-position: center; }
.event-cover-placeholder { background: linear-gradient(135deg, #1a1a2e, #16213e); }
.event-cover-icon { position: absolute; top: 50%; left: 50%; transform: translate(-50%, -60%); font-size: 3rem; }
.event-type-badge { font-family: var(--font-cond); font-size: 0.7rem; font-weight: 700; letter-spacing: 0.1em; text-transform: uppercase; padding: 0.2rem 0.65rem; background: rgba(229,25,26,0.85); color: #fff; border-radius: 3px; position: relative; z-index: 1; }

.event-body { padding: 1.25rem; }
.event-date { font-size: 0.78rem; margin-bottom: 0.25rem; }
.event-title { font-size: 1rem; font-weight: 700; line-height: 1.4; margin-bottom: 0.5rem; }
.event-meta { display: flex; gap: 1rem; flex-wrap: wrap; font-size: 0.82rem; margin-bottom: 0.5rem; }

.event-reg-status { font-size: 0.78rem; font-weight: 600; }
.reg-open     { color: #22c55e; }
.reg-upcoming { color: #f59e0b; }
.reg-closed   { color: #9ca3af; }
.reg-full     { color: #ef4444; }

/* ── Garmin Banner ───────────────────────────────────────── */
.garmin-banner { background: linear-gradient(135deg, #0d0d0d 0%, #1a0000 100%); border-top: 1px solid var(--color-border); border-bottom: 1px solid var(--color-border); padding: 5rem 0; }
.garmin-inner { display: flex; gap: 4rem; align-items: center; flex-wrap: wrap; }
.garmin-text { flex: 1; min-width: 280px; }
.garmin-text h2 { font-size: 2.2rem; margin-bottom: 1rem; }
.garmin-text p { color: var(--color-gray-1); line-height: 1.8; margin-bottom: 2rem; }
.garmin-visual { flex: 0 0 260px; display: flex; justify-content: center; }
.garmin-ring { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.ring { width: 100px; height: 100px; border-radius: 50%; border: 3px solid var(--color-primary); display: flex; align-items: center; justify-content: center; font-family: var(--font-display); font-size: 2rem; color: var(--color-primary); animation: pulse-red 3s infinite; }
.ring.bike { animation-delay: 1s; }
.ring.run  { animation-delay: 2s; grid-column: 1 / -1; justify-self: center; }
@keyframes pulse-red { 0%,100% { box-shadow: 0 0 0 0 rgba(229,25,26,.4); } 50% { box-shadow: 0 0 0 8px rgba(229,25,26,0); } }

/* ── Footer ──────────────────────────────────────────────── */
.footer { background: var(--color-bg-2); border-top: 1px solid var(--color-border); padding: 4rem 0 2rem; }
.footer-inner { display: flex; gap: 4rem; flex-wrap: wrap; }
.footer-brand { flex: 1; min-width: 200px; }
.footer-logo { font-size: 2.5rem; }
.footer-brand p { color: var(--color-gray-2); margin-top: 0.5rem; }
.footer-links { display: flex; gap: 4rem; flex-wrap: wrap; }
.footer-col h4 { font-family: var(--font-cond); font-size: 0.8rem; letter-spacing: 0.15em; text-transform: uppercase; color: var(--color-gray-2); margin-bottom: 1rem; }
.footer-col ul { list-style: none; }
.footer-col li { margin-bottom: 0.5rem; }
.footer-col a { color: var(--color-gray-1); font-size: 0.9rem; }
.footer-col a:hover { color: var(--color-primary); }
.training-feed { padding:4rem 0; background:var(--color-bg-2); }
/* Training grid */
.training-grid { display:grid; gap:1.25rem; grid-template-columns:repeat(auto-fill,minmax(280px,1fr)); }
.training-card { display:block; background:var(--color-bg-card); border:1px solid var(--color-border); border-radius:8px; overflow:hidden; transition:all .2s; }
.training-card:hover { border-color:var(--color-primary); transform:translateY(-3px); box-shadow:0 6px 24px rgba(229,25,26,.12); }
.training-cover { height:140px; display:flex; align-items:center; justify-content:center; position:relative; }
.sport-1 { background:linear-gradient(135deg,#1a1a2e,#2d1b4e); }
.sport-2 { background:linear-gradient(135deg,#0a1628,#0e3460); }
.sport-3 { background:linear-gradient(135deg,#1a1a1a,#2d2000); }
.sport-4 { background:linear-gradient(135deg,#0d1a0d,#1a3a1a); }
.sport-5, .sport-6 { background:linear-gradient(135deg,#1a0d0d,#3a1a1a); }
.cover-img { width:100%; height:100%; object-fit:cover; position:absolute; inset:0; }
.cover-icon { font-size:3rem; position:relative; z-index:1; }
.sport-badge { position:absolute; bottom:.5rem; left:.5rem; font-family:var(--font-cond); font-size:.65rem; font-weight:700; letter-spacing:.1em; text-transform:uppercase; padding:.15rem .5rem; background:rgba(0,0,0,.6); border-radius:3px; color:#fff; }
.training-body { padding:1rem; }
.training-meta { display:flex; justify-content:space-between; margin-bottom:.35rem; }
.training-title { font-size:.95rem; font-weight:700; line-height:1.4; margin-bottom:.4rem; }
.training-stats { display:flex; gap:.75rem; flex-wrap:wrap; font-size:.78rem; color:var(--color-gray-2); }

/* ── Latest Products ──────────────────────────────────────── */
.latest-products { padding:3.5rem 0; background:var(--color-bg-card); }
.products-mini-grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(220px,1fr)); gap:1.25rem; }
.product-mini-card { display:block; background:var(--color-bg-card); border:1px solid var(--color-border); border-radius:8px; overflow:hidden; box-shadow:0 1px 4px rgba(86,98,112,.08); transition:all .2s; }
.product-mini-card:hover { border-color:var(--color-primary); transform:translateY(-3px); box-shadow:0 4px 16px rgba(207,32,39,.12); }
.pmc-img { height:160px; overflow:hidden; background:var(--color-bg-2); display:flex; align-items:center; justify-content:center; }
.pmc-img img { width:100%; height:100%; object-fit:cover; }
.pmc-placeholder { font-size:3rem; }
.pmc-body { padding:.9rem 1rem; }
.pmc-title { font-weight:700; font-size:.9rem; margin-bottom:.3rem; color:var(--color-text); }
.pmc-price { font-family:var(--font-cond,'Barlow Condensed',sans-serif); font-size:1.1rem; font-weight:700; color:var(--color-primary); }

</style>
