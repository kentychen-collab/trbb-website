<template>
  <div>
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="page-title">一般會員</h1>
        <p class="page-subtitle">管理前台申請的一般會員</p>
      </div>
      <button class="btn btn-primary" @click="openCreate">＋ 新增會員</button>
    </div>

    <!-- Filters -->
    <div class="card mb-2">
      <div class="card-body" style="padding:1rem">
        <div class="filter-row">
          <input v-model="filters.keyword" placeholder="搜尋姓名 / Email / 手機..."
            @keyup.enter="fetchUsers" style="flex:1;min-width:180px" />
          <select v-model="filters.status" @change="fetchUsers">
            <option value="">全部狀態</option>
            <option value="0">待審核</option>
            <option value="1">已啟用</option>
            <option value="2">已停用</option>
            <option value="3">已拒絕</option>
          </select>
          <button class="btn btn-primary btn-sm" @click="fetchUsers">搜尋</button>
          <button class="btn btn-ghost btn-sm" @click="resetFilters">重設</button>
        </div>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-bar mb-2">
      <div class="stat-chip" v-for="s in quickStats" :key="s.label"
        :class="{ active: filters.status === s.val }" @click="setStatus(s.val)">
        <span class="stat-chip-num">{{ s.count }}</span>
        <span class="stat-chip-label">{{ s.label }}</span>
      </div>
    </div>

    <!-- Table -->
    <div class="card">
      <div class="card-body" style="padding:0">
        <div v-if="loading" class="loading-row">載入中...</div>
        <div v-else-if="!users.length" class="loading-row text-gray">查無資料</div>
        <table v-else class="table">
          <thead>
            <tr>
              <th>會員</th>
              <th>姓名</th>
              <th>Email / 手機</th>
              <th>狀態</th>
              <th>申請時間</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id">
              <td>
                <div class="user-cell">
                  <div class="avatar">{{ (u.display_name || u.username)[0] }}</div>
                  <div class="text-xs text-gray">@{{ u.username }}</div>
                </div>
              </td>
              <td>
                <div v-if="u.name_zh" class="fw-bold">{{ u.name_zh }}</div>
                <div v-if="u.name_en" class="text-gray text-xs">{{ u.name_en }}</div>
                <div v-if="u.display_name && u.display_name !== u.name_zh && u.display_name !== u.name_en"
                  class="text-gray text-xs">（{{ u.display_name }}）</div>
              </td>
              <td>
                <div>{{ u.email }}</div>
                <div class="text-gray text-xs">{{ u.phone }}</div>
              </td>
              <td><span class="badge" :class="statusBadge(u.status)">{{ statusLabel(u.status) }}</span></td>
              <td class="text-gray text-xs">{{ fmt(u.created_at) }}</td>
              <td>
                <div class="action-btns">
                  <button v-if="u.status === 0" class="btn btn-sm btn-primary" @click="updateStatus(u, 1)">✓ 核准</button>
                  <button v-if="u.status === 0" class="btn btn-sm btn-danger"  @click="updateStatus(u, 3)">✗ 拒絕</button>
                  <button v-if="u.status === 1" class="btn btn-sm btn-ghost"   @click="updateStatus(u, 2)">停用</button>
                  <button v-if="u.status === 2" class="btn btn-sm btn-ghost"   @click="updateStatus(u, 1)">恢復</button>
                  <button class="btn btn-sm btn-ghost" @click="openEdit(u)">編輯</button>
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

    <!-- Create / Edit Modal -->
    <div class="modal-overlay" v-if="editingUser || showCreate" @click.self="closeModal">
      <div class="edit-modal">
        <div class="modal-header">
          <h3>{{ showCreate ? '新增會員' : '編輯會員' }}</h3>
          <button @click="closeModal">✕</button>
        </div>
        <div class="modal-body">

          <!-- 新增專用欄位 -->
          <template v-if="showCreate">
            <div class="section-label">帳號資訊</div>
            <div class="form-row">
              <div class="form-group">
                <label>會員 ID <span class="req">*</span></label>
                <input v-model="editForm.username" placeholder="英數字及底線，3~50" />
              </div>
              <div class="form-group">
                <label>Email <span class="req">*</span></label>
                <input v-model="editForm.email" type="email" />
              </div>
            </div>
            <div class="form-group mb-1">
              <label>初始密碼 <span class="req">*</span></label>
              <input v-model="editForm.password" type="password" placeholder="至少 8 字元" />
            </div>
          </template>

          <!-- 共用：姓名 -->
          <div class="section-label">姓名資料</div>
          <div class="form-row">
            <div class="form-group">
              <label>顯示名稱</label>
              <input v-model="editForm.display_name" placeholder="暱稱" />
            </div>
            <div class="form-group">
              <label>中文姓名</label>
              <input v-model="editForm.name_zh" placeholder="真實中文姓名" />
            </div>
          </div>
          <div class="form-group mb-1">
            <label>英文姓名</label>
            <input v-model="editForm.name_en" placeholder="English Name" />
          </div>

          <!-- 個人資訊 -->
          <div class="section-label">個人資訊</div>
          <div class="form-row">
            <div class="form-group">
              <label>手機</label>
              <input v-model="editForm.phone" type="tel" />
            </div>
            <div class="form-group">
              <label>性別</label>
              <select v-model="editForm.gender">
                <option :value="null">-</option>
                <option :value="1">男</option>
                <option :value="2">女</option>
                <option :value="3">其他</option>
              </select>
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>出生年月日</label>
              <input v-model="editForm.birthday" type="date" />
            </div>
            <div class="form-group">
              <label>身份證字號</label>
              <input v-model="editForm.id_number" placeholder="A123456789" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>護照號碼</label>
              <input v-model="editForm.passport_number" />
            </div>
            <div class="form-group">
              <label>通訊地址</label>
              <input v-model="editForm.address" />
            </div>
          </div>

          <!-- 偏好 -->
          <div class="section-label">偏好設定</div>
          <div class="form-row">
            <div class="form-group">
              <label>衣服尺寸</label>
              <select v-model="editForm.shirt_size">
                <option value="">-</option>
                <option v-for="s in ['XS','S','M','L','XL','2XL','3XL']" :key="s" :value="s">{{ s }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>飲食習慣</label>
              <select v-model="editForm.food_type">
                <option :value="null">-</option>
                <option :value="1">葷</option><option :value="2">素</option><option :value="3">全素</option>
              </select>
            </div>
          </div>

          <!-- 緊急聯絡人 -->
          <div class="section-label">緊急聯絡人</div>
          <div class="form-row">
            <div class="form-group">
              <label>聯絡人姓名</label>
              <input v-model="editForm.emergency_contact" />
            </div>
            <div class="form-group">
              <label>聯絡人手機</label>
              <input v-model="editForm.emergency_phone" />
            </div>
          </div>
          <div class="form-group mb-1">
            <label>與本人關係</label>
            <input v-model="editForm.emergency_relation" placeholder="例如：配偶、父母、朋友等" />
          </div>

          <!-- 修改密碼（編輯時） -->
          <template v-if="!showCreate">
            <div class="section-label" style="border-top:1px solid var(--border);margin-top:.5rem;padding-top:.75rem">
              修改密碼（留空不修改）
            </div>
            <div class="form-group mb-1">
              <label>新密碼</label>
              <input v-model="editForm.new_password" type="password" placeholder="至少 8 字元，留空不修改" />
            </div>
          </template>

          <div v-if="modalError" class="form-error mt-1">{{ modalError }}</div>
          <div class="modal-footer">
            <button class="btn btn-primary" @click="saveModal" :disabled="modalLoading">
              {{ modalLoading ? '儲存中...' : (showCreate ? '建立會員' : '儲存變更') }}
            </button>
            <button class="btn btn-ghost" @click="closeModal">取消</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import api from '@/services/api'

const users = ref([])
const loading = ref(false)
const page = ref(1)
const totalPages = ref(1)
const filters = reactive({ keyword: '', status: '' })
const editingUser = ref(null)
const showCreate = ref(false)
const modalLoading = ref(false)
const modalError = ref('')

const quickStats = ref([
  { label:'全部', val:'', count:0 },
  { label:'待審核', val:'0', count:0 },
  { label:'已啟用', val:'1', count:0 },
  { label:'已停用', val:'2', count:0 },
])

const emptyForm = () => ({
  username:'', email:'', password:'',
  display_name:'', name_zh:'', name_en:'',
  phone:'', gender:null, birthday:'',
  id_number:'', passport_number:'', address:'',
  shirt_size:'', food_type:null,
  emergency_contact:'', emergency_phone:'', emergency_relation:'',
  new_password:'',
})
const editForm = reactive(emptyForm())

async function fetchUsers() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: 20 }
    if (filters.keyword) params.keyword = filters.keyword
    if (filters.status !== '') params.status = filters.status
    const { data } = await api.get('/members', { params })
    users.value = data.users || []
    totalPages.value = data.pages || 1
    updateStats()
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

async function updateStats() {
  try {
    const all = await api.get('/members', { params: { page_size:1 } })
    quickStats.value[0].count = all.data.total
    for (const [i, s] of [0,1,2].entries()) {
      const r = await api.get('/members', { params: { page_size:1, status:s } })
      quickStats.value[i+1].count = r.data.total
    }
  } catch {}
}

async function updateStatus(u, status) {
  try {
    await api.put(`/members/${u.id}/status`, { status })
    u.status = status
    await fetchUsers()
  } catch(e) { alert(e.response?.data?.error || '操作失敗') }
}

function openCreate() {
  Object.assign(editForm, emptyForm())
  modalError.value = ''
  showCreate.value = true
}

function openEdit(u) {
  Object.assign(editForm, {
    username: u.username, email: u.email, password: '',
    display_name: u.display_name || '', name_zh: u.name_zh || '', name_en: u.name_en || '',
    phone: u.phone || '', gender: u.gender, birthday: u.birthday || '',
    id_number: u.id_number || '', passport_number: u.passport_number || '',
    address: u.address || '',
    shirt_size: u.shirt_size || '', food_type: u.food_type,
    emergency_contact: u.emergency_contact || '',
    emergency_phone: u.emergency_phone || '',
    emergency_relation: u.emergency_relation || '',
    new_password: '',
  })
  editingUser.value = u
  modalError.value = ''
}

function closeModal() { editingUser.value = null; showCreate.value = false }

async function saveModal() {
  modalError.value = ''
  modalLoading.value = true
  try {
    if (showCreate.value) {
      if (!editForm.username || !editForm.email || !editForm.password) {
        modalError.value = '請填寫帳號、Email 和初始密碼'; return
      }
      await api.post('/members', {
        username: editForm.username, email: editForm.email,
        password: editForm.password,
        display_name: editForm.display_name || editForm.name_zh || editForm.name_en || editForm.username,
        phone: editForm.phone,
      })
    } else {
      await api.put(`/members/${editingUser.value.id}/profile`, editForm)
      if (editForm.new_password) {
        if (editForm.new_password.length < 8) { modalError.value = '密碼至少 8 字元'; return }
        await api.put(`/members/${editingUser.value.id}/password`, { password: editForm.new_password })
      }
    }
    closeModal()
    await fetchUsers()
  } catch(e) {
    modalError.value = e.response?.data?.error || '操作失敗'
  } finally {
    modalLoading.value = false
  }
}

function goPage(p) { page.value = p; fetchUsers() }
function setStatus(val) { filters.status = val; page.value = 1; fetchUsers() }
function resetFilters() { filters.keyword = ''; filters.status = ''; page.value = 1; fetchUsers() }
function fmt(d) { return d ? new Date(d).toLocaleDateString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit' }) : '-' }
function statusLabel(s) { return {0:'待審核',1:'已啟用',2:'已停用',3:'已拒絕'}[s]||'未知' }
function statusBadge(s) { return {0:'badge-warning',1:'badge-success',2:'badge-gray',3:'badge-danger'}[s]||'badge-gray' }

onMounted(fetchUsers)
</script>

<style scoped>
.filter-row { display:flex; gap:.75rem; flex-wrap:wrap; align-items:center; }
.filter-row input, .filter-row select { height:36px; font-size:.85rem; }
.stats-bar { display:flex; gap:.75rem; flex-wrap:wrap; }
.stat-chip { background:var(--bg-card); border:1px solid var(--border); border-radius:4px; padding:.5rem 1rem; display:flex; gap:.5rem; align-items:center; cursor:pointer; transition:all .15s; }
.stat-chip:hover, .stat-chip.active { border-color:var(--primary); }
.stat-chip.active .stat-chip-num { color:var(--primary); }
.stat-chip-num { font-family:var(--font-c); font-size:1.2rem; font-weight:700; }
.stat-chip-label { font-size:.78rem; color:var(--gray-2); }
.user-cell { display:flex; flex-direction:column; gap:.2rem; }
.avatar { width:28px; height:28px; border-radius:50%; background:var(--primary); display:flex; align-items:center; justify-content:center; font-weight:700; font-size:.8rem; margin-bottom:.2rem; }
.fw-bold { font-weight:600; font-size:.9rem; }
.text-xs { font-size:.75rem; }
.action-btns { display:flex; gap:.4rem; flex-wrap:wrap; }
.btn-danger { background:var(--danger); color:#fff; }
.btn-danger:hover { opacity:.85; }
.loading-row { padding:3rem; text-align:center; color:var(--gray-2); }
.pagination { display:flex; align-items:center; justify-content:center; gap:1rem; padding:1rem; border-top:1px solid var(--border); }

.modal-overlay { position:fixed; inset:0; background:rgba(0,0,0,.75); z-index:100; display:flex; align-items:center; justify-content:center; padding:1rem; }
.edit-modal { background:var(--bg-card); border:1px solid var(--border); border-radius:8px; width:100%; max-width:600px; max-height:92vh; overflow-y:auto; }
.modal-header { display:flex; align-items:center; justify-content:space-between; padding:1.25rem 1.5rem; border-bottom:1px solid var(--border); position:sticky; top:0; background:var(--bg-card); }
.modal-header h3 { font-family:var(--font-c); font-size:1.1rem; font-weight:700; }
.modal-header button { background:none; border:none; color:var(--gray-2); font-size:1.2rem; cursor:pointer; }
.modal-body { padding:1.5rem; }
.section-label { font-size:.7rem; font-weight:700; letter-spacing:.12em; text-transform:uppercase; color:var(--gray-2); margin:.75rem 0 .5rem; }
.form-row { display:grid; grid-template-columns:1fr 1fr; gap:.75rem; margin-bottom:.75rem; }
.form-group { display:flex; flex-direction:column; gap:.3rem; }
.form-group.mb-1 { margin-bottom:.75rem; }
.form-group label { font-size:.72rem; font-weight:600; text-transform:uppercase; letter-spacing:.06em; color:var(--gray-1); }
.form-group input, .form-group select { width:100%; }
.req { color:var(--primary); }
.form-error { background:rgba(239,68,68,.1); border:1px solid rgba(239,68,68,.3); border-radius:4px; color:#fca5a5; font-size:.83rem; padding:.5rem .75rem; }
.modal-footer { display:flex; gap:.75rem; margin-top:1.25rem; padding-top:1rem; border-top:1px solid var(--border); }
.mt-1 { margin-top:.5rem; }
</style>
