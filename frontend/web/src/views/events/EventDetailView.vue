<template>
  <div class="event-detail-page">

    <div v-if="loading" class="loading-full">
      <div class="loading-spinner"></div>
    </div>

    <template v-else-if="event">
      <!-- Hero -->
      <div class="event-hero" :style="event.cover_url ? `background-image:url(${event.cover_url})` : ''">
        <div class="event-hero-overlay"></div>
        <div class="event-hero-grid"></div>
        <div class="container event-hero-content">
          <div class="event-hero-type">{{ eventTypeLabel(event.event_type) }}</div>
          <h1 class="event-hero-title">{{ event.title }}</h1>
          <div class="event-hero-meta">
            <span>📅 {{ formatDate(event.start_at) }}</span>
            <span>📍 {{ event.location }}</span>
            <span v-if="event.fee > 0">💰 NT$ {{ event.fee.toLocaleString() }}</span>
            <span v-else>💰 免費</span>
          </div>
        </div>
      </div>

      <div class="container event-main">
        <!-- Left: info -->
        <div class="event-info">
          <!-- Status banner -->
          <div class="reg-status-banner" :class="regBannerClass">
            <span class="reg-status-icon">{{ regBannerIcon }}</span>
            <span>{{ regBannerText }}</span>
          </div>

          <!-- Description -->
          <div class="event-section" v-if="event.description">
            <h2 class="event-section-title">賽事說明</h2>
            <div class="event-desc" v-html="descHtml"></div>
          </div>

          <!-- Details -->
          <div class="event-section">
            <h2 class="event-section-title">賽事資訊</h2>
            <div class="detail-grid">
              <div class="detail-row"><span>賽事日期</span><strong>{{ formatDate(event.start_at) }} – {{ formatDate(event.end_at) }}</strong></div>
              <div class="detail-row"><span>報名期間</span><strong>{{ formatDate(event.reg_start_at) }} – {{ formatDate(event.reg_end_at) }}</strong></div>
              <div class="detail-row"><span>地點</span><strong>{{ event.location }}</strong></div>
              <div class="detail-row"><span>報名費用</span><strong class="text-red">{{ event.fee > 0 ? 'NT$ '+event.fee.toLocaleString() : '免費' }}</strong></div>
              <div class="detail-row" v-if="event.max_participants">
                <span>人數上限</span>
                <strong>{{ event.max_participants }} 人（已報名 {{ event.registered_count }}）</strong>
              </div>
            </div>
          </div>
        </div>

        <!-- Right: action -->
        <aside class="event-aside">
          <div class="aside-card">
            <div class="aside-fee">
              <span v-if="event.fee > 0">NT$ <strong>{{ event.fee.toLocaleString() }}</strong></span>
              <span v-else class="free-badge">免費報名</span>
            </div>

            <!-- Already registered -->
            <div v-if="myReg" class="reg-done-box">
              <div class="reg-done-icon">✅</div>
              <div class="reg-done-text">您已完成報名</div>
              <div class="reg-done-status">狀態：{{ regStatusLabel(myReg.status) }}</div>
              <button class="btn btn-ghost" style="width:100%;margin-top:1rem"
                @click="showCancelConfirm=true" v-if="myReg.status < 2">取消報名</button>
            </div>

            <!-- Register button -->
            <template v-else>
              <button v-if="!auth.isLoggedIn" class="btn btn-primary reg-btn"
                @click="router.push(`/login?redirect=/events/${event.id}`)">
                登入後報名
              </button>
              <button v-else-if="canRegister" class="btn btn-primary reg-btn" @click="openRegModal">
                立即報名
              </button>
              <div v-else class="reg-disabled">
                {{ regDisabledReason }}
              </div>
            </template>

            <div class="aside-spots" v-if="event.max_participants">
              剩餘名額：{{ Math.max(0, event.max_participants - event.registered_count) }} 名
            </div>
          </div>
        </aside>
      </div>
    </template>

    <!-- ── Registration Modal ──────────────────────────────── -->
    <div class="modal-overlay" v-if="showRegModal" @click.self="showRegModal=false">
      <div class="reg-modal">
        <div class="reg-modal-header">
          <h2>報名表單</h2>
          <div class="reg-modal-subtitle">{{ event?.title }}</div>
          <button class="close-btn" @click="showRegModal=false">✕</button>
        </div>
        <div class="reg-modal-body">
          <div class="prefill-hint">
            <span>💡 已自動帶入您的會員資料，可直接修改</span>
          </div>

          <div class="reg-form-grid">
            <div class="form-group required">
              <label>中文姓名</label>
              <input v-model="form.name_zh" placeholder="真實中文姓名" />
            </div>
            <div class="form-group">
              <label>英文姓名</label>
              <input v-model="form.name_en" placeholder="English Name" />
            </div>
            <div class="form-group required">
              <label>手機號碼</label>
              <input v-model="form.phone" type="tel" placeholder="09xxxxxxxx" />
            </div>
            <div class="form-group required">
              <label>Email</label>
              <input v-model="form.email" type="email" placeholder="your@email.com" />
            </div>
            <div class="form-group">
              <label>身份證字號</label>
              <input v-model="form.id_number" placeholder="A123456789" />
            </div>
            <div class="form-group">
              <label>出生年月日</label>
              <input v-model="form.birthday" type="date" />
            </div>
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
              <label>衣服尺寸</label>
              <select v-model="form.shirt_size">
                <option value="">請選擇</option>
                <option v-for="s in ['XS','S','M','L','XL','2XL','3XL']" :key="s" :value="s">{{ s }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>飲食習慣</label>
              <select v-model="form.food_type">
                <option :value="null">請選擇</option>
                <option :value="1">葷食</option>
                <option :value="2">素食</option>
                <option :value="3">全素</option>
              </select>
            </div>
            <div class="form-group full">
              <label>通訊地址</label>
              <input v-model="form.address" placeholder="縣市 + 完整地址" />
            </div>
            <div class="form-group required">
              <label>緊急聯絡人姓名</label>
              <input v-model="form.emergency_contact" placeholder="緊急聯絡人姓名" />
            </div>
            <div class="form-group required">
              <label>緊急聯絡人手機</label>
              <input v-model="form.emergency_phone" type="tel" placeholder="09xxxxxxxx" />
            </div>
            <div class="form-group required">
              <label>緊急聯絡人關係</label>
              <select v-model="form.emergency_relation">
                <option value="">請選擇</option>
                <option v-for="r in ['配偶','父親','母親','兄弟','姊妹','子女','朋友','其他']" :key="r" :value="r">{{ r }}</option>
              </select>
            </div>
            <div class="form-group full">
              <label>備註</label>
              <input v-model="form.note" placeholder="其他需要告知事項（選填）" />
            </div>
          </div>

          <div v-if="regError" class="reg-error">{{ regError }}</div>

          <div class="reg-modal-footer">
            <button class="btn btn-primary reg-submit-btn" @click="submitReg" :disabled="regLoading">
              <span v-if="regLoading" class="spinner"></span>
              {{ regLoading ? '報名中...' : `確認報名${event?.fee > 0 ? '（NT$ '+event?.fee.toLocaleString()+'）' : ''}` }}
            </button>
            <button class="btn btn-ghost" @click="showRegModal=false">取消</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Cancel confirm -->
    <div class="modal-overlay" v-if="showCancelConfirm" @click.self="showCancelConfirm=false">
      <div class="confirm-modal">
        <h3>確認取消報名？</h3>
        <p class="text-gray">取消後若要重新報名，需在報名截止前再次填寫表單。</p>
        <div class="confirm-actions">
          <button class="btn btn-danger" @click="cancelReg">確認取消</button>
          <button class="btn btn-ghost" @click="showCancelConfirm=false">返回</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'

const route  = useRoute()
const router = useRouter()
const auth   = useAuthStore()

const event   = ref(null)
const myReg   = ref(null)
const loading = ref(true)
const showRegModal      = ref(false)
const showCancelConfirm = ref(false)
const regLoading = ref(false)
const regError   = ref('')

const form = reactive({
  name_zh:'', name_en:'', id_number:'', passport_number:'',
  gender:null, birthday:'', phone:'', email:'',
  shirt_size:'', food_type:null, address:'',
  emergency_contact:'', emergency_phone:'', emergency_relation:'',
  note:'',
})

const descHtml = computed(() => {
  if (!event.value?.description) return ''
  return event.value.description.replace(/\n/g, '<br>')
})

const canRegister = computed(() => {
  if (!event.value) return false
  const now = new Date()
  return event.value.status === 1 &&
    now > new Date(event.value.reg_start_at) &&
    now < new Date(event.value.reg_end_at) &&
    (!event.value.max_participants || event.value.registered_count < event.value.max_participants)
})

const regBannerClass = computed(() => {
  if (!event.value) return ''
  const now = new Date()
  if (now < new Date(event.value.reg_start_at)) return 'banner-upcoming'
  if (now > new Date(event.value.reg_end_at))   return 'banner-closed'
  if (event.value.max_participants && event.value.registered_count >= event.value.max_participants) return 'banner-full'
  return 'banner-open'
})
const regBannerIcon = computed(() => ({ 'banner-open':'🟢','banner-upcoming':'🟡','banner-closed':'🔴','banner-full':'🔴' }[regBannerClass.value]))
const regBannerText = computed(() => {
  const ev = event.value
  if (!ev) return ''
  const now = new Date()
  const fmt = d => new Date(d).toLocaleDateString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit' })
  if (now < new Date(ev.reg_start_at)) return `報名尚未開放，${fmt(ev.reg_start_at)} 起開放報名`
  if (now > new Date(ev.reg_end_at))   return `報名已截止（${fmt(ev.reg_end_at)}）`
  if (ev.max_participants && ev.registered_count >= ev.max_participants) return '報名人數已額滿'
  return `報名進行中，截止日期 ${fmt(ev.reg_end_at)}`
})
const regDisabledReason = computed(() => {
  const cls = regBannerClass.value
  if (cls === 'banner-upcoming') return '報名尚未開放'
  if (cls === 'banner-closed')   return '報名已截止'
  if (cls === 'banner-full')     return '報名人數已額滿'
  return ''
})

async function openRegModal() {
  regError.value = ''
  // 帶入會員資料
  try {
    const { data } = await api.get('/me/registration-profile')
    Object.keys(form).forEach(k => { if (data[k] != null) form[k] = data[k] })
  } catch {}
  showRegModal.value = true
}

async function submitReg() {
  regError.value = ''
  if (!form.phone)   { regError.value = '請填寫手機號碼'; return }
  if (!form.email)   { regError.value = '請填寫 Email'; return }
  regLoading.value = true
  try {
    await api.post(`/events/${route.params.id}/register`, form)
    showRegModal.value = false
    // Reload
    await fetchData()
  } catch(e) {
    regError.value = e.response?.data?.error || '報名失敗，請稍後再試'
  } finally {
    regLoading.value = false
  }
}

async function cancelReg() {
  try {
    await api.delete(`/events/${route.params.id}/register`)
    showCancelConfirm.value = false
    await fetchData()
  } catch(e) {
    alert(e.response?.data?.error || '取消失敗')
  }
}

async function fetchData() {
  try {
    const { data } = await api.get(`/events/${route.params.id}`)
    event.value = data
    if (auth.isLoggedIn) {
      try {
        const r = await api.get(`/events/${route.params.id}/register`)
        myReg.value = r.data.registered ? r.data.registration : null
      } catch { myReg.value = null }
    }
  } catch {
    event.value = null
  } finally {
    loading.value = false
  }
}

const typeMap = { 1:'鐵人三項',2:'路跑',3:'游泳',4:'單車',5:'訓練',6:'其他' }
function eventTypeLabel(t) { return typeMap[t]||'其他' }
function formatDate(d) {
  return new Date(d).toLocaleDateString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit' })
}
function regStatusLabel(s) { return { 0:'待確認',1:'已確認',2:'已取消',3:'已退款' }[s]||'未知' }

onMounted(fetchData)
</script>

<style scoped>
.event-detail-page { background:var(--color-bg);min-height:100vh; }
.nav-link:hover { color:var(--color-primary); }

.loading-full { display:flex;align-items:center;justify-content:center;height:100vh; }
.loading-spinner { width:32px;height:32px;border:3px solid var(--color-border);border-top-color:var(--color-primary);border-radius:50%;animation:spin .7s linear infinite; }
@keyframes spin { to { transform:rotate(360deg) } }

/* Hero */
.event-hero { position:relative;height:400px;margin-top:64px;background:#0a0a1a;background-size:cover;background-position:center; }
.event-hero-overlay { position:absolute;inset:0;background:linear-gradient(to bottom,rgba(0,0,0,.3),rgba(0,0,0,.8)); }
.event-hero-grid { position:absolute;inset:0;background-image:linear-gradient(rgba(229,25,26,.04)1px,transparent 1px),linear-gradient(90deg,rgba(229,25,26,.04)1px,transparent 1px);background-size:50px 50px; }
.event-hero-content { position:relative;z-index:1;height:100%;display:flex;flex-direction:column;justify-content:flex-end;padding-bottom:2.5rem; }
.event-hero-type { font-family:var(--font-cond);font-size:.75rem;font-weight:700;letter-spacing:.2em;color:var(--color-primary);text-transform:uppercase;margin-bottom:.5rem; }
.event-hero-title { font-family:var(--font-display);font-size:clamp(2rem,6vw,4rem);line-height:1.1;margin-bottom:1rem; }
.event-hero-meta { display:flex;gap:1.5rem;flex-wrap:wrap;font-size:.9rem;color:var(--color-gray-1); }

/* Layout */
.event-main { display:grid;grid-template-columns:1fr 320px;gap:2.5rem;padding:2.5rem 0 4rem;align-items:start; }
@media(max-width:768px){ .event-main { grid-template-columns:1fr } .event-aside { order:-1 } }

/* Info */
.reg-status-banner { display:flex;align-items:center;gap:.75rem;padding:.85rem 1.25rem;border-radius:6px;margin-bottom:1.5rem;font-size:.9rem;font-weight:600; }
.banner-open     { background:rgba(34,197,94,.1);border:1px solid rgba(34,197,94,.3);color:#22c55e; }
.banner-upcoming { background:rgba(245,158,11,.1);border:1px solid rgba(245,158,11,.3);color:#f59e0b; }
.banner-closed,.banner-full { background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.3);color:#ef4444; }

.event-section { margin-bottom:2rem; }
.event-section-title { font-family:var(--font-cond);font-size:.8rem;font-weight:700;letter-spacing:.12em;text-transform:uppercase;color:var(--color-gray-2);margin-bottom:1rem;padding-bottom:.5rem;border-bottom:1px solid var(--color-border); }
.event-desc { color:var(--color-gray-1);line-height:1.9;font-size:.95rem; }
.detail-grid { display:flex;flex-direction:column; }
.detail-row { display:flex;justify-content:space-between;padding:.65rem 0;border-bottom:1px solid var(--color-border);font-size:.9rem; }
.detail-row span { color:var(--color-gray-2); }

/* Aside */
.event-aside { position:sticky;top:calc(64px + 1.5rem); }
.aside-card { background:var(--color-bg-card);border:1px solid var(--color-border);border-radius:8px;padding:1.5rem; }
.aside-fee { font-size:.9rem;color:var(--color-gray-2);margin-bottom:1rem; }
.aside-fee strong { font-family:var(--font-display);font-size:2.2rem;color:var(--color-primary); }
.free-badge { font-family:var(--font-cond);font-size:1.4rem;font-weight:700;color:#22c55e; }
.reg-btn { width:100%;padding:.9rem;font-size:1rem; }
.reg-disabled { text-align:center;color:var(--color-gray-2);font-size:.88rem;padding:.75rem;background:var(--color-bg);border-radius:4px; }
.aside-spots { text-align:center;font-size:.78rem;color:var(--color-gray-2);margin-top:.75rem; }
.reg-done-box { text-align:center;padding:1rem 0; }
.reg-done-icon { font-size:2.5rem; }
.reg-done-text { font-weight:700;font-size:1rem;margin:.5rem 0 .25rem; }
.reg-done-status { font-size:.82rem;color:var(--color-gray-2); }

/* Modal */
.modal-overlay { position:fixed;inset:0;background:rgba(0,0,0,.75);z-index:200;display:flex;align-items:center;justify-content:center;padding:1rem; }
.reg-modal { background:var(--color-bg-card);border:1px solid var(--color-border);border-radius:10px;width:100%;max-width:680px;max-height:90vh;overflow-y:auto; }
.reg-modal-header { padding:1.5rem;border-bottom:1px solid var(--color-border);position:sticky;top:0;background:var(--color-bg-card);z-index:1;display:flex;flex-direction:column; }
.reg-modal-header h2 { font-family:var(--font-cond);font-size:1.3rem;font-weight:700; }
.reg-modal-subtitle { font-size:.82rem;color:var(--color-gray-2);margin-top:.25rem; }
.close-btn { position:absolute;top:1.25rem;right:1.25rem;background:none;border:none;color:var(--color-gray-2);font-size:1.2rem;cursor:pointer; }
.reg-modal-body { padding:1.5rem; }
.prefill-hint { background:rgba(229,25,26,.08);border:1px solid rgba(229,25,26,.2);border-radius:4px;padding:.65rem 1rem;font-size:.82rem;color:var(--color-gray-1);margin-bottom:1.25rem; }
.reg-form-grid { display:grid;grid-template-columns:1fr 1fr;gap:1rem; }
@media(max-width:500px){ .reg-form-grid { grid-template-columns:1fr } }
.form-group { display:flex;flex-direction:column;gap:.35rem; }
.form-group.full { grid-column:1/-1; }
.form-group label { font-size:.75rem;font-weight:600;letter-spacing:.06em;text-transform:uppercase;color:var(--color-gray-1); }
.form-group.required label::after { content:' *';color:var(--color-primary); }
.form-group input,.form-group select { width:100%; }
.reg-error { background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.3);border-radius:4px;color:#fca5a5;font-size:.83rem;padding:.6rem .9rem;margin-top:1rem; }
.reg-modal-footer { display:flex;gap:1rem;margin-top:1.5rem;padding-top:1.25rem;border-top:1px solid var(--color-border); }
.reg-submit-btn { flex:1;padding:.85rem;font-size:.95rem;display:flex;align-items:center;justify-content:center;gap:.5rem; }
.spinner { width:14px;height:14px;border:2px solid rgba(255,255,255,.3);border-top-color:#fff;border-radius:50%;animation:spin .7s linear infinite; }

.confirm-modal { background:var(--color-bg-card);border:1px solid var(--color-border);border-radius:8px;padding:2rem;max-width:400px;width:100%; }
.confirm-modal h3 { font-size:1.1rem;margin-bottom:.75rem; }
.confirm-actions { display:flex;gap:.75rem;margin-top:1.5rem; }
.btn-danger { background:#ef4444;color:#fff; }
.btn-danger:hover { opacity:.85; }
</style>
