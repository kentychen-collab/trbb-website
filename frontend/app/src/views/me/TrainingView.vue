<template>
  <div class="training-page">
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="section-title">訓練日記</h1>
        <p class="text-gray">記錄每次訓練，串接 Garmin 自動同步</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-ghost btn-sm" @click="showUpload=true">⬆ 上傳 GPX/FIT</button>
        <button class="btn btn-primary btn-sm" @click="openCreate">＋ 新增紀錄</button>
      </div>
    </div>

    <!-- Filter -->
    <div class="training-filter">
      <button v-for="s in sportTypes" :key="s.value"
        class="sport-btn" :class="{ active: selectedSport === s.value }"
        @click="selectSport(s.value)">
        {{ s.icon }} {{ s.label }}
      </button>
    </div>

    <!-- 日期篩選 + 批次操作 -->
    <div class="filter-bar">
      <div class="date-filter">
        <label>從</label>
        <input v-model="dateFrom" type="date" class="date-input" @change="applyDateFilter" />
        <label>到</label>
        <input v-model="dateTo" type="date" class="date-input" @change="applyDateFilter" />
        <button v-if="dateFrom || dateTo" class="btn btn-ghost btn-xs" @click="clearDateFilter">✕ 清除</button>
      </div>
      <div class="batch-actions" v-if="logs.length">
        <label class="select-all-label">
          <input type="checkbox" v-model="selectAll" @change="toggleSelectAll" />
          全選
        </label>
        <template v-if="selectedIds.size > 0">
          <span class="selected-count">已選 {{ selectedIds.size }} 筆</span>
          <button class="btn btn-ghost btn-xs" @click="batchSetPublic(true)">🌐 設為公開</button>
          <button class="btn btn-ghost btn-xs" @click="batchSetPublic(false)">🔒 設為私人</button>
          <button class="btn btn-ghost btn-xs delete-btn" @click="batchDelete">🗑 刪除</button>
        </template>
      </div>
    </div>

    <!-- Loading / Empty -->
    <div v-if="loading" class="loading-state"><div class="loading-spinner"></div></div>
    <div v-else-if="!logs.length" class="empty-state">
      <div class="empty-icon">🏃</div>
      <p>還沒有訓練紀錄</p>
      <p class="text-gray text-sm">上傳 GPX 或手動新增第一筆</p>
    </div>

    <!-- List -->
    <div v-else class="training-list">
      <div v-for="log in logs" :key="log.id"
        class="training-card card"
        :class="{ selected: selectedIds.has(log.id) }">
        <!-- Checkbox -->
        <div class="card-select" @click.stop>
          <input type="checkbox"
            :checked="selectedIds.has(log.id)"
            @change="toggleSelect(log.id)" />
        </div>
        <!-- Card content -->
        <div class="card-body-click" @click="openDetail(log)">
          <div class="tc-header">
            <div class="tc-sport-badge" :class="`sport-${log.sport_type}`">
              {{ sportIcon(log.sport_type) }} {{ sportLabel(log.sport_type) }}
            </div>
            <div class="tc-privacy">
              <span v-if="log.is_public" class="public-tag">🌐 公開</span>
              <span v-else class="private-tag">🔒 私人</span>
            </div>
          </div>
          <h3 class="tc-title">{{ log.title }}</h3>
          <div class="tc-date">{{ log.date }}</div>
          <div class="tc-stats">
            <div class="stat-item" v-if="log.distance_km">
              <span class="stat-val">{{ log.distance_km.toFixed(2) }}</span>
              <span class="stat-unit">km</span>
            </div>
            <div class="stat-item" v-if="log.duration_min">
              <span class="stat-val">{{ formatDuration(log.duration_min) }}</span>
              <span class="stat-unit">時間</span>
            </div>
            <div class="stat-item" v-if="log.avg_heart_rate">
              <span class="stat-val">{{ log.avg_heart_rate }}</span>
              <span class="stat-unit">avg bpm</span>
            </div>
            <div class="stat-item" v-if="log.avg_pace">
              <span class="stat-val">{{ log.avg_pace }}</span>
              <span class="stat-unit">/km</span>
            </div>
            <div class="stat-item" v-if="log.elevation_m">
              <span class="stat-val">{{ log.elevation_m }}</span>
              <span class="stat-unit">m 爬升</span>
            </div>
            <div class="stat-item" v-if="log.calories">
              <span class="stat-val">{{ log.calories }}</span>
              <span class="stat-unit">kcal</span>
            </div>
          </div>
          <!-- Mini map preview if GPX route available -->
          <div v-if="log.route_points && log.route_points.length" class="tc-map-preview">
            <canvas :id="`map-${log.id}`" width="100%" height="80"></canvas>
          </div>
          <!-- Source badge -->
          <div class="tc-source" v-if="log.source !== 'manual'">
            <span class="source-badge">{{ sourceLabel(log.source) }}</span>
          </div>
        </div><!-- end card-body-click -->
      </div>
    </div>

    <div class="pagination" v-if="totalPages > 1">
      <button :disabled="page===1" @click="goPage(page-1)" class="btn btn-ghost">‹</button>
      <span class="text-gray">{{ page }} / {{ totalPages }}</span>
      <button :disabled="page===totalPages" @click="goPage(page+1)" class="btn btn-ghost">›</button>
    </div>

    <!-- ── Detail / Edit Modal ─────────────────────────────── -->
    <div class="modal-overlay" v-if="detailLog" @click.self="detailLog=null">
      <div class="detail-modal">
        <div class="detail-header">
          <div>
            <span class="tc-sport-badge" :class="`sport-${detailLog.sport_type}`">
              {{ sportIcon(detailLog.sport_type) }} {{ sportLabel(detailLog.sport_type) }}
            </span>
            <h2>{{ detailLog.title }}</h2>
            <div class="text-gray">{{ detailLog.date }}</div>
          </div>
          <div class="detail-header-actions">
            <button class="btn btn-ghost btn-sm" @click="editLog(detailLog)">編輯</button>
            <button class="btn btn-ghost btn-sm" @click="confirmDeleteLog(detailLog)" style="color:var(--color-primary)">刪除</button>
            <button @click="detailLog=null" class="close-btn">✕</button>
          </div>
        </div>
        <div class="detail-body">
          <!-- GPX Map -->
          <div v-if="detailLog.route_points && detailLog.route_points.length" class="detail-map-container">
            <TrainingMap :route-points="detailLog.route_points" :start-lat="detailLog.start_lat" :start-lng="detailLog.start_lng" />
          </div>

          <!-- Stats grid -->
          <div class="detail-stats">
            <div class="dstat" v-if="detailLog.distance_km"><span class="dstat-val">{{ detailLog.distance_km.toFixed(2) }}</span><span class="dstat-lbl">公里</span></div>
            <div class="dstat" v-if="detailLog.duration_min"><span class="dstat-val">{{ formatDuration(detailLog.duration_min) }}</span><span class="dstat-lbl">時間</span></div>
            <div class="dstat" v-if="detailLog.avg_pace"><span class="dstat-val">{{ detailLog.avg_pace }}</span><span class="dstat-lbl">配速 /km</span></div>
            <div class="dstat" v-if="detailLog.avg_heart_rate"><span class="dstat-val">{{ detailLog.avg_heart_rate }}</span><span class="dstat-lbl">平均心率</span></div>
            <div class="dstat" v-if="detailLog.max_heart_rate"><span class="dstat-val">{{ detailLog.max_heart_rate }}</span><span class="dstat-lbl">最高心率</span></div>
            <div class="dstat" v-if="detailLog.elevation_m"><span class="dstat-val">{{ detailLog.elevation_m }}</span><span class="dstat-lbl">爬升 m</span></div>
            <div class="dstat" v-if="detailLog.calories"><span class="dstat-val">{{ detailLog.calories }}</span><span class="dstat-lbl">卡路里</span></div>
            <div class="dstat" v-if="detailLog.avg_speed_kph"><span class="dstat-val">{{ detailLog.avg_speed_kph }}</span><span class="dstat-lbl">平均速度</span></div>
          </div>

          <!-- Photos -->
          <div v-if="detailLog.photos && detailLog.photos.length" class="detail-photos">
            <img v-for="(p, i) in detailLog.photos" :key="i" :src="imgUrl(p)" class="detail-photo" @click="openPhotoFull(p)" />
          </div>

          <!-- Note -->
          <div v-if="detailLog.note" class="detail-note">{{ detailLog.note }}</div>

          <!-- 公開/私人 toggle -->
          <div class="detail-privacy-row">
            <span class="text-gray text-sm">公開設定：</span>
            <button class="toggle-privacy-btn"
              :class="detailLog.is_public ? 'is-public' : 'is-private'"
              @click="togglePublic(detailLog)">
              {{ detailLog.is_public ? '🌐 公開' : '🔒 私人' }}
            </button>
          </div>

          <!-- Share -->
          <div v-if="detailLog.is_public" class="detail-share">
            <span class="share-label">分享：</span>
            <button class="share-btn" @click="copyLink(detailLog)">🔗 複製連結</button>
            <a :href="lineShareUrl(detailLog)" target="_blank" class="share-btn line">LINE</a>
            <a :href="fbShareUrl(detailLog)" target="_blank" class="share-btn fb">Facebook</a>
            <button class="share-btn ig" @click="shareIG(detailLog)">Instagram</button>
          </div>
          <div v-else class="share-private-hint text-gray text-sm">
            設定為公開後才可以分享連結
          </div>
        </div>
      </div>
    </div>

    <!-- ── Create / Edit Form Modal ───────────────────────── -->
    <div class="modal-overlay" v-if="showForm" @click.self="showForm=false">
      <div class="form-modal">
        <div class="modal-header">
          <h3>{{ formEditId ? '編輯訓練紀錄' : '新增訓練紀錄' }}</h3>
          <button @click="showForm=false">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-row">
            <div class="form-group full">
              <label>標題 *</label>
              <input v-model="form.title" placeholder="訓練名稱" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>運動類型</label>
              <select v-model.number="form.sport_type">
                <option v-for="s in sportTypes.slice(1)" :key="s.value" :value="s.value">{{ s.icon }} {{ s.label }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>日期 *</label>
              <input v-model="form.date" type="date" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group"><label>距離 (km)</label><input v-model.number="form.distance_km" type="number" step="0.01" /></div>
            <div class="form-group">
              <label>時間</label>
              <div class="duration-row">
                <input v-model.number="form.duration_hours" type="number" min="0" placeholder="0" class="duration-input" />
                <span class="duration-sep">時</span>
                <input v-model.number="form.duration_mins" type="number" min="0" max="59" placeholder="0" class="duration-input" />
                <span class="duration-sep">分</span>
              </div>
            </div>
          </div>
          <div class="form-row">
            <div class="form-group"><label>平均心率</label><input v-model.number="form.avg_heart_rate" type="number" /></div>
            <div class="form-group"><label>最高心率</label><input v-model.number="form.max_heart_rate" type="number" /></div>
          </div>
          <div class="form-row">
            <div class="form-group"><label>卡路里</label><input v-model.number="form.calories" type="number" /></div>
            <div class="form-group"><label>爬升 (m)</label><input v-model.number="form.elevation_m" type="number" /></div>
          </div>
          <div class="form-group mb-1">
            <label>備註</label>
            <textarea v-model="form.note" rows="3" class="form-textarea"></textarea>
          </div>
          <!-- Photos upload -->
          <div class="form-group mb-1">
            <label>照片</label>
            <div class="photos-upload-row">
              <div v-for="(p, i) in form.photos" :key="i" class="photo-thumb">
                <img :src="imgUrl(p)" />
                <button @click="form.photos.splice(i,1)" type="button">✕</button>
              </div>
              <div class="photo-add" @click="triggerPhotoUpload">
                <input ref="photoInput" type="file" accept="image/*" multiple style="display:none" @change="uploadPhotos" />
                <span v-if="photoUploading">...</span>
                <span v-else>＋</span>
              </div>
            </div>
          </div>
          <!-- 公開設定 -->
          <div class="form-group mb-1">
            <label>是否公開</label>
            <div class="toggle-row">
              <button type="button" class="toggle-btn" :class="{ active: !form.is_public }" @click="form.is_public=false">🔒 私人（預設）</button>
              <button type="button" class="toggle-btn" :class="{ active: form.is_public }" @click="form.is_public=true">🌐 公開</button>
            </div>
          </div>
          <div v-if="formError" class="form-error">{{ formError }}</div>
          <div class="modal-footer">
            <button class="btn btn-primary" @click="submitForm" :disabled="formLoading">
              {{ formLoading ? '儲存中...' : (formEditId ? '更新' : '建立') }}
            </button>
            <button class="btn btn-ghost" @click="showForm=false">取消</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ── GPX/FIT Upload Modal ────────────────────────────── -->
    <div class="modal-overlay" v-if="showUpload" @click.self="showUpload=false">
      <div class="upload-modal">
        <div class="modal-header">
          <h3>上傳訓練檔案</h3>
          <button @click="showUpload=false">✕</button>
        </div>
        <div class="modal-body">
          <div class="upload-zone" @dragover.prevent @drop.prevent="onDrop" @click="triggerFileUpload">
            <input ref="fileInput" type="file" accept=".gpx,.fit" style="display:none" @change="onFileSelected" />
            <div v-if="uploadingFile">
              <div class="loading-spinner"></div>
              <p>{{ uploadStatus }}</p>
            </div>
            <div v-else>
              <div class="upload-icon">📂</div>
              <p>點擊或拖放 <strong>.gpx</strong> / <strong>.fit</strong> 檔案</p>
              <p class="text-gray text-sm">GPX 會自動解析並顯示地圖路線</p>
            </div>
          </div>
          <div v-if="uploadError" class="form-error mt-1">{{ uploadError }}</div>
          <div v-if="uploadedLog" class="upload-success">
            ✓ 上傳成功！已建立「{{ uploadedLog.title }}」
            <button class="btn btn-ghost btn-sm" @click="openDetail(uploadedLog);showUpload=false">查看</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import TrainingMap from '@/components/TrainingMap.vue'
import api from '@/services/api'

const auth = useAuthStore()
const logs = ref([])
const dateFrom    = ref('')
const dateTo      = ref('')
const selectedIds = ref(new Set())
const selectAll   = ref(false)
const loading = ref(false)
const page = ref(1)
const totalPages = ref(1)
const selectedSport = ref(0)
const detailLog = ref(null)
const showForm = ref(false)
const showUpload = ref(false)
const formEditId = ref(null)
const formLoading = ref(false)
const formError = ref('')
const uploadingFile = ref(false)
const uploadStatus = ref('')
const uploadError = ref('')
const uploadedLog = ref(null)
const syncing = ref(false)
const garminConnected = ref(false)
const lastSyncLabel = ref('從未')
const photoUploading = ref(false)
const fileInput = ref(null)
const photoInput = ref(null)

const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || ''
function imgUrl(path) {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `${IMAGE_BASE}/images/${path}`
}

const sportTypes = [
  { value:0, icon:'🏅', label:'全部' },
  { value:1, icon:'🏃', label:'跑步' },
  { value:2, icon:'🏊', label:'游泳' },
  { value:3, icon:'🚴', label:'騎車' },
  { value:4, icon:'🏊🚴🏃', label:'鐵人三項' },
  { value:5, icon:'💪', label:'健身' },
  { value:6, icon:'🏅', label:'其他' },
]
function sportLabel(t) { return sportTypes.find(s=>s.value===t)?.label || '其他' }
function sportIcon(t)  { return sportTypes.find(s=>s.value===t)?.icon || '🏅' }
function sourceLabel(s) { return { gpx:'GPX匯入', fit:'FIT匯入', garmin:'Garmin同步' }[s] || s }

const emptyForm = () => ({
  title: '', sport_type: 1, date: new Date().toISOString().slice(0,10),
  distance_km: null, duration_min: null, duration_hours: null, duration_mins: null,
  avg_heart_rate: null, max_heart_rate: null,
  calories: null, elevation_m: null,
  avg_pace: '', note: '', is_public: false, photos: [],
})
const form = reactive(emptyForm())

function formatDuration(mins) {
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return h > 0 ? `${h}:${String(m).padStart(2,'0')}` : `${m} 分`
}

async function fetchLogs() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: 20 }
    if (selectedSport.value) params.sport_type = selectedSport.value
    const { data } = await api.get('/me/training', { params })
    logs.value = data.logs || []
    totalPages.value = data.pages || 1
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

async function checkGarminStatus() {
  try {
    const { data } = await api.get('/me/garmin/status')
    garminConnected.value = data.connected
    if (data.last_sync_at) {
      lastSyncLabel.value = new Date(data.last_sync_at).toLocaleDateString('zh-TW')
    }
  } catch {}
}

function selectSport(v) { selectedSport.value = v; page.value = 1; fetchLogs() }
function applyDateFilter() { page.value = 1; fetchLogs() }
function clearDateFilter() { dateFrom.value = ''; dateTo.value = ''; applyDateFilter() }

function toggleSelect(id) {
  const s = new Set(selectedIds.value)
  if (s.has(id)) s.delete(id)
  else s.add(id)
  selectedIds.value = s
  selectAll.value = s.size === logs.value.length
}

function toggleSelectAll() {
  if (selectAll.value) {
    selectedIds.value = new Set(logs.value.map(l => l.id))
  } else {
    selectedIds.value = new Set()
  }
}

async function batchSetPublic(isPublic) {
  const ids = [...selectedIds.value]
  if (!ids.length) return
  try {
    await Promise.all(ids.map(id =>
      api.patch(`/training/${id}/public`, { is_public: isPublic })
    ))
    logs.value.forEach(l => {
      if (selectedIds.value.has(l.id)) l.is_public = isPublic
    })
    selectedIds.value = new Set()
    selectAll.value = false
  } catch(e) {
    alert('操作失敗：' + (e.response?.data?.error || e.message))
  }
}
function goPage(p) { page.value = p; fetchLogs() }

function openCreate() {
  Object.assign(form, emptyForm())
  formEditId.value = null; formError.value = ''
  showForm.value = true
}

function editLog(log) {
  Object.assign(form, {
    title: log.title, sport_type: log.sport_type, date: log.date,
    distance_km: log.distance_km, duration_min: log.duration_min,
    duration_hours: log.duration_min ? Math.floor(log.duration_min / 60) : null,
    duration_mins:  log.duration_min ? log.duration_min % 60 : null,
    avg_heart_rate: log.avg_heart_rate, max_heart_rate: log.max_heart_rate,
    calories: log.calories, elevation_m: log.elevation_m,
    avg_pace: log.avg_pace || '', note: log.note || '',
    is_public: log.is_public, photos: [...(log.photos || [])],
  })
  formEditId.value = log.id
  formError.value = ''
  showForm.value = true
  detailLog.value = null
}

async function submitForm() {
  formError.value = ''
  if (!form.title) { formError.value = '請填寫標題'; return }
  if (!form.date)  { formError.value = '請選擇日期'; return }
  formLoading.value = true
  try {
    // 合併 hours + mins → duration_min
    const h = form.duration_hours || 0
    const m = form.duration_mins  || 0
    const durationMin = (h * 60 + m) || null

    // 建立乾淨的 payload（過濾 0 值和前端專用欄位）
    const payload = {
      title:          form.title,
      sport_type:     form.sport_type || 1,
      date:           form.date,
      duration_min:   durationMin,
      distance_km:    form.distance_km  || null,
      avg_heart_rate: form.avg_heart_rate || null,
      max_heart_rate: form.max_heart_rate || null,
      calories:       form.calories     || null,
      elevation_m:    form.elevation_m  || null,
      avg_pace:       form.avg_pace     || '',
      note:           form.note         || '',
      is_public:      form.is_public    || false,
      photos:         form.photos       || [],
    }

    if (formEditId.value) {
      await api.put(`/training/${formEditId.value}`, payload)
    } else {
      await api.post('/training', payload)
    }
    showForm.value = false
    await fetchLogs()
  } catch(e) {
    formError.value = e.response?.data?.error || '儲存失敗'
  } finally {
    formLoading.value = false
  }
}

async function confirmDeleteLog(log) {
  if (!confirm(`確認刪除「${log.title}」？`)) return
  try {
    await api.delete(`/training/${log.id}`)
    detailLog.value = null
    await fetchLogs()
  } catch(e) { alert(e.response?.data?.error || '刪除失敗') }
}

// Photos
function triggerPhotoUpload() { photoInput.value?.click() }
async function uploadPhotos(e) {
  const files = Array.from(e.target.files || [])
  photoUploading.value = true
  try {
    for (const file of files) {
      const fd = new FormData()
      fd.append('file', file)
      const { data } = await api.post('/upload/image?folder=training', fd, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      form.photos.push(data.path)
    }
  } catch(e) { formError.value = '照片上傳失敗' }
  finally { photoUploading.value = false; e.target.value = '' }
}

// GPX/FIT upload
function triggerFileUpload() { fileInput.value?.click() }
function onDrop(e) {
  const file = e.dataTransfer.files[0]
  if (file) handleFileUpload(file)
}
function onFileSelected(e) {
  const file = e.target.files[0]
  if (file) handleFileUpload(file)
}
async function handleFileUpload(file) {
  const ext = file.name.toLowerCase().split('.').pop()
  if (!['gpx','fit'].includes(ext)) {
    uploadError.value = '只支援 .gpx 和 .fit 格式'
    return
  }
  uploadingFile.value = true
  uploadError.value = ''
  uploadStatus.value = ext === 'gpx' ? '解析 GPX 並建立日記...' : '上傳 FIT 檔案...'
  uploadedLog.value = null

  const fd = new FormData()
  fd.append('file', file)
  try {
    const endpoint = `/training/upload/${ext}`
    const { data } = await api.post(endpoint, fd, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    uploadedLog.value = data.log
    await fetchLogs()
  } catch(e) {
    uploadError.value = e.response?.data?.error || '上傳失敗'
  } finally {
    uploadingFile.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

// Public toggle
async function batchDelete() {
  const ids = [...selectedIds.value]
  if (!ids.length) return
  if (!confirm(`確認刪除選取的 ${ids.length} 筆訓練記錄？此操作無法復原。`)) return
  try {
    await Promise.all(ids.map(id => api.delete(`/training/${id}`)))
    logs.value = logs.value.filter(l => !selectedIds.value.has(l.id))
    selectedIds.value = new Set()
    selectAll.value = false
  } catch(e) {
    alert('刪除失敗：' + (e.response?.data?.error || e.message))
  }
}

async function togglePublic(log) {
  try {
    await api.patch(`/training/${log.id}/public`, { is_public: !log.is_public })
    log.is_public = !log.is_public
  } catch(e) { alert('更新失敗') }
}

// Share
const SITE_URL = import.meta.env.VITE_API_BASE_URL?.replace('/v1/api','') || 'https://trbbtw.com'
function shareUrl(log) { return `${SITE_URL}/training/share/${log.uuid}` }
function lineShareUrl(log) { return `https://social-plugins.line.me/lineit/share?url=${encodeURIComponent(shareUrl(log))}` }
function fbShareUrl(log) { return `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(shareUrl(log))}` }
function copyLink(log) {
  navigator.clipboard.writeText(shareUrl(log)).then(() => alert('連結已複製！'))
}
function shareIG(log) {
  navigator.clipboard.writeText(shareUrl(log)).then(() => alert('連結已複製，請貼到 Instagram 限時動態！'))
}

// Garmin
async function connectGarmin() {
  try {
    const { data } = await api.get('/me/garmin/connect')
    if (data.auth_url) {
      alert(`Garmin API 尚未設定（${data.error}）`)
    }
  } catch(e) { alert(e.response?.data?.error || '連結失敗') }
}
async function disconnectGarmin() {
  if (!confirm('確認解除 Garmin 連結？')) return
  try {
    await api.delete('/me/garmin/disconnect')
    garminConnected.value = false
  } catch {}
}
async function syncGarmin() {
  syncing.value = true
  try {
    const { data } = await api.post('/training/garmin/sync')
    alert(data.message || '同步完成')
    await fetchLogs()
  } catch(e) {
    alert(e.response?.data?.error || '同步失敗，請確認 Garmin 連結狀態')
  } finally { syncing.value = false }
}

function openDetail(log) { detailLog.value = log }
function openPhotoFull(p) { window.open(imgUrl(p), '_blank') }

onMounted(async () => {
  await fetchLogs()
  await checkGarminStatus()
})
</script>

<style scoped>
.training-page { max-width:760px; }
.page-header { margin-bottom:1.25rem; }
.header-actions { display:flex; gap:.5rem; }
.btn-sm { padding:.4rem 1rem; font-size:.82rem; }

/* Garmin banner */
.garmin-icon { font-size:1.5rem; }
.garmin-status-title { font-weight:600; font-size:.9rem; }
.garmin-status-sub { font-size:.78rem; }

/* Filter */
.training-filter { display:flex; gap:.5rem; flex-wrap:wrap; margin-bottom:.75rem; }

/* Filter bar */
.filter-bar { display:flex; justify-content:space-between; align-items:center; flex-wrap:wrap; gap:.6rem; margin-bottom:1.25rem; padding:.75rem 1rem; background:var(--color-bg-card); border:1px solid var(--color-border); border-radius:6px; }
.date-filter { display:flex; align-items:center; gap:.5rem; flex-wrap:wrap; font-size:.82rem; color:var(--color-gray-2); }
.date-input { height:32px; padding:.2rem .6rem; font-size:.82rem; width:130px; }
.btn-xs { padding:.2rem .65rem; font-size:.76rem; }
.batch-actions { display:flex; align-items:center; gap:.5rem; flex-wrap:wrap; font-size:.82rem; }
.select-all-label { display:flex; align-items:center; gap:.3rem; cursor:pointer; font-weight:600; color:var(--color-gray-1); }
.selected-count { color:var(--color-primary); font-weight:600; font-size:.8rem; }

/* Card select checkbox */
.training-card { position:relative; }
.card-select { position:absolute; top:.75rem; left:.75rem; z-index:2; }
.card-select input[type=checkbox] { width:16px; height:16px; cursor:pointer; accent-color:var(--color-primary); }
.card-body-click { cursor:pointer; }
.training-card.selected { border-color:var(--color-primary); background:rgba(207,32,39,.03); }
.sport-btn { padding:.3rem .85rem; border-radius:4px; border:1px solid var(--color-border); font-size:.8rem; cursor:pointer; background:none; color:var(--color-gray-2); transition:all .15s; }
.sport-btn.active, .sport-btn:hover { border-color:var(--color-primary); color:var(--color-primary); }

.loading-state { display:flex; justify-content:center; padding:3rem; }
.loading-spinner { width:24px; height:24px; border:2px solid var(--color-border); border-top-color:var(--color-primary); border-radius:50%; animation:spin .7s linear infinite; }
@keyframes spin { to { transform:rotate(360deg) } }
.empty-state { text-align:center; padding:3rem; color:var(--color-gray-2); }
.empty-icon { font-size:3rem; margin-bottom:.75rem; }

/* Training cards */
.training-list { display:flex; flex-direction:column; gap:1rem; }
.training-card { padding:1.25rem; cursor:pointer; transition:all .2s; }
.training-card:hover { border-color:var(--color-primary); }
.tc-header { display:flex; justify-content:space-between; align-items:center; margin-bottom:.5rem; }
.tc-sport-badge { display:inline-flex; align-items:center; gap:.25rem; font-family:var(--font-cond); font-size:.7rem; font-weight:700; letter-spacing:.08em; text-transform:uppercase; padding:.2rem .65rem; border-radius:3px; }
.sport-1 { background:rgba(34,197,94,.1); color:#22c55e; }
.sport-2 { background:rgba(59,130,246,.1); color:#60a5fa; }
.sport-3 { background:rgba(245,158,11,.1); color:#f59e0b; }
.sport-4 { background:rgba(168,85,247,.1); color:#c084fc; }
.sport-5 { background:rgba(239,68,68,.1); color:#f87171; }
.sport-6 { background:rgba(107,114,128,.1); color:#9ca3af; }
.tc-privacy { font-size:.75rem; }
.public-tag { color:#22c55e; }
.private-tag { color:var(--color-gray-2); }
.tc-title { font-size:1rem; font-weight:700; margin-bottom:.2rem; }
.tc-date { font-size:.78rem; color:var(--color-gray-2); margin-bottom:.75rem; }
.tc-stats { display:flex; gap:1.25rem; flex-wrap:wrap; }
.stat-item { display:flex; flex-direction:column; }
.stat-val { font-family:var(--font-cond); font-size:1.2rem; font-weight:700; }
.stat-unit { font-size:.68rem; color:var(--color-gray-2); text-transform:uppercase; letter-spacing:.08em; }
.tc-source { margin-top:.5rem; }
.source-badge { font-size:.7rem; background:rgba(229,25,26,.08); color:var(--color-primary); padding:.15rem .5rem; border-radius:3px; }
.pagination { display:flex; align-items:center; justify-content:center; gap:1rem; margin-top:1.25rem; }

/* Detail modal */
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.75); z-index:200; display:flex; align-items:center; justify-content:center; padding:1rem; overflow-y:auto; }
.detail-modal { background:var(--color-bg-card); border:1px solid var(--color-border); border-radius:10px; width:100%; max-width:680px; max-height:92vh; overflow-y:auto; }
.detail-header { display:flex; justify-content:space-between; align-items:flex-start; padding:1.5rem; border-bottom:1px solid var(--color-border); position:sticky; top:0; background:var(--color-bg-card); z-index:1; }
.detail-header h2 { font-size:1.2rem; margin:.4rem 0 .2rem; }
.detail-header-actions { display:flex; gap:.4rem; align-items:center; }
.close-btn { background:none; border:none; color:var(--color-gray-2); font-size:1.2rem; cursor:pointer; }
.detail-body { padding:1.5rem; }
.detail-map-container { height:280px; border-radius:8px; overflow:hidden; margin-bottom:1.25rem; border:1px solid var(--color-border); }
.detail-stats { display:grid; grid-template-columns:repeat(auto-fill,minmax(100px,1fr)); gap:.75rem; margin-bottom:1.25rem; }
.dstat { background:var(--color-bg); border:1px solid var(--color-border); border-radius:6px; padding:.6rem .8rem; text-align:center; }
.dstat-val { display:block; font-family:var(--font-cond); font-size:1.3rem; font-weight:700; }
.dstat-lbl { font-size:.68rem; color:var(--color-gray-2); text-transform:uppercase; letter-spacing:.06em; }
.detail-photos { display:flex; gap:.5rem; flex-wrap:wrap; margin-bottom:1rem; }
.detail-photo { width:100px; height:100px; object-fit:cover; border-radius:4px; cursor:pointer; border:1px solid var(--color-border); }
.detail-note { color:var(--color-gray-1); font-size:.9rem; line-height:1.7; margin-bottom:1rem; padding:1rem; background:var(--color-bg); border-radius:6px; }
.detail-privacy-row { display:flex; align-items:center; gap:.75rem; padding:.75rem 0; border-top:1px solid var(--color-border); }
.toggle-privacy-btn { padding:.35rem .9rem; border-radius:4px; border:1px solid var(--color-border); font-size:.82rem; cursor:pointer; background:none; }
.toggle-privacy-btn.is-public { border-color:#22c55e; color:#22c55e; }
.toggle-privacy-btn.is-private { color:var(--color-gray-2); }
.detail-share { display:flex; align-items:center; gap:.5rem; flex-wrap:wrap; padding:.5rem 0; }
.share-label { font-size:.78rem; color:var(--color-gray-2); }
.share-btn { padding:.3rem .8rem; border-radius:4px; border:1px solid var(--color-border); font-size:.78rem; cursor:pointer; background:none; color:var(--color-gray-1); text-decoration:none; display:inline-flex; align-items:center; }
.share-btn.line { border-color:#06c755; color:#06c755; }
.share-btn.fb { border-color:#1877f2; color:#1877f2; }
.share-private-hint { padding:.5rem 0; }

/* Form modal */
.form-modal { background:var(--color-bg-card); border:1px solid var(--color-border); border-radius:10px; width:100%; max-width:560px; max-height:92vh; overflow-y:auto; }
.modal-header { display:flex; align-items:center; justify-content:space-between; padding:1.25rem 1.5rem; border-bottom:1px solid var(--color-border); position:sticky; top:0; background:var(--color-bg-card); }
.modal-header h3 { font-family:var(--font-cond); font-size:1.1rem; font-weight:700; }
.modal-header button { background:none; border:none; color:var(--color-gray-2); font-size:1.2rem; cursor:pointer; }
.modal-body { padding:1.5rem; }
.form-row { display:grid; grid-template-columns:1fr 1fr; gap:.75rem; margin-bottom:.75rem; }
.form-group { display:flex; flex-direction:column; gap:.3rem; }
.form-group.full { grid-column:1/-1; }
.form-group.mb-1 { margin-bottom:.75rem; }
.form-group label { font-size:.72rem; font-weight:600; text-transform:uppercase; letter-spacing:.06em; color:var(--color-gray-1); }
.form-group input, .form-group select { width:100%; }
.form-textarea { width:100%; background:var(--color-bg); color:#fff; border:1px solid var(--color-border); border-radius:4px; padding:.6rem .9rem; font-family:inherit; resize:vertical; }
.duration-row { display:flex; align-items:center; gap:.4rem; }
.duration-input { width:64px; }
.duration-sep { font-size:.85rem; color:var(--color-gray-2); flex-shrink:0; }
.photos-upload-row { display:flex; gap:.5rem; flex-wrap:wrap; }
.photo-thumb { width:64px; height:64px; border-radius:4px; overflow:hidden; position:relative; }
.photo-thumb img { width:100%; height:100%; object-fit:cover; }
.photo-thumb button { position:absolute; top:1px; right:1px; width:18px; height:18px; border-radius:50%; background:rgba(0,0,0,.65); border:none; color:#fff; font-size:.65rem; cursor:pointer; }
.photo-add { width:64px; height:64px; border-radius:4px; border:2px dashed var(--color-border); display:flex; align-items:center; justify-content:center; font-size:1.5rem; cursor:pointer; color:var(--color-gray-2); }
.photo-add:hover { border-color:var(--color-primary); }
.toggle-row { display:flex; gap:.5rem; }
.toggle-btn { padding:.4rem 1rem; border-radius:4px; border:1px solid var(--color-border); font-size:.82rem; font-weight:600; cursor:pointer; background:none; color:var(--color-gray-2); }
.toggle-btn.active { background:var(--color-primary); border-color:var(--color-primary); color:#fff; }
.form-error { background:rgba(239,68,68,.1); border:1px solid rgba(239,68,68,.3); border-radius:4px; color:#fca5a5; font-size:.83rem; padding:.5rem .75rem; }
.modal-footer { display:flex; gap:.75rem; margin-top:1.25rem; padding-top:1rem; border-top:1px solid var(--color-border); }
.mt-1 { margin-top:.5rem; }

/* Upload modal */
.upload-modal { background:var(--color-bg-card); border:1px solid var(--color-border); border-radius:10px; width:100%; max-width:440px; }
.upload-zone { border:2px dashed var(--color-border); border-radius:8px; padding:3rem 2rem; text-align:center; cursor:pointer; transition:border-color .2s; }
.upload-zone:hover { border-color:var(--color-primary); }
.upload-icon { font-size:3rem; margin-bottom:1rem; }
.upload-success { background:rgba(34,197,94,.1); border:1px solid rgba(34,197,94,.3); border-radius:4px; color:#86efac; font-size:.85rem; padding:.6rem 1rem; margin-top:.75rem; display:flex; align-items:center; justify-content:space-between; }
.delete-btn { color:var(--color-danger, #ef4444) !important; }
</style>
