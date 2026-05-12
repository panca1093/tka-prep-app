<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Question = components['schemas']['QuestionDetailResponse']
type Topic = components['schemas']['TopicResponse']
type CreateReq = components['schemas']['CreateQuestionRequest']

const questions = ref<Question[]>([])
const topics = ref<Topic[]>([])
const total = ref(0)
const isLoading = ref(false)
const showForm = ref(false)
const editTarget = ref<Question | null>(null)
const deleteError = ref('')

// Filters
const searchText = ref('')
const topicFilter = ref('')
const difficultyFilter = ref('')

// Form state
const form = ref<{
  topic_id: string
  text: string
  explanation: string
  difficulty: 'easy' | 'medium' | 'hard'
  options: { label: string; text: string; is_correct: boolean }[]
}>({
  topic_id: '',
  text: '',
  explanation: '',
  difficulty: 'medium',
  options: [
    { label: 'A', text: '', is_correct: false },
    { label: 'B', text: '', is_correct: false },
    { label: 'C', text: '', is_correct: false },
    { label: 'D', text: '', is_correct: false },
    { label: 'E', text: '', is_correct: false },
  ],
})
const formError = ref('')
const isSaving = ref(false)

async function fetchQuestions() {
  isLoading.value = true
  const { data } = await client.GET('/questions', {
    params: {
      query: {
        search: searchText.value || undefined,
        topic_id: topicFilter.value || undefined,
        difficulty: (difficultyFilter.value as 'easy' | 'medium' | 'hard') || undefined,
        limit: 50,
      },
    },
  })
  if (data) { questions.value = data.data; total.value = data.total }
  isLoading.value = false
}

onMounted(async () => {
  const { data } = await client.GET('/topics')
  if (data) topics.value = data.data
  await fetchQuestions()
})

function openCreate() {
  editTarget.value = null
  form.value = { topic_id: '', text: '', explanation: '', difficulty: 'medium', options: 'ABCDE'.split('').map(l => ({ label: l, text: '', is_correct: false })) }
  formError.value = ''
  showForm.value = true
}

function setCorrect(idx: number) {
  form.value.options.forEach((o, i) => { o.is_correct = i === idx })
}

async function saveQuestion() {
  formError.value = ''
  const correctCount = form.value.options.filter((o) => o.is_correct).length
  if (correctCount !== 1) { formError.value = 'Exactly one option must be marked correct.'; return }
  if (!form.value.topic_id) { formError.value = 'Please select a topic.'; return }

  isSaving.value = true
  try {
    const body: CreateReq = {
      topic_id: form.value.topic_id,
      text: form.value.text,
      explanation: form.value.explanation || undefined,
      difficulty: form.value.difficulty,
      options: form.value.options,
    }
    if (editTarget.value) {
      await client.PATCH('/questions/{questionId}', { params: { path: { questionId: editTarget.value.id } }, body })
    } else {
      await client.POST('/questions', { body })
    }
    showForm.value = false
    await fetchQuestions()
  } catch {
    formError.value = 'Failed to save question.'
  } finally {
    isSaving.value = false
  }
}

async function deleteQuestion(id: string) {
  deleteError.value = ''
  if (!confirm('Delete this question?')) return
  const { error } = await client.DELETE('/questions/{questionId}', { params: { path: { questionId: id } } })
  if (error) { deleteError.value = 'Cannot delete — question may be used in a published test.'; return }
  await fetchQuestions()
}

const diffColor: Record<string, string> = { easy: '#22c55e', medium: '#f59e0b', hard: '#ef4444' }
</script>

<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Question Bank</h1>
      <button class="btn-primary" @click="openCreate">+ Add Question</button>
    </div>

    <div class="filters">
      <input v-model="searchText" placeholder="Search questions…" class="search-input" @input="fetchQuestions" />
      <select v-model="topicFilter" class="filter-select" @change="fetchQuestions">
        <option value="">All Topics</option>
        <option v-for="t in topics" :key="t.id" :value="t.id">{{ t.name }}</option>
      </select>
      <select v-model="difficultyFilter" class="filter-select" @change="fetchQuestions">
        <option value="">All Difficulties</option>
        <option value="easy">Easy</option>
        <option value="medium">Medium</option>
        <option value="hard">Hard</option>
      </select>
    </div>

    <p v-if="deleteError" class="error-msg">{{ deleteError }}</p>

    <div v-if="isLoading" class="loading">Loading…</div>

    <div v-else-if="questions.length === 0" class="empty-state">No questions yet. Add your first question!</div>

    <div v-else class="question-list">
      <div v-for="q in questions" :key="q.id" class="question-card">
        <div class="q-header">
          <span class="q-difficulty" :style="{ color: diffColor[q.difficulty] }">{{ q.difficulty }}</span>
          <span class="q-topic">{{ topics.find(t => t.id === q.topic_id)?.name ?? '—' }}</span>
          <div class="q-actions">
            <button class="icon-btn" @click="deleteQuestion(q.id)">🗑</button>
          </div>
        </div>
        <p class="q-text">{{ q.text }}</p>
        <div class="q-options">
          <span v-for="o in q.options" :key="o.id" class="q-opt" :class="{ correct: o.is_correct }">
            {{ o.label }}
          </span>
        </div>
      </div>
    </div>

    <!-- Slide-in form -->
    <div v-if="showForm" class="form-backdrop" @click.self="showForm = false">
      <div class="form-panel">
        <h2>{{ editTarget ? 'Edit Question' : 'New Question' }}</h2>

        <div class="field">
          <label>Topic</label>
          <select v-model="form.topic_id" class="filter-select">
            <option value="">Select topic…</option>
            <option v-for="t in topics" :key="t.id" :value="t.id">{{ t.name }}</option>
          </select>
        </div>

        <div class="field">
          <label>Difficulty</label>
          <select v-model="form.difficulty" class="filter-select">
            <option value="easy">Easy</option>
            <option value="medium">Medium</option>
            <option value="hard">Hard</option>
          </select>
        </div>

        <div class="field">
          <label>Question Text</label>
          <textarea v-model="form.text" rows="3" class="text-area" placeholder="Enter question…" />
        </div>

        <div class="field">
          <label>Options (click ● to mark correct)</label>
          <div class="option-inputs">
            <div v-for="(opt, i) in form.options" :key="opt.label" class="opt-input-row">
              <button class="correct-dot" :class="{ active: opt.is_correct }" @click="setCorrect(i)">{{ opt.label }}</button>
              <input v-model="opt.text" class="opt-text-input" :placeholder="`Option ${opt.label}`" />
            </div>
          </div>
        </div>

        <div class="field">
          <label>Explanation (optional)</label>
          <textarea v-model="form.explanation" rows="2" class="text-area" placeholder="Explain the correct answer…" />
        </div>

        <p v-if="formError" class="error-msg">{{ formError }}</p>

        <div class="form-actions">
          <button class="btn-cancel" @click="showForm = false">Cancel</button>
          <button class="btn-primary" :disabled="isSaving" @click="saveQuestion">
            {{ isSaving ? 'Saving…' : 'Save Question' }}
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

.filters { display: flex; gap: 0.75rem; margin-bottom: 1.5rem; flex-wrap: wrap; }
.search-input {
  flex: 1; min-width: 200px;
  padding: 0.55rem 0.875rem; border-radius: 8px;
  border: 1px solid #1e2a45; background: #141c2e; color: #f1f5f9; font-size: 0.875rem;
  outline: none;
}
.search-input:focus { border-color: #4f8ef7; }
.filter-select {
  padding: 0.55rem 0.875rem; border-radius: 8px;
  border: 1px solid #1e2a45; background: #141c2e; color: #f1f5f9; font-size: 0.875rem;
  cursor: pointer; outline: none;
}

.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: #450a0a; color: #fca5a5; font-size: 0.825rem; margin-bottom: 1rem; }
.empty-state { color: #94a3b8; text-align: center; padding: 2rem; background: #141c2e; border: 1px solid #1e2a45; border-radius: 12px; }

.question-list { display: flex; flex-direction: column; gap: 0.625rem; }
.question-card { background: #141c2e; border: 1px solid #1e2a45; border-radius: 10px; padding: 1rem; }
.q-header { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.5rem; }
.q-difficulty { font-size: 0.75rem; font-weight: 700; text-transform: capitalize; }
.q-topic { font-size: 0.75rem; color: #64748b; }
.q-actions { margin-left: auto; }
.icon-btn { background: transparent; border: none; cursor: pointer; font-size: 0.9rem; color: #64748b; padding: 0.2rem; }
.icon-btn:hover { color: #ef4444; }
.q-text { margin: 0 0 0.625rem; font-size: 0.9rem; color: #e2e8f0; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.q-options { display: flex; gap: 0.375rem; }
.q-opt { width: 1.6rem; height: 1.6rem; border-radius: 50%; background: #1e2a45; display: flex; align-items: center; justify-content: center; font-size: 0.7rem; font-weight: 700; color: #94a3b8; }
.q-opt.correct { background: #22c55e; color: #000; }

.btn-primary {
  padding: 0.55rem 1.25rem; border-radius: 8px; border: none;
  background: #4f8ef7; color: #fff; font-size: 0.875rem; font-weight: 600;
  cursor: pointer; transition: background 0.15s;
}
.btn-primary:hover { background: #3b7be8; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

/* Form panel */
.form-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 50; }
.form-panel {
  position: fixed; right: 0; top: 0; bottom: 0; width: 480px; max-width: 100vw;
  background: #141c2e; border-left: 1px solid #1e2a45;
  padding: 2rem; overflow-y: auto;
  display: flex; flex-direction: column; gap: 1rem;
}
.form-panel h2 { margin: 0; font-size: 1.1rem; }
.field { display: flex; flex-direction: column; gap: 0.375rem; }
.field label { font-size: 0.78rem; font-weight: 600; color: #94a3b8; }
.text-area {
  padding: 0.65rem 0.875rem; border-radius: 8px; border: 1px solid #1e2a45;
  background: #0d1424; color: #f1f5f9; font-size: 0.875rem; resize: vertical; font-family: inherit;
  outline: none;
}
.text-area:focus { border-color: #4f8ef7; }

.option-inputs { display: flex; flex-direction: column; gap: 0.5rem; }
.opt-input-row { display: flex; align-items: center; gap: 0.5rem; }
.correct-dot {
  width: 2rem; height: 2rem; border-radius: 50%; border: 2px solid #1e2a45;
  background: #0d1424; color: #94a3b8; font-size: 0.75rem; font-weight: 700;
  cursor: pointer; flex-shrink: 0; display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.correct-dot.active { background: #22c55e; border-color: #22c55e; color: #000; }
.opt-text-input {
  flex: 1; padding: 0.55rem 0.75rem; border-radius: 8px; border: 1px solid #1e2a45;
  background: #0d1424; color: #f1f5f9; font-size: 0.875rem; outline: none;
}
.opt-text-input:focus { border-color: #4f8ef7; }

.form-actions { display: flex; gap: 0.75rem; margin-top: auto; padding-top: 1rem; }
.btn-cancel {
  flex: 1; padding: 0.65rem; border-radius: 8px; border: 1px solid #1e2a45;
  background: transparent; color: #94a3b8; cursor: pointer; font-size: 0.875rem;
  transition: all 0.15s;
}
.btn-cancel:hover { border-color: #ef4444; color: #ef4444; }
</style>
