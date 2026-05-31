<template>
  <div class="strava-page">
    <h1 class="section-title">Strava 串接</h1>
    <p class="text-gray" style="margin-bottom:2rem">
      連結您的 Strava 帳號，自動將訓練活動匯入訓練日記。
    </p>

    <!-- 連結狀態卡 -->
    <div class="status-card card">
      <div class="status-left">
        <div class="strava-logo">🟠</div>
        <div class="status-info">
          <div class="status-title">
            <span class="dot" :class="dotClass"></span>
            {{ statusTitle }}
          </div>
          <div class="text-gray">{{ statusSub }}</div>
          <div v-if="tokenInfo" class="token-meta text-gray">
            <span v-if="tokenInfo.athlete_name">運動員：{{ tokenInfo.athlete_name }}</span>
            <span v-if="tokenInfo.last_sync_at">　上次同步：{{ fmt(tokenInfo.last_sync_at) }}</span>
          </div>
        </div>
      </div>
      <div class="status-actions">
        <template v-if="stravaStatus === 'connected'">
          <div class="sync-options">
            <!-- 自動同步 -->
            <label class="pref-row">
              <span class="pref-label">Webhook 自動同步</span>
              <button class="pref-toggle" :class="{ on: prefs.auto_sync }"
                @click="togglePref('auto_sync')" type="button">
                {{ prefs.auto_sync ? '🟢 開啟' : '⭕ 關閉' }}
              </button>
            </label>
            <!-- 同步後公開 -->
            <label class="pref-row">
              <span class="pref-label">同步後設為</span>
              <button class="pref-toggle public-tog" :class="{ public: prefs.sync_public }"
                @click="togglePref('sync_public')" type="button">
                {{ prefs.sync_public ? '🌐 公開' : '🔒 私人' }}
              </button>
            </label>
            <div class="pref-hint" v-if="prefs.auto_sync">
              ✓ 每次在 Strava 完成活動後，系統將自動匯入
            </div>
            <div class="sync-btns">
              <button class="btn btn-primary btn-sm" @click="syncNow" :disabled="syncing">
                <span v-if="syncing" class="spinner"></span>
                {{ syncing ? '同步中...' : '🔄 手動同步近 90 天' }}
              </button>
              <button class="btn btn-ghost btn-sm" @click="confirmDisconnect">解除連結</button>
            </div>
          </div>
        </template>
        <template v-else-if="stravaStatus === 'disconnected'">
          <button class="btn btn-strava" @click="connectStrava" :disabled="connecting">
            <span v-if="connecting" class="spinner"></span>
            {{ connecting ? '連接中...' : '連結 Strava' }}
          </button>
        </template>
        <template v-else-if="stravaStatus === 'unavailable'">
          <div class="unavail-badge">Strava API 憑證未設定</div>
        </template>
      </div>
    </div>

    <!-- 同步結果 -->
    <div v-if="syncResult" class="sync-result" :class="syncResult.ok ? 'ok' : 'err'">
      <template v-if="syncResult.ok">
        <span v-if="syncResult.count > 0">
          ✓ 已匯入 {{ syncResult.count }} 筆訓練活動
          <span v-if="syncResult.public">（{{ syncResult.public }} 筆已設為公開）</span>
        </span>
        <span v-else>{{ syncResult.msg || '✓ 無新活動' }}</span>
      </template>
      <template v-else>✗ {{ syncResult.msg }}</template>
    </div>

    <!-- 說明卡片 -->
    <div class="info-cards">
      <div class="info-card card">
        <div class="info-icon">🔐</div>
        <h4>OAuth 2.0 授權</h4>
        <p>透過 Strava 官方 OAuth 2.0，TRBB 僅讀取您的活動資料。</p>
      </div>
      <div class="info-card card">
        <div class="info-icon">📊</div>
        <h4>同步的資料</h4>
        <p>距離、時間、配速、心率、卡路里、爬升、GPS 路線，支援跑步、游泳、騎車、鐵人三項等。</p>
      </div>
      <div class="info-card card">
        <div class="info-icon">🔔</div>
        <h4>Webhook 自動同步</h4>
        <p>開啟後，每次在 Strava 完成活動即自動匯入，無需手動操作。</p>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import api from '@/services/api'

const stravaStatus = ref('loading')
const tokenInfo   = ref(null)
const prefs       = ref({ sync_public: false, auto_sync: false })
const prefSaving  = ref(false)
const connecting   = ref(false)
const syncing      = ref(false)
const syncResult   = ref(null)
const apiConfigured = ref(false)

const compareRows = [
  { label: '申請難度',   strava: '簡單（個人即可）', garmin: '需企業申請' },
  { label: '授權協議',   strava: 'OAuth 2.0',         garmin: 'OAuth 1.0a' },
  { label: '資料項目',   strava: '全方位',             garmin: '全方位' },
  { label: 'GPS 路線',   strava: '✓ Polyline',         garmin: '✓ FIT/GPX' },
  { label: '心率資料',   strava: '✓',                  garmin: '✓' },
  { label: '裝置支援',   strava: '多廠牌都可推送',     garmin: '僅限 Garmin 裝置' },
]

const dotClass = computed(() => ({
  loading:     '',
  connected:   'connected',
  disconnected:'disconnected',
  unavailable: 'unavailable',
}[stravaStatus.value] || ''))

const statusTitle = computed(() => ({
  loading:     '檢查連結狀態中...',
  connected:   'Strava 已連結',
  disconnected:'Strava 尚未連結',
  unavailable: 'Strava API 設定中',
}[stravaStatus.value] || ''))

const statusSub = computed(() => ({
  loading:     '',
  connected:   '訓練活動將同步至日記',
  disconnected:'連結後可自動匯入 Strava 訓練資料',
  unavailable: '管理員尚未設定 Strava API 憑證，串接功能暫時無法使用',
}[stravaStatus.value] || ''))

async function checkStatus() {
  try {
    const { data } = await api.get('/me/strava/status')
    if (data.connected) {
      stravaStatus.value = 'connected'
      tokenInfo.value = data
      prefs.value = {
        sync_public: data.sync_public || false,
        auto_sync:   data.auto_sync   || false,
      }
    } else {
      apiConfigured.value = data.api_configured !== false
      stravaStatus.value = data.api_configured === false ? 'unavailable' : 'disconnected'
    }
  } catch {
    stravaStatus.value = 'disconnected'
  }
}

async function togglePref(key) {
  prefs.value[key] = !prefs.value[key]
  try {
    await api.patch('/me/strava/sync_prefs', {
      sync_public: prefs.value.sync_public,
      auto_sync:   prefs.value.auto_sync,
    })
  } catch { prefs.value[key] = !prefs.value[key] } // revert on error
}

async function connectStrava() {
  connecting.value = true
  try {
    const { data } = await api.get('/me/strava/connect')
    if (data.auth_url) {
      window.location.href = data.auth_url
    } else {
      stravaStatus.value = 'unavailable'
    }
  } catch(e) {
    alert(e.response?.data?.error || '連結失敗')
  } finally {
    connecting.value = false
  }
}

async function syncNow() {
  syncing.value = true
  syncResult.value = null
  try {
    // 1. 同步（後端預設存為私人）
    const { data } = await api.post('/training/strava/sync')
    const count = data.synced || 0

    // 2. 如果設定公開，用後端直接 batch update（不受分頁限制）
    if (prefs.value.sync_public && count > 0) {
      // 取所有 strava 來源的私人記錄，不限頁數
      let allStrava = []
      let pg = 1
      while (true) {
        const { data: pg_data } = await api.get('/me/training', {
          params: { page: pg, page_size: 200 }
        })
        const strava_private = (pg_data.logs || [])
          .filter(l => l.source === 'strava' && !l.is_public)
          .map(l => l.id)
        allStrava = allStrava.concat(strava_private)
        if (!pg_data.logs || pg_data.logs.length < 200 || pg >= pg_data.pages) break
        pg++
      }
      if (allStrava.length) {
        await Promise.all(allStrava.map(id =>
          api.patch(`/training/${id}/public`, { is_public: true })
        ))
      }
      syncResult.value = { ok: true, count, public: allStrava.length }
    } else {
      syncResult.value = { ok: true, count }
    }
    await checkStatus()
  } catch(e) {
    syncResult.value = { ok: false, msg: e.response?.data?.error || '同步失敗' }
  } finally {
    syncing.value = false
  }
}

function confirmDisconnect() {
  if (!confirm('確認解除 Strava 連結？\n已匯入的訓練資料不會刪除。')) return
  api.delete('/me/strava/disconnect').then(() => {
    stravaStatus.value = 'disconnected'
    tokenInfo.value = null
  }).catch(() => alert('解除失敗'))
}

function fmt(d) {
  return d ? new Date(d).toLocaleString('zh-TW', {
    year:'numeric', month:'2-digit', day:'2-digit',
    hour:'2-digit', minute:'2-digit',
  }) : '-'
}

onMounted(() => {
  // 處理 OAuth callback 結果
  const params = new URLSearchParams(window.location.search)
  if (params.get('success') === '1') {
    syncResult.value = { ok: true, count: 0, msg: '✓ 已成功連結 Strava！正在匯入近 30 天訓練資料...' }
    // Clean URL
    window.history.replaceState({}, '', '/me/strava')
  } else if (params.get('error')) {
    syncResult.value = { ok: false, msg: `連結失敗：${params.get('error')}` }
    window.history.replaceState({}, '', '/me/strava')
  }
  checkStatus()
})
</script>

<style scoped>
.strava-page { max-width:700px; }

.status-card { padding:1.5rem; display:flex; justify-content:space-between; align-items:center; gap:1.5rem; flex-wrap:wrap; margin-bottom:1rem; }
.status-left { display:flex; align-items:center; gap:1.25rem; }
.strava-logo { font-size:2.5rem; flex-shrink:0; }
.status-title { font-size:1.05rem; font-weight:700; display:flex; align-items:center; gap:.5rem; margin-bottom:.25rem; }
.dot { width:10px; height:10px; border-radius:50%; background:var(--color-gray-3); display:inline-block; }
.dot.connected    { background:#fc4c02; box-shadow:0 0 6px rgba(252,76,2,.5); }
.dot.disconnected { background:#9ca3af; }
.dot.unavailable  { background:#f59e0b; }
.token-meta { font-size:.78rem; margin-top:.35rem; display:flex; flex-wrap:wrap; gap:.75rem; }
.status-actions { display:flex; flex-direction:column; gap:.5rem; align-items:flex-end; }
.btn-strava { background:#fc4c02; color:#fff; box-shadow:0 2px 8px rgba(252,76,2,.3); }
.btn-strava:hover { background:#e04402; }
.unavail-badge { font-size:.75rem; padding:.3rem .8rem; background:rgba(245,158,11,.1); color:#f59e0b; border:1px solid rgba(245,158,11,.3); border-radius:4px; }
.btn-sm { padding:.4rem 1rem; font-size:.82rem; }

.sync-result { padding:.65rem 1rem; border-radius:6px; font-size:.88rem; margin-bottom:1rem; }
.sync-result.ok  { background:rgba(34,197,94,.1); border:1px solid rgba(34,197,94,.3); color:#86efac; }
.sync-result.err { background:rgba(239,68,68,.1); border:1px solid rgba(239,68,68,.3); color:#fca5a5; }

.info-cards { display:grid; grid-template-columns:repeat(auto-fill,minmax(200px,1fr)); gap:1rem; margin-bottom:1.25rem; }
.info-card { padding:1.25rem; }
.info-icon { font-size:1.8rem; margin-bottom:.6rem; }
.info-card h4 { font-size:.95rem; margin-bottom:.35rem; }
.info-card p { font-size:.82rem; color:var(--color-gray-2); line-height:1.6; }

.apply-section { padding:1.5rem; margin-bottom:1.25rem; }
.apply-title { font-size:1rem; font-weight:700; margin-bottom:1.25rem; }
.apply-steps { display:flex; flex-direction:column; gap:.85rem; }
.apply-step { display:flex; align-items:flex-start; gap:.9rem; font-size:.88rem; }
.step-icon { font-size:1.1rem; flex-shrink:0; margin-top:.05rem; }
.apply-step.pending { opacity:.7; }
.apply-step.disabled { opacity:.4; }

.compare-card { padding:1.5rem; margin-bottom:1.25rem; }
.compare-card h3 { font-size:1rem; margin-bottom:1rem; }
.compare-table { font-size:.85rem; }
.compare-header { display:grid; grid-template-columns:1fr 1.2fr 1.2fr; font-weight:700; padding:.4rem 0; border-bottom:2px solid var(--color-border); }
.compare-header div:nth-child(2) { color:#fc4c02; }
.compare-header div:nth-child(3) { color:#1A3A7A; }
.compare-row { display:grid; grid-template-columns:1fr 1.2fr 1.2fr; padding:.5rem 0; border-bottom:1px solid var(--color-border); }
.compare-label { color:var(--color-gray-2); }
.compare-val { }

.alt-section { padding:1.5rem; }
.alt-section h3 { font-size:1rem; margin-bottom:.5rem; }
.spinner { width:14px; height:14px; border:2px solid rgba(255,255,255,.3); border-top-color:#fff; border-radius:50%; animation:spin .7s linear infinite; display:inline-block; }
@keyframes spin { to { transform:rotate(360deg) } }
.sync-public-hint { font-size:.85rem; color:var(--color-gray-1); margin-bottom:1rem; padding:.6rem 1rem; background:var(--color-bg-2,#f5f5e8); border-radius:6px; border:1px solid var(--color-border); }
.sync-options { display:flex; flex-direction:column; gap:.5rem; align-items:flex-end; }
.public-switch { display:flex; align-items:center; gap:.5rem; font-size:.82rem; color:var(--color-gray-2); }
.switch-label { font-size:.78rem; }
.pub-toggle { padding:.3rem .85rem; border-radius:4px; border:1px solid var(--color-border); font-size:.8rem; font-weight:600; cursor:pointer; background:var(--color-bg-card,#fff); color:var(--color-gray-2); transition:all .15s; }
.pub-toggle.public { border-color:rgba(34,197,94,.5); color:#22c55e; background:rgba(34,197,94,.06); }
.pub-toggle:not(.public) { border-color:var(--color-border); }
.pref-row { display:flex; align-items:center; justify-content:space-between; gap:1rem; padding:.4rem 0; border-bottom:1px solid var(--color-border,#E0E3DA); }
.pref-row:last-of-type { border-bottom:none; }
.pref-label { font-size:.82rem; color:var(--color-gray-1,#566270); }
.pref-toggle { padding:.28rem .85rem; border-radius:20px; border:1px solid var(--color-border,#E0E3DA); font-size:.78rem; font-weight:600; cursor:pointer; background:var(--color-bg-card,#fff); color:var(--color-gray-2,#8A9099); transition:all .15s; }
.pref-toggle.on { border-color:rgba(34,197,94,.5); color:#22c55e; background:rgba(34,197,94,.06); }
.pref-toggle.public { border-color:rgba(34,197,94,.5); color:#22c55e; background:rgba(34,197,94,.06); }
.pref-hint { font-size:.75rem; color:var(--color-gray-2); background:rgba(34,197,94,.06); border:1px solid rgba(34,197,94,.2); border-radius:4px; padding:.4rem .75rem; }
.sync-btns { display:flex; gap:.5rem; margin-top:.25rem; }
</style>
