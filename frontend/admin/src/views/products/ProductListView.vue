<template>
  <div>
    <div class="page-header flex justify-between items-center">
      <div><h1 class="page-title">商品管理</h1><p class="page-subtitle">新增、編輯、管理商城商品</p></div>
      <button class="btn btn-primary" @click="openCreate">＋ 新增商品</button>
    </div>

    <!-- Filters -->
    <div class="card mb-2">
      <div class="card-body" style="padding:1rem">
        <div class="filter-row">
          <input v-model="filters.keyword" placeholder="搜尋商品名稱..." @keyup.enter="fetchProducts" style="flex:1;min-width:180px" />
          <select v-model="filters.status" @change="fetchProducts">
            <option value="">全部狀態</option>
            <option value="0">草稿</option><option value="1">已上架</option><option value="2">售完</option>
          </select>
          <button class="btn btn-primary btn-sm" @click="fetchProducts">搜尋</button>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-body" style="padding:0">
        <div v-if="loading" class="loading-row">載入中...</div>
        <table v-else class="table">
          <thead><tr><th style="width:70px">圖片</th><th>商品</th><th>分類</th><th>價格</th><th>庫存</th><th>狀態</th><th>操作</th></tr></thead>
          <tbody>
            <tr v-for="p in products" :key="p.id">
              <td>
                <div class="product-thumb">
                  <img v-if="p.images?.length" :src="imgUrl(p.images[0])" />
                  <span v-else>🛍</span>
                </div>
              </td>
              <td><div class="fw-bold">{{ p.title }}</div></td>
              <td><span class="text-gray text-xs">{{ catLabel(p.category) }}</span></td>
              <td class="text-red fw-bold">NT$ {{ Number(p.price).toLocaleString() }}</td>
              <td>{{ p.stock }}</td>
              <td><span class="badge" :class="statusBadge(p.status)">{{ statusLabel(p.status) }}</span></td>
              <td>
                <div class="action-btns">
                  <button class="btn btn-sm btn-ghost" @click="openEdit(p)">編輯</button>
                  <button class="btn btn-sm btn-ghost" style="color:var(--danger)" @click="confirmDelete(p)">刪除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create / Edit Modal -->
    <div class="modal-overlay" v-if="showForm" @click.self="showForm=false">
      <div class="product-form-modal">
        <div class="modal-header">
          <h3>{{ editingId ? '編輯商品' : '新增商品' }}</h3>
          <button @click="showForm=false">✕</button>
        </div>
        <div class="modal-body">
          <!-- 基本資訊 -->
          <div class="form-section-label">基本資訊</div>
          <div class="form-row">
            <div class="form-group full">
              <label>商品名稱 *</label>
              <input v-model="form.title" placeholder="商品名稱" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>分類</label>
              <select v-model.number="form.category">
                <option :value="1">服裝</option><option :value="2">裝備</option>
                <option :value="3">補給</option><option :value="4">配件</option>
              </select>
            </div>
            <div class="form-group">
              <label>狀態</label>
              <select v-model.number="form.status">
                <option :value="0">草稿</option><option :value="1">上架</option><option :value="2">售完</option>
              </select>
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>價格 (NT$) *</label>
              <input v-model.number="form.price" type="number" min="0" />
            </div>
            <div class="form-group">
              <label>庫存數量</label>
              <input v-model.number="form.stock" type="number" min="0" />
            </div>
          </div>

          <!-- 商品圖片（多張） -->
          <div class="form-section-label">商品圖片（可上傳多張）</div>
          <div class="images-grid">
            <div v-for="(img, i) in form.images" :key="i" class="image-slot">
              <img :src="imgUrl(img)" @error="removeImage(i)" />
              <button class="remove-img-btn" @click="removeImage(i)" type="button">✕</button>
            </div>
            <div class="image-upload-slot" @click="triggerImageUpload">
              <input ref="imgFileInput" type="file" accept="image/*" multiple style="display:none" @change="onImagesSelected" />
              <div v-if="imgUploading" class="uploading-hint">上傳中...</div>
              <div v-else class="upload-hint-inner">
                <span class="upload-plus">＋</span>
                <span>新增圖片</span>
              </div>
            </div>
          </div>

          <!-- 規格 -->
          <div class="form-section-label">規格選項（每行一個，如 S / M / L）</div>
          <textarea v-model="specsText" rows="3" placeholder="S&#10;M&#10;L&#10;XL" class="form-textarea"></textarea>

          <!-- 商品說明 -->
          <div class="form-section-label">商品說明</div>
          <textarea v-model="form.description" rows="5" placeholder="商品詳細說明..." class="form-textarea"></textarea>

          <div v-if="formError" class="form-error mt-1">{{ formError }}</div>
          <div class="modal-footer">
            <button class="btn btn-primary" @click="submitForm" :disabled="formLoading">
              <span v-if="formLoading" class="spinner-sm"></span>
              {{ formLoading ? '儲存中...' : (editingId ? '更新' : '建立') }}
            </button>
            <button class="btn btn-ghost" @click="showForm=false">取消</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete confirm -->
    <div class="modal-overlay" v-if="deletingProduct" @click.self="deletingProduct=null">
      <div class="confirm-modal">
        <h3>確認刪除商品？</h3>
        <p class="text-gray">「{{ deletingProduct?.title }}」將被永久刪除。</p>
        <div class="confirm-actions">
          <button class="btn btn-danger" @click="doDelete">確認刪除</button>
          <button class="btn btn-ghost" @click="deletingProduct=null">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api from '@/services/api'

const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || ''
function imgUrl(path) {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `${IMAGE_BASE}/images/${path}`
}

const products = ref([])
const loading = ref(false)
const filters = reactive({ keyword: '', status: '' })
const showForm = ref(false)
const editingId = ref(null)
const formLoading = ref(false)
const formError = ref('')
const deletingProduct = ref(null)
const imgFileInput = ref(null)
const imgUploading = ref(false)

const emptyForm = () => ({
  title: '', description: '', category: 1, price: 0, stock: 0, images: [], specs: [], status: 1,
})
const form = reactive(emptyForm())
const specsText = computed({
  get: () => form.specs.join('\n'),
  set: (v) => { form.specs = v.split('\n').map(s => s.trim()).filter(Boolean) }
})

async function fetchProducts() {
  loading.value = true
  try {
    const params = {}
    if (filters.keyword) params.keyword = filters.keyword
    if (filters.status !== '') params.status = filters.status
    const { data } = await api.get('/products', { params })
    products.value = data.products || []
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

function openCreate() {
  Object.assign(form, emptyForm())
  form.images = []; form.specs = []
  editingId.value = null; formError.value = ''
  showForm.value = true
}

function openEdit(p) {
  Object.assign(form, {
    title: p.title, description: p.description || '',
    category: p.category, price: p.price, stock: p.stock,
    status: p.status, images: [...(p.images || [])], specs: [...(p.specs || [])],
  })
  editingId.value = p.id; formError.value = ''
  showForm.value = true
}

function triggerImageUpload() { imgFileInput.value?.click() }

async function onImagesSelected(e) {
  const files = Array.from(e.target.files || [])
  if (!files.length) return
  imgUploading.value = true
  try {
    for (const file of files) {
      const fd = new FormData()
      fd.append('file', file)
      const { data } = await api.post('/upload/image?folder=products', fd, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      form.images.push(data.path)
    }
  } catch(e) {
    formError.value = e.response?.data?.error || '圖片上傳失敗'
  } finally {
    imgUploading.value = false
    e.target.value = ''
  }
}

function removeImage(i) { form.images.splice(i, 1) }

async function submitForm() {
  formError.value = ''
  if (!form.title) { formError.value = '請填寫商品名稱'; return }
  if (!form.price)  { formError.value = '請填寫價格'; return }

  formLoading.value = true
  try {
    if (editingId.value) {
      await api.put(`/products/${editingId.value}`, form)
    } else {
      await api.post('/products', form)
    }
    showForm.value = false
    await fetchProducts()
  } catch(e) {
    formError.value = e.response?.data?.error || '儲存失敗'
  } finally {
    formLoading.value = false
  }
}

function confirmDelete(p) { deletingProduct.value = p }
async function doDelete() {
  try {
    await api.delete(`/products/${deletingProduct.value.id}`)
    deletingProduct.value = null
    await fetchProducts()
  } catch(e) { alert(e.response?.data?.error || '刪除失敗') }
}

function catLabel(c) { return { 1:'服裝',2:'裝備',3:'補給',4:'配件' }[c]||'其他' }
function statusLabel(s) { return { 0:'草稿',1:'已上架',2:'售完' }[s]||'未知' }
function statusBadge(s) { return { 0:'badge-gray',1:'badge-success',2:'badge-warning' }[s]||'badge-gray' }

onMounted(fetchProducts)
</script>

<style scoped>
.filter-row { display:flex;gap:.75rem;flex-wrap:wrap;align-items:center; }
.filter-row input, .filter-row select { height:36px;font-size:.85rem; }
.product-thumb { width:60px;height:40px;border-radius:4px;overflow:hidden;background:var(--bg);border:1px solid var(--border);display:flex;align-items:center;justify-content:center;font-size:1.2rem; }
.product-thumb img { width:100%;height:100%;object-fit:cover; }
.fw-bold { font-weight:600; }.text-xs { font-size:.75rem; }
.action-btns { display:flex;gap:.35rem; }
.loading-row { padding:3rem;text-align:center;color:var(--gray-2); }
.modal-overlay { position:fixed;inset:0;background:rgba(0,0,0,.75);z-index:100;display:flex;align-items:center;justify-content:center;padding:1rem; }
.product-form-modal { background:var(--bg-card);border:1px solid var(--border);border-radius:8px;width:100%;max-width:680px;max-height:92vh;overflow-y:auto; }
.modal-header { display:flex;align-items:center;justify-content:space-between;padding:1.25rem 1.5rem;border-bottom:1px solid var(--border);position:sticky;top:0;background:var(--bg-card); }
.modal-header h3 { font-family:var(--font-c);font-size:1.1rem;font-weight:700; }
.modal-header button { background:none;border:none;color:var(--gray-2);font-size:1.2rem;cursor:pointer; }
.modal-body { padding:1.5rem; }
.form-section-label { font-size:.72rem;font-weight:700;letter-spacing:.1em;text-transform:uppercase;color:var(--gray-2);margin:.75rem 0 .5rem; }
.form-row { display:grid;grid-template-columns:1fr 1fr;gap:1rem;margin-bottom:.75rem; }
.form-group { display:flex;flex-direction:column;gap:.3rem; }
.form-group.full { grid-column:1/-1; }
.form-group label { font-size:.72rem;font-weight:600;text-transform:uppercase;letter-spacing:.06em;color:var(--gray-1); }
.form-group input, .form-group select { width:100%; }
.form-textarea { width:100%;background:var(--bg);color:#fff;border:1px solid var(--border);border-radius:4px;padding:.6rem .9rem;font-family:inherit;font-size:.85rem;resize:vertical; }

/* Images grid */
.images-grid { display:flex;flex-wrap:wrap;gap:.6rem;margin-bottom:.75rem; }
.image-slot { width:80px;height:80px;border-radius:6px;overflow:hidden;position:relative;border:1px solid var(--border); }
.image-slot img { width:100%;height:100%;object-fit:cover; }
.remove-img-btn { position:absolute;top:2px;right:2px;width:20px;height:20px;border-radius:50%;background:rgba(0,0,0,.65);border:none;color:#fff;font-size:.7rem;cursor:pointer;display:flex;align-items:center;justify-content:center; }
.image-upload-slot { width:80px;height:80px;border-radius:6px;border:2px dashed var(--border);display:flex;flex-direction:column;align-items:center;justify-content:center;cursor:pointer;transition:border-color .15s; }
.image-upload-slot:hover { border-color:var(--primary); }
.upload-hint-inner { display:flex;flex-direction:column;align-items:center;gap:.2rem;color:var(--gray-2);font-size:.7rem; }
.upload-plus { font-size:1.4rem;line-height:1; }
.uploading-hint { font-size:.7rem;color:var(--gray-2);text-align:center; }

.form-error { background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.3);border-radius:4px;color:#fca5a5;font-size:.83rem;padding:.5rem .75rem; }
.modal-footer { display:flex;gap:.75rem;margin-top:1.25rem;padding-top:1rem;border-top:1px solid var(--border); }
.spinner-sm { width:12px;height:12px;border:2px solid rgba(255,255,255,.3);border-top-color:#fff;border-radius:50%;animation:spin .7s linear infinite;display:inline-block; }
@keyframes spin { to { transform:rotate(360deg) } }
.mt-1 { margin-top:.5rem; }
.confirm-modal { background:var(--bg-card);border:1px solid var(--border);border-radius:8px;padding:2rem;max-width:400px;width:100%; }
.confirm-modal h3 { font-size:1.1rem;margin-bottom:.75rem; }
.confirm-actions { display:flex;gap:.75rem;margin-top:1.5rem; }
.btn-danger { background:var(--danger);color:#fff; }
</style>
