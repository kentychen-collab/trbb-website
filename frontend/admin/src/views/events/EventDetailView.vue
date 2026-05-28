<template>
  <div>
    <div class="page-header">
      <RouterLink to="/events" class="back-link">← 返回賽事列表</RouterLink>
      <div class="flex justify-between items-center mt-1">
        <div>
          <h1 class="page-title">{{ event?.title || '載入中...' }}</h1>
          <p class="page-subtitle" v-if="event">
            {{ fmt(event.start_at) }} · {{ event.location }} ·
            <span class="badge" :class="statusBadge(event.status)">{{ statusLabel(event.status) }}</span>
          </p>
        </div>
        <!-- 匯出按鈕 -->
        <div class="export-btns" v-if="registrations.length">
          <button class="btn btn-ghost btn-sm" @click="exportFile('csv')">⬇ CSV</button>
          <button class="btn btn-ghost btn-sm" @click="exportFile('xlsx')">⬇ XLSX</button>
          <button class="btn btn-ghost btn-sm" @click="printList">🖨 列印／PDF</button>
        </div>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-row mb-2" v-if="event">
      <div class="stat-chip"><span class="stat-num">{{ registrations.length }}</span><span class="stat-lbl">總報名數</span></div>
      <div class="stat-chip"><span class="stat-num confirmed">{{ countByStatus(1) }}</span><span class="stat-lbl">已確認</span></div>
      <div class="stat-chip"><span class="stat-num pending">{{ countByStatus(0) }}</span><span class="stat-lbl">待確認</span></div>
      <div class="stat-chip"><span class="stat-num cancelled">{{ countByStatus(2) }}</span><span class="stat-lbl">已取消</span></div>
    </div>

    <!-- Table -->
    <div class="card" id="print-area">
      <div class="card-body" style="padding:0">
        <div v-if="loading" class="loading-row">載入中...</div>
        <div v-else-if="!registrations.length" class="loading-row text-gray">尚無報名資料</div>
        <table v-else class="table" style="font-size:.82rem">
          <thead>
            <tr>
              <th>#</th>
              <th>姓名</th>
              <th>手機 / Email</th>
              <th>性別 / 生日</th>
              <th>衣服 / 飲食</th>
              <th>緊急聯絡人</th>
              <th>付款確認</th>
              <th class="no-print">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(reg, idx) in registrations" :key="reg.id">
              <td>{{ idx+1 }}</td>
              <td>
                <div class="fw-bold">{{ reg.reg_name_zh }}</div>
                <div class="text-gray">{{ reg.reg_name_en }}</div>
                <div class="text-gray" style="font-size:.72rem">@{{ reg.username }}</div>
              </td>
              <td>
                <div>{{ reg.reg_phone }}</div>
                <div class="text-gray">{{ reg.reg_email }}</div>
              </td>
              <td>
                <div>{{ genderLabel(reg.reg_gender) }}</div>
                <div class="text-gray">{{ reg.reg_birthday || '-' }}</div>
              </td>
              <td>
                <div>{{ reg.reg_shirt_size || '-' }}</div>
                <div class="text-gray">{{ foodLabel(reg.reg_food_type) }}</div>
              </td>
              <td>
                <div>{{ reg.reg_emergency_contact }}</div>
                <div class="text-gray">{{ reg.reg_emergency_phone }}</div>
                <div class="text-gray" style="font-size:.72rem">{{ reg.reg_emergency_relation }}</div>
              </td>
              <td>
                <!-- 付款確認狀態 + 切換 -->
                <div class="payment-status-cell">
                  <span class="badge" :class="regStatusBadge(reg.status)">{{ regStatusLabel(reg.status) }}</span>
                </div>
              </td>
              <td class="no-print">
                <div class="action-btns">
                  <!-- 快速狀態切換 -->
                  <button v-if="reg.status === 0"
                    class="btn btn-sm btn-success" @click="quickStatus(reg, 1)"
                    title="確認付款">✓ 確認</button>
                  <button v-if="reg.status === 1"
                    class="btn btn-sm btn-ghost" @click="quickStatus(reg, 0)"
                    title="取消確認">↩ 待確認</button>
                  <button v-if="reg.status < 2"
                    class="btn btn-sm btn-ghost" style="color:var(--danger)"
                    @click="quickStatus(reg, 2)" title="取消報名">✗ 取消</button>
                  <!-- 編輯詳細資料 -->
                  <button class="btn btn-sm btn-ghost" @click="openEdit(reg)">編輯</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ── Edit Modal ──────────────────────────────────────── -->
    <div class="modal-overlay" v-if="editingReg" @click.self="editingReg=null">
      <div class="edit-modal">
        <div class="modal-header">
          <h3>編輯報名資料</h3>
          <button @click="editingReg=null">✕</button>
        </div>
        <div class="modal-body">
          <!-- 付款確認狀態 -->
          <div class="form-group mb-1">
            <label>付款確認狀態</label>
            <div class="status-toggle-row">
              <button v-for="opt in regStatusOptions" :key="opt.value"
                class="status-btn" :class="{ active: editForm.status === opt.value }"
                @click="editForm.status = opt.value" type="button">
                {{ opt.label }}
              </button>
            </div>
          </div>

          <div class="form-grid">
            <div class="form-group"><label>中文姓名</label><input v-model="editForm.name_zh" /></div>
            <div class="form-group"><label>英文姓名</label><input v-model="editForm.name_en" /></div>
            <div class="form-group"><label>手機</label><input v-model="editForm.phone" /></div>
            <div class="form-group"><label>Email</label><input v-model="editForm.email" /></div>
            <div class="form-group">
              <label>衣服尺寸</label>
              <select v-model="editForm.shirt_size">
                <option value="">-</option>
                <option v-for="s in ['XS','S','M','L','XL','2XL','3XL']" :key="s" :value="s">{{ s }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>飲食</label>
              <select v-model="editForm.food_type">
                <option :value="null">-</option>
                <option :value="1">葷</option><option :value="2">素</option><option :value="3">全素</option>
              </select>
            </div>
            <div class="form-group"><label>緊急聯絡人</label><input v-model="editForm.emergency_contact" /></div>
            <div class="form-group"><label>緊急聯絡電話</label><input v-model="editForm.emergency_phone" /></div>
            <div class="form-group full"><label>備註</label><input v-model="editForm.note" /></div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-primary" @click="saveEdit" :disabled="editLoading">
              {{ editLoading ? '儲存中...' : '儲存' }}
            </button>
            <button class="btn btn-ghost" @click="editingReg=null">取消</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import api from '@/services/api'

const route = useRoute()
const eventId = route.params.id

const event = ref(null)
const registrations = ref([])
const loading = ref(true)
const editingReg = ref(null)
const editLoading = ref(false)

const regStatusOptions = [
  { value: 0, label: '⏳ 待確認' },
  { value: 1, label: '✓ 已確認' },
  { value: 2, label: '✗ 已取消' },
  { value: 3, label: '↩ 已退款' },
]

const editForm = reactive({
  name_zh:'', name_en:'', id_number:'', passport_number:'',
  gender:null, birthday:'', phone:'', email:'',
  shirt_size:'', food_type:null, address:'',
  emergency_contact:'', emergency_phone:'', emergency_relation:'',
  note:'', status: 0,
})

async function fetchData() {
  loading.value = true
  try {
    const [evRes, regRes] = await Promise.all([
      api.get(`/events/${eventId}`),
      api.get(`/events/${eventId}/registrations`),
    ])
    event.value = evRes.data
    registrations.value = regRes.data.registrations || []
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

// 快速切換狀態（無需開 modal）
async function quickStatus(reg, newStatus) {
  try {
    await api.put(`/events/${eventId}/registrations/${reg.id}`, {
      ...regToPayload(reg),
      status: newStatus,
    })
    reg.status = newStatus
  } catch(e) {
    alert(e.response?.data?.error || '操作失敗')
  }
}

function regToPayload(reg) {
  return {
    name_zh: reg.reg_name_zh, name_en: reg.reg_name_en,
    id_number: reg.reg_id_number, passport_number: reg.reg_passport_number,
    gender: reg.reg_gender, birthday: reg.reg_birthday,
    phone: reg.reg_phone, email: reg.reg_email,
    shirt_size: reg.reg_shirt_size, food_type: reg.reg_food_type,
    address: reg.reg_address,
    emergency_contact: reg.reg_emergency_contact,
    emergency_phone: reg.reg_emergency_phone,
    emergency_relation: reg.reg_emergency_relation,
    note: reg.note,
  }
}

function openEdit(reg) {
  editingReg.value = reg
  Object.assign(editForm, {
    name_zh: reg.reg_name_zh, name_en: reg.reg_name_en,
    id_number: reg.reg_id_number, passport_number: reg.reg_passport_number,
    gender: reg.reg_gender, birthday: reg.reg_birthday,
    phone: reg.reg_phone, email: reg.reg_email,
    shirt_size: reg.reg_shirt_size, food_type: reg.reg_food_type,
    address: reg.reg_address,
    emergency_contact: reg.reg_emergency_contact,
    emergency_phone: reg.reg_emergency_phone,
    emergency_relation: reg.reg_emergency_relation,
    note: reg.note, status: reg.status,
  })
}

async function saveEdit() {
  editLoading.value = true
  try {
    await api.put(`/events/${eventId}/registrations/${editingReg.value.id}`, editForm)
    // 更新本地資料
    const idx = registrations.value.findIndex(r => r.id === editingReg.value.id)
    if (idx >= 0) {
      registrations.value[idx] = {
        ...registrations.value[idx],
        status: editForm.status,
        reg_name_zh: editForm.name_zh, reg_name_en: editForm.name_en,
        reg_phone: editForm.phone, reg_email: editForm.email,
        reg_shirt_size: editForm.shirt_size, reg_food_type: editForm.food_type,
        reg_emergency_contact: editForm.emergency_contact,
        reg_emergency_phone: editForm.emergency_phone,
        reg_emergency_relation: editForm.emergency_relation,
        note: editForm.note,
      }
    }
    editingReg.value = null
  } catch(e) {
    alert(e.response?.data?.error || '更新失敗')
  } finally {
    editLoading.value = false
  }
}

// ── 匯出 ──────────────────────────────────────────────────
function exportFile(format) {
  const token = localStorage.getItem('trbb_admin_token')
  const eventTitle = event.value?.title || eventId
  const url = `/v1/admin/events/${eventId}/registrations/export?format=${format}`
  fetch(url, { headers: { Authorization: `Bearer ${token}` } })
    .then(r => r.blob())
    .then(blob => {
      const ext = format === 'xlsx' ? 'xlsx' : 'csv'
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = `${eventTitle}_報名名單.${ext}`
      a.click()
      URL.revokeObjectURL(a.href)
    })
    .catch(() => alert('下載失敗'))
}

function printList() {
  // 建立列印專用 iframe，只列印名單區域
  const printArea = document.getElementById('print-area')
  if (!printArea) return

  const iframe = document.createElement('iframe')
  iframe.style.cssText = 'position:fixed;top:-9999px;left:-9999px;width:1px;height:1px'
  document.body.appendChild(iframe)

  const doc = iframe.contentDocument
  const eventTitle = event.value?.title || ''
  const now = new Date().toLocaleString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit', hour:'2-digit', minute:'2-digit' })

  doc.open()
  doc.write(`<!DOCTYPE html>
<html><head>
<meta charset="UTF-8">
<title>${eventTitle}_報名名單</title>
<style>
  * { margin:0; padding:0; box-sizing:border-box; }
  body { font-family: 'PingFang TC', 'Microsoft JhengHei', sans-serif; font-size:10px; color:#000; }
  h2 { font-size:14px; margin-bottom:4px; }
  .meta { font-size:9px; color:#555; margin-bottom:10px; }
  table { width:100%; border-collapse:collapse; }
  th { background:#1F3864; color:#fff; padding:5px 4px; text-align:center; font-size:9px; }
  td { border:1px solid #ccc; padding:4px; font-size:9px; vertical-align:top; }
  tr:nth-child(even) td { background:#f9f9f9; }
  .badge-confirmed { color:#1a7a1a; font-weight:bold; }
  .badge-pending   { color:#7b5e00; }
  .badge-cancel    { color:#888; }
          { display:none; }
  @page { margin:1.5cm; size:A4 landscape; }
</style>
</head><body>
<h2>${eventTitle}｜報名名單</h2>
<p class="meta">共 ${registrations.value.length} 筆報名　匯出時間：${now}</p>
${printArea.querySelector('table').outerHTML}
</body></html>`)
  doc.close()

  iframe.onload = () => {
    iframe.contentWindow.print()
    setTimeout(() => document.body.removeChild(iframe), 1000)
  }
}

// ── Helpers ──────────────────────────────────────────────
function countByStatus(s) { return registrations.value.filter(r => r.status === s).length }
function fmt(d) { return d ? new Date(d).toLocaleDateString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit' }) : '-' }
function fmtFull(d) { return d ? new Date(d).toLocaleString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit', hour:'2-digit', minute:'2-digit' }) : '-' }
function genderLabel(g) { return {1:'男',2:'女',3:'其他'}[g] || '-' }
function foodLabel(f) { return {1:'葷',2:'素',3:'全素'}[f] || '-' }
function statusLabel(s) { return {0:'草稿',1:'已發布',2:'額滿',3:'已結束',4:'已取消'}[s] || '未知' }
function statusBadge(s) { return {0:'badge-gray',1:'badge-success',2:'badge-warning',3:'badge-gray',4:'badge-danger'}[s] || 'badge-gray' }
function regStatusLabel(s) { return {0:'待確認',1:'已確認',2:'已取消',3:'已退款'}[s] || '未知' }
function regStatusBadge(s) { return {0:'badge-warning',1:'badge-success',2:'badge-gray',3:'badge-gray'}[s] || 'badge-gray' }

onMounted(fetchData)
</script>

<style scoped>
.back-link { font-size:.85rem; color:var(--gray-2); transition:color .15s; }
.back-link:hover { color:var(--primary); }
.export-btns { display:flex; gap:.5rem; }

.stats-row { display:flex; gap:.75rem; flex-wrap:wrap; }
.stat-chip { background:var(--bg-card); border:1px solid var(--border); border-radius:4px; padding:.5rem 1.25rem; display:flex; flex-direction:column; align-items:center; gap:.15rem; }
.stat-num { font-family:var(--font-c); font-size:1.5rem; font-weight:700; }
.stat-num.confirmed { color:var(--success); }
.stat-num.pending { color:var(--warning); }
.stat-num.cancelled { color:var(--gray-2); }
.stat-lbl { font-size:.7rem; color:var(--gray-2); text-transform:uppercase; letter-spacing:.1em; }

.fw-bold { font-weight:600; }
.loading-row { padding:3rem; text-align:center; color:var(--gray-2); }

/* 付款狀態欄 */
.payment-status-cell { display:flex; flex-direction:column; gap:.3rem; }

/* 操作按鈕 */
.action-btns { display:flex; gap:.35rem; flex-wrap:wrap; }
.btn-success { background:var(--success); color:#fff; }
.btn-success:hover { opacity:.85; }

/* Modal */
.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.75); z-index:100; display:flex; align-items:center; justify-content:center; padding:1rem; }
.edit-modal { background:var(--bg-card); border:1px solid var(--border); border-radius:8px; width:100%; max-width:600px; max-height:90vh; overflow-y:auto; }
.modal-header { display:flex; align-items:center; justify-content:space-between; padding:1.25rem 1.5rem; border-bottom:1px solid var(--border); position:sticky; top:0; background:var(--bg-card); }
.modal-header h3 { font-family:var(--font-c); font-size:1.1rem; font-weight:700; }
.modal-header button { background:none; border:none; color:var(--gray-2); font-size:1.2rem; cursor:pointer; }
.modal-body { padding:1.5rem; }
.form-group { display:flex; flex-direction:column; gap:.3rem; }
.form-group.mb-1 { margin-bottom:.75rem; }
.form-group label { font-size:.72rem; font-weight:600; text-transform:uppercase; letter-spacing:.06em; color:var(--gray-1); }
.form-group input, .form-group select { width:100%; }
.form-grid { display:grid; grid-template-columns:1fr 1fr; gap:.75rem; }
.form-group.full { grid-column:1/-1; }
.modal-footer { display:flex; gap:.75rem; margin-top:1.25rem; padding-top:1rem; border-top:1px solid var(--border); }

/* 狀態 toggle */
.status-toggle-row { display:flex; gap:.5rem; flex-wrap:wrap; }
.status-btn { padding:.4rem 1rem; border-radius:4px; border:1px solid var(--border); font-size:.82rem; font-weight:600; cursor:pointer; background:var(--bg); color:var(--gray-2); transition:all .15s; }
.status-btn:hover { border-color:var(--white); color:var(--white); }
.status-btn.active { background:var(--primary); border-color:var(--primary); color:#fff; }
</style>
