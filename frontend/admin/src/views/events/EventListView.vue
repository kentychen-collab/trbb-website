<template>
  <div>
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="page-title">賽事管理</h1>
        <p class="page-subtitle">新增、編輯、管理所有賽事項目</p>
      </div>
      <button class="btn btn-primary" @click="openCreate">＋ 新增賽事</button>
    </div>

    <!-- Filters -->
    <div class="card mb-2">
      <div class="card-body" style="padding:1rem">
        <div class="filter-row">
          <input v-model="filters.keyword" placeholder="搜尋賽事名稱..." @keyup.enter="fetchEvents" style="flex:1;min-width:180px" />
          <select v-model="filters.status" @change="fetchEvents">
            <option value="">全部狀態</option>
            <option value="0">草稿</option>
            <option value="1">已發布</option>
            <option value="2">額滿</option>
            <option value="3">已結束</option>
            <option value="4">已取消</option>
          </select>
          <button class="btn btn-primary btn-sm" @click="fetchEvents">搜尋</button>
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="card">
      <div class="card-body" style="padding:0">
        <div v-if="loading" class="loading-row">載入中...</div>
        <div v-else-if="!events.length" class="loading-row text-gray">目前沒有賽事資料</div>
        <table v-else class="table">
          <thead>
            <tr>
              <th style="width:70px">圖片</th>
              <th>賽事名稱</th>
              <th>賽事日期</th>
              <th>報名期間</th>
              <th>費用</th>
              <th>報名人數</th>
              <th>狀態</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="ev in events" :key="ev.id">
              <td>
                <div class="event-thumb" :style="ev.cover_url ? `background-image:url(${ev.cover_url})` : ''">
                  <span v-if="!ev.cover_url">{{ typeIcon(ev.event_type) }}</span>
                </div>
              </td>
              <td>
                <div class="fw-bold">{{ ev.title }}</div>
                <div class="text-gray text-xs">{{ typeLabel(ev.event_type) }} · {{ ev.location }}</div>
              </td>
              <td class="text-xs">{{ fmt(ev.start_at) }}</td>
              <td class="text-xs">
                <div>{{ fmt(ev.reg_start_at) }}</div>
                <div class="text-gray">～ {{ fmt(ev.reg_end_at) }}</div>
              </td>
              <td>
                <span v-if="ev.fee > 0" class="text-red fw-bold">NT$ {{ Number(ev.fee).toLocaleString() }}</span>
                <span v-else class="text-gray">免費</span>
              </td>
              <td>
                <span class="fw-bold">{{ ev.registered_count }}</span>
                <span class="text-gray text-xs"> / {{ ev.max_participants || '∞' }}</span>
              </td>
              <td>
                <span class="badge" :class="statusBadge(ev.status)">{{ statusLabel(ev.status) }}</span>
              </td>
              <td>
                <div class="action-btns">
                  <button class="btn btn-sm btn-ghost" @click="openEdit(ev)">編輯</button>
                  <RouterLink :to="`/events/${ev.id}`" class="btn btn-sm btn-ghost">報名名單</RouterLink>
                  <button class="btn btn-sm btn-ghost" style="color:var(--danger)"
                    @click="confirmDelete(ev)">刪除</button>
                </div>
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

    <!-- ── Create / Edit Modal ─────────────────────────────── -->
    <div class="modal-overlay" v-if="showForm" @click.self="showForm=false">
      <div class="event-form-modal">
        <div class="modal-header">
          <h3>{{ editingId ? '編輯賽事' : '新增賽事' }}</h3>
          <button @click="showForm=false">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-grid">
            <!-- 賽事名稱 -->
            <div class="form-group full">
              <label>賽事名稱 <span class="req">*</span></label>
              <input v-model="form.title" placeholder="賽事名稱" />
            </div>
            <!-- 類型 + 地點 -->
            <div class="form-group">
              <label>賽事類型</label>
              <select v-model="form.event_type">
                <option :value="1">🏊 鐵人三項</option>
                <option :value="2">🏃 路跑</option>
                <option :value="3">🏊 游泳</option>
                <option :value="4">🚴 單車</option>
                <option :value="5">💪 訓練</option>
                <option :value="6">🏅 其他</option>
              </select>
            </div>
            <div class="form-group">
              <label>地點 <span class="req">*</span></label>
              <input v-model="form.location" placeholder="賽事地點" />
            </div>
            <!-- 賽事日期（點選） -->
            <div class="form-group">
              <label>賽事開始日期 <span class="req">*</span></label>
              <input v-model="form.start_at_str" type="datetime-local" />
            </div>
            <div class="form-group">
              <label>賽事結束日期</label>
              <input v-model="form.end_at_str" type="datetime-local" />
            </div>
            <!-- 報名日期（點選） -->
            <div class="form-group">
              <label>報名開始 <span class="req">*</span></label>
              <input v-model="form.reg_start_at_str" type="datetime-local" />
            </div>
            <div class="form-group">
              <label>報名截止 <span class="req">*</span></label>
              <input v-model="form.reg_end_at_str" type="datetime-local" />
            </div>
            <!-- 費用 + 人數 -->
            <div class="form-group">
              <label>報名費用（NT$，0 為免費）</label>
              <input v-model.number="form.fee" type="number" min="0" placeholder="0" />
            </div>
            <div class="form-group">
              <label>人數上限（留空不限）</label>
              <input v-model.number="form.max_participants" type="number" min="1" placeholder="不限" />
            </div>
            <!-- 封面圖片上傳 -->
            <div class="form-group full">
              <label>封面圖片</label>
              <ImageUpload v-model="form.cover_path" folder="events" />
            </div>
            <!-- 是否公開 — 用 toggle 切換 -->
            <div class="form-group full">
              <label>前台顯示狀態</label>
              <div class="status-toggle-row">
                <button v-for="opt in statusOptions" :key="opt.value"
                  class="status-btn"
                  :class="{ active: form.status === opt.value }"
                  @click="form.status = opt.value" type="button">
                  {{ opt.label }}
                </button>
              </div>
              <p class="hint">只有「已發布」才會在前台顯示給訪客</p>
            </div>
            <!-- 說明 -->
            <div class="form-group full">
              <label>賽事說明</label>
              <textarea v-model="form.description" rows="5"
                placeholder="賽事詳細說明、注意事項、交通資訊等..."
                class="form-textarea"></textarea>
            </div>
          </div>

          <div v-if="formError" class="form-error">{{ formError }}</div>

          <div class="modal-footer">
            <button class="btn btn-primary" @click="submitForm" :disabled="formLoading">
              <span v-if="formLoading" class="spinner"></span>
              {{ formLoading ? '儲存中...' : (editingId ? '更新賽事' : '建立賽事') }}
            </button>
            <button class="btn btn-ghost" @click="showForm=false">取消</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete confirm -->
    <div class="modal-overlay" v-if="deletingEvent" @click.self="deletingEvent=null">
      <div class="confirm-modal">
        <h3>確認刪除賽事？</h3>
        <p class="text-gray">「{{ deletingEvent?.title }}」將被永久刪除，此動作無法復原。</p>
        <div class="confirm-actions">
          <button class="btn btn-danger" @click="doDelete">確認刪除</button>
          <button class="btn btn-ghost" @click="deletingEvent=null">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import api from '@/services/api'
import ImageUpload from '@/components/ImageUpload.vue'

const events = ref([])
const loading = ref(false)
const page = ref(1)
const totalPages = ref(1)
const filters = reactive({ keyword: '', status: '' })
const showForm = ref(false)
const editingId = ref(null)
const formLoading = ref(false)
const formError = ref('')
const deletingEvent = ref(null)

const statusOptions = [
  { value: 0, label: '草稿（不公開）' },
  { value: 1, label: '✓ 已發布（公開顯示）' },
  { value: 4, label: '已取消' },
]

// 預設日期為今天
function todayStr() {
  const now = new Date()
  const pad = n => String(n).padStart(2, '0')
  return `${now.getFullYear()}-${pad(now.getMonth()+1)}-${pad(now.getDate())}T${pad(now.getHours())}:${pad(now.getMinutes())}`
}

function emptyForm() {
  const today = todayStr()
  return {
    title: '', description: '', event_type: 1, location: '',
    cover_url: '', cover_path: '', fee: 0, max_participants: null,
    status: 1,           // 預設「已發布」
    start_at_str:     today,
    end_at_str:       today,
    reg_start_at_str: today,
    reg_end_at_str:   today,
  }
}

const form = reactive(emptyForm())

async function fetchEvents() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: 20 }
    if (filters.keyword) params.keyword = filters.keyword
    if (filters.status !== '') params.status = filters.status
    const { data } = await api.get('/events', { params })
    events.value     = data.events || []
    totalPages.value = data.pages  || 1
  } catch(e) {
    console.error('fetchEvents error', e)
  } finally {
    loading.value = false
  }
}

function goPage(p) { page.value = p; fetchEvents() }

function openCreate() {
  Object.assign(form, emptyForm())
  editingId.value = null
  formError.value = ''
  showForm.value = true
}

function openEdit(ev) {
  Object.assign(form, {
    title: ev.title, description: ev.description || '',
    event_type: ev.event_type, location: ev.location,
    cover_url: ev.cover_url || '', cover_path: ev.cover_url || '',
    fee: ev.fee, max_participants: ev.max_participants, status: ev.status,
    start_at_str:     toLocal(ev.start_at),
    end_at_str:       toLocal(ev.end_at),
    reg_start_at_str: toLocal(ev.reg_start_at),
    reg_end_at_str:   toLocal(ev.reg_end_at),
  })
  editingId.value = ev.id
  formError.value = ''
  showForm.value = true
}

async function submitForm() {
  formError.value = ''
  if (!form.title)            { formError.value = '請填寫賽事名稱'; return }
  if (!form.location)         { formError.value = '請填寫地點'; return }
  if (!form.start_at_str)     { formError.value = '請選擇賽事日期'; return }
  if (!form.reg_start_at_str) { formError.value = '請選擇報名開始日期'; return }
  if (!form.reg_end_at_str)   { formError.value = '請選擇報名截止日期'; return }

  const toISO = s => s ? new Date(s).toISOString() : null
  const payload = {
    title:       form.title,
    description: form.description,
    event_type:  form.event_type,
    location:    form.location,
    // cover_path 是 objectPath（如 "events/xxx.jpg"），組合成完整 URL
    cover_url: form.cover_path
      ? (form.cover_path.startsWith('http')
          ? form.cover_path
          : `${import.meta.env.VITE_IMAGE_BASE_URL}/images/${form.cover_path}`)
      : '',
    fee:         form.fee || 0,
    max_participants: form.max_participants || null,
    status:      form.status,
    start_at:     toISO(form.start_at_str),
    end_at:       toISO(form.end_at_str || form.start_at_str),
    reg_start_at: toISO(form.reg_start_at_str),
    reg_end_at:   toISO(form.reg_end_at_str),
  }

  formLoading.value = true
  try {
    if (editingId.value) {
      await api.put(`/events/${editingId.value}`, payload)
    } else {
      await api.post('/events', payload)
    }
    showForm.value = false
    await fetchEvents()
  } catch(e) {
    formError.value = e.response?.data?.error || '儲存失敗，請確認所有欄位並重試'
  } finally {
    formLoading.value = false
  }
}

function confirmDelete(ev) { deletingEvent.value = ev }
async function doDelete() {
  try {
    await api.delete(`/events/${deletingEvent.value.id}`)
    deletingEvent.value = null
    await fetchEvents()
  } catch(e) {
    alert(e.response?.data?.error || '刪除失敗')
  }
}

// datetime-local format: YYYY-MM-DDTHH:mm
function toLocal(d) {
  if (!d) return ''
  const dt = new Date(d)
  const pad = n => String(n).padStart(2, '0')
  return `${dt.getFullYear()}-${pad(dt.getMonth()+1)}-${pad(dt.getDate())}T${pad(dt.getHours())}:${pad(dt.getMinutes())}`
}

function fmt(d) {
  return d ? new Date(d).toLocaleDateString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit' }) : '-'
}

const typeMap  = { 1:'鐵人三項', 2:'路跑', 3:'游泳', 4:'單車', 5:'訓練', 6:'其他' }
const typeIconMap = { 1:'🏊', 2:'🏃', 3:'🏊', 4:'🚴', 5:'💪', 6:'🏅' }
function typeLabel(t) { return typeMap[t]  || '其他' }
function typeIcon(t)  { return typeIconMap[t] || '🏅' }
function statusLabel(s) { return { 0:'草稿', 1:'已發布', 2:'額滿', 3:'已結束', 4:'已取消' }[s] || '未知' }
function statusBadge(s) { return { 0:'badge-gray', 1:'badge-success', 2:'badge-warning', 3:'badge-gray', 4:'badge-danger' }[s] || 'badge-gray' }

onMounted(fetchEvents)
</script>

<style scoped>
.filter-row { display:flex; gap:.75rem; flex-wrap:wrap; align-items:center; }
.filter-row input, .filter-row select { height:36px; font-size:.85rem; }
.event-thumb { width:60px; height:40px; border-radius:4px; background:var(--bg); background-size:cover; background-position:center; display:flex; align-items:center; justify-content:center; font-size:1.2rem; border:1px solid var(--border); }
.fw-bold { font-weight:600; font-size:.9rem; }
.text-xs { font-size:.78rem; }
.action-btns { display:flex; gap:.35rem; flex-wrap:wrap; }
.loading-row { padding:3rem; text-align:center; color:var(--gray-2); }
.pagination { display:flex; align-items:center; justify-content:center; gap:1rem; padding:1rem; border-top:1px solid var(--border); }

/* Modal */
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.75); z-index:100; display:flex; align-items:center; justify-content:center; padding:1rem; }
.event-form-modal { background:var(--bg-card); border:1px solid var(--border); border-radius:8px; width:100%; max-width:700px; max-height:92vh; overflow-y:auto; }
.modal-header { display:flex; align-items:center; justify-content:space-between; padding:1.25rem 1.5rem; border-bottom:1px solid var(--border); position:sticky; top:0; background:var(--bg-card); z-index:1; }
.modal-header h3 { font-family:var(--font-c); font-size:1.1rem; font-weight:700; }
.modal-header button { background:none; border:none; color:var(--gray-2); font-size:1.2rem; cursor:pointer; }
.modal-body { padding:1.5rem; }
.form-grid { display:grid; grid-template-columns:1fr 1fr; gap:1rem; }
.form-group { display:flex; flex-direction:column; gap:.35rem; }
.form-group.full { grid-column:1/-1; }
.form-group label { font-size:.75rem; font-weight:600; text-transform:uppercase; letter-spacing:.06em; color:var(--gray-1); }
.form-group input, .form-group select { width:100%; }
.req { color:var(--primary); }
.hint { font-size:.72rem; color:var(--gray-2); margin-top:.2rem; }
.form-textarea { width:100%; background:var(--bg); color:#fff; border:1px solid var(--border); border-radius:4px; padding:.6rem .9rem; font-family:inherit; font-size:.85rem; resize:vertical; }
.form-textarea:focus { outline:none; border-color:var(--primary); }

/* Cover preview */
.cover-preview { margin-top:.5rem; border-radius:4px; overflow:hidden; max-height:120px; }
.cover-preview img { width:100%; height:120px; object-fit:cover; }

/* Status toggle */
.status-toggle-row { display:flex; gap:.5rem; flex-wrap:wrap; }
.status-btn { padding:.45rem 1rem; border-radius:4px; border:1px solid var(--border); font-size:.82rem; font-weight:600; cursor:pointer; background:var(--bg); color:var(--gray-2); transition:all .15s; }
.status-btn:hover { border-color:var(--white); color:var(--white); }
.status-btn.active { background:var(--primary); border-color:var(--primary); color:#fff; }

.form-error { background:rgba(239,68,68,.1); border:1px solid rgba(239,68,68,.3); border-radius:4px; color:#fca5a5; font-size:.83rem; padding:.6rem .9rem; margin-top:.75rem; }
.modal-footer { display:flex; gap:.75rem; margin-top:1.5rem; padding-top:1rem; border-top:1px solid var(--border); }
.spinner { width:14px; height:14px; border:2px solid rgba(255,255,255,.3); border-top-color:#fff; border-radius:50%; animation:spin .7s linear infinite; display:inline-block; }
@keyframes spin { to { transform:rotate(360deg) } }

.confirm-modal { background:var(--bg-card); border:1px solid var(--border); border-radius:8px; padding:2rem; max-width:400px; width:100%; }
.confirm-modal h3 { font-size:1.1rem; margin-bottom:.75rem; }
.confirm-actions { display:flex; gap:.75rem; margin-top:1.5rem; }
.btn-danger { background:var(--danger); color:#fff; }
.btn-danger:hover { opacity:.85; }
</style>
