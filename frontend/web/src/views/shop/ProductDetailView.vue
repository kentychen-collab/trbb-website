<template>
  <div class="product-page">

    <div v-if="loading" class="loading-full"><div class="loading-spinner"></div></div>

    <template v-else-if="product">
      <div class="container product-main">
        <!-- 圖片區 -->
        <div class="product-gallery">
          <div class="main-image">
            <img v-if="activeImage" :src="imgUrl(activeImage)" :alt="product.title" />
            <div v-else class="no-image">🛍</div>
          </div>
          <div class="thumbnails" v-if="product.images && product.images.length > 1">
            <div v-for="(img, i) in product.images" :key="i"
              class="thumb" :class="{ active: activeImage === img }"
              @click="activeImage = img">
              <img :src="imgUrl(img)" :alt="`圖片 ${i+1}`" />
            </div>
          </div>
        </div>

        <!-- 資訊區 -->
        <div class="product-info">
          <div class="product-cat">{{ catLabel(product.category) }}</div>
          <h1 class="product-title">{{ product.title }}</h1>
          <div class="product-price">NT$ {{ Number(product.price).toLocaleString() }}</div>
          <div class="product-stock" :class="product.stock > 0 ? 'in-stock' : 'out-stock'">
            {{ product.stock > 0 ? `庫存剩 ${product.stock} 件` : '已售完' }}
          </div>

          <!-- 規格 -->
          <div v-if="product.specs && product.specs.length" class="product-specs">
            <div class="specs-label">規格選項</div>
            <div class="specs-btns">
              <button v-for="spec in product.specs" :key="spec"
                class="spec-btn" :class="{ active: selectedSpec === spec }"
                @click="selectedSpec = spec" type="button">
                {{ spec }}
              </button>
            </div>
          </div>

          <!-- 數量 -->
          <div class="qty-row">
            <button class="qty-btn" @click="qty = Math.max(1, qty-1)">−</button>
            <span class="qty-num">{{ qty }}</span>
            <button class="qty-btn" @click="qty = Math.min(product.stock, qty+1)">＋</button>
          </div>

          <!-- 按鈕 -->
          <div v-if="!auth.isLoggedIn">
            <RouterLink to="/login" class="btn btn-primary buy-btn">登入後購買</RouterLink>
          </div>
          <div v-else>
            <button class="btn btn-primary buy-btn"
              :disabled="product.stock === 0"
              @click="openOrder">
              {{ product.stock > 0 ? '立即購買' : '已售完' }}
            </button>
          </div>

          <!-- 描述 -->
          <div class="product-desc" v-if="product.description">
            <div class="desc-title">商品說明</div>
            <div class="desc-body" v-html="descHtml"></div>
          </div>
        </div>
      </div>
    </template>

    <!-- ── 訂購 Modal ──────────────────────────────────────── -->
    <div class="modal-overlay" v-if="showOrder" @click.self="showOrder=false">
      <div class="order-modal">
        <div class="order-modal-header">
          <h2>填寫訂購資訊</h2>
          <button @click="showOrder=false">✕</button>
        </div>
        <div class="order-modal-body">
          <!-- 訂購摘要 -->
          <div class="order-summary">
            <div class="summary-img">
              <img v-if="product?.images?.length" :src="imgUrl(product.images[0])" />
              <span v-else>🛍</span>
            </div>
            <div class="summary-info">
              <div class="summary-title">{{ product?.title }}</div>
              <div v-if="selectedSpec" class="summary-spec text-gray">規格：{{ selectedSpec }}</div>
              <div class="summary-qty text-gray">數量：{{ qty }}</div>
              <div class="summary-price text-red">NT$ {{ (Number(product?.price) * qty).toLocaleString() }}</div>
            </div>
          </div>

          <!-- 收件資訊 -->
          <div class="order-section-title">收件資訊</div>
          <div class="order-form-row">
            <div class="form-group">
              <label>收件人姓名 <span class="req">*</span></label>
              <input v-model="orderForm.shipping_name" placeholder="收件人全名" />
            </div>
            <div class="form-group">
              <label>收件人手機 <span class="req">*</span></label>
              <input v-model="orderForm.shipping_phone" type="tel" placeholder="09xxxxxxxx" />
            </div>
          </div>

          <!-- 取貨方式 -->
          <div class="order-section-title">取貨方式 <span class="req">*</span></div>
          <div class="toggle-row">
            <button v-for="d in deliveryOptions" :key="d.value"
              class="toggle-btn" :class="{ active: orderForm.delivery_method === d.value }"
              @click="orderForm.delivery_method = d.value" type="button">
              {{ d.label }}
            </button>
          </div>
          <div v-if="orderForm.delivery_method === 1" class="form-group mt-1">
            <label>收件地址 <span class="req">*</span></label>
            <input v-model="orderForm.shipping_addr" placeholder="縣市 + 完整地址" />
          </div>
          <div v-if="orderForm.delivery_method === 2" class="form-group mt-1">
            <label>自取地點</label>
            <input v-model="orderForm.pickup_location" placeholder="自取地點說明" readonly class="disabled-input" value="TRBB 社團辦公室（詳情請洽管理員）" />
          </div>

          <!-- 付款方式 -->
          <div class="order-section-title">付款方式 <span class="req">*</span></div>
          <div class="toggle-row">
            <button v-for="pm in paymentOptions" :key="pm.value"
              class="toggle-btn" :class="{ active: orderForm.payment_method === pm.value }"
              @click="orderForm.payment_method = pm.value" type="button">
              {{ pm.label }}
            </button>
          </div>
          <div v-if="orderForm.payment_method === 2" class="payment-hint">
            銀行轉帳資訊將於訂單成立後顯示，請於 3 個工作天內完成匯款。
          </div>
          <div v-if="orderForm.payment_method === 3" class="payment-hint">
            LINE Pay 付款功能即將上線，目前暫不提供。
          </div>

          <!-- 備註 -->
          <div class="form-group mt-1">
            <label>備註（選填）</label>
            <input v-model="orderForm.note" placeholder="其他需要告知事項" />
          </div>

          <div v-if="orderError" class="order-error">{{ orderError }}</div>

          <!-- 總計 -->
          <div class="order-total">
            <span>訂單總計</span>
            <span class="total-price">NT$ {{ (Number(product?.price) * qty).toLocaleString() }}</span>
          </div>

          <div class="order-modal-footer">
            <button class="btn btn-primary order-submit" @click="submitOrder" :disabled="orderLoading">
              <span v-if="orderLoading" class="spinner"></span>
              {{ orderLoading ? '建立中...' : '確認下單' }}
            </button>
            <button class="btn btn-ghost" @click="showOrder=false">取消</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 訂單成功 -->
    <div class="modal-overlay" v-if="orderDone" @click.self="orderDone=false">
      <div class="success-modal">
        <div class="success-icon">✅</div>
        <h2>訂單已建立！</h2>
        <p>訂單編號：<strong>{{ createdOrder?.uuid?.slice(0,8).toUpperCase() }}</strong></p>
        <p class="text-gray">可至會員中心查看訂單狀態</p>
        <div class="success-actions">
          <RouterLink to="/me/orders" class="btn btn-primary">查看我的訂單</RouterLink>
          <RouterLink to="/shop" class="btn btn-ghost">繼續購物</RouterLink>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'

const route   = useRoute()
const auth    = useAuthStore()
const product = ref(null)
const loading = ref(true)
const activeImage = ref('')
const selectedSpec = ref('')
const qty = ref(1)
const showOrder = ref(false)
const orderLoading = ref(false)
const orderError = ref('')
const orderDone = ref(false)
const createdOrder = ref(null)

const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || ''
function imgUrl(path) {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `${IMAGE_BASE}/images/${path}`
}

const deliveryOptions = [
  { value: 1, label: '🚚 宅配' },
  { value: 2, label: '🏪 自取' },
]
const paymentOptions = [
  { value: 1, label: '💳 信用卡' },
  { value: 2, label: '🏦 轉帳' },
  { value: 3, label: '💚 LINE Pay' },
  { value: 4, label: '💵 現金' },
]

const orderForm = reactive({
  shipping_name: '', shipping_phone: '', shipping_addr: '',
  note: '', payment_method: 1, delivery_method: 1, pickup_location: '',
})

const descHtml = computed(() => (product.value?.description || '').replace(/\n/g, '<br>'))

async function openOrder() {
  orderError.value = ''
  // 自動帶入會員資料
  try {
    const { data } = await api.get('/me/registration-profile')
    orderForm.shipping_name  = data.name_zh || data.name_en || data.display_name || ''
    orderForm.shipping_phone = data.phone || ''
    orderForm.shipping_addr  = data.address || ''
  } catch {}
  showOrder.value = true
}

async function submitOrder() {
  orderError.value = ''
  if (!orderForm.shipping_name)  { orderError.value = '請填寫收件人姓名'; return }
  if (!orderForm.shipping_phone) { orderError.value = '請填寫收件人手機'; return }
  if (orderForm.delivery_method === 1 && !orderForm.shipping_addr) {
    orderError.value = '宅配請填寫收件地址'; return
  }
  if (!orderForm.payment_method) { orderError.value = '請選擇付款方式'; return }

  orderLoading.value = true
  try {
    const payload = {
      items: [{ product_id: product.value.id, qty: qty.value, spec: selectedSpec.value }],
      ...orderForm,
    }
    const { data } = await api.post('/orders', payload)
    createdOrder.value = data.order
    showOrder.value = false
    orderDone.value = true
  } catch(e) {
    orderError.value = e.response?.data?.error || '建立訂單失敗，請稍後再試'
  } finally {
    orderLoading.value = false
  }
}

function catLabel(c) { return { 1:'服裝', 2:'裝備', 3:'補給', 4:'配件' }[c] || '其他' }

onMounted(async () => {
  try {
    const { data } = await api.get(`/products/${route.params.id}`)
    product.value = data
    if (data.images?.length) activeImage.value = data.images[0]
  } catch {}
  finally { loading.value = false }
})
</script>

<style scoped>
.product-page { background:var(--color-bg);min-height:100vh; }
.nav-link:hover { color:var(--color-primary); }
.loading-full { display:flex;align-items:center;justify-content:center;height:100vh; }
.loading-spinner { width:32px;height:32px;border:3px solid var(--color-border);border-top-color:var(--color-primary);border-radius:50%;animation:spin .7s linear infinite; }
@keyframes spin { to { transform:rotate(360deg) } }

.product-main { display:grid;grid-template-columns:1fr 1fr;gap:3rem;padding-top:calc(64px + 2rem);padding-bottom:4rem;align-items:start; }
@media(max-width:768px){ .product-main { grid-template-columns:1fr } }

/* Gallery */
.product-gallery {}
.main-image { width:100%;aspect-ratio:1;border-radius:8px;overflow:hidden;background:var(--color-bg-card);border:1px solid var(--color-border);display:flex;align-items:center;justify-content:center;font-size:6rem; }
.main-image img { width:100%;height:100%;object-fit:cover; }
.thumbnails { display:flex;gap:.5rem;margin-top:.75rem;flex-wrap:wrap; }
.thumb { width:70px;height:70px;border-radius:4px;overflow:hidden;cursor:pointer;border:2px solid transparent;transition:border-color .15s; }
.thumb.active { border-color:var(--color-primary); }
.thumb img { width:100%;height:100%;object-fit:cover; }

/* Info */
.product-cat { font-family:var(--font-cond);font-size:.7rem;letter-spacing:.15em;text-transform:uppercase;color:var(--color-primary);margin-bottom:.5rem; }
.product-title { font-size:1.8rem;font-weight:700;line-height:1.3;margin-bottom:.75rem; }
.product-price { font-family:var(--font-display);font-size:2.5rem;color:var(--color-primary);margin-bottom:.5rem; }
.product-stock { font-size:.85rem;margin-bottom:1.25rem;font-weight:600; }
.in-stock { color:#22c55e; }
.out-stock { color:#9ca3af; }

.product-specs { margin-bottom:1.25rem; }
.specs-label { font-size:.78rem;font-weight:600;text-transform:uppercase;letter-spacing:.08em;color:var(--color-gray-2);margin-bottom:.5rem; }
.specs-btns { display:flex;gap:.5rem;flex-wrap:wrap; }
.spec-btn { padding:.35rem .9rem;border-radius:4px;border:1px solid var(--color-border);font-size:.85rem;cursor:pointer;background:none;color:var(--color-gray-1);transition:all .15s; }
.spec-btn.active { border-color:var(--color-primary);color:var(--color-primary); }

.qty-row { display:flex;align-items:center;gap:1rem;margin-bottom:1.5rem; }
.qty-btn { width:36px;height:36px;border-radius:4px;border:1px solid var(--color-border);background:none;color:#fff;font-size:1.2rem;cursor:pointer;display:flex;align-items:center;justify-content:center; }
.qty-btn:hover { border-color:var(--color-primary); }
.qty-num { font-family:var(--font-cond);font-size:1.3rem;font-weight:700;min-width:32px;text-align:center; }
.buy-btn { width:100%;padding:1rem;font-size:1rem; }

.product-desc { margin-top:1.5rem;padding-top:1.5rem;border-top:1px solid var(--color-border); }
.desc-title { font-family:var(--font-cond);font-size:.78rem;font-weight:700;letter-spacing:.1em;text-transform:uppercase;color:var(--color-gray-2);margin-bottom:.75rem; }
.desc-body { color:var(--color-gray-1);line-height:1.9;font-size:.95rem; }

/* Order Modal */
.modal-overlay { position:fixed;inset:0;background:rgba(0,0,0,.75);z-index:200;display:flex;align-items:center;justify-content:center;padding:1rem; }
.order-modal { background:var(--color-bg-card);border:1px solid var(--color-border);border-radius:10px;width:100%;max-width:560px;max-height:90vh;overflow-y:auto; }
.order-modal-header { display:flex;align-items:center;justify-content:space-between;padding:1.25rem 1.5rem;border-bottom:1px solid var(--color-border);position:sticky;top:0;background:var(--color-bg-card);z-index:1; }
.order-modal-header h2 { font-family:var(--font-cond);font-size:1.2rem;font-weight:700; }
.order-modal-header button { background:none;border:none;color:var(--color-gray-2);font-size:1.2rem;cursor:pointer; }
.order-modal-body { padding:1.5rem; }

.order-summary { display:flex;gap:1rem;align-items:center;background:var(--color-bg);border:1px solid var(--color-border);border-radius:6px;padding:.75rem;margin-bottom:1.25rem; }
.summary-img { width:60px;height:60px;border-radius:4px;overflow:hidden;flex-shrink:0;background:var(--color-bg-card);display:flex;align-items:center;justify-content:center;font-size:1.5rem; }
.summary-img img { width:100%;height:100%;object-fit:cover; }
.summary-title { font-weight:700;margin-bottom:.2rem; }
.summary-price { font-family:var(--font-cond);font-size:1.1rem;font-weight:700;margin-top:.25rem; }

.order-section-title { font-size:.78rem;font-weight:700;text-transform:uppercase;letter-spacing:.1em;color:var(--color-gray-2);margin:.75rem 0 .5rem; }
.order-form-row { display:grid;grid-template-columns:1fr 1fr;gap:.75rem; }
.form-group { display:flex;flex-direction:column;gap:.3rem; }
.form-group label { font-size:.75rem;font-weight:600;text-transform:uppercase;letter-spacing:.06em;color:var(--color-gray-1); }
.form-group input { width:100%; }
.mt-1 { margin-top:.75rem; }
.disabled-input { opacity:.6;cursor:default; }
.req { color:var(--color-primary); }

.toggle-row { display:flex;gap:.5rem;flex-wrap:wrap; }
.toggle-btn { padding:.45rem 1rem;border-radius:4px;border:1px solid var(--color-border);font-size:.85rem;font-weight:600;cursor:pointer;background:none;color:var(--color-gray-1);transition:all .15s; }
.toggle-btn.active { background:var(--color-primary);border-color:var(--color-primary);color:#fff; }

.payment-hint { font-size:.78rem;color:var(--color-gray-2);background:rgba(229,25,26,.06);border:1px solid rgba(229,25,26,.15);border-radius:4px;padding:.5rem .75rem;margin-top:.5rem; }

.order-total { display:flex;justify-content:space-between;align-items:center;padding:1rem;background:var(--color-bg);border:1px solid var(--color-border);border-radius:6px;margin:.75rem 0;font-weight:700; }
.total-price { font-family:var(--font-display);font-size:1.5rem;color:var(--color-primary); }
.order-error { background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.3);border-radius:4px;color:#fca5a5;font-size:.83rem;padding:.5rem .75rem;margin-bottom:.75rem; }
.order-modal-footer { display:flex;gap:.75rem; }
.order-submit { flex:1;padding:.85rem;font-size:.95rem;display:flex;align-items:center;justify-content:center;gap:.5rem; }
.spinner { width:14px;height:14px;border:2px solid rgba(255,255,255,.3);border-top-color:#fff;border-radius:50%;animation:spin .7s linear infinite; }

.success-modal { background:var(--color-bg-card);border:1px solid var(--color-border);border-radius:10px;padding:2.5rem;max-width:400px;width:100%;text-align:center; }
.success-icon { font-size:3rem;margin-bottom:1rem; }
.success-modal h2 { font-size:1.5rem;margin-bottom:.75rem; }
.success-modal p { color:var(--color-gray-1);margin-bottom:.5rem; }
.success-actions { display:flex;gap:.75rem;justify-content:center;margin-top:1.5rem; }
</style>
