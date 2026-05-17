<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'
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
const eduLevelFilter = ref('')

const activeFilterCount = computed(() =>
  [topicFilter.value, difficultyFilter.value, typeFilter.value, eduLevelFilter.value].filter(Boolean).length
)

function clearFilters() {
  topicFilter.value = ''
  difficultyFilter.value = ''
  typeFilter.value = ''
  eduLevelFilter.value = ''
  fetchQuestions()
}

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
        education_level: (eduLevelFilter.value as 'sd' | 'smp' | 'sma') || undefined,
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
      <h1 class="page-title">Bank Soal</h1>
      <button class="btn-primary" @click="openCreate">+ Tambah Soal</button>
    </div>

    <div class="filter-toolbar">
      <div class="search-wrap">
        <svg class="search-icon" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd"/>
        </svg>
        <input v-model="searchText" class="search-input" placeholder="Cari soal…" @input="fetchQuestions" />
      </div>
      <select v-model="topicFilter" class="filter-select" :class="{ 'has-value': topicFilter }" @change="fetchQuestions">
        <option value="">Semua Topik</option>
        <option v-for="t in topics" :key="t.id" :value="t.id">{{ t.name }}</option>
      </select>
      <select v-model="difficultyFilter" class="filter-select" :class="{ 'has-value': difficultyFilter }" @change="fetchQuestions">
        <option value="">Semua Kesulitan</option>
        <option value="easy">Mudah</option>
        <option value="medium">Sedang</option>
        <option value="hard">Sulit</option>
      </select>
      <select v-model="typeFilter" class="filter-select" :class="{ 'has-value': typeFilter }" @change="fetchQuestions">
        <option value="">Semua Tipe</option>
        <option value="mcq">PG</option>
        <option value="multi_correct">PGK</option>
        <option value="true_false">B/S</option>
      </select>
      <select v-model="eduLevelFilter" class="filter-select" :class="{ 'has-value': eduLevelFilter }" @change="fetchQuestions">
        <option value="">Semua Jenjang</option>
        <option value="sd">SD</option>
        <option value="smp">SMP</option>
        <option value="sma">SMA</option>
      </select>
    </div>

    <div v-if="activeFilterCount > 0" class="filter-summary">
      <span v-if="topicFilter" class="f-chip">
        {{ topics.find(t => t.id === topicFilter)?.name }}
        <button @click="topicFilter = ''; fetchQuestions()">✕</button>
      </span>
      <span v-if="difficultyFilter" class="f-chip">
        {{ ({ easy: 'Mudah', medium: 'Sedang', hard: 'Sulit' } as Record<string,string>)[difficultyFilter] }}
        <button @click="difficultyFilter = ''; fetchQuestions()">✕</button>
      </span>
      <span v-if="typeFilter" class="f-chip">
        {{ ({ mcq: 'PG', multi_correct: 'PGK', true_false: 'B/S' } as Record<string,string>)[typeFilter] }}
        <button @click="typeFilter = ''; fetchQuestions()">✕</button>
      </span>
      <span v-if="eduLevelFilter" class="f-chip">
        {{ (eduLevelFilter as string).toUpperCase() }}
        <button @click="eduLevelFilter = ''; fetchQuestions()">✕</button>
      </span>
      <button class="clear-all-btn" @click="clearFilters">Hapus semua</button>
    </div>

    <p v-if="deleteError" class="error-msg">{{ deleteError }}</p>

    <div v-if="isLoading" class="loading">Loading…</div>

    <div v-else-if="questions.length === 0" class="empty-state">Belum ada soal. Tambahkan soal pertama Anda!</div>

    <div v-else class="question-list">
      <div v-for="q in questions" :key="q.id" class="question-card">
        <div class="q-header">
          <span
            class="q-type-badge"
            :style="{ color: typeColor[q.question_type], borderColor: typeColor[q.question_type] + '44', background: typeColor[q.question_type] + '18' }"
          >{{ typeLabel[q.question_type] ?? q.question_type }}</span>
          <span class="q-difficulty" :style="{ color: diffColor[q.difficulty] }">{{ q.difficulty }}</span>
          <span v-if="(q as any).education_level" class="q-edu-badge">{{ ((q as any).education_level as string).toUpperCase() }}</span>
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

    <!-- Centered modal dialog -->
    <Teleport to="body">
      <div v-if="showForm" class="modal-backdrop" @click.self="showForm = false">
        <div class="modal-dialog">
          <!-- Header -->
          <div class="modal-header">
            <h2 class="modal-title">{{ editTarget ? 'Edit Soal' : 'Soal Baru' }}</h2>
            <button class="modal-close" @click="showForm = false">×</button>
          </div>

          <!-- Scrollable body -->
          <div class="modal-body">

            <!-- Question type pills -->
            <div class="field">
              <label class="field-label">Tipe Soal</label>
              <div class="type-tabs">
                <button
                  v-for="t in [{ v: 'mcq', label: 'Pilihan Ganda' }, { v: 'multi_correct', label: 'PGK' }, { v: 'true_false', label: 'Benar / Salah' }]"
                  :key="t.v"
                  class="type-tab"
                  :class="{ active: form.question_type === t.v }"
                  @click="form.question_type = t.v as QuestionType; onTypeChange()"
                >{{ t.label }}</button>
              </div>
            </div>

            <div class="form-row">
              <div class="field">
                <label class="field-label">Topik</label>
                <select v-model="form.topic_id" class="modal-select">
                  <option value="">Pilih topik…</option>
                  <option v-for="t in topics" :key="t.id" :value="t.id">{{ t.name }}</option>
                </select>
              </div>

              <div class="field">
                <label class="field-label">Kesulitan</label>
                <select v-model="form.difficulty" class="modal-select">
                  <option value="easy">Mudah</option>
                  <option value="medium">Sedang</option>
                  <option value="hard">Sulit</option>
                </select>
              </div>

              <div class="field">
                <label class="field-label">Jenjang</label>
                <select v-model="form.education_level" class="modal-select">
                  <option value="">Semua</option>
                  <option value="sd">SD</option>
                  <option value="smp">SMP</option>
                  <option value="sma">SMA</option>
                </select>
              </div>
            </div>

            <div class="section-divider" />

            <div class="field">
              <label class="field-label">Teks Soal</label>
              <RichTextEditor v-model="form.text" />
            </div>

            <!-- MCQ / PGK options -->
            <div v-if="!isTrueFalse" class="field">
              <label class="field-label">
                Pilihan Jawaban
                <span class="hint-inline">{{ isMultiCorrect ? '— tandai semua yang benar' : '— klik label untuk menandai benar' }}</span>
              </label>
              <div class="option-inputs">
                <div v-for="(opt, i) in form.options" :key="opt.label" class="opt-card">
                  <div class="opt-input-row">
                    <button
                      class="correct-dot"
                      :class="{ active: opt.is_correct, pgk: isMultiCorrect }"
                      @click="isMultiCorrect ? toggleCorrectPGK(i) : setCorrectMCQ(i)"
                    >{{ opt.is_correct && isMultiCorrect ? '✓' : opt.label }}</button>
                    <RichTextEditor v-model="opt.text" class="opt-text-input" />
                  </div>
                </div>
                <button type="button" class="add-option-btn" @click="addOption()">+ Tambah Pilihan</button>
              </div>
            </div>

            <!-- B/S statements -->
            <div v-if="isTrueFalse" class="field">
              <label class="field-label">Pernyataan ({{ form.statements.length }}/6)</label>
              <div class="stmt-inputs">
                <div v-for="(stmt, i) in form.statements" :key="i" class="stmt-card">
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
                </div>
              </div>
              <button class="btn-add-stmt" :disabled="form.statements.length >= 6" @click="addStatement">
                + Tambah Pernyataan
              </button>
            </div>

            <div class="section-divider" />

            <div class="field">
              <label class="field-label">Penjelasan (opsional)</label>
              <RichTextEditor v-model="form.explanation" />
            </div>

            <p v-if="formError" class="error-msg">{{ formError }}</p>
          </div>

          <!-- Footer -->
          <div class="modal-footer">
            <button class="btn-cancel" @click="showForm = false">Batal</button>
            <button class="btn-primary" :disabled="isSaving" @click="saveQuestion">
              {{ isSaving ? 'Menyimpan…' : 'Simpan Soal' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
.page-title { margin: 0; font-size: 1.5rem; font-weight: 800; }
.loading { color: var(--text-muted); }

/* ── Filter toolbar ─────────────────────────────────────────────── */
.filter-toolbar { display: flex; gap: 0.625rem; align-items: center; margin-bottom: 0.875rem; flex-wrap: wrap; }
.search-wrap { flex: 1; min-width: 180px; position: relative; }
.search-icon { position: absolute; left: 0.75rem; top: 50%; transform: translateY(-50%); color: var(--text-muted); width: 15px; height: 15px; pointer-events: none; }
.search-input { width: 100%; padding: 0.55rem 0.875rem 0.55rem 2.25rem; border-radius: 8px; border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-primary); font-size: 0.875rem; font-family: inherit; outline: none; transition: border-color 0.15s; }
.search-input:focus { border-color: var(--accent); }
.search-input::placeholder { color: var(--text-muted); }
.filter-select { padding: 0.55rem 2rem 0.55rem 0.75rem; border-radius: 8px; border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-muted); font-size: 0.85rem; font-family: inherit; cursor: pointer; outline: none; appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 20 20' fill='%2394a3b8'%3E%3Cpath fill-rule='evenodd' d='M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z' clip-rule='evenodd'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 0.65rem center; transition: border-color 0.15s, color 0.15s; }
.filter-select:focus { border-color: var(--accent); color: var(--text-primary); }
.filter-select.has-value { border-color: rgba(79,142,247,0.4); color: var(--accent); }
.filter-summary { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 1rem; flex-wrap: wrap; }
.f-chip { display: inline-flex; align-items: center; gap: 0.3rem; padding: 0.2rem 0.5rem 0.2rem 0.625rem; border-radius: 5px; font-size: 0.72rem; font-weight: 600; background: rgba(79,142,247,0.1); color: var(--accent); border: 1px solid rgba(79,142,247,0.2); }
.f-chip button { background: none; border: none; cursor: pointer; color: var(--accent); opacity: 0.5; font-size: 0.65rem; line-height: 1; padding: 0; font-family: inherit; transition: opacity 0.1s; }
.f-chip button:hover { opacity: 1; }
.clear-all-btn { margin-left: auto; background: none; border: none; color: var(--text-muted); font-size: 0.72rem; font-weight: 500; cursor: pointer; font-family: inherit; transition: color 0.12s; padding: 0; }
.clear-all-btn:hover { color: var(--danger); }

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
.q-edu-badge {
  font-size: 0.65rem; font-weight: 700; padding: 0.15rem 0.45rem;
  border-radius: 3px; background: color-mix(in srgb, var(--accent) 15%, transparent);
  color: var(--accent); letter-spacing: 0.03em;
}
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

/* Modal */
@keyframes modal-in {
  from { opacity: 0; transform: scale(0.95) translateY(8px); }
  to   { opacity: 1; transform: scale(1) translateY(0); }
}

.modal-backdrop {
  position: fixed; inset: 0; z-index: 200;
  background: rgba(0,0,0,0.55);
  backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center;
  padding: 1rem;
}

.modal-dialog {
  width: 100%; max-width: 680px; max-height: 88vh;
  background: var(--bg-surface);
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(0,0,0,0.4), 0 4px 16px rgba(0,0,0,0.2);
  display: flex; flex-direction: column;
  animation: modal-in 0.25s cubic-bezier(0.22, 1, 0.36, 1) both;
  overflow: hidden;
}

.modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.modal-title { margin: 0; font-size: 1.05rem; font-weight: 700; color: var(--text-heading); }

.modal-close {
  width: 2rem; height: 2rem; border-radius: 8px;
  border: 1px solid var(--border); background: transparent;
  color: var(--text-muted); font-size: 1.2rem; line-height: 1;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.modal-close:hover { border-color: var(--danger); color: var(--danger); background: color-mix(in srgb, var(--danger) 10%, transparent); }

.modal-body {
  flex: 1; overflow-y: auto; padding: 1.5rem;
  display: flex; flex-direction: column; gap: 1.25rem;
}

.modal-footer {
  display: flex; gap: 0.75rem; justify-content: flex-end;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
}

.section-divider {
  height: 1px; background: var(--border);
  margin: 0.25rem 0;
}

.form-row { display: flex; gap: 0.75rem; }
.form-row .field { flex: 1; }

.field { display: flex; flex-direction: column; gap: 0.375rem; }
.field-label { font-size: 0.72rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; }
.hint-inline { font-size: 0.72rem; font-weight: 400; color: var(--text-muted); margin-left: 0.25rem; }

/* Type tabs — pill style */
.type-tabs {
  display: flex; gap: 0; background: var(--bg-input);
  border: 1px solid var(--border); border-radius: 10px; padding: 3px;
}
.type-tab {
  flex: 1; padding: 0.45rem 0.5rem; border-radius: 8px;
  border: none; background: transparent;
  color: var(--text-muted); font-size: 0.78rem; font-weight: 600; cursor: pointer;
  transition: all 0.18s; text-align: center;
}
.type-tab.active {
  background: var(--accent); color: #fff;
  box-shadow: 0 1px 4px rgba(79,142,247,0.35);
}
.type-tab:not(.active):hover { color: var(--text-primary); }

/* Modal select */
.modal-select {
  width: 100%; padding: 0.55rem 0.75rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-input); color: var(--text-primary);
  font-size: 0.875rem; cursor: pointer; outline: none; transition: border-color 0.15s;
}
.modal-select:focus { border-color: var(--accent); }

/* Options */
.option-inputs { display: flex; flex-direction: column; gap: 0.5rem; }
.opt-card { display: flex; flex-direction: column; gap: 0.25rem; padding: 0.625rem 0.75rem; border: 1px solid var(--border); border-radius: 10px; background: var(--bg-input); transition: border-color 0.15s; }
.opt-card:focus-within { border-color: color-mix(in srgb, var(--accent) 50%, var(--border)); }
.opt-input-row { display: flex; align-items: center; gap: 0.625rem; }
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
.stmt-inputs { display: flex; flex-direction: column; gap: 0.5rem; margin-bottom: 0.5rem; }
.stmt-card { display: flex; flex-direction: column; gap: 0.25rem; padding: 0.625rem 0.75rem; border: 1px solid var(--border); border-radius: 10px; background: var(--bg-input); }
.stmt-input-row { display: flex; align-items: center; gap: 0.5rem; }
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

.btn-cancel {
  padding: 0.55rem 1.25rem; border-radius: 8px; border: 1px solid var(--border);
  background: transparent; color: var(--text-muted); cursor: pointer; font-size: 0.875rem;
  font-weight: 600; transition: all 0.15s;
}
.btn-cancel:hover { border-color: var(--danger); color: var(--danger); }

.add-option-btn {
  margin-top: 0.25rem; padding: 0.5rem; border-radius: 8px;
  border: 1px dashed var(--border); background: transparent;
  color: var(--text-muted); font-size: 0.8rem; cursor: pointer; width: 100%; transition: all 0.15s;
}
.add-option-btn:hover { border-color: var(--accent); color: var(--accent); }

@media (max-width: 640px) {
  .modal-dialog { max-height: 95vh; border-radius: 12px; }
  .form-row { flex-direction: column; }
}
</style>
