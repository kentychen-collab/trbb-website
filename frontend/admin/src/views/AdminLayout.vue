<template>
  <div class="admin-layout">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-logo">
        <span class="tr">TR</span><span class="bb">BB</span>
        <small>管理後台 v1.0</small>
      </div>
      <nav class="sidebar-nav">
        <div class="nav-section">總覽</div>
        <RouterLink to="/" class="nav-item" :class="{ active: route.path === '/' }">
          <span class="nav-icon">📊</span> 儀表板
        </RouterLink>

        <div class="nav-section">會員管理</div>
        <RouterLink to="/members" class="nav-item" :class="{ active: route.path.startsWith('/members') }">
          <span class="nav-icon">👤</span> 一般會員
        </RouterLink>
        <RouterLink to="/admins" class="nav-item" :class="{ active: route.path.startsWith('/admins') }" v-if="store.admin?.role >= 9">
          <span class="nav-icon">🛡</span> 管理員列表
        </RouterLink>

        <div class="nav-section">賽事管理</div>
        <RouterLink to="/events" class="nav-item" :class="{ active: route.path.startsWith('/events') }">
          <span class="nav-icon">🏅</span> 賽事列表
        </RouterLink>

        <div class="nav-section">商城管理</div>
        <RouterLink to="/products" class="nav-item" :class="{ active: route.path.startsWith('/products') }">
          <span class="nav-icon">🛒</span> 商品管理
        </RouterLink>
        <RouterLink to="/orders" class="nav-item" :class="{ active: route.path.startsWith('/orders') }">
          <span class="nav-icon">📦</span> 訂單管理
        </RouterLink>
        <RouterLink to="/training" class="nav-item" :class="{ active: route.path.startsWith('/training') }">
          <span class="nav-icon">📔</span> 訓練日記
        </RouterLink>

        <div class="nav-section">內容</div>
        <RouterLink to="/announcements" class="nav-item" :class="{ active: route.path.startsWith('/announcements') }">
          <span class="nav-icon">📢</span> 公告管理
        </RouterLink>
      </nav>
      <div style="padding:1rem 1.5rem;border-top:1px solid var(--border)">
        <button class="nav-item" style="width:100%;color:var(--danger)" @click="handleLogout">
          <span class="nav-icon">🚪</span> 登出
        </button>
      </div>
    </aside>

    <!-- Main -->
    <div class="main-content">
      <header class="topbar">
        <div style="font-weight:600">{{ pageTitle }}</div>
        <div class="flex items-center gap-2">
          <span class="text-gray" style="font-size:0.85rem">{{ admin.admin?.display_name || 'Admin' }}</span>
          <div style="width:32px;height:32px;border-radius:50%;background:var(--primary);display:flex;align-items:center;justify-content:center;font-weight:700">
            {{ (admin.admin?.display_name || 'A')[0] }}
          </div>
        </div>
      </header>
      <div class="page-content">
        <RouterView />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { RouterView, RouterLink, useRoute, useRouter } from 'vue-router'
import { useAdminStore } from '@/stores/admin'

const route = useRoute()
const router = useRouter()
const admin = useAdminStore()
const store = admin  // alias for template v-if checks

const pageTitles = {
  '/': '儀表板',
  '/members': '一般會員',
  '/admins': '管理員列表',
  '/events': '賽事管理',
  '/products': '商品管理',
  '/orders': '訂單管理',
  '/training': '訓練日記管理',
  '/announcements': '公告管理',
}
const pageTitle = computed(() => pageTitles[route.path] || 'TRBB Admin')

function handleLogout() {
  admin.logout()
  router.push('/login')
}
</script>
