<template>
  <div class="shop-page">

    <div class="container" style="padding-top:calc(64px + 2rem);padding-bottom:4rem">
      <div class="page-hero">
        <div class="page-hero-tag">TRBB SHOP</div>
        <h1 class="page-hero-title">裝備商城</h1>
        <p class="page-hero-desc">官方認證裝備、訓練補給，會員享專屬優惠。</p>
      </div>

      <!-- Filter -->
      <div class="shop-filter">
        <button v-for="cat in categories" :key="cat.value"
          class="cat-btn" :class="{ active: selectedCat === cat.value }"
          @click="selectCat(cat.value)">
          {{ cat.label }}
        </button>
        <input v-model="keyword" placeholder="搜尋商品..." @keyup.enter="fetchProducts" class="search-input" />
      </div>

      <div v-if="loading" class="loading-state"><div class="loading-spinner"></div></div>
      <div v-else-if="!products.length" class="empty-state"><p>目前沒有商品</p></div>
      <div v-else class="products-grid">
        <RouterLink v-for="p in products" :key="p.id" :to="`/shop/${p.id}`" class="product-card card">
          <div class="product-cover">
            <img v-if="p.images && p.images.length" :src="imgUrl(p.images[0])" :alt="p.title" />
            <div v-else class="product-cover-placeholder">🛍</div>
            <div v-if="p.stock === 0" class="sold-out-badge">已售完</div>
          </div>
          <div class="product-body">
            <div class="product-cat">{{ catLabel(p.category) }}</div>
            <h3 class="product-title">{{ p.title }}</h3>
            <div class="product-price">NT$ {{ Number(p.price).toLocaleString() }}</div>
            <div class="product-stock text-gray text-sm">庫存：{{ p.stock }}</div>
          </div>
        </RouterLink>
      </div>

      <div class="pagination" v-if="totalPages > 1">
        <button :disabled="page===1" @click="goPage(page-1)" class="btn btn-ghost">‹</button>
        <span class="text-gray">{{ page }} / {{ totalPages }}</span>
        <button :disabled="page===totalPages" @click="goPage(page+1)" class="btn btn-ghost">›</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/services/api'

const auth = useAuthStore()
const products = ref([])
const loading = ref(false)
const page = ref(1)
const totalPages = ref(1)
const keyword = ref('')
const selectedCat = ref(0)

const categories = [
  { value: 0, label: '全部' },
  { value: 1, label: '服裝' },
  { value: 2, label: '裝備' },
  { value: 3, label: '營養補給' },
  { value: 4, label: '配件' },
]

const IMAGE_BASE = import.meta.env.VITE_IMAGE_BASE_URL || ''
function imgUrl(path) {
  if (!path) return ''
  if (path.startsWith('http')) return path
  return `${IMAGE_BASE}/images/${path}`
}

async function fetchProducts() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: 12 }
    if (keyword.value) params.keyword = keyword.value
    if (selectedCat.value) params.category = selectedCat.value
    const { data } = await api.get('/products', { params })
    products.value = data.products || []
    totalPages.value = data.pages || 1
  } catch(e) { console.error(e) }
  finally { loading.value = false }
}

function selectCat(val) { selectedCat.value = val; page.value = 1; fetchProducts() }
function goPage(p) { page.value = p; fetchProducts() }
function catLabel(c) { return { 1:'服裝', 2:'裝備', 3:'補給', 4:'配件' }[c] || '其他' }

onMounted(fetchProducts)
</script>

<style scoped>
.nav-link:hover { color:var(--color-primary); }

.page-hero { padding:3rem 0 2rem;text-align:center; }
.page-hero-tag { font-family:var(--font-cond);font-size:.75rem;font-weight:700;letter-spacing:.3em;color:var(--color-primary);margin-bottom:.75rem; }
.page-hero-title { font-family:var(--font-display);font-size:clamp(3rem,8vw,5rem);margin-bottom:.75rem; }
.page-hero-desc { color:var(--color-gray-1);max-width:400px;margin:0 auto; }

.shop-filter { display:flex;gap:.75rem;align-items:center;flex-wrap:wrap;margin-bottom:2rem; }
.cat-btn { padding:.4rem 1.2rem;border-radius:4px;border:1px solid var(--color-border);font-family:var(--font-cond);font-size:.82rem;font-weight:600;letter-spacing:.05em;cursor:pointer;background:none;color:var(--color-gray-2);transition:all .15s; }
.cat-btn:hover,.cat-btn.active { border-color:var(--color-primary);color:var(--color-primary); }
.search-input { height:36px;max-width:240px;font-size:.85rem; }

.loading-state { display:flex;align-items:center;justify-content:center;padding:4rem; }
.loading-spinner { width:24px;height:24px;border:2px solid var(--color-border);border-top-color:var(--color-primary);border-radius:50%;animation:spin .7s linear infinite; }
@keyframes spin { to { transform:rotate(360deg) } }
.empty-state { text-align:center;padding:4rem;color:var(--color-gray-2); }

.products-grid { display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:1.5rem;margin-bottom:2rem; }
.product-card { display:block;background:var(--color-bg-card);border:1px solid var(--color-border);border-radius:8px;overflow:hidden;transition:all .2s; }
.product-card:hover { border-color:var(--color-primary);transform:translateY(-3px); }
.product-cover { height:200px;overflow:hidden;background:var(--color-bg-2);position:relative; }
.product-cover img { width:100%;height:100%;object-fit:cover; }
.product-cover-placeholder { width:100%;height:100%;display:flex;align-items:center;justify-content:center;font-size:4rem; }
.sold-out-badge { position:absolute;top:.75rem;right:.75rem;background:rgba(107,114,128,.85);color:#fff;font-size:.72rem;font-weight:700;padding:.2rem .65rem;border-radius:3px; }
.product-body { padding:1.25rem; }
.product-cat { font-family:var(--font-cond);font-size:.7rem;letter-spacing:.12em;text-transform:uppercase;color:var(--color-primary);margin-bottom:.25rem; }
.product-title { font-size:1rem;font-weight:700;margin-bottom:.5rem;line-height:1.4; }
.product-price { font-family:var(--font-cond);font-size:1.3rem;font-weight:700;color:var(--color-primary); }
.product-stock { font-size:.78rem;margin-top:.25rem; }
.pagination { display:flex;align-items:center;justify-content:center;gap:1rem; }
</style>
