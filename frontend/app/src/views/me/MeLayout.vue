<template>
  <div class="me-page">
    <!-- Topbar -->

    <div class="me-layout container">
      <!-- Sidebar -->
      <aside class="me-sidebar">
        <div class="me-user-card">
          <div class="me-avatar">
            {{ (auth.user?.display_name || auth.user?.username || '?')[0] }}
          </div>
          <div class="me-user-info">
            <div class="me-username">{{ auth.user?.display_name || auth.user?.username }}</div>
            <div class="me-email">{{ auth.user?.email }}</div>
            <div class="me-status" :class="`status-${auth.user?.status}`">
              {{ statusLabel(auth.user?.status) }}
            </div>
          </div>
        </div>

        <nav class="me-nav">
          <RouterLink v-for="item in navItems" :key="item.to"
            :to="item.to" class="me-nav-item"
            :class="{ active: $route.path === item.to }">
            <span class="me-nav-icon">{{ item.icon }}</span>
            {{ item.label }}
          </RouterLink>
        </nav>
      </aside>

      <!-- Content -->
      <main class="me-content">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup>
import { RouterLink, RouterView, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth   = useAuthStore()
const router = useRouter()

const navItems = [
  { to: '/me',           icon: '👤', label: '個人資料' },
  { to: '/me/orders',    icon: '📦', label: '我的訂單' },
  { to: '/me/training',  icon: '📔', label: '訓練日記' },
  { to: '/me/garmin',    icon: '⌚', label: 'Garmin 串接' },
  { to: '/me/strava',    icon: '🟠', label: 'Strava 串接' },
  { to: '/me/my-items',  icon: '🔄', label: '我的二手品' },
  { to: '/me/membership',icon: '🏅', label: '會員狀態' },
]

function statusLabel(s) {
  return { 0: '審核中', 1: '會員', 2: '已停用', 3: '已拒絕' }[s] ?? '未知'
}

async function handleLogout() {
  await auth.logout()
  router.push('/login')
}
</script>

<style scoped>
.me-page { background: var(--color-bg); min-height: 100vh; }
.nav-link { font-family: var(--font-cond); font-weight: 600; letter-spacing: 0.05em; text-transform: uppercase; font-size: 0.85rem; color: var(--color-gray-1); transition: color 0.2s; }
.nav-link:hover { color: var(--color-primary); }
.btn-sm { padding: 0.4rem 1rem; font-size: 0.82rem; }

.me-layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 2rem;
  padding-top: calc(64px + 2rem);
  padding-bottom: 4rem;
  align-items: start;
}

/* Sidebar */
.me-sidebar { position: sticky; top: calc(64px + 2rem); }
.me-user-card {
  background: var(--color-bg-card); border: 1px solid var(--color-border);
  border-radius: 8px; padding: 1.25rem;
  display: flex; flex-direction: column; align-items: center; gap: 0.75rem;
  text-align: center; margin-bottom: 1rem;
}
.me-avatar {
  width: 64px; height: 64px; border-radius: 50%;
  background: var(--color-primary);
  display: flex; align-items: center; justify-content: center;
  font-family: var(--font-display); font-size: 2rem; color: #fff;
}
.me-username { font-weight: 700; font-size: 1rem; }
.me-email { font-size: 0.78rem; color: var(--color-gray-2); margin-top: 0.15rem; }
.me-status {
  font-size: 0.72rem; font-weight: 700; letter-spacing: 0.08em; text-transform: uppercase;
  padding: 0.2rem 0.6rem; border-radius: 3px; margin-top: 0.25rem;
}
.status-0 { background: rgba(245,158,11,0.15); color: #f59e0b; }
.status-1 { background: rgba(34,197,94,0.15);  color: #22c55e; }
.status-2 { background: rgba(107,114,128,0.15); color: #9ca3af; }
.status-3 { background: rgba(239,68,68,0.15);   color: #ef4444; }

.me-nav { display: flex; flex-direction: column; gap: 0.25rem; }
.me-nav-item {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.65rem 1rem; border-radius: 6px;
  font-size: 0.9rem; color: var(--color-gray-1);
  transition: all 0.15s;
}
.me-nav-item:hover { background: var(--color-bg-hover); color: var(--color-white); }
.me-nav-item.active { background: rgba(229,25,26,0.1); color: var(--color-primary); font-weight: 600; }
.me-nav-icon { font-size: 1rem; width: 20px; text-align: center; }

/* Content */
.me-content { min-width: 0; }

@media (max-width: 768px) {
  .me-layout { grid-template-columns: 1fr; }
  .me-sidebar { position: static; }
}
</style>
