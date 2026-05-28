<template>
  <div class="image-upload">
    <!-- 預覽區 -->
    <div class="preview-box" :class="{ 'has-image': previewUrl }" @click="triggerPick">
      <img v-if="previewUrl" :src="previewUrl" alt="preview" class="preview-img" />
      <div v-else class="preview-placeholder">
        <span class="upload-icon">🖼</span>
        <span>點擊選擇圖片</span>
        <span class="upload-hint">JPG、PNG、WebP，最大 10MB</span>
      </div>
      <!-- 覆蓋層 -->
      <div v-if="previewUrl" class="preview-overlay">
        <span>更換圖片</span>
      </div>
    </div>

    <!-- 隱藏 file input -->
    <input ref="fileInput" type="file"
      accept=".jpg,.jpeg,.png,.gif,.webp"
      style="display:none"
      @change="onFileSelected" />

    <!-- 操作按鈕 -->
    <div class="upload-actions">
      <button type="button" class="btn btn-ghost btn-sm" @click="triggerPick" :disabled="uploading">
        {{ previewUrl ? '更換' : '選擇圖片' }}
      </button>
      <button type="button" class="btn btn-primary btn-sm" @click="doUpload"
        :disabled="!selectedFile || uploading">
        <span v-if="uploading" class="spinner-sm"></span>
        {{ uploading ? '上傳中...' : '上傳' }}
      </button>
      <button type="button" class="btn btn-ghost btn-sm" @click="clearImage"
        v-if="previewUrl" :disabled="uploading">
        清除
      </button>
    </div>

    <!-- 進度 / 錯誤 -->
    <div v-if="uploadError" class="upload-error">{{ uploadError }}</div>
    <div v-if="uploadDone" class="upload-done">✓ 上傳成功</div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import api from '@/services/api'

const props = defineProps({
  modelValue: { type: String, default: '' },  // 當前的 objectPath
  folder: { type: String, default: 'general' },
})
const emit = defineEmits(['update:modelValue'])

const fileInput  = ref(null)
const selectedFile = ref(null)
const previewUrl = ref('')
const uploading  = ref(false)
const uploadError = ref('')
const uploadDone  = ref(false)

// 如果 modelValue 已有值（編輯模式），顯示現有圖片
watch(() => props.modelValue, (val) => {
  if (val && !selectedFile.value) {
    // 組合完整 URL 顯示預覽
    const base = import.meta.env.VITE_IMAGE_BASE_URL || ''
    previewUrl.value = val.startsWith('http') ? val : `${base}/${val}`
  }
}, { immediate: true })

function triggerPick() {
  fileInput.value?.click()
}

function onFileSelected(e) {
  const file = e.target.files?.[0]
  if (!file) return
  selectedFile.value = file
  uploadError.value = ''
  uploadDone.value  = false

  // 本地預覽
  const reader = new FileReader()
  reader.onload = (ev) => { previewUrl.value = ev.target.result }
  reader.readAsDataURL(file)
}

async function doUpload() {
  if (!selectedFile.value) return
  uploading.value  = true
  uploadError.value = ''
  uploadDone.value  = false

  const formData = new FormData()
  formData.append('file', selectedFile.value)

  try {
    const { data } = await api.post(
      `/upload/image?folder=${props.folder}`,
      formData,
      { headers: { 'Content-Type': 'multipart/form-data' } }
    )
    // 通知父元件：存 objectPath (e.g. "events/123_cover.jpg")
    emit('update:modelValue', data.path)
    // 顯示真實 CDN URL
    previewUrl.value = data.url
    uploadDone.value = true
    selectedFile.value = null
    setTimeout(() => { uploadDone.value = false }, 3000)
  } catch(e) {
    uploadError.value = e.response?.data?.error || '上傳失敗，請稍後再試'
  } finally {
    uploading.value = false
  }
}

function clearImage() {
  previewUrl.value   = ''
  selectedFile.value = null
  uploadError.value  = ''
  if (fileInput.value) fileInput.value.value = ''
  emit('update:modelValue', '')
}
</script>

<style scoped>
.image-upload { display:flex; flex-direction:column; gap:.6rem; }

.preview-box {
  width:100%; height:180px; border-radius:6px;
  border:2px dashed var(--border);
  cursor:pointer; overflow:hidden; position:relative;
  display:flex; align-items:center; justify-content:center;
  background:var(--bg); transition:border-color .2s;
}
.preview-box:hover { border-color:var(--primary); }
.preview-box.has-image { border-style:solid; }

.preview-img { width:100%; height:100%; object-fit:cover; display:block; }

.preview-placeholder {
  display:flex; flex-direction:column; align-items:center; gap:.35rem;
  color:var(--gray-2); font-size:.82rem; pointer-events:none;
}
.upload-icon { font-size:2.5rem; }
.upload-hint { font-size:.72rem; color:var(--gray-2); }

.preview-overlay {
  position:absolute; inset:0;
  background:rgba(0,0,0,.5);
  display:flex; align-items:center; justify-content:center;
  color:#fff; font-size:.85rem; font-weight:600;
  opacity:0; transition:opacity .2s;
}
.preview-box:hover .preview-overlay { opacity:1; }

.upload-actions { display:flex; gap:.5rem; }

.upload-error { font-size:.78rem; color:#fca5a5; background:rgba(239,68,68,.1); border:1px solid rgba(239,68,68,.3); border-radius:4px; padding:.4rem .75rem; }
.upload-done  { font-size:.78rem; color:#86efac; background:rgba(34,197,94,.1); border:1px solid rgba(34,197,94,.3); border-radius:4px; padding:.4rem .75rem; }

.spinner-sm { width:12px; height:12px; border:2px solid rgba(255,255,255,.3); border-top-color:#fff; border-radius:50%; animation:spin .7s linear infinite; display:inline-block; margin-right:.25rem; }
@keyframes spin { to { transform:rotate(360deg) } }
</style>
