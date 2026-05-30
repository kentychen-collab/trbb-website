<template>
  <div class="gradient-editor">
    <!-- Type selector -->
    <div class="type-row">
      <button v-for="t in types" :key="t.value"
        class="type-btn" :class="{ active: parsed.type === t.value }"
        @click="setType(t.value)" type="button">
        {{ t.label }}
      </button>
    </div>

    <!-- Solid -->
    <div v-if="parsed.type === 'solid'" class="color-list">
      <div class="color-input-row">
        <input type="color" v-model="parsed.colors[0]" class="color-swatch" @input="emit()" />
        <input v-model="parsed.colors[0]" class="color-text" @input="emit()" placeholder="#FFFFF3" />
      </div>
    </div>

    <!-- Linear / Radial gradient -->
    <div v-else>
      <div v-if="parsed.type === 'linear'" class="angle-row">
        <label>角度</label>
        <input v-model="parsed.angle" type="range" min="0" max="360" @input="emit()" class="angle-slider" />
        <span class="angle-val">{{ parsed.angle }}°</span>
      </div>

      <div class="color-stops">
        <div v-for="(color, i) in parsed.colors" :key="i" class="color-stop-row">
          <span class="stop-num">{{ i+1 }}</span>
          <input type="color" :value="color" @input="e => updateColor(i, e.target.value)" class="color-swatch" />
          <input :value="color" @input="e => updateColor(i, e.target.value)" class="color-text" />
          <input :value="parsed.stops?.[i] ?? Math.round(i*100/(parsed.colors.length-1))"
            type="number" min="0" max="100" @input="e => updateStop(i, e.target.value)"
            class="stop-pct" placeholder="%" />
          <span class="stop-pct-sym">%</span>
          <button @click="removeColor(i)" type="button" class="remove-btn"
            :disabled="parsed.colors.length <= 2">✕</button>
        </div>
      </div>

      <button class="btn btn-ghost btn-sm add-stop-btn" @click="addColor" type="button">
        ＋ 新增色點
      </button>
    </div>

    <!-- Preview -->
    <div class="preview-bar" :style="{ background: previewCSS }"></div>
    <div class="preview-label text-gray">預覽</div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({ modelValue: { type: String, default: '' } })
const emits = defineEmits(['update:modelValue'])

const types = [
  { value: 'solid',  label: '純色' },
  { value: 'linear', label: '線性漸變' },
  { value: 'radial', label: '放射漸變' },
]

// Parse incoming JSON or plain color string
function parse(val) {
  if (!val) return { type: 'solid', colors: ['#FFFFF3'], stops: [0], angle: '135' }
  try {
    const g = JSON.parse(val)
    return {
      type:   g.type   || 'solid',
      colors: g.colors || ['#FFFFF3'],
      stops:  g.stops  || null,
      angle:  g.angle  || '135',
    }
  } catch {
    return { type: 'solid', colors: [val], stops: [0], angle: '135' }
  }
}

const parsed = ref(parse(props.modelValue))

watch(() => props.modelValue, v => {
  const p = parse(v)
  if (JSON.stringify(p) !== JSON.stringify(parsed.value)) {
    parsed.value = p
  }
})

const previewCSS = computed(() => {
  const p = parsed.value
  if (p.type === 'solid') return p.colors[0]
  const stops = p.colors.map((c, i) => {
    const pct = p.stops?.[i] ?? Math.round(i * 100 / (p.colors.length - 1))
    return `${c} ${pct}%`
  }).join(', ')
  if (p.type === 'linear') return `linear-gradient(${p.angle}deg, ${stops})`
  return `radial-gradient(circle, ${stops})`
})

function emit() {
  emits('update:modelValue', JSON.stringify(parsed.value))
}

function setType(t) {
  parsed.value.type = t
  if (parsed.value.colors.length < 2 && t !== 'solid') {
    parsed.value.colors.push('#566270')
  }
  emit()
}

function updateColor(i, val) {
  parsed.value.colors[i] = val
  emit()
}

function updateStop(i, val) {
  if (!parsed.value.stops) {
    parsed.value.stops = parsed.value.colors.map((_, idx) =>
      Math.round(idx * 100 / (parsed.value.colors.length - 1))
    )
  }
  parsed.value.stops[i] = parseInt(val) || 0
  emit()
}

function addColor() {
  parsed.value.colors.push('#A593E0')
  if (parsed.value.stops) {
    parsed.value.stops.push(100)
  }
  emit()
}

function removeColor(i) {
  if (parsed.value.colors.length <= 2) return
  parsed.value.colors.splice(i, 1)
  if (parsed.value.stops) parsed.value.stops.splice(i, 1)
  emit()
}
</script>

<style scoped>
.gradient-editor { display:flex; flex-direction:column; gap:.6rem; }
.type-row { display:flex; gap:.4rem; }
.type-btn { padding:.3rem .85rem; border-radius:4px; border:1px solid var(--border); font-size:.78rem; font-weight:600; cursor:pointer; background:var(--bg); color:var(--gray-2); transition:all .15s; }
.type-btn.active { background:var(--navy); border-color:var(--navy); color:#fff; }
.color-list { display:flex; flex-direction:column; gap:.4rem; }
.color-input-row { display:flex; align-items:center; gap:.5rem; }
.color-swatch { width:40px; height:34px; padding:2px 4px; cursor:pointer; flex-shrink:0; border-radius:4px; border:1px solid var(--border); }
.color-text { flex:1; }
.angle-row { display:flex; align-items:center; gap:.6rem; font-size:.78rem; color:var(--gray-2); }
.angle-slider { flex:1; }
.angle-val { min-width:36px; font-weight:600; color:var(--text); }
.color-stops { display:flex; flex-direction:column; gap:.4rem; }
.color-stop-row { display:flex; align-items:center; gap:.4rem; }
.stop-num { width:16px; font-size:.72rem; color:var(--gray-2); flex-shrink:0; }
.stop-pct { width:48px; }
.stop-pct-sym { font-size:.75rem; color:var(--gray-2); }
.remove-btn { width:24px; height:24px; border-radius:4px; border:1px solid var(--border); font-size:.75rem; cursor:pointer; display:flex; align-items:center; justify-content:center; background:none; color:var(--gray-2); flex-shrink:0; }
.remove-btn:hover:not(:disabled) { background:rgba(207,32,39,.1); border-color:var(--primary); color:var(--primary); }
.remove-btn:disabled { opacity:.3; cursor:not-allowed; }
.add-stop-btn { margin-top:.2rem; }
.preview-bar { height:32px; border-radius:6px; border:1px solid var(--border); }
.preview-label { font-size:.68rem; }
</style>
