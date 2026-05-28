<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">訓練日記管理</h1>
      <p class="page-subtitle">查看所有會員的訓練記錄與統計</p>
    </div>

    <!-- 全局統計 -->
    <div class="stats-row mb-2" v-if="!statsLoading && stats.length">
      <div class="stat-chip">
        <span class="stat-num">{{ stats.length }}</span>
        <span class="stat-lbl">活躍會員</span>
      </div>
      <div class="stat-chip">
        <span class="stat-num">{{ totalLogs }}</span>
        <span class="stat-lbl">總訓練筆數</span>
      </div>
      <div class="stat-chip">
        <span class="stat-num">{{ totalKm.toFixed(0) }}</span>
        <span class="stat-lbl">總累積公里</span>
      </div>
      <div class="stat-chip">
        <span class="stat-num">{{ Math.round(totalMin / 60) }}</span>
        <span class="stat-lbl">總訓練小時</span>
      </div>
    </div>

    <!-- Tabs -->
    <div class="tabs mb-2">
      <button class="tab-btn" :class="{ active: tab === 'logs' }"   @click="tab='logs'">📋 訓練記錄</button>
      <button class="tab-btn" :class="{ active: tab === 'stats' }"  @click="tab='stats'">📊 會員統計</button>
    </div>

    <!-- ── Tab: Logs ──────────────────────────────────────── -->
    <template v-if="tab === 'logs'">
      <!-- 篩選列 -->
      <div class="card mb-2">
        <div class="card-body" style="padding:1rem">
          <div class="filter-row">
            <!-- 關鍵字（搜尋標題或會員） -->
            <input v-model="filters.keyword" placeholder="搜尋標題 / 帳號 / 姓名..."
              @keyup.enter="fetchLogs" style="flex:1;min-width:160px" />
            <!-- 運動項目 -->
            <select v-model="filters.sport_type" @change="fetchLogs">
              <option value="">全部運動</option>
              <option value="1">🏃 路跑</option>
              <option value="2">🏊 游泳</option>
              <option value="3">🚴 單車</option>
              <option value="4">🏅 鐵人三項</option>
              <option value="5">💪 重訓</option>
              <option value="6">其他</option>
            </select>
            <!-- 日期區間 -->
            <input v-model="filters.date_from" type="date" placeholder="開始日期" style="width:140px" />
            <span class="text-gray" style="flex-shrink:0">～</span>
            <input v-model="filters.date_to"   type="date" placeholder="結束日期" style="width:140px" />
            <button class="btn btn-primary btn-sm" @click="fetchLogs">搜尋</button>
            <button class="btn btn-ghost btn-sm" @click="resetFilters">重設</button>
          </div>
          <!-- 篩選會員（下拉清單，選擇後篩選） -->
          <div class="filter-row" style="margin-top:.6rem">
            <div style="font-size:.78rem;color:var(--gray-2);flex-shrink:0">篩選會員：</div>
            <select v-model="filters.user_id" @change="fetchLogs" style="flex:1;max-width:300px">
              <option value="">全部會員</option>
              <option v-for="s in stats" :key="s.user_id" :value="s.user_id">
                {{ s.display_name || s.username }} (@{{ s.username }}) — {{ s.total_logs }} 筆
              </option>
            </select>
            <span v-if="filters.user_id" class="filter-tag">
              {{ currentMemberName }}
              <button @click="clearMember">✕</button>
            </span>
          </div>
        </div>
      </div>

      <!-- Table -->
      <div class="card">
        <div class="card-body" style="padding:0">
          <div v-if="logsLoading" class="loading-row">載入中...</div>
          <div v-else-if="!logs.length" class="loading-row text-gray">查無訓練記錄</div>
          <table v-else class="table" style="font-size:.82rem">
            <thead>
              <tr>
                <th>會員</th>
                <th>標題</th>
                <th>運動</th>
                <th>日期</th>
                <th>距離</th>
                <th>時間</th>
                <th>均心率</th>
                <th>公開</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in logs" :key="log.id">
                <td>
                  <div class="fw-bold">{{ log.display_name || log.username }}</div>
                  <div class="text-gray text-xs">@{{ log.username }}</div>
                </td>
                <td>
                  <div class="fw-bold">{{ log.title }}</div>
                  <div class="text-gray text-xs">{{ sourceLabel(log.source) }}</div>
                </td>
                <td><span class="badge badge-gray">{{ sportLabel(log.sport_type) }}</span></td>
                <td class="text-gray">{{ log.date }}</td>
                <td>{{ log.distance_km ? Number(log.distance_km).toFixed(2)+' km' : '-' }}</td>
                <td>{{ fmtDuration(log.duration_min) }}</td>
                <td>{{ log.avg_heart_rate ? log.avg_heart_rate+' bpm' : '-' }}</td>
                <td>
                  <span class="badge" :class="log.is_public ? 'badge-success' : 'badge-gray'">
                    {{ log.is_public ? '🌐 公開' : '🔒 私人' }}
                  </span>
                </td>
                <td>
                  <button class="btn btn-sm btn-ghost" @click="viewLog(log)">查看</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="pagination" v-if="totalPages > 1">
          <button :disabled="page===1" @click="goPage(page-1)" class="btn btn-ghost btn-sm">‹</button>
          <span class="text-gray">{{ page }} / {{ totalPages }}</span>
          <button :disabled="page===totalPages" @click="goPage(page+1)" class="btn btn-ghost btn-sm">›</button>
        </div>
      </div>
    </template>

    <!-- ── Tab: Stats ─────────────────────────────────────── -->
    <template v-if="tab === 'stats'">
      <div class="card">
        <div class="card-body" style="padding:0">
          <div v-if="statsLoading" class="loading-row">載入中...</div>
          <div v-else-if="!stats.length" class="loading-row text-gray">尚無訓練資料</div>
          <table v-else class="table">
            <thead>
              <tr><th>#</th><th>會員</th><th>總筆數</th><th>累積距離</th><th>累積時間</th><th>最近訓練</th><th>操作</th></tr>
            </thead>
            <tbody>
              <tr v-for="(s, i) in stats" :key="s.user_id">
                <td class="text-gray">{{ i+1 }}</td>
                <td>
                  <div class="fw-bold">{{ s.display_name || s.username }}</div>
                  <div class="text-gray text-xs">@{{ s.username }}</div>
                </td>
                <td><span class="fw-bold">{{ s.total_logs }}</span> 筆</td>
                <td class="fw-bold text-red">{{ Number(s.total_km).toFixed(1) }} km</td>
                <td>{{ Math.floor(s.total_min/60) }}h {{ s.total_min%60 }}m</td>
                <td class="text-gray text-xs">{{ s.last_date || '-' }}</td>
                <td>
                  <button class="btn btn-sm btn-ghost" @click="filterByUser(s)">查看記錄</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- ── Log Detail Modal ───────────────────────────────── -->
    <div class="modal-overlay" v-if="viewingLog" @click.self="viewingLog=null">
      <div class="log-modal">
        <div class="modal-header">
          <div>
            <h3>{{ viewingLog.title }}</h3>
            <div class="text-gray text-xs">
              {{ viewingLog.display_name || viewingLog.username }} (@{{ viewingLog.username }})
              · {{ viewingLog.date }}
              · {{ sportLabel(viewingLog.sport_type) }}
            </div>
          </div>
          <button @click="viewingLog=null">✕</button>
        </div>
        <div class="modal-body">
          <!-- Stats -->
          <div class="log-stats-row">
            <div class="log-stat" v-if="viewingLog.distance_km">
              <span class="ls-val">{{ Number(viewingLog.distance_km).toFixed(2) }}</span>
              <span class="ls-unit">km</span>
            </div>
            <div class="log-stat" v-if="viewingLog.duration_min">
              <span class="ls-val">{{ fmtDuration(viewingLog.duration_min) }}</span>
              <span class="ls-unit">時間</span>
            </div>
            <div class="log-stat" v-if="viewingLog.avg_pace">
              <span class="ls-val">{{ viewingLog.avg_pace }}</span>
              <span class="ls-unit">/km</span>
            </div>
            <div class="log-stat" v-if="viewingLog.avg_heart_rate">
              <span class="ls-val">{{ viewingLog.avg_heart_rate }}</span>
              <span class="ls-unit">bpm</span>
            </div>
            <div class="log-stat" v-if="viewingLog.elevation_m">
              <span class="ls-val">{{ viewingLog.elevation_m }}</span>
              <span class="ls-unit">m 爬升</span>
            </div>
            <div class="log-stat" v-if="viewingLog.calories">
              <span class="ls-val">{{ viewingLog.calories }}</span>
              <span class="ls-unit">kcal</span>
            </div>
          </div>
          <!-- Route info -->
          <div v-if="viewingLog.route_points?.length" class="route-info text-gray">
            🗺 GPS 路線 {{ viewingLog.route_points.length }} 點
          </div>
          <!-- Photos -->
          <div v-if="viewingLog.photos?.length" class="log-photos">
            <img v-for="(p,i) in viewingLog.photos" :key="i"
              :src="imgUrl(p)" class="log-photo" />
          </div>
          <!-- Note -->
          <div v-if="viewingLog.note" class="log-note">{{ viewingLog.note }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api from '@/services/api'

const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || ''
function imgUrl(p) { return p?.startsWith('http') ? p : `${IMAGE_BASE}/images/${p}` }

const tab = ref('logs')
const logs = ref([])
const stats = ref([])
const logsLoading = ref(false)
const statsLoading = ref(false)
const page = ref(1)
const totalPages = ref(1)
const viewingLog = ref(null)

const filters = reactive({
  keyword: '', sport_type: '',
  date_from: '', date_to: '',
  user_id: '',
})

const totalLogs = computed(() => stats.value.reduce((s, r) => s + r.total_logs, 0))
const totalKm   = computed(() => stats.value.reduce((s, r) => s + Number(r.total_km), 0))
const totalMin  = computed(() => stats.value.reduce((s, r) => s + r.total_min, 0))
const currentMemberName = computed(() => {
  if (!filters.user_id) return ''
  const s = stats.value.find(s => s.user_id == filters.user_id)
  return s ? (s.display_name || s.username) : ''
})

async function fetchLogs() {
  logsLoading.value = true
  try {
    const params = { page: page.value, page_size: 20 }
    if (filters.keyword)    params.keyword    = filters.keyword
    if (filters.sport_type) params.sport_type = filters.sport_type
    if (filters.date_from)  params.date_from  = filters.date_from
    if (filters.date_to)    params.date_to    = filters.date_to
    if (filters.user_id)    params.user_id    = filters.user_id
    const { data } = await api.get('/training', { params })
    logs.value = data.logs || []
    totalPages.value = data.pages || 1
  } catch(e) { console.error(e) }
  finally { logsLoading.value = false }
}

async function fetchStats() {
  statsLoading.value = true
  try {
    const { data } = await api.get('/training/stats')
    stats.value = data.stats || []
  } catch(e) { console.error(e) }
  finally { statsLoading.value = false }
}

function filterByUser(s) {
  filters.user_id = s.user_id
  tab.value = 'logs'
  page.value = 1
  fetchLogs()
}
function clearMember() { filters.user_id = ''; fetchLogs() }
function resetFilters() {
  Object.assign(filters, { keyword:'', sport_type:'', date_from:'', date_to:'', user_id:'' })
  page.value = 1
  fetchLogs()
}
function goPage(p) { page.value = p; fetchLogs() }
function viewLog(log) { viewingLog.value = log }

function sportLabel(t) { return { 1:'路跑',2:'游泳',3:'單車',4:'鐵人',5:'重訓',6:'其他' }[t]||'其他' }
function sourceLabel(s) { return { gpx:'GPX',fit:'FIT',garmin:'Garmin',manual:'' }[s]||'' }
function fmtDuration(min) {
  if (!min) return '-'
  const h = Math.floor(min/60), m = min%60
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}

onMounted(() => { fetchLogs(); fetchStats() })
</script>

<style scoped>
.filter-row { display:flex; gap:.6rem; flex-wrap:wrap; align-items:center; }
.filter-row input, .filter-row select { height:36px; font-size:.85rem; }
.filter-tag { display:flex; align-items:center; gap:.35rem; padding:.2rem .65rem; background:rgba(229,25,26,.1); border:1px solid rgba(229,25,26,.3); border-radius:3px; font-size:.78rem; color:var(--primary); }
.filter-tag button { background:none; border:none; cursor:pointer; color:var(--primary); font-size:.9rem; line-height:1; }

.stats-row { display:flex; gap:.75rem; flex-wrap:wrap; }
.stat-chip { background:var(--bg-card); border:1px solid var(--border); border-radius:4px; padding:.6rem 1.25rem; display:flex; flex-direction:column; align-items:center; }
.stat-num { font-family:var(--font-c); font-size:1.6rem; font-weight:700; color:var(--primary); }
.stat-lbl { font-size:.7rem; color:var(--gray-2); text-transform:uppercase; letter-spacing:.08em; }

.tabs { display:flex; gap:.5rem; }
.tab-btn { padding:.5rem 1.25rem; border-radius:4px; border:1px solid var(--border); font-size:.85rem; font-weight:600; cursor:pointer; background:var(--bg); color:var(--gray-2); transition:all .15s; }
.tab-btn.active { background:var(--primary); border-color:var(--primary); color:#fff; }

.fw-bold { font-weight:600; font-size:.9rem; }
.text-xs { font-size:.75rem; }
.text-red { color:var(--primary); }
.loading-row { padding:3rem; text-align:center; color:var(--gray-2); }
.pagination { display:flex; align-items:center; justify-content:center; gap:1rem; padding:1rem; border-top:1px solid var(--border); }

/* Modal */
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.75); z-index:100; display:flex; align-items:center; justify-content:center; padding:1rem; }
.log-modal { background:var(--bg-card); border:1px solid var(--border); border-radius:8px; width:100%; max-width:560px; max-height:88vh; overflow-y:auto; }
.modal-header { display:flex; align-items:flex-start; justify-content:space-between; padding:1.25rem 1.5rem; border-bottom:1px solid var(--border); position:sticky; top:0; background:var(--bg-card); }
.modal-header h3 { font-family:var(--font-c); font-size:1.1rem; font-weight:700; }
.modal-header button { background:none; border:none; color:var(--gray-2); font-size:1.2rem; cursor:pointer; flex-shrink:0; margin-left:1rem; }
.modal-body { padding:1.5rem; }

.log-stats-row { display:flex; gap:1rem; flex-wrap:wrap; background:var(--bg); border:1px solid var(--border); border-radius:6px; padding:.75rem 1rem; margin-bottom:1rem; }
.log-stat { display:flex; flex-direction:column; align-items:center; min-width:60px; }
.ls-val { font-family:var(--font-c); font-size:1.3rem; font-weight:700; color:var(--primary); }
.ls-unit { font-size:.68rem; color:var(--gray-2); text-transform:uppercase; letter-spacing:.06em; }

.route-info { font-size:.82rem; margin-bottom:.75rem; }
.log-photos { display:flex; gap:.5rem; flex-wrap:wrap; margin-bottom:.75rem; }
.log-photo { width:80px; height:80px; border-radius:4px; object-fit:cover; }
.log-note { font-size:.88rem; color:var(--gray-1); line-height:1.7; white-space:pre-wrap; background:var(--bg); border:1px solid var(--border); border-radius:4px; padding:.75rem; }
</style>
