<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Test = components['schemas']['TestDetailResponse']

const tests = ref<Test[]>([])
const isLoading = ref(true)
const actionError = ref('')

// Create form
const showCreate = ref(false)
const creating = ref(false)
const createForm = ref({ title: '', description: '', category: 'tka_saintek' as 'tka_saintek' | 'tka_soshum' | 'smbt', duration_minutes: 90, difficulty: 'medium' as 'easy' | 'medium' | 'hard' })

// Publish / unpublish
async function setStatus(testId: string, publish: boolean) {
  actionError.value = ''
  const path = publish ? '/tests/{testId}/publish' : '/tests/{testId}/unpublish'
  const { error } = await client.POST(path as '/tests/{testId}/publish', { params: { path: { testId } } })
  if (error) { actionError.value = 'Failed to update status.'; return }
  await fetchTests()
}

async function deleteTest(id: string) {
  if (!confirm('Delete this test?')) return
  const { error } = await client.DELETE('/tests/{testId}', { params: { path: { testId: id } } })
  if (error) { actionError.value = 'Cannot delete published test. Unpublish first.'; return }
  await fetchTests()
}

async function fetchTests() {
  const { data } = await client.GET('/tests', { params: { query: { limit: 50 } } })
  if (data) tests.value = data.data
}

async function createTest() {
  creating.value = true
  const { error } = await client.POST('/tests', { body: { ...createForm.value, scoring_config: { correct_points: 4, wrong_points: 0, blank_points: 0 } } })
  creating.value = false
  if (error) { actionError.value = 'Failed to create test.'; return }
  showCreate.value = false
  await fetchTests()
}

const categoryLabel: Record<string, string> = { tka_saintek: 'TKA Saintek', tka_soshum: 'TKA Soshum', smbt: 'SMBT' }
const diffColor: Record<string, string> = { easy: '#22c55e', medium: '#f59e0b', hard: '#ef4444' }

onMounted(async () => { await fetchTests(); isLoading.value = false })
</script>

<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Tests</h1>
      <button class="btn-primary" @click="showCreate = true">+ New Test</button>
    </div>

    <p v-if="actionError" class="error-msg">{{ actionError }}</p>
    <div v-if="isLoading" class="loading">Loading…</div>

    <div v-else-if="tests.length === 0" class="empty-state">No tests yet. Create your first test!</div>

    <div v-else class="test-list">
      <div v-for="t in tests" :key="t.id" class="test-card">
        <div class="tc-header">
          <span class="tc-category">{{ categoryLabel[t.category] }}</span>
          <span class="tc-status" :class="t.status">{{ t.status }}</span>
        </div>
        <h3 class="tc-title">{{ t.title }}</h3>
        <div class="tc-meta">
          <span :style="{ color: diffColor[t.difficulty] }">{{ t.difficulty }}</span>
          <span>{{ t.duration_minutes }} min</span>
          <span>{{ t.questions.length }} questions</span>
        </div>
        <div class="tc-actions">
          <button v-if="t.status === 'draft'" class="btn-sm btn-publish" @click="setStatus(t.id, true)">Publish</button>
          <button v-else class="btn-sm btn-unpublish" @click="setStatus(t.id, false)">Unpublish</button>
          <button v-if="t.status === 'draft'" class="btn-sm btn-delete" @click="deleteTest(t.id)">Delete</button>
        </div>
      </div>
    </div>

    <!-- Create modal -->
    <div v-if="showCreate" class="modal-backdrop" @click.self="showCreate = false">
      <div class="modal">
        <h3>New Test</h3>
        <div class="field">
          <label>Title</label>
          <input v-model="createForm.title" class="form-input" placeholder="Test title…" />
        </div>
        <div class="field">
          <label>Description (optional)</label>
          <input v-model="createForm.description" class="form-input" placeholder="Brief description…" />
        </div>
        <div class="form-row">
          <div class="field">
            <label>Category</label>
            <select v-model="createForm.category" class="form-select">
              <option value="tka_saintek">TKA Saintek</option>
              <option value="tka_soshum">TKA Soshum</option>
              <option value="smbt">SMBT</option>
            </select>
          </div>
          <div class="field">
            <label>Difficulty</label>
            <select v-model="createForm.difficulty" class="form-select">
              <option value="easy">Easy</option>
              <option value="medium">Medium</option>
              <option value="hard">Hard</option>
            </select>
          </div>
        </div>
        <div class="field">
          <label>Duration (minutes)</label>
          <input v-model.number="createForm.duration_minutes" type="number" class="form-input" min="5" max="300" />
        </div>
        <div class="modal-actions">
          <button class="btn-cancel" @click="showCreate = false">Cancel</button>
          <button class="btn-primary" :disabled="!createForm.title || creating" @click="createTest">
            {{ creating ? 'Creating…' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
.page-title { margin: 0; font-size: 1.5rem; font-weight: 800; }
.loading { color: #94a3b8; }
.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: #450a0a; color: #fca5a5; font-size: 0.825rem; margin-bottom: 1rem; }
.empty-state { color: #94a3b8; text-align: center; padding: 2rem; background: #141c2e; border: 1px solid #1e2a45; border-radius: 12px; }

.test-list { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1rem; }
.test-card { background: #141c2e; border: 1px solid #1e2a45; border-radius: 12px; padding: 1.25rem; display: flex; flex-direction: column; gap: 0.5rem; }
.tc-header { display: flex; align-items: center; justify-content: space-between; }
.tc-category { font-size: 0.72rem; font-weight: 600; color: #4f8ef7; text-transform: uppercase; }
.tc-status { font-size: 0.72rem; font-weight: 700; padding: 0.2rem 0.6rem; border-radius: 4px; text-transform: capitalize; }
.tc-status.published { background: rgba(34,197,94,0.15); color: #22c55e; }
.tc-status.draft { background: rgba(148,163,184,0.15); color: #94a3b8; }
.tc-title { margin: 0; font-size: 0.95rem; font-weight: 700; }
.tc-meta { display: flex; gap: 0.75rem; font-size: 0.78rem; color: #64748b; }
.tc-actions { display: flex; gap: 0.5rem; margin-top: 0.25rem; }
.btn-sm { padding: 0.3rem 0.75rem; border-radius: 6px; border: 1px solid; font-size: 0.78rem; cursor: pointer; transition: all 0.15s; background: transparent; }
.btn-publish { border-color: #22c55e; color: #22c55e; }
.btn-publish:hover { background: rgba(34,197,94,0.1); }
.btn-unpublish { border-color: #f59e0b; color: #f59e0b; }
.btn-unpublish:hover { background: rgba(245,158,11,0.1); }
.btn-delete { border-color: #ef4444; color: #ef4444; }
.btn-delete:hover { background: rgba(239,68,68,0.1); }

.btn-primary { padding: 0.55rem 1.25rem; border-radius: 8px; border: none; background: #4f8ef7; color: #fff; font-size: 0.875rem; font-weight: 600; cursor: pointer; transition: background 0.15s; }
.btn-primary:hover { background: #3b7be8; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.6); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal { background: #141c2e; border: 1px solid #1e2a45; border-radius: 16px; padding: 2rem; width: 420px; max-width: calc(100vw - 2rem); display: flex; flex-direction: column; gap: 1rem; }
.modal h3 { margin: 0; font-size: 1.1rem; }
.field { display: flex; flex-direction: column; gap: 0.375rem; }
.field label { font-size: 0.78rem; font-weight: 600; color: #94a3b8; }
.form-input, .form-select { padding: 0.6rem 0.875rem; border-radius: 8px; border: 1px solid #1e2a45; background: #0d1424; color: #f1f5f9; font-size: 0.875rem; outline: none; }
.form-input:focus, .form-select:focus { border-color: #4f8ef7; }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.modal-actions { display: flex; gap: 0.75rem; margin-top: 0.5rem; }
.btn-cancel { flex: 1; padding: 0.65rem; border-radius: 8px; border: 1px solid #1e2a45; background: transparent; color: #94a3b8; cursor: pointer; transition: all 0.15s; }
.btn-cancel:hover { border-color: #ef4444; color: #ef4444; }
</style>
