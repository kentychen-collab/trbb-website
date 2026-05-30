<template>
  <RouterLink to="/" class="site-logo-wrap">
    <img v-if="logoUrl" :src="logoUrl" :alt="logoText" class="site-logo-img" />
    <span v-else class="trbb-logo-text">
      <span class="tr">TR</span><span class="bb">BB</span>
    </span>
  </RouterLink>
</template>

<script setup>
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useSiteSettingsStore } from '@/stores/siteSettings'

const site = useSiteSettingsStore()
const logoUrl  = computed(() => site.get('logo_image', ''))
const logoText = computed(() => site.get('logo_text', 'TRBB'))
</script>

<style scoped>
.site-logo-wrap {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  text-decoration: none;
  height: 40px;
}
.site-logo-img {
  height: 36px;
  max-height: 36px;
  width: auto;
  max-width: 160px;
  object-fit: contain;
  display: block;
}
.trbb-logo-text {
  font-family: var(--font-display, 'Bebas Neue', sans-serif);
  letter-spacing: .05em;
  font-size: 2rem;
  line-height: 1;
}
.tr { color: var(--color-navy, #1A3A7A); }
.bb { color: var(--color-primary, #CF2027); }
</style>
