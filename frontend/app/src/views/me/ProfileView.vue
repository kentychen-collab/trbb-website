<template>
  <div class="profile-page">
    <div class="page-header">
      <h1 class="section-title">個人資料</h1>
      <p class="text-gray">完善資料後，報名賽事時可快速帶入</p>
    </div>

    <div v-if="pageLoading" class="loading-box">載入中...</div>

    <form v-else @submit.prevent="handleSave" novalidate>

      <!-- ── 帳號資訊（不可修改） ───────────────────────── -->
      <div class="form-section">
        <h3 class="form-section-title">帳號資訊 <span class="readonly-badge">不可修改</span></h3>
        <div class="form-row-2">
          <div class="form-group">
            <label>會員 ID</label>
            <input :value="auth.user?.username" disabled class="disabled-input" />
          </div>
          <div class="form-group">
            <label>Email（登入帳號）</label>
            <input :value="auth.user?.email" disabled class="disabled-input" />
          </div>
        </div>
      </div>

      <!-- ── 姓名資料 ──────────────────────────────────── -->
      <div class="form-section">
        <h3 class="form-section-title">姓名資料</h3>
        <div class="form-row-3">
          <div class="form-group">
            <label>暱稱 / 顯示名稱</label>
            <input v-model="form.display_name" placeholder="顯示於社群的名稱" />
          </div>
          <div class="form-group">
            <label>中文姓名</label>
            <input v-model="form.name_zh" placeholder="真實中文姓名（選填）" />
          </div>
          <div class="form-group">
            <label>英文姓名</label>
            <input v-model="form.name_en" placeholder="English Name（選填）" />
          </div>
        </div>
      </div>

      <!-- ── 個人資訊 ───────────────────────────────────── -->
      <div class="form-section">
        <h3 class="form-section-title">個人資訊</h3>
        <div class="form-row-3">
          <div class="form-group">
            <label>性別</label>
            <select v-model="form.gender">
              <option :value="null">請選擇</option>
              <option :value="1">男</option>
              <option :value="2">女</option>
              <option :value="3">其他</option>
            </select>
          </div>
          <div class="form-group">
            <label>出生年月日</label>
            <input v-model="form.birthday" type="date" />
          </div>
          <div class="form-group">
            <label>手機號碼 <span class="req">*</span></label>
            <input v-model="form.phone" type="tel" placeholder="09xxxxxxxx" />
          </div>
        </div>
        <div class="form-row-3">
          <div class="form-group">
            <label>身份證字號</label>
            <input v-model="form.id_number" placeholder="A123456789" />
          </div>
          <div class="form-group">
            <label>護照號碼</label>
            <input v-model="form.passport_number" placeholder="如有護照請填寫" />
          </div>
          <div class="form-group">
            <label>通訊地址</label>
            <input v-model="form.address" placeholder="縣市 + 完整地址" />
          </div>
        </div>
      </div>

      <!-- ── 偏好設定 ───────────────────────────────────── -->
      <div class="form-section">
        <h3 class="form-section-title">偏好設定</h3>
        <div class="form-row-2">
          <div class="form-group">
            <label>衣服尺寸</label>
            <select v-model="form.shirt_size">
              <option value="">請選擇</option>
              <option v-for="s in shirtSizes" :key="s" :value="s">{{ s }}</option>
            </select>
          </div>
          <div class="form-group">
            <label>飲食習慣</label>
            <select v-model="form.food_type">
              <option :value="null">請選擇</option>
              <option :value="1">葷食</option>
              <option :value="2">素食</option>
              <option :value="3">全素（純素）</option>
            </select>
          </div>
        </div>
      </div>

      <!-- ── 緊急聯絡人 ─────────────────────────────────── -->
      <div class="form-section">
        <h3 class="form-section-title">緊急聯絡人 <span class="optional-badge">選填</span></h3>
        <div class="form-row-3">
          <div class="form-group">
            <label>聯絡人姓名</label>
            <input v-model="form.emergency_contact" placeholder="緊急聯絡人姓名" />
          </div>
          <div class="form-group">
            <label>聯絡人手機</label>
            <input v-model="form.emergency_phone" type="tel" placeholder="09xxxxxxxx" />
          </div>
          <div class="form-group">
            <label>與本人關係</label>
            <input v-model="form.emergency_relation" placeholder="例如：配偶、父母、朋友等" />
          </div>
        </div>
      </div>

      <!-- ── 資料完整度 ─────────────────────────────────── -->
      <div class="completeness-bar">
        <div class="completeness-label">
          報名資料完整度
          <span :style="{ color: completenessColor }">{{ completeness }}%</span>
        </div>
        <div class="completeness-track">
          <div class="completeness-fill" :style="{ width: completeness + '%', background: completenessColor }"></div>
        </div>
        <div class="completeness-hint" v-if="completeness < 100">
          還差：{{ missingFields.join('、') }}
        </div>
      </div>

      <div v-if="error"   class="form-msg error">{{ error }}</div>
      <div v-if="success" class="form-msg success">✓ {{ success }}</div>

      <div class="form-actions">
        <button type="submit" class="btn btn-primary" :disabled="saving">
          <span v-if="saving" class="spinner"></span>
          {{ saving ? '儲存中...' : '儲存變更' }}
        </button>
        <button type="button" class="btn btn-ghost" @click="resetForm">重設</button>
      </div>
    </form>

    <!-- ── 修改密碼 ────────────────────────────────────── -->
    <div class="form-section mt-4">
      <h3 class="form-section-title">修改密碼</h3>
      <div class="form-row-3">
        <div class="form-group">
          <label>目前密碼 <span class="req">*</span></label>
          <input v-model="pwdForm.old_password" type="password" placeholder="輸入目前密碼" />
        </div>
        <div class="form-group">
          <label>新密碼 <span class="req">*</span></label>
          <input v-model="pwdForm.new_password" type="password" placeholder="至少 8 字元" />
        </div>
        <div class="form-group">
          <label>確認新密碼 <span class="req">*</span></label>
          <input v-model="pwdForm.confirm_password" type="password" placeholder="再次輸入新密碼" />
        </div>
      </div>
      <div v-if="pwdError"   class="form-msg error">{{ pwdError }}</div>
      <div v-if="pwdSuccess" class="form-msg success">✓ {{ pwdSuccess }}</div>
      <div class="form-actions">
        <button type="button" class="btn btn-primary" @click="handleChangePwd" :disabled="pwdSaving">
          <span v-if="pwdSaving" class="spinner"></span>
          {{ pwdSaving ? '更新中...' : '更新密碼' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'

const auth       = useAuthStore()
const saving     = ref(false)
const pageLoading = ref(true)
const error      = ref('')
const success    = ref('')
const pwdSaving  = ref(false)
const pwdError   = ref('')
const pwdSuccess = ref('')

const shirtSizes = ['XS', 'S', 'M', 'L', 'XL', '2XL', '3XL']

const form = reactive({
  display_name: '', name_zh: '', name_en: '',
  id_number: '', passport_number: '',
  gender: null, birthday: '', phone: '',
  shirt_size: '', food_type: null, address: '',
  emergency_contact: '', emergency_phone: '', emergency_relation: '',
})

const pwdForm = reactive({
  old_password: '', new_password: '', confirm_password: '',
})

// 完整度計算（緊急聯絡人改為選填，不計入必填）
const completenessFields = [
  { key: 'phone',              label: '手機號碼' },
  { key: 'id_number',          label: '身份證字號' },
  { key: 'gender',             label: '性別' },
  { key: 'birthday',           label: '出生年月日' },
  { key: 'shirt_size',         label: '衣服尺寸' },
  { key: 'food_type',          label: '飲食習慣' },
  { key: 'address',            label: '通訊地址' },
  { key: 'emergency_contact',  label: '緊急聯絡人姓名' },
  { key: 'emergency_phone',    label: '緊急聯絡人手機' },
  { key: 'emergency_relation', label: '緊急聯絡人關係' },
]
// 至少填一個姓名
const hasName = computed(() => !!(form.name_zh || form.name_en))

const missingFields = computed(() => {
  const missing = completenessFields
    .filter(f => !form[f.key] && form[f.key] !== 0)
    .map(f => f.label)
  if (!hasName.value) missing.unshift('姓名（中文或英文）')
  return missing
})
const completeness = computed(() => {
  const total = completenessFields.length + 1 // +1 for name
  const filled = total - missingFields.value.length
  return Math.round((filled / total) * 100)
})
const completenessColor = computed(() =>
  completeness.value >= 100 ? '#22c55e' : completeness.value >= 60 ? '#f59e0b' : '#ef4444'
)

function fillForm(user) {
  Object.keys(form).forEach(k => {
    if (user[k] !== undefined && user[k] !== null) form[k] = user[k]
  })
}

function resetForm() {
  if (auth.user) fillForm(auth.user)
  error.value = ''
  success.value = ''
}

async function handleSave() {
  error.value = ''
  success.value = ''
  if (!form.phone) { error.value = '請填寫手機號碼'; return }

  saving.value = true
  try {
    const { data } = await api.put('/me', form)
    auth.user = { ...auth.user, ...data.user }
    localStorage.setItem('trbb_user', JSON.stringify(auth.user))
    success.value = '個人資料已成功儲存'
    setTimeout(() => { success.value = '' }, 4000)
  } catch(e) {
    error.value = e.response?.data?.error || '儲存失敗，請稍後再試'
  } finally {
    saving.value = false
  }
}

async function handleChangePwd() {
  pwdError.value = ''
  pwdSuccess.value = ''
  if (!pwdForm.old_password) { pwdError.value = '請輸入目前密碼'; return }
  if (!pwdForm.new_password) { pwdError.value = '請輸入新密碼'; return }
  if (pwdForm.new_password.length < 8) { pwdError.value = '新密碼至少需 8 字元'; return }
  if (pwdForm.new_password !== pwdForm.confirm_password) { pwdError.value = '兩次輸入的新密碼不一致'; return }

  pwdSaving.value = true
  try {
    await api.put('/me/password', {
      old_password: pwdForm.old_password,
      new_password: pwdForm.new_password,
    })
    pwdSuccess.value = '密碼已更新成功'
    Object.assign(pwdForm, { old_password: '', new_password: '', confirm_password: '' })
    setTimeout(() => { pwdSuccess.value = '' }, 4000)
  } catch(e) {
    pwdError.value = e.response?.data?.error || '密碼更新失敗'
  } finally {
    pwdSaving.value = false
  }
}

onMounted(async () => {
  try {
    const { data } = await api.get('/me')
    auth.user = { ...auth.user, ...data }
    fillForm(data)
  } catch {}
  finally { pageLoading.value = false }
})
</script>

<style scoped>
.profile-page { max-width:800px; }
.page-header { margin-bottom:2rem; }
.loading-box { padding:4rem; text-align:center; color:var(--color-gray-2); }
.mt-4 { margin-top:2rem; }

.form-section { background:var(--color-bg-card); border:1px solid var(--color-border); border-radius:8px; padding:1.5rem; margin-bottom:1.25rem; }
.form-section-title { font-family:var(--font-cond); font-size:.85rem; font-weight:700; letter-spacing:.12em; text-transform:uppercase; color:var(--color-gray-2); margin-bottom:1.25rem; display:flex; align-items:center; gap:.75rem; }
.readonly-badge { font-size:.65rem; padding:.15rem .5rem; border-radius:3px; background:rgba(107,114,128,.15); color:var(--color-gray-2); letter-spacing:.05em; }
.optional-badge { font-size:.65rem; padding:.15rem .5rem; border-radius:3px; background:rgba(34,197,94,.1); color:#22c55e; letter-spacing:.05em; text-transform:none; }

.form-row-2 { display:grid; grid-template-columns:1fr 1fr; gap:1rem; }
.form-row-3 { display:grid; grid-template-columns:1fr 1fr 1fr; gap:1rem; }
@media (max-width:640px) { .form-row-2, .form-row-3 { grid-template-columns:1fr; } }
.form-group { display:flex; flex-direction:column; gap:.35rem; }
.form-group label { font-size:.78rem; font-weight:600; letter-spacing:.06em; text-transform:uppercase; color:var(--color-gray-1); }
.form-group input, .form-group select { width:100%; }
.disabled-input { opacity:.5; cursor:not-allowed; }
.req { color:var(--color-primary); }

.completeness-bar { background:var(--color-bg-card); border:1px solid var(--color-border); border-radius:8px; padding:1.25rem; margin-bottom:1.25rem; }
.completeness-label { display:flex; justify-content:space-between; font-size:.85rem; font-weight:600; margin-bottom:.6rem; }
.completeness-track { height:6px; background:var(--color-border); border-radius:3px; overflow:hidden; }
.completeness-fill { height:100%; border-radius:3px; transition:width .5s ease; }
.completeness-hint { font-size:.78rem; color:var(--color-gray-2); margin-top:.6rem; }

.form-msg { padding:.7rem 1rem; border-radius:6px; font-size:.88rem; margin-bottom:1rem; }
.form-msg.error   { background:rgba(239,68,68,.1); border:1px solid rgba(239,68,68,.3); color:#fca5a5; }
.form-msg.success { background:rgba(34,197,94,.1); border:1px solid rgba(34,197,94,.3); color:#86efac; }

.form-actions { display:flex; gap:1rem; align-items:center; }
.spinner { width:14px; height:14px; border:2px solid rgba(255,255,255,.3); border-top-color:#fff; border-radius:50%; animation:spin .7s linear infinite; display:inline-block; }
@keyframes spin { to { transform:rotate(360deg) } }
</style>
