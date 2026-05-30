<template>
  <div id="trbb-app">
    <!-- 全域 Navbar：在所有 view 上方，不受各 view CSS 影響 -->
    <SiteNavbar v-if="showNavbar" />
    <RouterView />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, RouterView } from 'vue-router'
import SiteNavbar from '@/components/SiteNavbar.vue'

const route = useRoute()

// 登入、註冊頁不顯示 navbar（它們有自己的樣式）
const noNavbarRoutes = ['Login', 'Register', 'ForgotPassword']
const showNavbar = computed(() => !noNavbarRoutes.includes(route.name))
</script>

<style>
/* 所有頁面頂部預留 navbar 高度 */
#trbb-app > :not(.site-navbar) {
  /* 由各頁面自己控制 padding-top，這裡不強制 */
}
</style>
