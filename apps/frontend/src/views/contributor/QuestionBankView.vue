<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'
import ImageUpload from '@/components/ImageUpload.vue'

type Question = components['schemas']['QuestionDetailResponse']
type Topic = components['schemas']['TopicResponse']
type CreateReq = components['schemas']['CreateQuestionRequest']
type QuestionType = 'mcq' | 'multi_correct' | 'true_false'

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
const typeFilter = ref('')

// Form state
interface FormState {
  question_type: QuestionType
  topic_id: string
  text: string
  explanation: string
  image_url: string | null
  difficulty: 'easy' | 'medium' | 'hard'
  options: { label: string; text: string; is_correct: boolean; image_url: string | null }[]
  statements: { text: string; is_correct: boolean; image_url: string | null }[]
}

const emptyForm = (): FormState => ({
  question_type: 'mcq',
  topic_id: '',
  text: '',
  explanation: '',
  image_url: null,
  difficulty: 'medium',
  options: 'ABCDE'.split('').map(l => ({ label: l, text: '', is_correct: false, image_url: null })),
  statements: [
    { text: '', is_correct: true, image_url: null },
    { text: '', is_correct: false, image_url: null },
  ],
})

const form = ref<FormState>(emptyForm())
const formError = ref('')
const isSaving = ref(false)

// Computed helpers
const isMultiCorrect = computed(() => form.value.question_type === 'multi_correct')
const isTrueFalse = computed(() => form.value.question_type === 'true_false')

async function fetchQuestions() {
  isLoading.value = true
  const { data } = await client.GET('/questions', {
    params: {
      query: {
        search: searchText.value || undefined,
        topic_id: topicFilter.value || undefined,
        difficulty: (difficultyFilter.value as 'easy' | 'medium' | 'hard') || undefined,
        question_type: (typeFilter.value as QuestionType) || undefined,
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
  form.value = emptyForm()
  formError.value = ''
  showForm.value = true
}

function onTypeChange() {
  formError.value = ''
}

// MCQ: only one correct at a time
function setCorrectMCQ(idx: number) {
  form.value.options.forEach((o, i) => { o.is_correct = i === idx })
}

// PGK: toggle correct
function toggleCorrectPGK(idx: number) {
  form.value.options[idx].is_correct = !form.value.options[idx].is_correct
}

// B/S statements
function addStatement() {
  if (form.value.statements.length >= 6) return
  form.value.statements.push({ text: '', is_correct: false, image_url: null })
}

function removeStatement(idx: number) {
  if (form.value.statements.length <= 2) return
  form.value.statements.splice(idx, 1)
}

function toggleStatementCorrect(idx: number) {
  form.value.statements[idx].is_correct = !form.value.statements[idx].is_correct
}

async function saveQuestion() {
  formError.value = ''
  if (!form.value.topic_id) { formError.value = 'Please select a topic.'; return }
  if (!form.value.text.trim()) { formError.value = 'Question text is required.'; return }

  let body: CreateReq

  if (form.value.question_type === 'mcq') {
    const correctCount = form.value.options.filter((o) => o.is_correct).length
    if (correctCount !== 1) { formError.value = 'Exactly one option must be marked correct.'; return }
    body = {
      question_type: 'mcq',
      topic_id: form.value.topic_id,
      text: form.value.text,
      explanation: form.value.explanation || undefined,
      image_url: form.value.image_url ?? undefined,
      difficulty: form.value.difficulty,
      options: form.value.options.map(o => ({ label: o.label, text: o.text, is_correct: o.is_correct, image_url: o.image_url ?? undefined })),
    }
  } else if (form.value.question_type === 'multi_correct') {
    const correctCount = form.value.options.filter((o) => o.is_correct).length
    if (correctCount < 1) { formError.value = 'At least one option must be marked correct for PGK.'; return }
    body = {
      question_type: 'multi_correct',
      topic_id: form.value.topic_id,
      text: form.value.text,
      explanation: form.value.explanation || undefined,
      image_url: form.value.image_url ?? undefined,
      difficulty: form.value.difficulty,
      options: form.value.options.map(o => ({ label: o.label, text: o.text, is_correct: o.is_correct, image_url: o.image_url ?? undefined })),
    }
  } else {
    const stmts = form.value.statements
    if (stmts.length < 2) { formError.value = 'Add at least 2 statements.'; return }
    if (stmts.some(s => !s.text.trim())) { formError.value = 'All statement texts are required.'; return }
    body = {
      question_type: 'true_false',
      topic_id: form.value.topic_id,
      text: form.value.text,
      explanation: form.value.explanation || undefined,
      image_url: form.value.image_url ?? undefined,
      difficulty: form.value.difficulty,
      statements: stmts.map((s, i) => ({ text: s.text, is_correct: s.is_correct, position: i, image_url: s.image_url ?? undefined })),
    }
  }

  isSaving.value = true
  try {
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
const typeLabel: Record<string, string> = { mcq: 'PG', multi_correct: 'PGK', true_false: 'B/S' }
const typeColor: Record<string, string> = {
  mcq: '#4f8ef7',
  multi_correct: '#a855f7',
  true_false: '#22c55e',
}
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
      <select v-model="typeFilter" class="filter-select" @change="fetchQuestions">
        <option value="">All Types</option>
        <option value="mcq">Pilihan Ganda (PG)</option>
        <option value="multi_correct">PGK</option>
        <option value="true_false">Benar/Salah</option>
      </select>
    </div>

    <p v-if="deleteError" class="error-msg">{{ deleteError }}</p>

    <div v-if="isLoading" class="loading">Loading…</div>

    <div v-else-if="questions.length === 0" class="empty-state">No questions yet. Add your first question!</div>

    <div v-else class="question-list">
      <div v-for="q in questions" :key="q.id" class="question-card">
        <div class="q-header">
          <span
            class="q-type-badge"
            :style="{ color: typeColor[q.question_type], borderColor: typeColor[q.question_type] + '44', background: typeColor[q.question_type] + '18' }"
          >{{ typeLabel[q.question_type] ?? q.question_type }}</span>
          <span class="q-difficulty" :style="{ color: diffColor[q.difficulty] }">{{ q.difficulty }}</span>
          <span class="q-topic">{{ topics.find(t => t.id === q.topic_id)?.name ?? '—' }}</span>
          <div class="q-actions">
            <button class="icon-btn" @click="deleteQuestion(q.id)">🗑</button>
          </div>
        </div>
        <p class="q-text">{{ q.text }}</p>
        <!-- Options summary for MCQ / PGK -->
        <div v-if="q.question_type !== 'true_false'" class="q-options">
          <span
            v-for="o in q.options"
            :key="o.id"
            class="q-opt"
            :class="{ correct: o.is_correct }"
          >{{ o.label }}</span>
        </div>
        <!-- Statement count for B/S -->
        <div v-else class="q-stmt-count">
          {{ q.statements.length }} pernyataan
        </div>
      </div>
    </div>

    <!-- Slide-in form -->
    <div v-if="showForm" class="form-backdrop" @click.self="showForm = false">
      <div class="form-panel">
        <h2>{{ editTarget ? 'Edit Question' : 'New Question' }}</h2>

        <!-- Question type -->
        <div class="field">
          <label>Question Type</label>
          <div class="type-tabs">
            <button
              v-for="t in [{ v: 'mcq', label: 'Pilihan Ganda (PG)' }, { v: 'multi_correct', label: 'PGK' }, { v: 'true_false', label: 'Benar / Salah' }]"
              :key="t.v"
              class="type-tab"
              :class="{ active: form.question_type === t.v }"
              @click="form.question_type = t.v as QuestionType; onTypeChange()"
            >{{ t.label }}</button>
          </div>
        </div>

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

        <ImageUpload v-model="form.image_url" label="Question image (optional)" />

        <!-- MCQ options: single correct -->
        <div v-if="!isTrueFalse" class="field">
          <label>
            Options
            <span v-if="isMultiCorrect" class="hint-inline"> — mark all correct answers</span>
            <span v-else class="hint-inline"> — click label to mark correct</span>
          </label>
          <div class="option-inputs">
            <div v-for="(opt, i) in form.options" :key="opt.label" class="opt-input-row-wrap">
              <div class="opt-input-row">
                <button
                  class="correct-dot"
                  :class="{ active: opt.is_correct, pgk: isMultiCorrect }"
                  @click="isMultiCorrect ? toggleCorrectPGK(i) : setCorrectMCQ(i)"
                >{{ opt.is_correct && isMultiCorrect ? '✓' : opt.label }}</button>
                <input v-model="opt.text" class="opt-text-input" :placeholder="`Option ${opt.label}`" />
              </div>
              <ImageUpload v-model="opt.image_url" class="opt-img-upload" />
            </div>
          </div>
        </div>

        <!-- B/S statements -->
        <div v-if="isTrueFalse" class="field">
          <label>Pernyataan ({{ form.statements.length }}/6)</label>
          <div class="stmt-inputs">
            <div v-for="(stmt, i) in form.statements" :key="i" class="stmt-input-block">
              <div class="stmt-input-row">
                <span class="stmt-idx">{{ i + 1 }}</span>
                <input v-model="stmt.text" class="opt-text-input" :placeholder="`Pernyataan ${i + 1}`" />
                <button
                  class="bs-toggle"
                  :class="{ benar: stmt.is_correct, salah: !stmt.is_correct }"
                  @click="toggleStatementCorrect(i)"
                >{{ stmt.is_correct ? 'Benar' : 'Salah' }}</button>
                <button class="remove-btn" :disabled="form.statements.length <= 2" @click="removeStatement(i)">✕</button>
              </div>
              <ImageUpload v-model="stmt.image_url" class="stmt-img-upload" />
            </div>
          </div>
          <button class="btn-add-stmt" :disabled="form.statements.length >= 6" @click="addStatement">
            + Tambah Pernyataan
          </button>
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
.q-header { display: flex; align-items: center; gap: 0.625rem; margin-bottom: 0.5rem; }
.q-type-badge {
  font-size: 0.68rem; font-weight: 800; padding: 0.2rem 0.5rem;
  border-radius: 4px; border: 1px solid; text-transform: uppercase; letter-spacing: 0.03em;
}
.q-difficulty { font-size: 0.75rem; font-weight: 700; text-transform: capitalize; }
.q-topic { font-size: 0.75rem; color: #64748b; }
.q-actions { margin-left: auto; }
.icon-btn { background: transparent; border: none; cursor: pointer; font-size: 0.9rem; color: #64748b; padding: 0.2rem; }
.icon-btn:hover { color: #ef4444; }
.q-text { margin: 0 0 0.625rem; font-size: 0.9rem; color: #e2e8f0; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.q-options { display: flex; gap: 0.375rem; }
.q-opt { width: 1.6rem; height: 1.6rem; border-radius: 50%; background: #1e2a45; display: flex; align-items: center; justify-content: center; font-size: 0.7rem; font-weight: 700; color: #94a3b8; }
.q-opt.correct { background: #22c55e; color: #000; }
.q-stmt-count { font-size: 0.75rem; color: #64748b; }

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
  position: fixed; right: 0; top: 0; bottom: 0; width: 500px; max-width: 100vw;
  background: #141c2e; border-left: 1px solid #1e2a45;
  padding: 2rem; overflow-y: auto;
  display: flex; flex-direction: column; gap: 1rem;
}
.form-panel h2 { margin: 0; font-size: 1.1rem; }
.field { display: flex; flex-direction: column; gap: 0.375rem; }
.field label { font-size: 0.78rem; font-weight: 600; color: #94a3b8; }
.hint-inline { font-size: 0.72rem; font-weight: 400; color: #64748b; margin-left: 0.25rem; }
.text-area {
  padding: 0.65rem 0.875rem; border-radius: 8px; border: 1px solid #1e2a45;
  background: #0d1424; color: #f1f5f9; font-size: 0.875rem; resize: vertical; font-family: inherit;
  outline: none;
}
.text-area:focus { border-color: #4f8ef7; }

/* Type tabs */
.type-tabs { display: flex; gap: 0.375rem; }
.type-tab {
  flex: 1; padding: 0.5rem 0.25rem; border-radius: 8px;
  border: 1px solid #1e2a45; background: #0d1424;
  color: #64748b; font-size: 0.78rem; font-weight: 600; cursor: pointer;
  transition: all 0.15s; text-align: center;
}
.type-tab.active { border-color: #4f8ef7; background: rgba(79,142,247,0.12); color: #4f8ef7; }

/* Options */
.option-inputs { display: flex; flex-direction: column; gap: 0.625rem; }
.opt-input-row-wrap { display: flex; flex-direction: column; gap: 0.25rem; padding: 0.5rem; border: 1px solid #1e2a45; border-radius: 8px; background: #0a0f1e; }
.opt-input-row { display: flex; align-items: center; gap: 0.5rem; }
.opt-img-upload { margin-left: 2.5rem; }
.correct-dot {
  width: 2rem; height: 2rem; border-radius: 50%; border: 2px solid #1e2a45;
  background: #0d1424; color: #94a3b8; font-size: 0.75rem; font-weight: 700;
  cursor: pointer; flex-shrink: 0; display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.correct-dot.pgk { border-radius: 5px; }
.correct-dot.active { background: #22c55e; border-color: #22c55e; color: #000; }
.correct-dot.pgk.active { background: #a855f7; border-color: #a855f7; }
.opt-text-input {
  flex: 1; padding: 0.55rem 0.75rem; border-radius: 8px; border: 1px solid #1e2a45;
  background: #0d1424; color: #f1f5f9; font-size: 0.875rem; outline: none;
}
.opt-text-input:focus { border-color: #4f8ef7; }

/* Statements */
.stmt-inputs { display: flex; flex-direction: column; gap: 0.625rem; margin-bottom: 0.5rem; }
.stmt-input-block { display: flex; flex-direction: column; gap: 0.25rem; padding: 0.5rem; border: 1px solid #1e2a45; border-radius: 8px; background: #0a0f1e; }
.stmt-input-row { display: flex; align-items: center; gap: 0.5rem; }
.stmt-img-upload { margin-left: 2rem; }
.stmt-idx {
  width: 1.5rem; height: 1.5rem; border-radius: 50%; background: #1e2a45;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.72rem; font-weight: 700; color: #94a3b8; flex-shrink: 0;
}
.bs-toggle {
  padding: 0.35rem 0.625rem; border-radius: 6px; font-size: 0.75rem; font-weight: 700;
  border: 1px solid #1e2a45; background: transparent; cursor: pointer; flex-shrink: 0; transition: all 0.15s;
}
.bs-toggle.benar { border-color: #22c55e; background: rgba(34,197,94,0.15); color: #22c55e; }
.bs-toggle.salah { border-color: #ef4444; background: rgba(239,68,68,0.15); color: #ef4444; }
.remove-btn {
  padding: 0.3rem 0.5rem; border-radius: 6px; background: transparent;
  border: 1px solid #1e2a45; color: #64748b; cursor: pointer; font-size: 0.75rem;
  flex-shrink: 0; transition: all 0.15s;
}
.remove-btn:hover:not(:disabled) { border-color: #ef4444; color: #ef4444; }
.remove-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.btn-add-stmt {
  padding: 0.45rem 0.875rem; border-radius: 8px; border: 1px dashed #1e2a45;
  background: transparent; color: #64748b; font-size: 0.8rem; cursor: pointer; transition: all 0.15s;
}
.btn-add-stmt:hover:not(:disabled) { border-color: #4f8ef7; color: #4f8ef7; }
.btn-add-stmt:disabled { opacity: 0.4; cursor: not-allowed; }

.form-actions { display: flex; gap: 0.75rem; margin-top: auto; padding-top: 1rem; }
.btn-cancel {
  flex: 1; padding: 0.65rem; border-radius: 8px; border: 1px solid #1e2a45;
  background: transparent; color: #94a3b8; cursor: pointer; font-size: 0.875rem;
  transition: all 0.15s;
}
.btn-cancel:hover { border-color: #ef4444; color: #ef4444; }
</style>
