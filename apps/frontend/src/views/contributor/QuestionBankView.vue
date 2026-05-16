<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'
import ImageUpload from '@/components/ImageUpload.vue'
import RichTextEditor from '@/components/editor/RichTextEditor.vue'
import RichTextViewer from '@/components/editor/RichTextViewer.vue'

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
  education_level: string
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
  education_level: '',
  options: 'ABCD'.split('').map(l => ({ label: l, text: '', is_correct: false, image_url: null })),
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

function openEdit(q: Question) {
  editTarget.value = q
  form.value = {
    question_type: q.question_type as QuestionType,
    topic_id: q.topic_id,
    text: q.text,
    explanation: q.explanation ?? '',
    image_url: q.image_url ?? null,
    difficulty: q.difficulty as 'easy' | 'medium' | 'hard',
    education_level: (q as any).education_level ?? '',
    options: q.question_type !== 'true_false'
      ? q.options.map(o => ({ label: o.label, text: o.text, is_correct: o.is_correct, image_url: o.image_url ?? null }))
      : 'ABCD'.split('').map(l => ({ label: l, text: '', is_correct: false, image_url: null })),
    statements: q.question_type === 'true_false'
      ? q.statements.map(s => ({ text: s.text, is_correct: s.is_correct, image_url: s.image_url ?? null }))
      : [{ text: '', is_correct: true, image_url: null }, { text: '', is_correct: false, image_url: null }],
  }
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

function addOption() {
  const labels = "ABCDEFGHIJ"
  form.value.options.push({
    label: labels[form.value.options.length] ?? String(form.value.options.length + 1),
    text: "",
    is_correct: false,
    image_url: null,
  })
}

async function saveQuestion() {
  formError.value = ''
  if (!form.value.topic_id) { formError.value = 'Please select a topic.'; return }
  const plainText = form.value.text.replace(/<[^>]*>/g, '').trim()
  if (!plainText) { formError.value = 'Question text is required.'; return }

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
      education_level: (form.value.education_level || undefined) as 'sd' | 'smp' | 'sma' | undefined,
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
      education_level: (form.value.education_level || undefined) as 'sd' | 'smp' | 'sma' | undefined,
      options: form.value.options.map(o => ({ label: o.label, text: o.text, is_correct: o.is_correct, image_url: o.image_url ?? undefined })),
    }
  } else {
    const stmts = form.value.statements
    if (stmts.length < 2) { formError.value = 'Add at least 2 statements.'; return }
    if (stmts.some(s => !s.text.replace(/<[^>]*>/g, '').trim())) { formError.value = 'All statement texts are required.'; return }
    body = {
      question_type: 'true_false',
      topic_id: form.value.topic_id,
      text: form.value.text,
      explanation: form.value.explanation || undefined,
      image_url: form.value.image_url ?? undefined,
      difficulty: form.value.difficulty,
      education_level: (form.value.education_level || undefined) as 'sd' | 'smp' | 'sma' | undefined,
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

const diffColor: Record<string, string> = { easy: 'var(--success)', medium: 'var(--warning)', hard: 'var(--danger)' }
const typeLabel: Record<string, string> = { mcq: 'PG', multi_correct: 'PGK', true_false: 'B/S' }
const typeColor: Record<string, string> = {
  mcq: 'var(--accent)',
  multi_correct: 'var(--accent)',
  true_false: 'var(--success)',
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
            <button class="icon-btn" @click="openEdit(q)">✏️</button>
            <button class="icon-btn" @click="deleteQuestion(q.id)">🗑</button>
          </div>
        </div>
        <div class="q-text"><RichTextViewer :html="q.text" /></div>
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
          <label>Education Level</label>
          <select v-model="form.education_level" class="filter-select">
            <option value="">All Levels</option>
            <option value="sd">SD</option>
            <option value="smp">SMP</option>
            <option value="sma">SMA</option>
          </select>
        </div>

        <div class="field">
          <label>Question Text</label>
          <RichTextEditor v-model="form.text" />
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
                <RichTextEditor v-model="opt.text" class="opt-text-input" />
              </div>
              <ImageUpload v-model="opt.image_url" class="opt-img-upload" />
            </div>
            <button
              v-if="!isTrueFalse"
              type="button"
              class="add-option-btn"
              @click="addOption()"
            >+ Add option</button>
          </div>
        </div>

        <!-- B/S statements -->
        <div v-if="isTrueFalse" class="field">
          <label>Pernyataan ({{ form.statements.length }}/6)</label>
          <div class="stmt-inputs">
            <div v-for="(stmt, i) in form.statements" :key="i" class="stmt-input-block">
              <div class="stmt-input-row">
                <span class="stmt-idx">{{ i + 1 }}</span>
                <RichTextEditor v-model="stmt.text" class="opt-text-input" />
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
          <RichTextEditor v-model="form.explanation" />
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
.loading { color: var(--text-muted); }

.filters { display: flex; gap: 0.75rem; margin-bottom: 1.5rem; flex-wrap: wrap; }
.search-input {
  flex: 1; min-width: 200px;
  padding: 0.55rem 0.875rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-primary); font-size: 0.875rem;
  outline: none;
}
.search-input:focus { border-color: var(--accent); }
.filter-select {
  padding: 0.55rem 0.875rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-primary); font-size: 0.875rem;
  cursor: pointer; outline: none;
}

.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: var(--danger-bg); color: var(--danger-text); font-size: 0.825rem; margin-bottom: 1rem; }
.empty-state { color: var(--text-muted); text-align: center; padding: 2rem; background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; }

.question-list { display: flex; flex-direction: column; gap: 0.625rem; }
.question-card { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 10px; padding: 1rem; }
.q-header { display: flex; align-items: center; gap: 0.625rem; margin-bottom: 0.5rem; }
.q-type-badge {
  font-size: 0.68rem; font-weight: 800; padding: 0.2rem 0.5rem;
  border-radius: 4px; border: 1px solid; text-transform: uppercase; letter-spacing: 0.03em;
}
.q-difficulty { font-size: 0.75rem; font-weight: 700; text-transform: capitalize; }
.q-topic { font-size: 0.75rem; color: var(--text-muted); }
.q-actions { margin-left: auto; }
.icon-btn { background: transparent; border: none; cursor: pointer; font-size: 0.9rem; color: var(--text-muted); padding: 0.2rem; }
.icon-btn:hover { color: var(--danger); }
.q-text { margin: 0 0 0.625rem; font-size: 0.9rem; color: var(--text-secondary); line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.q-options { display: flex; gap: 0.375rem; }
.q-opt { width: 1.6rem; height: 1.6rem; border-radius: 50%; background: var(--border); display: flex; align-items: center; justify-content: center; font-size: 0.7rem; font-weight: 700; color: var(--text-muted); }
.q-opt.correct { background: var(--success); color: #000; }
.q-stmt-count { font-size: 0.75rem; color: var(--text-muted); }

.btn-primary {
  padding: 0.55rem 1.25rem; border-radius: 8px; border: none;
  background: var(--accent); color: var(--text-on-accent); font-size: 0.875rem; font-weight: 600;
  cursor: pointer; transition: background 0.15s;
}
.btn-primary:hover { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

/* Form panel */
.form-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 50; }
.form-panel {
  position: fixed; right: 0; top: 0; bottom: 0; width: 500px; max-width: 100vw;
  background: var(--bg-surface); border-left: 1px solid var(--border);
  padding: 2rem; overflow-y: auto;
  display: flex; flex-direction: column; gap: 1rem;
}
.form-panel h2 { margin: 0; font-size: 1.1rem; }
.field { display: flex; flex-direction: column; gap: 0.375rem; }
.field label { font-size: 0.78rem; font-weight: 600; color: var(--text-muted); }
.hint-inline { font-size: 0.72rem; font-weight: 400; color: var(--text-muted); margin-left: 0.25rem; }
.text-area {
  padding: 0.65rem 0.875rem; border-radius: 8px; border: 1px solid var(--border);
  background: var(--bg-input); color: var(--text-primary); font-size: 0.875rem; resize: vertical; font-family: inherit;
  outline: none;
}
.text-area:focus { border-color: var(--accent); }

/* Type tabs */
.type-tabs { display: flex; gap: 0.375rem; }
.type-tab {
  flex: 1; padding: 0.5rem 0.25rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-input);
  color: var(--text-muted); font-size: 0.78rem; font-weight: 600; cursor: pointer;
  transition: all 0.15s; text-align: center;
}
.type-tab.active { border-color: var(--accent); background: rgba(79,142,247,0.12); color: var(--accent); }

/* Options */
.option-inputs { display: flex; flex-direction: column; gap: 0.625rem; }
.opt-input-row-wrap { display: flex; flex-direction: column; gap: 0.25rem; padding: 0.5rem; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-input); }
.opt-input-row { display: flex; align-items: center; gap: 0.5rem; }
.opt-img-upload { margin-left: 2.5rem; }
.correct-dot {
  width: 2rem; height: 2rem; border-radius: 50%; border: 2px solid var(--border);
  background: var(--bg-input); color: var(--text-muted); font-size: 0.75rem; font-weight: 700;
  cursor: pointer; flex-shrink: 0; display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.correct-dot.pgk { border-radius: 5px; }
.correct-dot.active { background: var(--success); border-color: var(--success); color: #000; }
.correct-dot.pgk.active { background: var(--accent); border-color: var(--accent); }
.opt-text-input {
  flex: 1; padding: 0.55rem 0.75rem; border-radius: 8px; border: 1px solid var(--border);
  background: var(--bg-input); color: var(--text-primary); font-size: 0.875rem; outline: none;
}
.opt-text-input:focus { border-color: var(--accent); }

/* Statements */
.stmt-inputs { display: flex; flex-direction: column; gap: 0.625rem; margin-bottom: 0.5rem; }
.stmt-input-block { display: flex; flex-direction: column; gap: 0.25rem; padding: 0.5rem; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-input); }
.stmt-input-row { display: flex; align-items: center; gap: 0.5rem; }
.stmt-img-upload { margin-left: 2rem; }
.stmt-idx {
  width: 1.5rem; height: 1.5rem; border-radius: 50%; background: var(--border);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.72rem; font-weight: 700; color: var(--text-muted); flex-shrink: 0;
}
.bs-toggle {
  padding: 0.35rem 0.625rem; border-radius: 6px; font-size: 0.75rem; font-weight: 700;
  border: 1px solid var(--border); background: transparent; cursor: pointer; flex-shrink: 0; transition: all 0.15s;
}
.bs-toggle.benar { border-color: var(--success); background: rgba(34,197,94,0.15); color: var(--success); }
.bs-toggle.salah { border-color: var(--danger); background: rgba(239,68,68,0.15); color: var(--danger); }
.remove-btn {
  padding: 0.3rem 0.5rem; border-radius: 6px; background: transparent;
  border: 1px solid var(--border); color: var(--text-muted); cursor: pointer; font-size: 0.75rem;
  flex-shrink: 0; transition: all 0.15s;
}
.remove-btn:hover:not(:disabled) { border-color: var(--danger); color: var(--danger); }
.remove-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.btn-add-stmt {
  padding: 0.45rem 0.875rem; border-radius: 8px; border: 1px dashed var(--border);
  background: transparent; color: var(--text-muted); font-size: 0.8rem; cursor: pointer; transition: all 0.15s;
}
.btn-add-stmt:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
.btn-add-stmt:disabled { opacity: 0.4; cursor: not-allowed; }

.form-actions { display: flex; gap: 0.75rem; margin-top: auto; padding-top: 1rem; }
.btn-cancel {
  flex: 1; padding: 0.65rem; border-radius: 8px; border: 1px solid var(--border);
  background: transparent; color: var(--text-muted); cursor: pointer; font-size: 0.875rem;
  transition: all 0.15s;
}
.btn-cancel:hover { border-color: var(--danger); color: var(--danger); }
</style>
@media (max-width: 768px) { .modal { width: 100%; max-width: 100%; } .form-row { flex-direction: column; } }

.add-option-btn { margin-top: 0.25rem; padding: 0.5rem; border-radius: 8px; border: 1px dashed var(--border); background: transparent; color: var(--text-muted); font-size: 0.8rem; cursor: pointer; width: 100%; transition: all 0.15s; }
.add-option-btn:hover { border-color: var(--accent); color: var(--accent); }
