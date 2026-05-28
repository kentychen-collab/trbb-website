<template>
  <div class="auth-page">
    <div class="auth-bg"><div class="auth-grid"></div></div>
    <div class="auth-box wide">
      <RouterLink to="/" class="auth-logo">
        <span class="tr">TR</span><span class="bb">BB</span>
      </RouterLink>
      <h1 class="auth-title">申請會員</h1>

      <!-- Step indicator -->
      <div class="steps">
        <div class="step" :class="{ active: step >= 1, done: step > 1 }">
          <span class="step-num">1</span><span class="step-label">填寫資料</span>
        </div>
        <div class="step-line" :class="{ done: step > 1 }"></div>
        <div class="step" :class="{ active: step >= 2 }">
          <span class="step-num">2</span><span class="step-label">等待審核</span>
        </div>
      </div>

      <!-- Step 1: Form -->
      <template v-if="step === 1">
        <div class="form-row">
          <div class="form-group">
            <label>會員 ID <span class="req">*</span></label>
            <input v-model="form.username" type="text" placeholder="英數字及底線，3~50 字元" />
          </div>
          <div class="form-group">
            <label>名稱 <span class="req">*</span></label>
            <input v-model="form.real_name" type="text" placeholder="中文或英文名稱" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>Email <span class="req">*</span></label>
            <input v-model="form.email" type="email" placeholder="your@email.com" />
          </div>
          <div class="form-group">
            <label>手機號碼 <span class="req">*</span></label>
            <input v-model="form.phone" type="tel" placeholder="09xxxxxxxx" />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>密碼 <span class="req">*</span></label>
            <div class="input-wrap">
              <input v-model="form.password" :type="showPwd ? 'text' : 'password'" placeholder="至少 8 個字元" />
              <button class="eye-btn" @click="showPwd = !showPwd" type="button">{{ showPwd ? '🙈' : '👁' }}</button>
            </div>
          </div>
          <div class="form-group">
            <label>確認密碼 <span class="req">*</span></label>
            <input v-model="form.confirm_password" :type="showPwd ? 'text' : 'password'" placeholder="再次輸入密碼" />
          </div>
        </div>

        <div v-if="error" class="auth-error">{{ error }}</div>

        <button class="btn btn-primary auth-submit" @click="handleSubmit" :disabled="loading">
          <span v-if="loading" class="spinner"></span>
          {{ loading ? '送出中...' : '送出申請' }}
        </button>
      </template>

      <!-- Step 2: Pending -->
      <template v-else>
        <div class="pending-box">
          <div class="pending-icon">⏳</div>
          <h2>申請已送出！</h2>
          <p>您的會員申請已成功送出，管理員將盡快審核。</p>
          <p>審核完成後將以 Email 通知您，請留意收件匣。</p>
          <div class="pending-info">
            <div class="info-row"><span>會員 ID</span><strong>{{ form.username }}</strong></div>
            <div class="info-row"><span>Email</span><strong>{{ form.email }}</strong></div>
          </div>
          <RouterLink to="/login" class="btn btn-outline" style="margin-top:1.5rem">返回登入</RouterLink>
        </div>
      </template>

      <div class="auth-footer" v-if="step === 1">
        已有帳號？<RouterLink to="/login">立即登入</RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import api from '@/services/api'

const step    = ref(1)
const loading = ref(false)
const error   = ref('')
const showPwd = ref(false)
const form    = ref({
  username: '', real_name: '', email: '', phone: '',
  password: '', confirm_password: ''
})

async function handleSubmit() {
  error.value = ''
  const { username, real_name, email, phone, password, confirm_password } = form.value

  if (!username)  { error.value = '請填寫會員 ID'; return }
  if (!real_name) { error.value = '請填寫名稱'; return }
  if (!email)     { error.value = '請填寫 Email'; return }
  if (!phone)     { error.value = '請填寫手機號碼'; return }
  if (!password)  { error.value = '請填寫密碼'; return }

  if (username.length < 3)               { error.value = '會員 ID 至少需 3 個字元'; return }
  if (!/^[a-zA-Z0-9_]+$/.test(username)) { error.value = '會員 ID 只能包含英文、數字和底線'; return }
  if (password.length < 8)               { error.value = '密碼至少需 8 個字元'; return }
  if (password !== confirm_password)     { error.value = '兩次輸入的密碼不一致'; return }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) { error.value = 'Email 格式不正確'; return }

  loading.value = true
  try {
    await api.post('/auth/register', { username, real_name, email, phone, password })
    step.value = 2
  } catch(e) {
    error.value = e.response?.data?.error || '送出失敗，請稍後再試'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page { min-height:100vh; display:flex; align-items:center; justify-content:center; padding:2rem 1rem; position:relative; background:var(--color-bg); }
.auth-bg { position:absolute; inset:0; }
.auth-grid { position:absolute; inset:0; background-image: linear-gradient(rgba(229,25,26,0.04)1px,transparent 1px), linear-gradient(90deg,rgba(229,25,26,0.04)1px,transparent 1px); background-size:50px 50px; }
.auth-box { position:relative; z-index:1; background:var(--color-bg-card); border:1px solid var(--color-border); border-radius:8px; padding:2.5rem; width:100%; max-width:420px; box-shadow:0 20px 60px rgba(0,0,0,.6); }
.auth-box.wide { max-width:600px; }
.auth-logo { display:block; font-family:var(--font-display); font-size:2.8rem; text-align:center; margin-bottom:.5rem; }
.auth-logo .tr { color:#fff; }
.auth-logo .bb { color:var(--color-primary); }
.auth-title { font-family:var(--font-cond); font-size:1.1rem; letter-spacing:.1em; text-align:center; color:var(--color-gray-2); text-transform:uppercase; margin-bottom:1.5rem; }

.steps { display:flex; align-items:center; justify-content:center; margin-bottom:2rem; }
.step { display:flex; flex-direction:column; align-items:center; gap:.25rem; }
.step-num { width:28px; height:28px; border-radius:50%; border:2px solid var(--color-border); display:flex; align-items:center; justify-content:center; font-family:var(--font-cond); font-size:.85rem; font-weight:700; color:var(--color-gray-2); transition:all .3s; }
.step.active .step-num { border-color:var(--color-primary); color:var(--color-primary); }
.step.done .step-num { background:var(--color-primary); border-color:var(--color-primary); color:#fff; }
.step-label { font-size:.72rem; color:var(--color-gray-2); }
.step.active .step-label { color:var(--color-white); }
.step-line { flex:1; height:2px; background:var(--color-border); margin:0 .75rem; min-width:60px; margin-bottom:1rem; transition:background .3s; }
.step-line.done { background:var(--color-primary); }

.form-row { display:grid; grid-template-columns:1fr 1fr; gap:1rem; }
@media (max-width:500px) { .form-row { grid-template-columns:1fr; } }
.form-group { margin-bottom:1rem; }
.form-group label { display:block; font-size:.78rem; font-weight:600; letter-spacing:.08em; text-transform:uppercase; color:var(--color-gray-1); margin-bottom:.4rem; }
.form-group input { width:100%; }
.req { color:var(--color-primary); }
.input-wrap { position:relative; }
.input-wrap input { width:100%; padding-right:2.5rem; }
.eye-btn { position:absolute; right:.6rem; top:50%; transform:translateY(-50%); background:none; border:none; cursor:pointer; font-size:1rem; opacity:.6; }
.auth-error { background:rgba(229,25,26,.1); border:1px solid rgba(229,25,26,.3); border-radius:4px; color:#ff6b6b; font-size:.85rem; padding:.6rem .9rem; margin-bottom:1rem; }
.auth-submit { width:100%; margin-top:.5rem; padding:.85rem; font-size:1rem; display:flex; align-items:center; justify-content:center; gap:.5rem; }
.spinner { width:16px; height:16px; border:2px solid rgba(255,255,255,.3); border-top-color:#fff; border-radius:50%; animation:spin .7s linear infinite; }
@keyframes spin { to { transform:rotate(360deg) } }
.auth-footer { text-align:center; margin-top:1.5rem; font-size:.88rem; color:var(--color-gray-2); }
.auth-footer a { color:var(--color-primary); font-weight:600; }

.pending-box { text-align:center; padding:1rem 0; }
.pending-icon { font-size:3rem; margin-bottom:1rem; }
.pending-box h2 { font-size:1.5rem; margin-bottom:.75rem; }
.pending-box p { color:var(--color-gray-1); font-size:.9rem; line-height:1.7; margin-bottom:.5rem; }
.pending-info { background:var(--color-bg); border:1px solid var(--color-border); border-radius:4px; padding:1rem 1.5rem; margin-top:1.5rem; text-align:left; }
.info-row { display:flex; justify-content:space-between; padding:.4rem 0; font-size:.88rem; }
.info-row span { color:var(--color-gray-2); }
.info-row strong { color:var(--color-white); }
</style>
