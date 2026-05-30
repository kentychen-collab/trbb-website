<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">儀表板</h1>
      <p class="page-subtitle">歡迎回來，{{ store.admin?.display_name }}。</p>
    </div>

    <div v-if="loading" class="loading-row">載入中...</div>
    <template v-else>
      <!-- 統計卡片 -->
      <div class="grid-4 mb-2">
        <div class="card stat-card">
          <div class="stat-label">總會員數</div>
          <div class="stat-value">{{ data.stats.total_members.toLocaleString() }}</div>
          <div class="stat-change">本月新增 {{ data.stats.new_members_month }} 人</div>
        </div>
        <div class="card stat-card">
          <div class="stat-label">本月報名</div>
          <div class="stat-value" style="color:var(--primary)">{{ data.stats.month_registrations }}</div>
          <div class="stat-change">共 {{ data.stats.total_events }} 個賽事</div>
        </div>
        <div class="card stat-card">
          <div class="stat-label">本月營收</div>
          <div class="stat-value" style="color:var(--success)">
            NT$ {{ data.stats.month_revenue.toLocaleString() }}
          </div>
          <div class="stat-change" :class="{ neg: data.stats.pending_orders > 0 }">
            待處理訂單 {{ data.stats.pending_orders }} 筆
          </div>
        </div>
        <div class="card stat-card">
          <div class="stat-label">訓練日記</div>
          <div class="stat-value">{{ data.stats.total_training_logs.toLocaleString() }}</div>
          <div class="stat-change">公開 {{ data.stats.public_logs }} 筆</div>
        </div>
      </div>

      <div class="grid-2 mt-4">
        <!-- 最新報名 -->
        <div class="card">
          <div class="card-header">
            <span style="font-weight:600">最新報名</span>
            <RouterLink to="/events" class="btn btn-ghost btn-sm">查看全部</RouterLink>
          </div>
          <div class="card-body" style="padding:0">
            <div v-if="!data.recent_regs.length" class="empty-hint">目前無報名記錄</div>
            <table v-else class="table">
              <thead>
                <tr><th>姓名</th><th>賽事</th><th>狀態</th><th>時間</th></tr>
              </thead>
              <tbody>
                <tr v-for="r in data.recent_regs" :key="r.id">
                  <td>{{ r.name }}</td>
                  <td class="text-ellipsis">{{ r.event_name }}</td>
                  <td>
                    <span class="badge" :class="regBadge(r.status)">{{ regLabel(r.status) }}</span>
                  </td>
                  <td class="text-gray" style="font-size:.78rem">{{ r.created_at }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- 最新訂單 -->
        <div class="card">
          <div class="card-header">
            <span style="font-weight:600">最新訂單</span>
            <RouterLink to="/orders" class="btn btn-ghost btn-sm">查看全部</RouterLink>
          </div>
          <div class="card-body" style="padding:0">
            <div v-if="!data.recent_orders.length" class="empty-hint">目前無訂單</div>
            <table v-else class="table">
              <thead>
                <tr><th>會員</th><th>商品</th><th>金額</th><th>狀態</th></tr>
              </thead>
              <tbody>
                <tr v-for="o in data.recent_orders" :key="o.id">
                  <td>{{ o.member_name }}</td>
                  <td class="text-ellipsis">{{ o.product_name }}</td>
                  <td>NT$ {{ Number(o.amount).toLocaleString() }}</td>
                  <td>
                    <span class="badge" :class="orderBadge(o.status)">{{ orderLabel(o.status) }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAdminStore } from '@/stores/admin'
import api from '@/services/api'

const store = useAdminStore()
const loading = ref(true)
const data = ref({
  stats: {
    total_members: 0, new_members_month: 0,
    total_events: 0, month_registrations: 0,
    pending_orders: 0, month_revenue: 0,
    total_training_logs: 0, public_logs: 0,
  },
  recent_regs: [],
  recent_orders: [],
})

function regBadge(s) {
  return { confirmed:'badge-success', paid:'badge-success', pending:'badge-warning', cancelled:'badge-danger' }[s] || 'badge-gray'
}
function regLabel(s) {
  return { confirmed:'已確認', paid:'已付款', pending:'待付款', cancelled:'已取消' }[s] || s
}
function orderBadge(s) {
  return { completed:'badge-success', paid:'badge-success', pending:'badge-warning', cancelled:'badge-danger' }[s] || 'badge-gray'
}
function orderLabel(s) {
  return { completed:'已完成', paid:'已付款', pending:'待處理', cancelled:'已取消', shipped:'已出貨' }[s] || s
}

onMounted(async () => {
  try {
    const { data: res } = await api.get('/dashboard')
    data.value = res
  } catch(e) {
    console.error('dashboard load error', e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.loading-row { padding:4rem; text-align:center; color:var(--gray-2); }
.empty-hint { padding:2rem; text-align:center; color:var(--gray-2); font-size:.85rem; }
.text-ellipsis { max-width:120px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
</style>
