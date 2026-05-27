<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type QuestionUsageEntry = components['schemas']['QuestionUsageEntry']

const entries = ref<QuestionUsageEntry[]>([])
const isLoading = ref(false)

onMounted(async () => {
  isLoading.value = true
  const { data } = await client.GET('/questions/usage-stats', { params: { query: { limit: 10 } } })
  if (data) entries.value = data.data
  isLoading.value = false
})
</script>

<template>
  <div class="muq-widget">
    <h3 class="muq-title">Pertanyaan Paling Banyak Digunakan</h3>
    <div v-if="isLoading" class="muq-loading">Memuat…</div>
    <div v-else-if="entries.length === 0" class="muq-empty">
      Belum ada pertanyaan yang digunakan dalam tes.
    </div>
    <ol v-else class="muq-list">
      <li v-for="e in entries" :key="e.question_id" class="muq-item">
        <div class="muq-text">{{ e.question_text }}</div>
        <div class="muq-meta">
          <span class="muq-topic">{{ e.topic_name }}</span>
          <span class="muq-counts">{{ e.own_test_count }} tes kamu · {{ e.other_test_count }} tes lain</span>
        </div>
      </li>
    </ol>
  </div>
</template>

<style scoped>
.muq-widget {
  background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: 14px; padding: 1.25rem 1.375rem;
}
.muq-title { margin: 0 0 0.875rem; font-size: 0.85rem; font-weight: 700; color: var(--text-heading); }
.muq-loading, .muq-empty { font-size: 0.78rem; color: var(--text-muted); }
.muq-list { margin: 0; padding: 0; list-style: none; display: flex; flex-direction: column; gap: 0.5rem; }
.muq-item { padding: 0.55rem 0.65rem; border-radius: 8px; background: var(--bg-input); border: 1px solid var(--border); }
.muq-text { font-size: 0.8rem; font-weight: 600; color: var(--text-primary); margin-bottom: 0.2rem; }
.muq-meta { display: flex; gap: 0.75rem; font-size: 0.68rem; color: var(--text-muted); }
.muq-topic { color: var(--accent); font-weight: 600; }
.muq-counts { color: var(--text-muted); }
</style>
