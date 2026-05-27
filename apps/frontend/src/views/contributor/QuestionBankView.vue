<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

const auth = useAuthStore()
import RichTextEditor from '@/components/editor/RichTextEditor.vue'
import RichTextViewer from '@/components/editor/RichTextViewer.vue'
import QuestionPreviewModal from '@/components/QuestionPreviewModal.vue'

type Question = components['schemas']['QuestionDetailResponse']
type Topic = components['schemas']['TopicResponse']
type CreateReq = components['schemas']['CreateQuestionRequest']
type UpdateReq = components['schemas']['UpdateQuestionRequest']
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
  [difficultyFilter.value, typeFilter.value, eduLevelFilter.value].filter(Boolean).length
)

function clearFilters() {
  difficultyFilter.value = ''
  typeFilter.value = ''
  eduLevelFilter.value = ''
  fetchQuestions()
}

function selectTopic(topicId: string) {
  if (topicFilter.value === topicId) {
    topicFilter.value = '' // deselect
  } else {
    topicFilter.value = topicId
  }
  fetchQuestions()
}

// Topic management
const showTopicCreate = ref(false)
const newTopicName = ref('')
const newTopicSaving = ref(false)
const newTopicError = ref('')
const editingTopicId = ref('')
const editingTopicName = ref('')
const topicEditError = ref('')
const topicEditSaving = ref(false)

// Delete confirmation modal
const deleteModal = ref<{ id: string; name: string; questionCount: number } | null>(null)

// Preview modal
const previewQuestion = ref<Question | null>(null)

async function createTopic() {
  const name = newTopicName.value.trim()
  if (!name) return
  newTopicSaving.value = true; newTopicError.value = ''
  const { error } = await client.POST('/topics', { body: { name } })
  if (error) { newTopicError.value = 'Nama sudah ada atau tidak valid.'; newTopicSaving.value = false; return }
  newTopicSaving.value = false; showTopicCreate.value = false; newTopicName.value = ''
  await fetchTopics()
}

function startEditTopic(t: Topic) {
  editingTopicId.value = t.id
  editingTopicName.value = t.name
  topicEditError.value = ''
}

function cancelEditTopic() {
  editingTopicId.value = ''
  editingTopicName.value = ''
  topicEditError.value = ''
}

async function saveEditTopic(id: string) {
  const name = editingTopicName.value.trim()
  if (!name) return
  topicEditSaving.value = true; topicEditError.value = ''
  const { error } = await client.PATCH('/topics/{topicId}', { params: { path: { topicId: id } }, body: { name } })
  if (error) { topicEditError.value = 'Gagal atau nama sudah ada.'; topicEditSaving.value = false; return }
  topicEditSaving.value = false; cancelEditTopic()
  await fetchTopics()
}

function openDeleteModal(t: Topic) {
  // Count questions for this topic
  const count = questions.value.filter(q => q.topic_id === t.id).length
  deleteModal.value = { id: t.id, name: t.name, questionCount: count }
}

async function confirmDeleteTopic() {
  if (!deleteModal.value) return
  const { error } = await client.DELETE('/topics/{topicId}', { params: { path: { topicId: deleteModal.value.id } } })
  if (error) { deleteModal.value = null; return }
  if (topicFilter.value === deleteModal.value.id) topicFilter.value = ''
  deleteModal.value = null
  await fetchTopics()
  await fetchQuestions()
}

function getTopicDotColor(index: number): string {
  const colors = ['#6c5ce7','#00d2a0','#ff9f43','#ff5252','#54a0ff','#a29bfe','#2ed573','#ffa502','#e056a0','#48dbfb','#feca57','#ff6b6b','#1dd1a1','#c8d6e5']
  return colors[index % colors.length]
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

async function fetchTopics() {
  const { data } = await client.GET('/topics')
  if (data) topics.value = data.data
}

onMounted(async () => {
  await fetchTopics()
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

async function saveOrFail(fn: () => Promise<any>) {
  isSaving.value = true
  try {
    const result = await fn()
    if (result.error) {
      const msg = (result.error as any)?.error?.message || (result.error as any)?.message
      formError.value = msg ? String(msg) : 'Failed to save question.'
      return
    }
    showForm.value = false
    await fetchQuestions()
  } catch {
    formError.value = 'Failed to save question.'
  } finally {
    isSaving.value = false
  }
}

async function saveQuestion() {
  formError.value = ''
  if (!form.value.topic_id) { formError.value = 'Please select a topic.'; return }
  const plainText = form.value.text.replace(/<[^>]*>/g, '').trim()
  if (!plainText) { formError.value = 'Question text is required.'; return }

  const eduLevel = (form.value.education_level || undefined) as 'sd' | 'smp' | 'sma' | undefined

  if (form.value.question_type === 'mcq') {
    const correctCount = form.value.options.filter((o) => o.is_correct).length
    if (correctCount !== 1) { formError.value = 'Exactly one option must be marked correct.'; return }
    const opts = form.value.options.map(o => ({ label: o.label, text: o.text, is_correct: o.is_correct, image_url: o.image_url ?? undefined }))
    if (editTarget.value) {
      const body: UpdateReq = {
        topic_id: form.value.topic_id,
        text: form.value.text,
        explanation: form.value.explanation || undefined,
        image_url: form.value.image_url ?? undefined,
        difficulty: form.value.difficulty,
        education_level: eduLevel,
        options: opts,
      }
      await saveOrFail(() => client.PATCH('/questions/{questionId}', { params: { path: { questionId: editTarget.value!.id } }, body }))
      return
    }
    const body: CreateReq = {
      question_type: 'mcq',
      topic_id: form.value.topic_id,
      text: form.value.text,
      explanation: form.value.explanation || undefined,
      image_url: form.value.image_url ?? undefined,
      difficulty: form.value.difficulty,
      education_level: eduLevel,
      options: opts,
    }
    await saveOrFail(() => client.POST('/questions', { body }))
    return
  }

  if (form.value.question_type === 'multi_correct') {
    const correctCount = form.value.options.filter((o) => o.is_correct).length
    if (correctCount < 1) { formError.value = 'At least one option must be marked correct for PGK.'; return }
    const opts = form.value.options.map(o => ({ label: o.label, text: o.text, is_correct: o.is_correct, image_url: o.image_url ?? undefined }))
    if (editTarget.value) {
      const body: UpdateReq = {
        topic_id: form.value.topic_id,
        text: form.value.text,
        explanation: form.value.explanation || undefined,
        image_url: form.value.image_url ?? undefined,
        difficulty: form.value.difficulty,
        education_level: eduLevel,
        options: opts,
      }
      await saveOrFail(() => client.PATCH('/questions/{questionId}', { params: { path: { questionId: editTarget.value!.id } }, body }))
      return
    }
    const body: CreateReq = {
      question_type: 'multi_correct',
      topic_id: form.value.topic_id,
      text: form.value.text,
      explanation: form.value.explanation || undefined,
      image_url: form.value.image_url ?? undefined,
      difficulty: form.value.difficulty,
      education_level: eduLevel,
      options: opts,
    }
    await saveOrFail(() => client.POST('/questions', { body }))
    return
  }

  // true_false
  {
    const stmts = form.value.statements
    if (stmts.length < 2) { formError.value = 'Add at least 2 statements.'; return }
    if (stmts.some(s => !s.text.replace(/<[^>]*>/g, '').trim())) { formError.value = 'All statement texts are required.'; return }
    if (editTarget.value) {
      const body: UpdateReq = {
        topic_id: form.value.topic_id,
        text: form.value.text,
        explanation: form.value.explanation || undefined,
        image_url: form.value.image_url ?? undefined,
        difficulty: form.value.difficulty,
        education_level: eduLevel,
        statements: stmts.map((s, i) => ({ text: s.text, is_correct: s.is_correct, position: i, image_url: s.image_url ?? undefined })),
      }
      await saveOrFail(() => client.PATCH('/questions/{questionId}', { params: { path: { questionId: editTarget.value!.id } }, body }))
      return
    }
    const body: CreateReq = {
      question_type: 'true_false',
      topic_id: form.value.topic_id,
      text: form.value.text,
      explanation: form.value.explanation || undefined,
      image_url: form.value.image_url ?? undefined,
      difficulty: form.value.difficulty,
      education_level: eduLevel,
      statements: stmts.map((s, i) => ({ text: s.text, is_correct: s.is_correct, position: i, image_url: s.image_url ?? undefined })),
    }
    await saveOrFail(() => client.POST('/questions', { body }))
  }
}

async function deleteQuestion(id: string) {
  deleteError.value = ''
  if (!confirm('Delete this question?')) return
  const { error } = await client.DELETE('/questions/{questionId}', { params: { path: { questionId: id } } })
  if (error) { deleteError.value = 'Cannot delete — question may be used in a published test.'; return }
  await fetchQuestions()
}

const typeLabel: Record<string, string> = { mcq: 'PG', multi_correct: 'PGK', true_false: 'B/S' }
const typeColor: Record<string, string> = {
  mcq: 'var(--accent)',
  multi_correct: 'var(--accent)',
  true_false: 'var(--success)',
}
</script>

<template>
  <div class="qb-page">
    <div class="page-header">
      <h1 class="page-title">Bank Soal</h1>
      <button class="btn-primary" @click="openCreate">+ Tambah Soal</button>
    </div>

    <div class="qb-layout">
      <!-- ═══════════════ TOPIC SIDEBAR ═══════════════ -->
      <aside class="qb-sidebar">
        <div class="sb-head">
          <span class="sb-title">Topik <span class="sb-count">{{ topics.length }}</span></span>
          <button class="sb-add-btn" @click="showTopicCreate = !showTopicCreate" :title="showTopicCreate ? 'Batal' : 'Tambah Topik'">
            {{ showTopicCreate ? '×' : '+' }}
          </button>
        </div>

        <!-- inline create -->
        <div v-if="showTopicCreate" class="sb-create">
          <input
            v-model="newTopicName"
            type="text"
            class="sb-create-input"
            placeholder="Nama topik baru…"
            maxlength="100"
            @keyup.enter="createTopic"
          />
          <p v-if="newTopicError" class="sb-err">{{ newTopicError }}</p>
          <div class="sb-create-actions">
            <button class="sb-create-cancel" @click="showTopicCreate = false; newTopicError = ''">Batal</button>
            <button class="sb-create-save" :disabled="newTopicSaving || !newTopicName.trim()" @click="createTopic">
              {{ newTopicSaving ? '...' : 'Simpan' }}
            </button>
          </div>
        </div>

        <!-- topic list -->
        <div class="sb-list">
          <button
            class="sb-item sb-item--all"
            :class="{ 'sb-item--active': !topicFilter }"
            @click="topicFilter = ''; fetchQuestions()"
          >
            <span class="sb-item-name">Semua Topik</span>
            <span class="sb-item-num">{{ questions.length }}</span>
          </button>

          <template v-for="(t, i) in topics" :key="t.id">
            <!-- editing row -->
            <div v-if="editingTopicId === t.id" class="sb-edit-row">
              <input
                v-model="editingTopicName"
                type="text"
                class="sb-edit-input"
                maxlength="100"
                @keyup.enter="saveEditTopic(t.id)"
                @keyup.escape="cancelEditTopic"
              />
              <p v-if="topicEditError" class="sb-err">{{ topicEditError }}</p>
              <div class="sb-edit-actions">
                <button class="sb-edit-cancel" @click="cancelEditTopic">Batal</button>
                <button class="sb-edit-save" :disabled="topicEditSaving" @click="saveEditTopic(t.id)">Simpan</button>
              </div>
            </div>

            <!-- normal row -->
            <button
              v-else
              class="sb-item"
              :class="{ 'sb-item--active': topicFilter === t.id }"
              @click="selectTopic(t.id)"
            >
              <span class="sb-dot" :style="{ background: getTopicDotColor(i) }" />
              <span class="sb-item-name">{{ t.name }}</span>
              <span class="sb-item-actions">
                <button class="sb-act-btn" title="Edit" @click.stop="startEditTopic(t)">✎</button>
                <button class="sb-act-btn sb-act-btn--del" title="Hapus" @click.stop="openDeleteModal(t)">×</button>
              </span>
            </button>
          </template>
        </div>
      </aside>

      <!-- ═══════════════ MAIN: Question List ═══════════════ -->
      <main class="qb-main">
        <!-- Filter toolbar — no topic dropdown -->
        <div class="filter-toolbar">
          <div class="search-wrap">
            <svg class="search-icon" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd"/>
            </svg>
            <input v-model="searchText" class="search-input" placeholder="Cari soal…" @input="fetchQuestions" />
          </div>
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

        <!-- Filter summary + active topic indicator -->
        <div v-if="activeFilterCount > 0 || topicFilter" class="filter-summary">
          <span v-if="topicFilter" class="f-chip f-chip--topic">
            <span class="f-chip-dot" :style="{ background: getTopicDotColor(topics.findIndex(t => t.id === topicFilter)) }" />
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
          <button class="clear-all-btn" @click="clearFilters(); topicFilter = ''; fetchQuestions()">Hapus semua</button>
        </div>

        <p v-if="deleteError" class="error-msg">{{ deleteError }}</p>

        <div v-if="isLoading" class="loading">Loading…</div>

        <div v-else-if="questions.length === 0" class="empty-state">Belum ada soal. Tambahkan soal pertama Anda!</div>

        <div v-else class="question-list">
          <div v-for="q in questions" :key="q.id" class="question-card">
            <div class="q-top">
              <span class="q-type" :style="{ color: typeColor[q.question_type], borderColor: typeColor[q.question_type] }">{{ typeLabel[q.question_type] ?? q.question_type }}</span>
              <span class="q-topic-name">
                <span class="q-topic-dot" :style="{ background: getTopicDotColor(topics.findIndex(t => t.id === q.topic_id)) }" />
                {{ topics.find(t => t.id === q.topic_id)?.name ?? '—' }}
              </span>
              <span v-if="(q as any).education_level" class="q-edu">{{ ((q as any).education_level as string).toUpperCase() }}</span>
              <span class="q-diff" :class="'q-diff--' + q.difficulty">{{ q.difficulty }}</span>
              <span class="q-spacer" />
              <span v-if="(q as any).contributor_name" class="q-owner">oleh: {{ (q as any).contributor_name }}</span>
              <span v-if="(q as any).usage && auth.user?.id === q.contributor_id" class="q-usage">
                {{ (q as any).usage.own_test_count }} tes kamu · {{ (q as any).usage.other_test_count }} kontributor lain
              </span>
              <button class="q-act q-act--preview" @click="previewQuestion = q">Preview</button>
              <button v-if="auth.user?.id === q.contributor_id || auth.role === 'admin'" class="q-act" @click="openEdit(q)">Edit</button>
              <button v-if="auth.user?.id === q.contributor_id || auth.role === 'admin'" class="q-act q-act--del" @click="deleteQuestion(q.id)">Hapus</button>
            </div>
            <div class="q-text"><RichTextViewer :html="q.text" /></div>
            <div v-if="q.question_type !== 'true_false'" class="q-options">
              <span v-for="o in q.options" :key="o.id" class="q-opt" :class="{ correct: o.is_correct }">{{ o.label }}</span>
            </div>
            <div v-else class="q-stmt-count">{{ q.statements.length }} pernyataan</div>
          </div>
        </div>
      </main>
    </div>

    <!-- Question preview modal -->
    <QuestionPreviewModal
      v-if="previewQuestion"
      :question="previewQuestion"
      :topic-name="topics.find(t => t.id === previewQuestion!.topic_id)?.name ?? '—'"
      @close="previewQuestion = null"
    />

    <!-- Delete topic confirmation modal -->
    <Teleport to="body">
      <div v-if="deleteModal" class="modal-backdrop" @click.self="deleteModal = null">
        <div class="modal-dialog" style="max-width: 420px;">
          <div class="modal-header">
            <h2 class="modal-title">Hapus Topik</h2>
            <button class="modal-close" @click="deleteModal = null">×</button>
          </div>
          <div class="modal-body" style="text-align: center; gap: 0.75rem;">
            <!-- Blocked: topic has questions -->
            <template v-if="deleteModal.questionCount > 0">
              <div class="del-icon del-icon--warn">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="22" height="22">
                  <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
              </div>
              <h3 style="margin:0;font-size:0.95rem;">Tidak Dapat Menghapus</h3>
              <p style="margin:0;font-size:0.78rem;color:var(--text-muted);">
                Topik <strong>{{ deleteModal.name }}</strong> sedang digunakan oleh soal.
              </p>
              <div class="del-count-box">
                <span class="del-count-num">{{ deleteModal.questionCount }}</span>
                <span class="del-count-lbl">soal menggunakan topik ini</span>
              </div>
              <p style="margin:0;font-size:0.7rem;color:var(--text-muted);">
                Hapus atau pindahkan soal terlebih dahulu.
              </p>
              <button class="btn-primary" style="width:100%;" @click="deleteModal = null">Mengerti</button>
            </template>
            <!-- Can delete: no questions -->
            <template v-else>
              <div class="del-icon del-icon--danger">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="22" height="22">
                  <path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14z"/>
                </svg>
              </div>
              <h3 style="margin:0;font-size:0.95rem;">Hapus Topik</h3>
              <p style="margin:0;font-size:0.78rem;color:var(--text-muted);">
                Anda akan menghapus topik <strong>{{ deleteModal.name }}</strong>.
              </p>
              <div class="del-safe-badge">Aman untuk dihapus — tidak ada soal terkait</div>
            </template>
          </div>
          <div v-if="deleteModal.questionCount === 0" class="modal-footer">
            <button class="btn-cancel" @click="deleteModal = null">Batal</button>
            <button class="btn-primary" style="background: var(--danger);" @click="confirmDeleteTopic">Ya, Hapus</button>
          </div>
        </div>
      </div>
    </Teleport>

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

/* ── Two-column layout ── */
.qb-page { display: flex; flex-direction: column; gap: 1.25rem; }
.qb-layout { display: flex; gap: 1.25rem; align-items: flex-start; }

/* ── Sidebar ── */
.qb-sidebar {
  width: 220px; flex-shrink: 0;
  background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: 12px; overflow: hidden; position: sticky; top: 5rem;
}
.sb-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.65rem 0.875rem; border-bottom: 1px solid var(--border);
}
.sb-title { font-size: 0.6rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.1em; color: var(--text-muted); display: flex; align-items: center; gap: 0.35rem; }
.sb-count { font-size: 0.55rem; color: var(--text-muted); background: var(--bg-input); padding: 0.08rem 0.3rem; border-radius: 99px; }
.sb-add-btn {
  width: 22px; height: 22px; border-radius: 5px; border: 1px solid var(--border);
  background: transparent; color: var(--text-muted); font-size: 0.9rem; cursor: pointer;
  display: flex; align-items: center; justify-content: center; line-height: 1; transition: all 0.15s;
}
.sb-add-btn:hover { background: var(--accent); color: #fff; border-color: var(--accent); }

/* inline create */
.sb-create { padding: 0.5rem 0.75rem; border-bottom: 1px solid var(--border); display: flex; flex-direction: column; gap: 0.3rem; background: color-mix(in srgb, var(--accent) 6%, transparent); }
.sb-create-input { padding: 0.32rem 0.45rem; border-radius: 5px; border: 1px solid var(--accent); background: var(--bg-input); color: var(--text-primary); font-size: 0.72rem; outline: none; font-family: inherit; }
.sb-create-actions { display: flex; gap: 0.3rem; justify-content: flex-end; }
.sb-create-save, .sb-create-cancel {
  padding: 0.22rem 0.5rem; border-radius: 4px; border: none; font-size: 0.65rem; font-weight: 700; cursor: pointer; font-family: inherit;
}
.sb-create-save { background: var(--accent); color: #fff; }
.sb-create-save:disabled { opacity: 0.5; cursor: not-allowed; }
.sb-create-cancel { background: transparent; border: 1px solid var(--border); color: var(--text-muted); }
.sb-err { margin: 0; font-size: 0.62rem; color: var(--danger); }

/* topic list */
.sb-list { display: flex; flex-direction: column; }
.sb-item {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.45rem 0.875rem; border: none; border-bottom: 1px solid var(--bg-input);
  background: transparent; cursor: pointer; font-size: 0.75rem; color: var(--text-primary);
  transition: background 0.1s; width: 100%; text-align: left; font-family: inherit;
  position: relative;
}
.sb-item:last-child { border-bottom: none; }
.sb-item:hover { background: color-mix(in srgb, var(--accent) 3%, transparent); }
.sb-item--active { background: color-mix(in srgb, var(--accent) 8%, transparent); font-weight: 600; border-right: 2px solid var(--accent); }
.sb-item--all { border-bottom: 1px solid var(--border); font-weight: 600; }
.sb-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.sb-item-name { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.sb-item-num {
  font-size: 0.58rem; font-weight: 700; color: var(--text-muted);
  background: var(--bg-input); padding: 0.06rem 0.3rem; border-radius: 99px;
}
.sb-item .sb-item-actions { display: none; gap: 0.1rem; position: absolute; right: 0.35rem; background: var(--bg-surface); padding: 0.05rem 0.2rem; border-radius: 4px; }
.sb-item:hover .sb-item-actions { display: flex; }
.sb-item:hover .sb-item-num { visibility: hidden; }
.sb-act-btn {
  width: 20px; height: 20px; border-radius: 3px; border: none;
  background: transparent; color: var(--text-muted); cursor: pointer;
  font-size: 0.65rem; display: flex; align-items: center; justify-content: center; font-family: inherit;
}
.sb-act-btn:hover { background: var(--bg-input); color: var(--accent); }
.sb-act-btn--del:hover { color: var(--danger); }

/* edit inline in sidebar */
.sb-edit-row { padding: 0.3rem 0.75rem; border-bottom: 1px solid var(--bg-input); background: color-mix(in srgb, var(--accent) 6%, transparent); display: flex; flex-direction: column; gap: 0.25rem; }
.sb-edit-input { padding: 0.28rem 0.4rem; border-radius: 4px; border: 1px solid var(--accent); background: var(--bg-input); color: var(--text-primary); font-size: 0.7rem; outline: none; font-family: inherit; }
.sb-edit-actions { display: flex; gap: 0.2rem; justify-content: flex-end; }
.sb-edit-save, .sb-edit-cancel { padding: 0.18rem 0.4rem; border-radius: 3px; border: none; font-size: 0.62rem; font-weight: 700; cursor: pointer; font-family: inherit; }
.sb-edit-save { background: var(--accent); color: #fff; }
.sb-edit-save:disabled { opacity: 0.5; }
.sb-edit-cancel { background: transparent; border: 1px solid var(--border); color: var(--text-muted); }

/* ── Main content area ── */
.qb-main { flex: 1; min-width: 0; }

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

.question-list { display: flex; flex-direction: column; gap: 0.5rem; }
.question-card { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 10px; padding: 0.9rem 1.1rem; display: flex; flex-direction: column; gap: 0.45rem; transition: border-color 0.15s; }
.question-card:hover { border-color: color-mix(in srgb, var(--accent) 25%, var(--border)); }
.q-top { display: flex; align-items: center; gap: 0.5rem; }
.q-type { font-size: 0.58rem; font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em; padding: 0.15rem 0.45rem; border-radius: 3px; border: 1px solid; background: color-mix(in srgb, currentColor 10%, transparent); }
.q-topic-name { font-size: 0.68rem; font-weight: 600; color: var(--text-muted); display: flex; align-items: center; gap: 0.25rem; }
.q-topic-dot { width: 6px; height: 6px; border-radius: 50%; flex-shrink: 0; }
.q-edu { font-size: 0.6rem; font-weight: 700; padding: 0.1rem 0.35rem; border-radius: 3px; background: color-mix(in srgb, var(--accent) 12%, transparent); color: var(--accent); letter-spacing: 0.03em; }
.q-diff { font-size: 0.62rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.04em; padding: 0.1rem 0.35rem; border-radius: 3px; }
.q-diff--easy { color: var(--success); background: rgba(0,210,160,0.08); }
.q-diff--medium { color: var(--warning); background: rgba(245,166,35,0.08); }
.q-diff--hard { color: var(--danger); background: rgba(255,82,82,0.08); }
.q-spacer { flex: 1; }
.q-act { padding: 0.25rem 0.55rem; border-radius: 5px; border: 1px solid var(--border); background: transparent; color: var(--text-muted); font-size: 0.68rem; font-weight: 600; cursor: pointer; transition: all 0.12s; font-family: inherit; }
.q-act:hover { border-color: var(--accent); color: var(--accent); }
.q-act--del:hover { border-color: var(--danger); color: var(--danger); }
.q-act--preview:hover { border-color: var(--success); color: var(--success); }
.q-owner { font-size: 0.65rem; color: var(--text-muted); font-style: italic; }
.q-usage { font-size: 0.62rem; font-weight: 600; color: var(--success); background: rgba(0,210,160,0.08); padding: 0.1rem 0.4rem; border-radius: 4px; }
.q-text { font-size: 0.85rem; color: var(--text-primary); line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.q-options { display: flex; gap: 0.3rem; }
.q-opt { width: 1.5rem; height: 1.5rem; border-radius: 50%; background: var(--bg-input); border: 1px solid var(--border); display: flex; align-items: center; justify-content: center; font-size: 0.65rem; font-weight: 700; color: var(--text-muted); }
.q-opt.correct { background: var(--success); border-color: var(--success); color: #000; }
.q-stmt-count { font-size: 0.72rem; color: var(--text-muted); }
.f-chip--topic { gap: 0.35rem; }
.f-chip-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }

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

/* ── Delete confirmation modal ── */
.del-icon {
  width: 44px; height: 44px; border-radius: 50%; margin: 0 auto;
  display: flex; align-items: center; justify-content: center;
}
.del-icon--warn { background: rgba(245,166,35,0.12); color: var(--warning); }
.del-icon--danger { background: rgba(255,82,82,0.1); color: var(--danger); }
.del-count-box {
  padding: 0.5rem; background: rgba(245,166,35,0.06);
  border: 1px solid rgba(245,166,35,0.15); border-radius: 8px;
  display: flex; flex-direction: column; align-items: center; gap: 0.1rem;
}
.del-count-num { font-size: 1.5rem; font-weight: 800; color: var(--warning); line-height: 1; }
.del-count-lbl { font-size: 0.65rem; color: var(--text-muted); }
.del-safe-badge {
  text-align: center; padding: 0.35rem 0.5rem; border-radius: 6px;
  font-size: 0.7rem; color: var(--success); background: rgba(0,210,160,0.06);
  border: 1px solid rgba(0,210,160,0.12);
}

@media (max-width: 800px) {
  .qb-layout { flex-direction: column; }
  .qb-sidebar { width: 100%; position: static; max-height: 260px; overflow-y: auto; }
}
@media (max-width: 640px) {
  .modal-dialog { max-height: 95vh; border-radius: 12px; }
  .form-row { flex-direction: column; }
}
</style>
