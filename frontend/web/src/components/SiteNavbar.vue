<template>
  <nav class="site-navbar">
    <div class="container nav-inner">
      <SiteLogo />
      <div class="nav-right">
        <RouterLink to="/" class="nav-link">首頁</RouterLink>
        <RouterLink to="/events" class="nav-link">賽事</RouterLink>
        <RouterLink to="/shop" class="nav-link">商城</RouterLink>
        <RouterLink to="/training" class="nav-link">訓練</RouterLink>
        <template v-if="auth.isLoggedIn">
          <RouterLink to="/me" class="nav-link">會員中心</RouterLink>
          <button class="btn btn-ghost btn-sm" @click="auth.logout(); $router.push('/login')">登出</button>
        </template>
        <template v-else>
          <RouterLink to="/login" class="btn btn-primary btn-sm">登入</RouterLink>
        </template>
      </div>
      <!-- Mobile toggle -->
      <button class="nav-mobile-btn" @click="open = !open">☰</button>
    </div>
    <!-- Mobile menu -->
    <div v-if="open" class="nav-mobile-menu">
      <RouterLink to="/"         class="mob-link" @click="open=false">首頁</RouterLink>
      <RouterLink to="/events"   class="mob-link" @click="open=false">賽事</RouterLink>
      <RouterLink to="/shop"     class="mob-link" @click="open=false">商城</RouterLink>
      <RouterLink to="/training" class="mob-link" @click="open=false">訓練</RouterLink>
      <RouterLink v-if="auth.isLoggedIn"  to="/me"    class="mob-link" @click="open=false">會員中心</RouterLink>
      <RouterLink v-if="!auth.isLoggedIn" to="/login" class="mob-link" @click="open=false">登入</RouterLink>
    </div>
  </nav>
</template>

<script setup>
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import SiteLogo from './SiteLogo.vue'

const auth = useAuthStore()
const open = ref(false)
</script>

<style scoped>
.site-navbar {
  position: fixed; top: 0; left: 0; right: 0; z-index: 100;
  background: var(--navbar-bg, rgba(255,255,243,0.95));
  backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--color-border, #E0E3DA);
  box-shadow: 0 1px 6px rgba(86,98,112,.08);
}
.nav-inner {
  display: flex; align-items: center;
  justify-content: space-between; height: 64px;
}
.nav-right {
  display: flex; align-items: center; gap: 1.5rem;
}
.nav-link {
  font-family: var(--font-cond, 'Barlow Condensed', sans-serif);
  font-weight: 600; letter-spacing: .06em; text-transform: uppercase;
  font-size: .88rem;
  color: var(--navbar-text, var(--color-navy, #1A3A7A));
  transition: color .15s;
}
.nav-link:hover { color: var(--color-primary, #CF2027); }
.btn-sm { padding: .4rem 1.1rem; font-size: .82rem; }
.nav-mobile-btn {
  display: none; font-size: 1.4rem;
  color: var(--color-navy, #1A3A7A);
}
.nav-mobile-menu {
  display: flex; flex-direction: column;
  background: var(--color-bg-card, #fff);
  border-top: 1px solid var(--color-border, #E0E3DA);
  padding: .5rem 0;
}
.mob-link {
  padding: .75rem 1.5rem;
  font-family: var(--font-cond, 'Barlow Condensed', sans-serif);
  font-weight: 600; text-transform: uppercase; font-size: .9rem;
  color: var(--color-navy, #1A3A7A);
}
.mob-link:hover { background: var(--color-bg-hover, #EEF0E8); }
@media (max-width: 768px) {
  .nav-right { display: none; }
  .nav-mobile-btn { display: block; }
}
</style>
