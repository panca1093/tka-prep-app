<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Test = components['schemas']['TestDetailResponse']
type Question = components['schemas']['QuestionDetailResponse']
type Topic = components['schemas']['TopicResponse']

const tests = ref<Test[]>([])
const isLoading = ref(true)
const actionError = ref('')

// Create form
const showCreate = ref(false)
const creating = ref(false)
const createForm = ref({ title: '', description: '', category: 'tka_saintek' as 'tka_saintek' | 'tka_soshum' | 'smbt', duration_minutes: 90, difficulty: 'medium' as 'easy' | 'medium' | 'hard', education_level: '' as '' | 'sd' | 'smp' | 'sma' | 'smk' })

// Question selector
const showQSelector = ref(false)
const selectorTestId = ref('')
const selectorTestTitle = ref('')
const selectorSelectedIds = ref<Set<string>>(new Set())
const selectorQuestions = ref<Question[]>([])
const selectorTopics = ref<Topic[]>([])
const selectorSearch = ref('')
const selectorTopic = ref('')
const selectorDifficulty = ref('')
const selectorLoading = ref(false)
const selectorSaving = ref(false)
const selectorError = ref('')

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
  const el = createForm.value.education_level || undefined
  const { error } = await client.POST('/tests', { body: { ...createForm.value, education_level: el as 'sd' | 'smp' | 'sma' | 'smk' | undefined, scoring_config: { correct_points: 4, wrong_points: 0, blank_points: 0 } } })
  creating.value = false
  if (error) { actionError.value = 'Failed to create test.'; return }
  showCreate.value = false
  await fetchTests()
}

// ─── Question selector ─────────────────────────────────────────────────────
async function openQSelector(t: Test) {
  selectorTestId.value = t.id
  selectorTestTitle.value = t.title
  selectorSelectedIds.value = new Set(t.questions.map((q) => q.question_id))
  selectorSearch.value = ''
  selectorTopic.value = ''
  selectorDifficulty.value = ''
  selectorError.value = ''
  showQSelector.value = true
  await Promise.all([fetchSelectorQuestions(), fetchSelectorTopics()])
}

async function fetchSelectorTopics() {
  const { data } = await client.GET('/topics')
  if (data) selectorTopics.value = data.data
}

async function fetchSelectorQuestions() {
  selectorLoading.value = true
  const { data } = await client.GET('/questions', {
    params: {
      query: {
        search: selectorSearch.value || undefined,
        topic_id: selectorTopic.value || undefined,
        difficulty: (selectorDifficulty.value as 'easy' | 'medium' | 'hard') || undefined,
        limit: 100,
      },
    },
  })
  if (data) selectorQuestions.value = data.data
  selectorLoading.value = false
}

function toggleQ(id: string) {
  if (selectorSelectedIds.value.has(id)) {
    selectorSelectedIds.value.delete(id)
  } else {
    selectorSelectedIds.value.add(id)
  }
  selectorSelectedIds.value = new Set(selectorSelectedIds.value)
}

async function saveQSelection() {
  selectorError.value = ''
  selectorSaving.value = true
  const { error } = await client.PUT('/tests/{testId}/questions', {
    params: { path: { testId: selectorTestId.value } },
    body: { question_ids: [...selectorSelectedIds.value] },
  })
  selectorSaving.value = false
  if (error) { selectorError.value = 'Failed to save questions.'; return }
  showQSelector.value = false
  await fetchTests()
}

const topicName = computed(() => {
  const map = Object.fromEntries(selectorTopics.value.map((t) => [t.id, t.name]))
  return (id: string) => map[id] ?? '—'
})

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
          <button v-if="t.status === 'draft'" class="btn-sm btn-questions" @click="openQSelector(t)">
            Select Questions
          </button>
          <button v-if="t.status === 'draft'" class="btn-sm btn-publish" :disabled="t.questions.length === 0" @click="setStatus(t.id, true)">Publish</button>
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
        <div class="field">
          <label>Education Level</label>
          <select v-model="createForm.education_level" class="form-select">
            <option value="">All Levels</option>
            <option value="sd">SD</option>
            <option value="smp">SMP</option>
            <option value="sma">SMA</option>
            <option value="smk">SMK</option>
          </select>
        </div>
        <div class="modal-actions">
          <button class="btn-cancel" @click="showCreate = false">Cancel</button>
          <button class="btn-primary" :disabled="!createForm.title || creating" @click="createTest">
            {{ creating ? 'Creating…' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Question selector panel -->
    <div v-if="showQSelector" class="qs-backdrop" @click.self="showQSelector = false">
      <div class="qs-panel">
        <div class="qs-header">
          <div>
            <h2 class="qs-title">Select Questions</h2>
            <p class="qs-sub">{{ selectorTestTitle }} · {{ selectorSelectedIds.size }} selected</p>
          </div>
          <button class="qs-close" @click="showQSelector = false">✕</button>
        </div>

        <div class="qs-filters">
          <input
            v-model="selectorSearch"
            class="search-input"
            placeholder="Search questions…"
            @input="fetchSelectorQuestions"
          />
          <select v-model="selectorTopic" class="filter-select" @change="fetchSelectorQuestions">
            <option value="">All Topics</option>
            <option v-for="t in selectorTopics" :key="t.id" :value="t.id">{{ t.name }}</option>
          </select>
          <select v-model="selectorDifficulty" class="filter-select" @change="fetchSelectorQuestions">
            <option value="">All Difficulties</option>
            <option value="easy">Easy</option>
            <option value="medium">Medium</option>
            <option value="hard">Hard</option>
          </select>
        </div>

        <div v-if="selectorLoading" class="qs-loading">Loading…</div>

        <div v-else class="qs-list">
          <div
            v-for="q in selectorQuestions"
            :key="q.id"
            class="qs-item"
            :class="{ selected: selectorSelectedIds.has(q.id) }"
            @click="toggleQ(q.id)"
          >
            <div class="qs-checkbox">
              <span v-if="selectorSelectedIds.has(q.id)" class="check-mark">✓</span>
            </div>
            <div class="qs-content">
              <div class="qs-meta">
                <span class="qs-difficulty" :style="{ color: diffColor[q.difficulty] }">{{ q.difficulty }}</span>
                <span class="qs-topic">{{ topicName(q.topic_id) }}</span>
              </div>
              <p class="qs-text">{{ q.text }}</p>
            </div>
          </div>

          <div v-if="selectorQuestions.length === 0" class="qs-empty">No questions found.</div>
        </div>

        <div class="qs-footer">
          <p v-if="selectorError" class="error-msg">{{ selectorError }}</p>
          <div class="qs-footer-actions">
            <button class="btn-cancel" @click="showQSelector = false">Cancel</button>
            <button class="btn-primary" :disabled="selectorSaving" @click="saveQSelection">
              {{ selectorSaving ? 'Saving…' : `Save (${selectorSelectedIds.size})` }}
            </button>
          </div>
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
.tc-actions { display: flex; gap: 0.5rem; margin-top: 0.25rem; flex-wrap: wrap; }
.btn-sm { padding: 0.3rem 0.75rem; border-radius: 6px; border: 1px solid; font-size: 0.78rem; cursor: pointer; transition: all 0.15s; background: transparent; }
.btn-questions { border-color: #4f8ef7; color: #4f8ef7; }
.btn-questions:hover { background: rgba(79,142,247,0.1); }
.btn-publish { border-color: #22c55e; color: #22c55e; }
.btn-publish:hover:not(:disabled) { background: rgba(34,197,94,0.1); }
.btn-publish:disabled { opacity: 0.4; cursor: not-allowed; }
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

/* ─── Question selector ──────────────────────────────────────────────────── */
.qs-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 200; }
.qs-panel {
  position: fixed; right: 0; top: 0; bottom: 0; width: 560px; max-width: 100vw;
  background: #0d1424; border-left: 1px solid #1e2a45;
  display: flex; flex-direction: column;
}

.qs-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  padding: 1.5rem 1.5rem 1rem;
  border-bottom: 1px solid #1e2a45;
}
.qs-title { margin: 0; font-size: 1.05rem; font-weight: 700; }
.qs-sub { margin: 0.25rem 0 0; font-size: 0.8rem; color: #4f8ef7; }
.qs-close { background: transparent; border: none; color: #64748b; font-size: 1.1rem; cursor: pointer; padding: 0.25rem; }
.qs-close:hover { color: #f1f5f9; }

.qs-filters {
  display: flex; gap: 0.625rem; padding: 1rem 1.5rem;
  border-bottom: 1px solid #1e2a45; flex-wrap: wrap;
}
.search-input {
  flex: 1; min-width: 160px;
  padding: 0.5rem 0.75rem; border-radius: 8px;
  border: 1px solid #1e2a45; background: #141c2e; color: #f1f5f9; font-size: 0.85rem; outline: none;
}
.search-input:focus { border-color: #4f8ef7; }
.filter-select {
  padding: 0.5rem 0.75rem; border-radius: 8px;
  border: 1px solid #1e2a45; background: #141c2e; color: #f1f5f9; font-size: 0.85rem;
  cursor: pointer; outline: none;
}

.qs-loading { padding: 2rem; color: #94a3b8; text-align: center; }

.qs-list { flex: 1; overflow-y: auto; padding: 0.75rem 1.5rem; display: flex; flex-direction: column; gap: 0.5rem; }

.qs-item {
  display: flex; align-items: flex-start; gap: 0.875rem;
  padding: 0.875rem; border-radius: 10px; border: 1px solid #1e2a45;
  background: #141c2e; cursor: pointer; transition: all 0.12s;
}
.qs-item:hover { border-color: #4f8ef7; }
.qs-item.selected { border-color: #4f8ef7; background: rgba(79,142,247,0.08); }

.qs-checkbox {
  width: 1.25rem; height: 1.25rem; flex-shrink: 0; border-radius: 4px;
  border: 2px solid #1e2a45; background: #0d1424;
  display: flex; align-items: center; justify-content: center;
  margin-top: 0.1rem;
}
.qs-item.selected .qs-checkbox { border-color: #4f8ef7; background: #4f8ef7; }
.check-mark { color: #fff; font-size: 0.7rem; font-weight: 900; }

.qs-content { flex: 1; min-width: 0; }
.qs-meta { display: flex; gap: 0.625rem; margin-bottom: 0.35rem; }
.qs-difficulty { font-size: 0.72rem; font-weight: 700; text-transform: capitalize; }
.qs-topic { font-size: 0.72rem; color: #64748b; }
.qs-text { margin: 0; font-size: 0.875rem; color: #e2e8f0; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.qs-empty { padding: 2rem; text-align: center; color: #94a3b8; font-size: 0.875rem; }

.qs-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #1e2a45;
}
.qs-footer-actions { display: flex; gap: 0.75rem; }
.qs-footer-actions .btn-cancel { flex: 0 0 auto; padding: 0.65rem 1.25rem; }
.qs-footer-actions .btn-primary { flex: 1; }
</style>
