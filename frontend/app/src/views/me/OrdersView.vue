<template>
  <div class="orders-page">
    <h1 class="section-title">我的訂單</h1>

    <div v-if="loading" class="loading-box">載入中...</div>
    <div v-else-if="!orders.length" class="empty-box">
      <div class="empty-icon">📦</div>
      <p>還沒有任何訂單</p>
      <RouterLink to="/shop" class="btn btn-primary" style="margin-top:1rem">去逛商城</RouterLink>
    </div>

    <div v-else class="orders-list">
      <div v-for="o in orders" :key="o.id" class="order-card card">
        <div class="order-header">
          <div class="order-meta">
            <span class="order-no">訂單 #{{ o.uuid?.slice(0,8).toUpperCase() }}</span>
            <span class="order-date text-gray">{{ fmt(o.created_at) }}</span>
          </div>
          <div class="order-badges">
            <span class="badge" :class="statusBadge(o.status)">{{ statusLabel(o.status) }}</span>
            <span class="badge" :class="paymentBadge(o.payment_status)">{{ paymentLabel(o.payment_status) }}</span>
          </div>
        </div>

        <!-- 商品列表 -->
        <div class="order-items" v-if="o.items?.length">
          <div v-for="it in o.items" :key="it.id" class="order-item">
            <span class="item-title">{{ it.title }}</span>
            <span v-if="it.spec" class="item-spec text-gray">（{{ it.spec }}）</span>
            <span class="item-qty text-gray"> × {{ it.qty }}</span>
            <span class="item-price">NT$ {{ (it.price * it.qty).toLocaleString() }}</span>
          </div>
        </div>

        <div class="order-footer">
          <div class="order-delivery text-gray text-sm">
            <span>{{ deliveryLabel(o.delivery_method) }}</span>
            <span v-if="o.tracking_number"> · 物流：{{ o.tracking_number }}</span>
          </div>
          <div class="order-total">
            合計：<strong class="text-red">NT$ {{ Number(o.total_amount).toLocaleString() }}</strong>
          </div>
        </div>

        <!-- 付款資訊 -->
        <div v-if="o.payment_status === 0 && o.payment_method === 2" class="transfer-hint">
          💳 請於 3 個工作天內完成銀行轉帳，帳號資訊請聯繫管理員。
        </div>
      </div>
    </div>

    <div class="pagination" v-if="totalPages > 1">
      <button :disabled="page===1" @click="goPage(page-1)" class="btn btn-ghost">‹</button>
      <span class="text-gray">{{ page }} / {{ totalPages }}</span>
      <button :disabled="page===totalPages" @click="goPage(page+1)" class="btn btn-ghost">›</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import api from '@/services/api'

const orders = ref([])
const loading = ref(true)
const page = ref(1)
const totalPages = ref(1)

async function fetchOrders() {
  loading.value = true
  try {
    const { data } = await api.get('/orders', { params: { page: page.value, page_size: 10 } })
    // 取得每張訂單的 items
    const list = data.orders || []
    for (const o of list) {
      try {
        const r = await api.get(`/orders/${o.id}`)
        o.items = r.data.items || []
      } catch {}
    }
    orders.value = list
    totalPages.value = data.pages || 1
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

function goPage(p) { page.value = p; fetchOrders() }
function fmt(d) { return d ? new Date(d).toLocaleDateString('zh-TW', { year:'numeric', month:'2-digit', day:'2-digit' }) : '-' }
function statusLabel(s) { return {0:'待處理',1:'已付款',2:'已出貨',3:'已完成',4:'已取消',5:'已退款'}[s]||'未知' }
function statusBadge(s) { return {0:'badge-warning',1:'badge-primary',2:'badge-primary',3:'badge-success',4:'badge-gray',5:'badge-gray'}[s]||'badge-gray' }
function paymentLabel(s) { return {0:'未付款',1:'已付款',2:'已退款'}[s]||'' }
function paymentBadge(s) { return {0:'badge-warning',1:'badge-success',2:'badge-gray'}[s]||'badge-gray' }
function deliveryLabel(d) { return {1:'宅配',2:'自取'}[d]||'-' }

onMounted(fetchOrders)
</script>

<style scoped>
.orders-page { max-width:700px; }
.loading-box, .empty-box { padding:3rem; text-align:center; color:var(--color-gray-2); }
.empty-icon { font-size:3rem; margin-bottom:.75rem; }
.orders-list { display:flex; flex-direction:column; gap:1rem; }
.order-card { padding:1.25rem; }
.order-header { display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:.75rem; }
.order-no { font-family:var(--font-cond); font-weight:700; font-size:.9rem; }
.order-date { font-size:.78rem; display:block; margin-top:.15rem; }
.order-badges { display:flex; gap:.4rem; flex-wrap:wrap; }
.order-items { border-top:1px solid var(--color-border); border-bottom:1px solid var(--color-border); padding:.5rem 0; margin:.5rem 0; display:flex; flex-direction:column; gap:.35rem; }
.order-item { display:flex; align-items:center; gap:.25rem; font-size:.88rem; }
.item-title { font-weight:600; }
.item-spec { font-size:.78rem; }
.item-qty { font-size:.78rem; }
.item-price { margin-left:auto; font-family:var(--font-cond); font-weight:700; color:var(--color-primary); }
.order-footer { display:flex; justify-content:space-between; align-items:center; font-size:.85rem; }
.transfer-hint { margin-top:.75rem; padding:.6rem .9rem; background:rgba(245,158,11,.1); border:1px solid rgba(245,158,11,.3); border-radius:4px; font-size:.8rem; color:#f59e0b; }
.pagination { display:flex; align-items:center; justify-content:center; gap:1rem; margin-top:1.5rem; }
</style>
