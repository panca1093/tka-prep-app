<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Topic = components['schemas']['TopicResponse']

defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const topics = ref<Topic[]>([])
const isLoading = ref(false)

// Inline create
const showCreate = ref(false)
const newName = ref('')
const newSaving = ref(false)
const newError = ref('')

async function fetchTopics() {
  isLoading.value = true
  const { data } = await client.GET('/topics')
  if (data) topics.value = data.data
  isLoading.value = false
}

async function createTopic() {
  const name = newName.value.trim()
  if (!name) return
  newSaving.value = true
  newError.value = ''
  const { data, error } = await client.POST('/topics', { body: { name } })
  if (error) {
    newError.value = 'Nama sudah ada atau tidak valid.'
    newSaving.value = false
    return
  }
  newSaving.value = false
  showCreate.value = false
  newName.value = ''
  await fetchTopics()
  // Auto-select the newly created topic
  if (data) {
    emit('update:modelValue', data.id)
  }
}

function cancelCreate() {
  showCreate.value = false
  newName.value = ''
  newError.value = ''
}

onMounted(fetchTopics)

defineExpose({ refresh: fetchTopics })
</script>

<template>
  <div class="ts-wrap">
    <select
      class="ts-select"
      :value="modelValue"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option value="">Pilih topik…</option>
      <option v-for="t in topics" :key="t.id" :value="t.id">{{ t.name }}</option>
    </select>

    <!-- Inline create toggle -->
    <button
      v-if="!showCreate"
      type="button"
      class="ts-add-btn"
      @click="showCreate = true"
      title="Buat Topik Baru"
    >+ Baru</button>

    <!-- Inline create form -->
    <div v-if="showCreate" class="ts-create">
      <input
        v-model="newName"
        type="text"
        class="ts-create-input"
        placeholder="Nama topik baru…"
        maxlength="100"
        @keyup.enter="createTopic"
        @keyup.escape="cancelCreate"
      />
      <p v-if="newError" class="ts-err">{{ newError }}</p>
      <div class="ts-create-actions">
        <button type="button" class="ts-create-cancel" @click="cancelCreate">Batal</button>
        <button type="button" class="ts-create-save" :disabled="newSaving || !newName.trim()" @click="createTopic">
          {{ newSaving ? '…' : 'Simpan' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ts-wrap { display: flex; flex-direction: column; gap: 0.4rem; }

.ts-select {
  width: 100%; padding: 0.55rem 0.75rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-input); color: var(--text-primary);
  font-size: 0.875rem; cursor: pointer; outline: none; transition: border-color 0.15s;
  font-family: inherit;
}
.ts-select:focus { border-color: var(--accent); }

.ts-add-btn {
  align-self: flex-start;
  padding: 0.25rem 0.6rem; border-radius: 5px;
  border: 1px dashed var(--border); background: transparent;
  color: var(--accent); font-size: 0.7rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s; font-family: inherit;
}
.ts-add-btn:hover { border-color: var(--accent); background: var(--accent-dim); }

.ts-create {
  display: flex; flex-direction: column; gap: 0.3rem;
  padding: 0.5rem 0.75rem; border-radius: 8px;
  border: 1px solid var(--accent); background: color-mix(in srgb, var(--accent) 6%, transparent);
}
.ts-create-input {
  padding: 0.35rem 0.5rem; border-radius: 5px;
  border: 1px solid var(--border); background: var(--bg-input);
  color: var(--text-primary); font-size: 0.75rem;
  outline: none; font-family: inherit;
}
.ts-create-input:focus { border-color: var(--accent); }
.ts-err { margin: 0; font-size: 0.65rem; color: var(--danger); }

.ts-create-actions { display: flex; gap: 0.3rem; justify-content: flex-end; }
.ts-create-save, .ts-create-cancel {
  padding: 0.25rem 0.55rem; border-radius: 4px; border: none;
  font-size: 0.68rem; font-weight: 700; cursor: pointer; font-family: inherit;
}
.ts-create-save { background: var(--accent); color: #fff; }
.ts-create-save:disabled { opacity: 0.5; cursor: not-allowed; }
.ts-create-cancel { background: transparent; border: 1px solid var(--border); color: var(--text-muted); }
</style>
