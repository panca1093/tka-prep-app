<script setup lang="ts">
import { ref } from 'vue'
import { uploadImage } from '@/api/upload'

defineProps<{
  modelValue: string | null | undefined
  label?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', val: string | null): void
}>()

const isUploading = ref(false)
const error = ref('')

const MAX_SIZE = 5 * 1024 * 1024 // 5 MB

async function onFileChange(evt: Event) {
  const input = evt.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  error.value = ''
  if (file.size > MAX_SIZE) {
    error.value = 'File too large (max 5 MB)'
    input.value = ''
    return
  }
  isUploading.value = true
  try {
    const url = await uploadImage(file)
    emit('update:modelValue', url)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Upload failed'
  } finally {
    isUploading.value = false
    input.value = ''
  }
}

function remove() {
  emit('update:modelValue', null)
}
</script>

<template>
  <div class="img-upload">
    <span v-if="label" class="img-upload-label">{{ label }}</span>
    <div v-if="modelValue" class="img-preview">
      <img :src="modelValue" alt="uploaded" class="preview-img" />
      <button class="remove-btn" type="button" @click="remove">✕</button>
    </div>
    <label v-else class="upload-zone" :class="{ uploading: isUploading }">
      <span v-if="isUploading">Uploading…</span>
      <span v-else>📷 {{ label ?? 'Add image' }}</span>
      <input type="file" accept="image/*" class="file-input" :disabled="isUploading" @change="onFileChange" />
    </label>
    <p v-if="error" class="upload-error">{{ error }}</p>
  </div>
</template>

<style scoped>
.img-upload { display: flex; flex-direction: column; gap: 0.375rem; }
.img-upload-label { font-size: 0.72rem; color: #94a3b8; font-weight: 600; }

.img-preview {
  position: relative;
  display: inline-block;
  max-width: 100%;
}
.preview-img {
  max-width: 100%;
  max-height: 200px;
  border-radius: 8px;
  border: 1px solid #1e2a45;
  display: block;
}
.remove-btn {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  background: rgba(0,0,0,0.65);
  border: none;
  color: #fff;
  font-size: 0.7rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.upload-zone {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.4rem 0.875rem;
  border-radius: 8px;
  border: 1px dashed #1e2a45;
  background: transparent;
  color: #64748b;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.15s;
}
.upload-zone:hover:not(.uploading) { border-color: #4f8ef7; color: #4f8ef7; }
.upload-zone.uploading { opacity: 0.6; cursor: wait; }

.file-input { display: none; }

.upload-error { font-size: 0.75rem; color: #ef4444; margin: 0; }
</style>
