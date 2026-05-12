<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type AdminTest = components['schemas']['AdminTestResponse']

const tests = ref<AdminTest[]>([])
const total = ref(0)
const page = ref(1)
const isLoading = ref(false)
const actionError = ref('')

async function fetchTests() {
  isLoading.value = true
  const { data } = await client.GET('/admin/tests', { params: { query: { page: page.value, limit: 20 } } })
  if (data) { tests.value = data.data; total.value = data.total }
  isLoading.value = false
}

async function unpublish(testId: string) {
  actionError.value = ''
  const { error } = await client.POST('/tests/{testId}/unpublish', { params: { path: { testId } } })
  if (error) { actionError.value = 'Failed to unpublish.'; return }
  await fetchTests()
}

onMounted(fetchTests)

const categoryLabel: Record<string, string> = { tka_saintek: 'TKA Saintek', tka_soshum: 'TKA Soshum', smbt: 'SMBT' }
const diffColor: Record<string, string> = { easy: '#22c55e', medium: '#f59e0b', hard: '#ef4444' }
</script>

<template>
  <div>
    <h1 class="page-title">All Tests</h1>
    <p v-if="actionError" class="error-msg">{{ actionError }}</p>

    <div v-if="isLoading" class="loading">Loading…</div>

    <div v-else class="table-wrap">
      <table class="test-table">
        <thead>
          <tr>
            <th>Title</th><th>Category</th><th>Difficulty</th><th>Status</th><th>Attempts</th><th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="t in tests" :key="t.id">
            <td class="title-cell">{{ t.title }}</td>
            <td>{{ categoryLabel[t.category] }}</td>
            <td :style="{ color: diffColor[t.difficulty] }">{{ t.difficulty }}</td>
            <td>
              <span class="status-badge" :class="t.status">{{ t.status }}</span>
            </td>
            <td class="attempts-cell">{{ t.attempt_count }}</td>
            <td>
              <button v-if="t.status === 'published'" class="btn-action" @click="unpublish(t.id)">Unpublish</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 1.5rem; font-size: 1.5rem; font-weight: 800; }
.loading { color: #94a3b8; }
.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: #450a0a; color: #fca5a5; font-size: 0.825rem; margin-bottom: 0.75rem; }

.table-wrap { background: #141c2e; border: 1px solid #1e2a45; border-radius: 12px; overflow: auto; }
.test-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
.test-table th { padding: 0.75rem 1rem; text-align: left; font-size: 0.75rem; font-weight: 700; color: #64748b; text-transform: uppercase; border-bottom: 1px solid #1e2a45; }
.test-table td { padding: 0.75rem 1rem; border-bottom: 1px solid #0d1424; vertical-align: middle; }
.test-table tbody tr:last-child td { border-bottom: none; }
.test-table tbody tr:hover td { background: #1a2535; }
.title-cell { font-weight: 600; max-width: 260px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.attempts-cell { color: #4f8ef7; font-weight: 700; }
.status-badge { font-size: 0.75rem; font-weight: 700; padding: 0.2rem 0.5rem; border-radius: 4px; text-transform: capitalize; }
.status-badge.published { background: rgba(34,197,94,0.15); color: #22c55e; }
.status-badge.draft { background: rgba(148,163,184,0.1); color: #94a3b8; }
.btn-action { padding: 0.3rem 0.75rem; border-radius: 6px; font-size: 0.78rem; font-weight: 600; cursor: pointer; background: transparent; border: 1px solid #f59e0b; color: #f59e0b; transition: background 0.15s; }
.btn-action:hover { background: rgba(245,158,11,0.1); }
</style>
