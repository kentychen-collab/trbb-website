<template>
  <div class="training-map-wrap">
    <div ref="mapEl" class="map-container"></div>
    <div v-if="!hasPoints" class="map-no-data">📍 無 GPS 路線資料</div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'

const props = defineProps({
  routePoints: { type: Array, default: () => [] }, // [[lat,lng], ...]
  startLat:    { type: Number, default: null },
  startLng:    { type: Number, default: null },
  height:      { type: String, default: '280px' },
})

const mapEl   = ref(null)
const hasPoints = ref(false)
let mapInstance = null
let L = null

async function initMap() {
  if (!mapEl.value) return
  const points = props.routePoints || []
  hasPoints.value = points.length > 0
  if (!hasPoints.value) return

  // Dynamically load Leaflet from CDN
  if (!window.L) {
    await loadLeaflet()
  }
  L = window.L

  // Destroy previous instance
  if (mapInstance) { mapInstance.remove(); mapInstance = null }

  mapInstance = L.map(mapEl.value, { zoomControl: true, attributionControl: false })
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© OpenStreetMap',
  }).addTo(mapInstance)

  // Draw route polyline
  const latlngs = points.map(p => [p[0], p[1]])
  const polyline = L.polyline(latlngs, {
    color: '#e5191a', weight: 3, opacity: 0.85,
  }).addTo(mapInstance)

  // Start marker
  if (latlngs.length) {
    const startIcon = L.divIcon({
      html: '<div style="width:12px;height:12px;background:#22c55e;border:2px solid #fff;border-radius:50%;"></div>',
      className: '', iconAnchor: [6, 6],
    })
    const endIcon = L.divIcon({
      html: '<div style="width:12px;height:12px;background:#e5191a;border:2px solid #fff;border-radius:50%;"></div>',
      className: '', iconAnchor: [6, 6],
    })
    L.marker(latlngs[0], { icon: startIcon }).addTo(mapInstance)
    L.marker(latlngs[latlngs.length - 1], { icon: endIcon }).addTo(mapInstance)
  }

  mapInstance.fitBounds(polyline.getBounds(), { padding: [20, 20] })
}

function loadLeaflet() {
  return new Promise((resolve) => {
    if (window.L) { resolve(); return }
    // CSS
    const link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css'
    document.head.appendChild(link)
    // JS
    const script = document.createElement('script')
    script.src = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.js'
    script.onload = resolve
    document.head.appendChild(script)
  })
}

watch(() => props.routePoints, () => { initMap() }, { deep: true })
onMounted(initMap)
onBeforeUnmount(() => { if (mapInstance) { mapInstance.remove(); mapInstance = null } })
</script>

<style scoped>
.training-map-wrap { position:relative; width:100%; height:v-bind(height); }
.map-container { width:100%; height:100%; border-radius:8px; overflow:hidden; }
.map-no-data { position:absolute; inset:0; display:flex; align-items:center; justify-content:center; color:var(--color-gray-2); background:var(--color-bg-card); border-radius:8px; font-size:.85rem; }
</style>
