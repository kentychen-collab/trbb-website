<template>
  <div class="settings-page">
    <div class="page-header flex justify-between items-center">
      <div>
        <h1 class="page-title">網站設定</h1>
        <p class="page-subtitle">調整前台的視覺樣式、品牌識別與顯示內容</p>
      </div>
      <button class="btn btn-primary" @click="saveAll" :disabled="saving">
        <span v-if="saving" class="spinner-sm"></span>
        {{ saving ? '儲存中...' : '💾 儲存所有設定' }}
      </button>
    </div>

    <div v-if="loading" class="loading-row">載入中...</div>
    <div v-else class="settings-grid">

      <!-- ── 品牌識別 ─────────────────────────────────────── -->
      <div class="settings-card card">
        <div class="card-header"><h3>🏷 品牌識別</h3></div>
        <div class="card-body">
          <div class="setting-item">
            <label>Logo 圖片</label>
            <div class="image-upload-row">
              <div class="img-preview" v-if="form.logo_image">
                <img :src="form.logo_image" alt="logo" />
              </div>
              <div class="img-upload-btns">
                <label class="btn btn-ghost btn-sm upload-label">
                  <input type="file" accept="image/*" style="display:none"
                    @change="e => uploadImage(e,'logo','logo_image')" />
                  {{ form.logo_image ? '更換圖片' : '上傳 Logo' }}
                </label>
                <button v-if="form.logo_image" class="btn btn-ghost btn-sm"
                  @click="form.logo_image=''">移除</button>
              </div>
            </div>
            <p class="hint">建議尺寸 200×60px，PNG 透明背景</p>
          </div>
          <div class="setting-item">
            <label>Logo 說明文字</label>
            <input v-model="form.logo_text" placeholder="TRBB 鐵人拔巴" />
            <p class="hint">顯示於 Logo 圖片下方或旁邊</p>
          </div>
          <div class="setting-row">
            <div class="setting-item">
              <label>文字大小</label>
              <input v-model="form.logo_text_size" placeholder="1rem" />
            </div>
          </div>

          <div class="setting-divider"></div>

          <!-- Icon / Favicon -->
          <div class="setting-item">
            <label>Favicon（瀏覽器分頁小圖示）</label>
            <div class="image-upload-row">
              <div class="icon-preview" v-if="form.site_icon">
                <img :src="form.site_icon" alt="favicon" />
              </div>
              <div class="img-upload-btns">
                <label class="btn btn-ghost btn-sm upload-label">
                  <input type="file" accept="image/png,image/ico,image/svg+xml,image/x-icon"
                    style="display:none" @change="e => uploadImage(e,'icon','site_icon')" />
                  {{ form.site_icon ? '更換 Favicon' : '上傳 Favicon' }}
                </label>
                <button v-if="form.site_icon" class="btn btn-ghost btn-sm"
                  @click="form.site_icon=''">移除</button>
              </div>
            </div>
            <p class="hint">建議 32×32 或 64×64 PNG，顯示於瀏覽器分頁</p>
          </div>

          <div class="setting-item">
            <label>應用 Icon（手機加入主畫面用）</label>
            <div class="image-upload-row">
              <div class="icon-preview lg" v-if="form.site_icon_lg">
                <img :src="form.site_icon_lg" alt="app icon" />
              </div>
              <div class="img-upload-btns">
                <label class="btn btn-ghost btn-sm upload-label">
                  <input type="file" accept="image/png" style="display:none"
                    @change="e => uploadImage(e,'icon_lg','site_icon_lg')" />
                  {{ form.site_icon_lg ? '更換應用 Icon' : '上傳應用 Icon' }}
                </label>
                <button v-if="form.site_icon_lg" class="btn btn-ghost btn-sm"
                  @click="form.site_icon_lg=''">移除</button>
              </div>
            </div>
            <p class="hint">建議 192×192 PNG，用於 iOS/Android 主畫面捷徑</p>
          </div>
        </div>
      </div>

      <!-- ── 橫幅 ─────────────────────────────────────────── -->
      <div class="settings-card card">
        <div class="card-header"><h3>🖼 橫幅設定</h3></div>
        <div class="card-body">
          <div class="setting-item">
            <div class="banner-toggle">
              <label>顯示橫幅</label>
              <button class="toggle-btn" :class="{ on: form.banner_visible === '1' }"
                @click="form.banner_visible = form.banner_visible==='1' ? '0' : '1'" type="button">
                {{ form.banner_visible === '1' ? '🟢 顯示中' : '⭕ 已隱藏' }}
              </button>
            </div>
          </div>
          <div class="setting-item">
            <label>橫幅圖片 1</label>
            <div class="image-upload-row">
              <div class="banner-preview" v-if="form.banner_image">
                <img :src="form.banner_image" alt="banner" />
              </div>
              <div class="img-upload-btns">
                <label class="btn btn-ghost btn-sm upload-label">
                  <input type="file" accept="image/*" style="display:none"
                    @change="e => uploadImage(e,'banner','banner_image')" />
                  {{ form.banner_image ? '更換' : '上傳橫幅 1' }}
                </label>
                <button v-if="form.banner_image" class="btn btn-ghost btn-sm"
                  @click="form.banner_image=''">移除</button>
              </div>
            </div>
          </div>
          <div class="setting-item">
            <label>橫幅圖片 2（輪播備用）</label>
            <div class="image-upload-row">
              <div class="banner-preview" v-if="form.banner_image_2">
                <img :src="form.banner_image_2" alt="banner2" />
              </div>
              <div class="img-upload-btns">
                <label class="btn btn-ghost btn-sm upload-label">
                  <input type="file" accept="image/*" style="display:none"
                    @change="e => uploadImage(e,'banner2','banner_image_2')" />
                  {{ form.banner_image_2 ? '更換' : '上傳橫幅 2' }}
                </label>
                <button v-if="form.banner_image_2" class="btn btn-ghost btn-sm"
                  @click="form.banner_image_2=''">移除</button>
              </div>
            </div>
          </div>
          <div class="setting-item">
            <label>橫幅說明文字</label>
            <input v-model="form.banner_text" placeholder="TRBB 10周年 FINISH STRONG" />
          </div>
          <div class="setting-item">
            <label>橫幅連結 URL</label>
            <input v-model="form.banner_link" placeholder="https://..." />
          </div>
        </div>
      </div>

      <!-- ── 主題顏色 ────────────────────────────────────── -->
      <div class="settings-card card">
        <div class="card-header"><h3>🎨 主題顏色</h3></div>
        <div class="card-body">
          <!-- Simple colors -->
          <div class="color-row">
            <div class="setting-item" v-for="c in simpleColors" :key="c.key">
              <label>{{ c.label }}</label>
              <div class="color-input-row">
                <input type="color" v-model="form[c.key]" class="color-swatch" />
                <input v-model="form[c.key]" class="color-text" placeholder="#000000" />
              </div>
            </div>
          </div>

          <!-- Gradient fields -->
          <div v-for="gf in gradientFields" :key="gf.key" class="setting-item" style="margin-top:1.25rem">
            <label>{{ gf.label }}</label>
            <GradientEditor v-model="form[gf.key]" :label="gf.label" />
          </div>
        </div>
      </div>

      <!-- ── 文字排版 ────────────────────────────────────── -->
      <div class="settings-card card">
        <div class="card-header"><h3>✍️ 文字排版</h3></div>
        <div class="card-body">

          <!-- Body Text -->
          <div class="typo-group">
            <div class="typo-group-title">內文</div>
            <div class="typo-grid">
              <div class="setting-item">
                <label>字體名稱</label>
                <input v-model="form.font_body" placeholder="Barlow" />
                <p class="hint">Google Fonts 名稱</p>
              </div>
              <div class="setting-item">
                <label>字體大小</label>
                <input v-model="form.font_body_size" placeholder="16px" />
              </div>
              <div class="setting-item">
                <label>字重</label>
                <select v-model="form.font_body_weight">
                  <option value="300">細（300）</option>
                  <option value="400">標準（400）</option>
                  <option value="500">中（500）</option>
                  <option value="600">半粗（600）</option>
                  <option value="700">粗（700）</option>
                </select>
              </div>
              <div class="setting-item">
                <label>文字顏色</label>
                <div class="color-input-row">
                  <input type="color" v-model="form.font_body_color" class="color-swatch" />
                  <input v-model="form.font_body_color" class="color-text" />
                </div>
              </div>
            </div>
          </div>

          <!-- Heading -->
          <div class="typo-group">
            <div class="typo-group-title">標題</div>
            <div class="typo-grid">
              <div class="setting-item">
                <label>字體名稱</label>
                <input v-model="form.font_heading" placeholder="Barlow Condensed" />
              </div>
              <div class="setting-item">
                <label>字體大小</label>
                <input v-model="form.font_heading_size" placeholder="1.6rem" />
              </div>
              <div class="setting-item">
                <label>字重</label>
                <select v-model="form.font_heading_weight">
                  <option value="600">半粗（600）</option>
                  <option value="700">粗（700）</option>
                  <option value="800">特粗（800）</option>
                  <option value="900">黑體（900）</option>
                </select>
              </div>
              <div class="setting-item">
                <label>文字顏色</label>
                <div class="color-input-row">
                  <input type="color" v-model="form.font_heading_color" class="color-swatch" />
                  <input v-model="form.font_heading_color" class="color-text" />
                </div>
              </div>
            </div>
          </div>

          <!-- Display -->
          <div class="typo-group">
            <div class="typo-group-title">展示字（大標 / Hero）</div>
            <div class="typo-grid">
              <div class="setting-item">
                <label>字體名稱</label>
                <input v-model="form.font_display" placeholder="Bebas Neue" />
              </div>
              <div class="setting-item">
                <label>文字顏色</label>
                <div class="color-input-row">
                  <input type="color" v-model="form.font_display_color" class="color-swatch" />
                  <input v-model="form.font_display_color" class="color-text" />
                </div>
              </div>
            </div>
          </div>

          <!-- Font presets -->
          <div class="font-presets">
            <div class="presets-label">快速套用字體組合</div>
            <div class="preset-btns">
              <button v-for="p in fontPresets" :key="p.name"
                class="preset-btn" @click="applyFontPreset(p)" type="button">
                {{ p.name }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- ── 導覽列 ──────────────────────────────────────── -->
      <div class="settings-card card">
        <div class="card-header"><h3>📌 導覽列</h3></div>
        <div class="card-body">
          <div class="setting-item">
            <label>導覽列文字顏色</label>
            <div class="color-input-row">
              <input type="color" v-model="form.navbar_text" class="color-swatch" />
              <input v-model="form.navbar_text" class="color-text" />
            </div>
          </div>
          <div class="setting-item" style="margin-top:.75rem">
            <label>導覽列背景</label>
            <GradientEditor v-model="form.navbar_bg" label="導覽列背景" />
          </div>
        </div>
      </div>

    </div><!-- end settings-grid -->

    <!-- 成功訊息 -->
    <div v-if="saved" class="save-toast">✓ 設定已儲存，前台即時生效</div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import api from '@/services/api'
import GradientEditor from '@/components/GradientEditor.vue'

const loading = ref(true)
const saving  = ref(false)
const saved   = ref(false)

const form = reactive({
  logo_image: '', logo_text: 'TRBB 鐵人拔巴', logo_text_size: '1rem',
  banner_image: '', banner_image_2: '', banner_text: '', banner_link: '', banner_visible: '1',
  bg_color:    '{"type":"solid","colors":["#FFFFF3"]}',
  bg2_color:   '{"type":"solid","colors":["#F5F5E8"]}',
  card_color:  '{"type":"solid","colors":["#FFFFFF"]}',
  border_color: '#E0E3DA',
  primary_color: '#CF2027', navy_color: '#1A3A7A', accent_color: '#A593E0',
  font_body: 'Barlow', font_body_size: '16px', font_body_color: '#566270', font_body_weight: '400',
  font_heading: 'Barlow Condensed', font_heading_size: '1.6rem', font_heading_color: '#1A3A7A', font_heading_weight: '700',
  font_display: 'Bebas Neue', font_display_color: '#1A3A7A',
  navbar_bg: '{"type":"solid","colors":["rgba(255,255,243,0.95)"]}', navbar_text: '#1A3A7A',
  site_icon: '', site_icon_lg: '',
})

const simpleColors = [
  { key: 'primary_color', label: '主色（品牌紅）' },
  { key: 'navy_color',    label: '海軍藍' },
  { key: 'accent_color',  label: '強調色（紫）' },
  { key: 'border_color',  label: '邊框顏色' },
]

const gradientFields = [
  { key: 'bg_color',   label: '頁面底色' },
  { key: 'bg2_color',  label: '次要底色（交替區塊）' },
  { key: 'card_color', label: '卡片底色' },
]

const fontPresets = [
  { name: '現代簡約', font_body: 'Barlow', font_heading: 'Barlow Condensed', font_display: 'Bebas Neue' },
  { name: '傳統優雅', font_body: 'Noto Serif TC', font_heading: 'Noto Serif TC', font_display: 'Playfair Display' },
  { name: '科技感', font_body: 'IBM Plex Sans', font_heading: 'IBM Plex Sans Condensed', font_display: 'Oswald' },
  { name: '活力運動', font_body: 'Inter', font_heading: 'Barlow Condensed', font_display: 'Anton' },
  { name: '系統預設', font_body: 'system-ui', font_heading: 'system-ui', font_display: 'system-ui' },
]

function applyFontPreset(p) {
  form.font_body     = p.font_body
  form.font_heading  = p.font_heading
  form.font_display  = p.font_display
}

async function uploadImage(e, purpose, key) {
  const file = e.target.files?.[0]
  if (!file) return
  const fd = new FormData()
  fd.append('file', file)
  try {
    const { data } = await api.post(`/settings/upload/${purpose}`, fd, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    form[key] = data.url
  } catch(err) {
    alert(err.response?.data?.error || '上傳失敗')
  }
  e.target.value = ''
}

async function saveAll() {
  saving.value = true
  try {
    const payload = {}
    Object.keys(form).forEach(k => { payload[k] = form[k] || '' })
    await api.post('/settings', payload)
    saved.value = true
    setTimeout(() => { saved.value = false }, 3000)
  } catch(e) {
    alert(e.response?.data?.error || '儲存失敗')
  } finally {
    saving.value = false
  }
}

async function load() {
  loading.value = true
  try {
    // GET /v1/admin/settings 回傳 grouped: { brand: [{key,value,...}], ... }
    const { data } = await api.get('/settings')
    // 攤平所有 group
    const allSettings = Object.values(data).flat()
    allSettings.forEach(s => {
      if (Object.prototype.hasOwnProperty.call(form, s.key) && s.value != null) {
        form[s.key] = s.value
      }
    })
  } catch(e) { console.error('load settings error', e) }
  finally { loading.value = false }
}

onMounted(load)
</script>

<style scoped>
.settings-page { max-width:1100px; }
.settings-grid { display:grid; grid-template-columns:1fr 1fr; gap:1.25rem; }
@media(max-width:900px){ .settings-grid { grid-template-columns:1fr } }

.settings-card { height:fit-content; }
.card-header h3 { font-family:var(--font-c); font-size:1rem; font-weight:700; }
.card-body { padding:1.5rem; display:flex; flex-direction:column; gap:1rem; }

.setting-item { display:flex; flex-direction:column; gap:.3rem; }
.setting-item label { font-size:.72rem; font-weight:600; text-transform:uppercase; letter-spacing:.08em; color:var(--gray-2); }
.setting-item input, .setting-item select { width:100%; }
.hint { font-size:.7rem; color:var(--gray-2); }
.setting-row { display:grid; grid-template-columns:1fr 1fr; gap:.75rem; }

/* Image upload */
.image-upload-row { display:flex; align-items:flex-start; gap:.75rem; flex-wrap:wrap; }
.img-preview { width:80px; height:80px; border-radius:6px; overflow:hidden; border:1px solid var(--border); flex-shrink:0; }
.img-preview img { width:100%; height:100%; object-fit:contain; background:#f9f9f9; }
.banner-preview { width:160px; height:60px; border-radius:4px; overflow:hidden; border:1px solid var(--border); }
.banner-preview img { width:100%; height:100%; object-fit:cover; }
.img-upload-btns { display:flex; flex-direction:column; gap:.35rem; }
.upload-label { cursor:pointer; }
.setting-divider { height:1px; background:var(--border); margin:.25rem 0; }
.icon-preview { width:48px; height:48px; border-radius:6px; overflow:hidden; border:1px solid var(--border); flex-shrink:0; background:#f9f9f9; display:flex; align-items:center; justify-content:center; }
.icon-preview.lg { width:72px; height:72px; border-radius:10px; }
.icon-preview img { width:100%; height:100%; object-fit:contain; }

/* Banner toggle */
.banner-toggle { display:flex; align-items:center; justify-content:space-between; }
.toggle-btn { padding:.35rem 1rem; border-radius:4px; border:1px solid var(--border); font-size:.82rem; font-weight:600; cursor:pointer; background:var(--bg); color:var(--gray-2); transition:all .15s; }
.toggle-btn.on { background:rgba(46,139,87,.1); border-color:var(--success); color:var(--success); }

/* Color input */
.color-row { display:grid; grid-template-columns:1fr 1fr; gap:.75rem; }
.color-input-row { display:flex; align-items:center; gap:.5rem; }
.color-swatch { width:40px; height:36px; padding:2px 4px; cursor:pointer; flex-shrink:0; border-radius:4px; }
.color-text { flex:1; }

/* Typography */
.typo-group { border:1px solid var(--border); border-radius:6px; padding:1rem; margin-bottom:.75rem; }
.typo-group:last-of-type { margin-bottom:0; }
.typo-group-title { font-size:.72rem; font-weight:700; text-transform:uppercase; letter-spacing:.1em; color:var(--gray-2); margin-bottom:.75rem; }
.typo-grid { display:grid; grid-template-columns:1fr 1fr; gap:.6rem; }

/* Font presets */
.font-presets { border-top:1px solid var(--border); padding-top:.75rem; margin-top:.25rem; }
.presets-label { font-size:.72rem; color:var(--gray-2); margin-bottom:.4rem; }
.preset-btns { display:flex; gap:.4rem; flex-wrap:wrap; }
.preset-btn { padding:.3rem .85rem; border-radius:4px; border:1px solid var(--border); font-size:.78rem; cursor:pointer; background:var(--bg); color:var(--gray-1); transition:all .15s; }
.preset-btn:hover { border-color:var(--navy); color:var(--navy); }

/* Toast */
.save-toast { position:fixed; bottom:2rem; right:2rem; background:var(--success); color:#fff; padding:.75rem 1.5rem; border-radius:6px; font-weight:600; box-shadow:0 4px 16px rgba(0,0,0,.2); animation:fadeIn .3s ease; }
@keyframes fadeIn { from { opacity:0; transform:translateY(8px); } to { opacity:1; transform:none; } }

.loading-row { padding:4rem; text-align:center; color:var(--gray-2); }
.spinner-sm { width:12px; height:12px; border:2px solid rgba(255,255,255,.3); border-top-color:#fff; border-radius:50%; animation:spin .7s linear infinite; display:inline-block; }
@keyframes spin { to { transform:rotate(360deg) } }
</style>
