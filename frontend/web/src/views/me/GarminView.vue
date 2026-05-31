<template>
  <div class="garmin-page">
    <h1 class="section-title">Garmin 串接</h1>
    <p class="text-gray" style="margin-bottom:2rem">
      連結您的 Garmin 帳號，自動將訓練活動匯入訓練日記。
    </p>

    <!-- 連結狀態卡 -->
    <div class="status-card card">
      <div class="status-left">
        <div class="garmin-logo">⌚</div>
        <div class="status-info">
          <div class="status-title">
            <span class="dot" :class="dotClass"></span>
            {{ statusTitle }}
          </div>
          <div class="text-gray">{{ statusSub }}</div>
          <div v-if="tokenInfo" class="token-meta text-gray">
            <span v-if="tokenInfo.last_sync_at">上次同步：{{ fmt(tokenInfo.last_sync_at) }}</span>
          </div>
        </div>
      </div>
      <div class="status-actions">
        <template v-if="garminStatus === 'connected'">
          <div class="sync-options">
            <label class="pref-row">
              <span class="pref-label">同步後設為</span>
              <button class="pref-toggle public-tog" :class="{ public: syncPublic }"
                @click="syncPublic = !syncPublic" type="button">
                {{ syncPublic ? '🌐 公開' : '🔒 私人' }}
              </button>
            </label>
            <button class="btn btn-primary btn-sm" @click="syncNow" :disabled="syncing">
              <span v-if="syncing" class="spinner"></span>
              {{ syncing ? '同步中...' : '🔄 同步近 90 天' }}
            </button>
            <button class="btn btn-ghost btn-sm" @click="confirmDisconnect">解除連結</button>
          </div>
        </template>
        <template v-else-if="garminStatus === 'disconnected'">
          <button class="btn btn-garmin" @click="connectGarmin" :disabled="connecting">
            {{ connecting ? '連接中...' : '連結 Garmin' }}
          </button>
        </template>
        <template v-else-if="garminStatus === 'unavailable'">
          <div class="unavail-badge">Garmin API 尚未開放</div>
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
        <span v-else>無新活動</span>
      </template>
      <template v-else>✗ {{ syncResult.msg }}</template>
    </div>

    <!-- 說明卡片 -->
    <div class="info-cards">
      <div class="info-card card">
        <div class="info-icon">🔐</div>
        <h4>OAuth 1.0a 授權</h4>
        <p>透過 Garmin Health API，安全授權存取您的訓練資料。</p>
      </div>
      <div class="info-card card">
        <div class="info-icon">📊</div>
        <h4>同步的資料</h4>
        <p>距離、時間、配速、心率、海拔、卡路里、GPS 路線等訓練數據。</p>
      </div>
      <div class="info-card card">
        <div class="info-icon">🔄</div>
        <h4>自動同步</h4>
        <p>連結後每次訓練完成 Garmin 上傳，即可在 TRBB 訓練日記自動顯示。</p>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import api from '@/services/api'

const garminStatus = ref('loading') // loading | connected | disconnected | unavailable
const syncPublic   = ref(false)     // 同步後是否公開
const tokenInfo    = ref(null)
const connecting   = ref(false)
const syncing      = ref(false)
const syncResult   = ref(null)
const apiConfigured = ref(false)

const statusTitle = computed(() => ({
  loading:      '檢查連結狀態中...',
  connected:    'Garmin Connect 已連結',
  disconnected: 'Garmin Connect 尚未連結',
  unavailable:  'Garmin API 設定中',
}[garminStatus.value] || ''))

const statusSub = computed(() => ({
  loading:      '',
  connected:    '訓練資料將自動同步至日記',
  disconnected: '連結後可自動匯入 Garmin 訓練資料',
  unavailable:  '管理員尚未設定 Garmin API 憑證，串接功能暫時無法使用',
}[garminStatus.value] || ''))

async function checkStatus() {
  try {
    const { data } = await api.get('/me/garmin/status')
    if (data.connected) {
      garminStatus.value = 'connected'
      tokenInfo.value = data
    } else {
      garminStatus.value = data.api_configured === false ? 'unavailable' : 'disconnected'
      apiConfigured.value = data.api_configured !== false
    }
  } catch {
    garminStatus.value = 'disconnected'
  }
}

async function connectGarmin() {
  connecting.value = true
  try {
    const { data } = await api.get('/me/garmin/connect')
    if (data.auth_url) {
      // Redirect to Garmin OAuth page
      window.location.href = data.auth_url
    } else {
      // API not yet configured
      garminStatus.value = 'unavailable'
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
    const { data } = await api.post('/training/garmin/sync')
    const count = data.count || 0

    // 若選擇公開，批次 PATCH 剛同步的 garmin 記錄
    if (syncPublic.value && count > 0) {
      const { data: logs } = await api.get('/me/training', {
        params: { page: 1, page_size: count + 10 }
      })
      const recentIds = (logs.logs || [])
        .filter(l => l.source === 'garmin' && !l.is_public)
        .slice(0, count)
        .map(l => l.id)
      if (recentIds.length) {
        await Promise.all(recentIds.map(id =>
          api.patch(`/training/${id}/public`, { is_public: true })
        ))
      }
      syncResult.value = { ok: true, count, public: recentIds.length }
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
  if (!confirm('確認解除 Garmin Connect 連結？\n已匯入的訓練資料不會刪除。')) return
  api.delete('/me/garmin/disconnect').then(() => {
    garminStatus.value = 'disconnected'
    tokenInfo.value = null
  }).catch(() => alert('解除失敗'))
}

function fmt(d) {
  return d ? new Date(d).toLocaleString('zh-TW', {
    year:'numeric', month:'2-digit', day:'2-digit', hour:'2-digit', minute:'2-digit'
  }) : '-'
}

onMounted(checkStatus)
</script>

<style scoped>
.garmin-page { max-width:700px; }
.status-card { padding:1.5rem; display:flex; justify-content:space-between; align-items:center; gap:1.5rem; flex-wrap:wrap; margin-bottom:1rem; }
.status-left { display:flex; align-items:center; gap:1.25rem; }
.garmin-logo { font-size:3rem; flex-shrink:0; }
.status-title { font-size:1.05rem; font-weight:700; display:flex; align-items:center; gap:.5rem; margin-bottom:.25rem; }
.dot { width:10px; height:10px; border-radius:50%; background:var(--color-gray-2); display:inline-block; }
.dot.connected { background:#22c55e; box-shadow:0 0 6px rgba(34,197,94,.5); }
.dot.disconnected { background:#9ca3af; }
.token-meta { font-size:.78rem; margin-top:.35rem; display:flex; flex-wrap:wrap; gap:.75rem; }
.status-actions { display:flex; flex-direction:column; gap:.5rem; align-items:flex-end; }
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
.apply-step strong { font-size:.9rem; }
.step-icon { font-size:1.1rem; flex-shrink:0; margin-top:.05rem; }
.apply-step.pending { opacity:.7; }
.apply-step.disabled { opacity:.4; }
.apply-step.done { opacity:1; }

.alt-section { padding:1.5rem; }
.alt-section h3 { font-size:1rem; margin-bottom:.5rem; }

.spinner { width:14px; height:14px; border:2px solid rgba(255,255,255,.3); border-top-color:#fff; border-radius:50%; animation:spin .7s linear infinite; display:inline-block; }
@keyframes spin { to { transform:rotate(360deg) } }
.sync-options { display:flex; flex-direction:column; gap:.5rem; align-items:flex-end; }
.public-switch { display:flex; align-items:center; gap:.5rem; font-size:.82rem; color:var(--color-gray-2); }
.switch-label { font-size:.78rem; }
.pub-toggle { padding:.3rem .85rem; border-radius:4px; border:1px solid var(--color-border); font-size:.8rem; font-weight:600; cursor:pointer; background:var(--color-bg-card,#fff); color:var(--color-gray-2); transition:all .15s; }
.pub-toggle.public { border-color:rgba(34,197,94,.5); color:#22c55e; background:rgba(34,197,94,.06); }
</style>
