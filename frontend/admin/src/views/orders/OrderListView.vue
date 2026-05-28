<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">訂單管理</h1>
      <p class="page-subtitle">查看並管理所有商城訂單</p>
    </div>

    <div class="card mb-2">
      <div class="card-body" style="padding:1rem">
        <div class="filter-row">
          <input v-model="filters.keyword" placeholder="搜尋帳號 / 收件人..." @keyup.enter="fetchOrders" style="flex:1;min-width:180px" />
          <select v-model="filters.status" @change="fetchOrders">
            <option value="">全部狀態</option>
            <option value="0">待處理</option><option value="1">已付款</option>
            <option value="2">已出貨</option><option value="3">已完成</option>
            <option value="4">已取消</option>
          </select>
          <button class="btn btn-primary btn-sm" @click="fetchOrders">搜尋</button>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-body" style="padding:0">
        <div v-if="loading" class="loading-row">載入中...</div>
        <table v-else class="table" style="font-size:.82rem">
          <thead>
            <tr><th>訂單</th><th>會員</th><th>商品</th><th>金額</th><th>取貨/付款</th><th>狀態</th><th>操作</th></tr>
          </thead>
          <tbody>
            <tr v-for="o in orders" :key="o.id">
              <td>
                <div class="fw-bold">{{ o.uuid?.slice(0,8).toUpperCase() }}</div>
                <div class="text-gray text-xs">{{ fmt(o.created_at) }}</div>
              </td>
              <td>{{ o.username || o.shipping_name }}</td>
              <td class="text-gray text-xs">{{ o.items?.map(i=>i.title).join(', ') || '—' }}</td>
              <td class="fw-bold text-red">NT$ {{ Number(o.total_amount).toLocaleString() }}</td>
              <td>
                <div class="text-xs">{{ deliveryLabel(o.delivery_method) }}</div>
                <div class="text-xs text-gray">{{ paymentMethodLabel(o.payment_method) }}</div>
              </td>
              <td>
                <div class="d-col">
                  <span class="badge" :class="statusBadge(o.status)">{{ statusLabel(o.status) }}</span>
                  <span class="badge" :class="payBadge(o.payment_status)" style="margin-top:.2rem">{{ payLabel(o.payment_status) }}</span>
                </div>
              </td>
              <td><button class="btn btn-sm btn-ghost" @click="openEdit(o)">管理</button></td>
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

    <!-- Edit Modal -->
    <div class="modal-overlay" v-if="editingOrder" @click.self="editingOrder=null">
      <div class="edit-modal">
        <div class="modal-header">
          <h3>管理訂單 #{{ editingOrder.uuid?.slice(0,8).toUpperCase() }}</h3>
          <button @click="editingOrder=null">✕</button>
        </div>
        <div class="modal-body">
          <!-- 狀態 -->
          <div class="form-section-label">訂單狀態</div>
          <div class="toggle-row">
            <button v-for="s in orderStatuses" :key="s.value"
              class="toggle-btn" :class="{ active: editForm.status === s.value }"
              @click="editForm.status = s.value" type="button">{{ s.label }}</button>
          </div>
          <div class="form-section-label" style="margin-top:.75rem">付款狀態</div>
          <div class="toggle-row">
            <button v-for="s in paymentStatuses" :key="s.value"
              class="toggle-btn" :class="{ active: editForm.payment_status === s.value }"
              @click="editForm.payment_status = s.value" type="button">{{ s.label }}</button>
          </div>

          <!-- 商品列表（唯讀） -->
          <div class="form-section-label">訂購商品</div>
          <div class="order-items-list">
            <div v-for="it in editingOrder.items" :key="it.id" class="order-item-row">
              <span>{{ it.title }}</span>
              <span v-if="it.spec" class="text-gray"> ({{ it.spec }})</span>
              <span class="text-gray"> × {{ it.qty }}</span>
              <span class="fw-bold text-red" style="margin-left:auto">NT$ {{ (it.price*it.qty).toLocaleString() }}</span>
            </div>
          </div>

          <!-- 收件資訊 -->
          <div class="form-section-label">收件資訊</div>
          <div class="form-row">
            <div class="form-group">
              <label>收件人</label><input v-model="editForm.shipping_name" />
            </div>
            <div class="form-group">
              <label>收件手機</label><input v-model="editForm.shipping_phone" />
            </div>
          </div>
          <div class="form-group mb-1">
            <label>收件地址</label><input v-model="editForm.shipping_addr" />
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>物流單號</label><input v-model="editForm.tracking_number" placeholder="填寫後自動顯示給會員" />
            </div>
            <div class="form-group">
              <label>備註</label><input v-model="editForm.note" />
            </div>
          </div>

          <div class="modal-footer">
            <button class="btn btn-primary" @click="saveOrder" :disabled="saveLoading">
              {{ saveLoading ? '儲存中...' : '儲存變更' }}
            </button>
            <button class="btn btn-ghost" @click="editingOrder=null">取消</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import api from '@/services/api'

const orders = ref([])
const loading = ref(false)
const page = ref(1)
const totalPages = ref(1)
const filters = reactive({ keyword: '', status: '' })
const editingOrder = ref(null)
const saveLoading = ref(false)
const editForm = reactive({
  status: null, payment_status: null,
  shipping_name: '', shipping_phone: '', shipping_addr: '',
  tracking_number: '', note: '',
})

const orderStatuses = [
  { value:0, label:'待處理' }, { value:1, label:'已付款' },
  { value:2, label:'已出貨' }, { value:3, label:'已完成' },
  { value:4, label:'已取消' },
]
const paymentStatuses = [
  { value:0, label:'未付款' }, { value:1, label:'已付款' }, { value:2, label:'已退款' },
]

async function fetchOrders() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: 20 }
    if (filters.keyword) params.keyword = filters.keyword
    if (filters.status !== '') params.status = filters.status
    const { data } = await api.get('/orders', { params })
    const list = data.orders || []
    // Load items for each order
    for (const o of list) {
      try { const r = await api.get(`/orders/${o.id}`); o.items = r.data.items || [] } catch {}
    }
    orders.value = list
    totalPages.value = data.pages || 1
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

function openEdit(o) {
  editingOrder.value = o
  Object.assign(editForm, {
    status: o.status, payment_status: o.payment_status,
    shipping_name: o.shipping_name || '', shipping_phone: o.shipping_phone || '',
    shipping_addr: o.shipping_addr || '', tracking_number: o.tracking_number || '',
    note: o.note || '',
  })
}

async function saveOrder() {
  saveLoading.value = true
  try {
    const r = await api.put(`/orders/${editingOrder.value.id}`, editForm)
    // Update local
    const idx = orders.value.findIndex(o => o.id === editingOrder.value.id)
    if (idx >= 0) Object.assign(orders.value[idx], r.data.order)
    editingOrder.value = null
  } catch(e) {
    alert(e.response?.data?.error || '更新失敗')
  } finally {
    saveLoading.value = false
  }
}

function goPage(p) { page.value = p; fetchOrders() }
function fmt(d) { return d ? new Date(d).toLocaleDateString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit' }) : '-' }
function statusLabel(s) { return {0:'待處理',1:'已付款',2:'已出貨',3:'已完成',4:'已取消'}[s]||'未知' }
function statusBadge(s) { return {0:'badge-warning',1:'badge-primary',2:'badge-primary',3:'badge-success',4:'badge-gray'}[s]||'badge-gray' }
function payLabel(s)  { return {0:'未付款',1:'已付款',2:'已退款'}[s]||'' }
function payBadge(s)  { return {0:'badge-warning',1:'badge-success',2:'badge-gray'}[s]||'badge-gray' }
function deliveryLabel(d) { return {1:'宅配',2:'自取'}[d]||'-' }
function paymentMethodLabel(m) { return {1:'信用卡',2:'轉帳',3:'LINE Pay',4:'現金'}[m]||'-' }

onMounted(fetchOrders)
</script>

<style scoped>
.filter-row { display:flex;gap:.75rem;flex-wrap:wrap;align-items:center; }
.filter-row input, .filter-row select { height:36px;font-size:.85rem; }
.fw-bold { font-weight:600; }.text-xs { font-size:.75rem; }
.loading-row { padding:3rem;text-align:center;color:var(--gray-2); }
.d-col { display:flex;flex-direction:column; }
.pagination { display:flex;align-items:center;justify-content:center;gap:1rem;padding:1rem;border-top:1px solid var(--border); }
.modal-overlay { position:fixed;inset:0;background:rgba(0,0,0,.75);z-index:100;display:flex;align-items:center;justify-content:center;padding:1rem; }
.edit-modal { background:var(--bg-card);border:1px solid var(--border);border-radius:8px;width:100%;max-width:580px;max-height:90vh;overflow-y:auto; }
.modal-header { display:flex;align-items:center;justify-content:space-between;padding:1.25rem 1.5rem;border-bottom:1px solid var(--border);position:sticky;top:0;background:var(--bg-card); }
.modal-header h3 { font-family:var(--font-c);font-size:1.1rem;font-weight:700; }
.modal-header button { background:none;border:none;color:var(--gray-2);font-size:1.2rem;cursor:pointer; }
.modal-body { padding:1.5rem; }
.form-section-label { font-size:.72rem;font-weight:700;letter-spacing:.1em;text-transform:uppercase;color:var(--gray-2);margin:.75rem 0 .5rem; }
.toggle-row { display:flex;gap:.5rem;flex-wrap:wrap; }
.toggle-btn { padding:.4rem .9rem;border-radius:4px;border:1px solid var(--border);font-size:.8rem;font-weight:600;cursor:pointer;background:var(--bg);color:var(--gray-2);transition:all .15s; }
.toggle-btn.active { background:var(--primary);border-color:var(--primary);color:#fff; }
.order-items-list { background:var(--bg);border:1px solid var(--border);border-radius:4px;padding:.75rem;margin-bottom:.5rem; }
.order-item-row { display:flex;align-items:center;gap:.25rem;font-size:.85rem;padding:.2rem 0; }
.form-row { display:grid;grid-template-columns:1fr 1fr;gap:.75rem;margin-bottom:.75rem; }
.form-group { display:flex;flex-direction:column;gap:.3rem; }
.form-group.mb-1 { margin-bottom:.75rem; }
.form-group label { font-size:.72rem;font-weight:600;text-transform:uppercase;letter-spacing:.06em;color:var(--gray-1); }
.form-group input { width:100%; }
.modal-footer { display:flex;gap:.75rem;margin-top:1.25rem;padding-top:1rem;border-top:1px solid var(--border); }
</style>
