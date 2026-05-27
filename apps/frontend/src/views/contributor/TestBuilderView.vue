<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { RouterLink } from 'vue-router'
import client from '@/api/client'
import { updateOwnedCategory } from '@/api/contributor'
import { useAuthStore } from '@/stores/auth'
import type { components } from '@tkaprep/shared-types'

type Test = components['schemas']['TestDetailResponse']
type Question = components['schemas']['QuestionDetailResponse']
type Topic = components['schemas']['TopicResponse']
type Category = components['schemas']['Category']

const tests = ref<Test[]>([])
const categories = ref<Category[]>([])
const isLoading = ref(true)
const actionError = ref('')

const showCreate = ref(false)
const creating = ref(false)
const createForm = ref({ title: '', description: '', category_id: '', duration_minutes: 90, difficulty: 'medium' as 'easy' | 'medium' | 'hard', education_level: '' as '' | 'sd' | 'smp' | 'sma' | 'smk' })

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
const showNewTopic = ref(false)
const newTopicName = ref('')
const newTopicSaving = ref(false)
const newTopicError = ref('')

async function setStatus(testId: string, publish: boolean) {
  actionError.value = ''
  const path = publish ? '/tests/{testId}/publish' : '/tests/{testId}/unpublish'
  const { error } = await client.POST(path as '/tests/{testId}/publish', { params: { path: { testId } } })
  if (error) {
    const msg = (error as any)?.error?.message || (error as any)?.message
    actionError.value = msg ? String(msg) : (publish ? 'Gagal menerbitkan tes.' : 'Gagal membatalkan terbit.')
    return
  }
  await fetchTests()
}

async function deleteTest(id: string) {
  if (!confirm('Hapus tes ini?')) return
  const { error } = await client.DELETE('/tests/{testId}', { params: { path: { testId: id } } })
  if (error) { actionError.value = 'Tidak bisa menghapus tes yang sudah diterbitkan. Unpublish terlebih dahulu.'; return }
  await fetchTests()
}

async function fetchCategories() {
  const { data } = await client.GET('/admin/categories')
  if (data) categories.value = data.data
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
  if (error) { actionError.value = 'Gagal membuat tes.'; return }
  showCreate.value = false
  await fetchTests()
}

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

async function createTopic() {
  const name = newTopicName.value.trim()
  if (!name) return
  newTopicSaving.value = true; newTopicError.value = ''
  const { data, error } = await client.POST('/topics', { body: { name } })
  if (error) { newTopicError.value = 'Gagal atau nama sudah ada.'; newTopicSaving.value = false; return }
  newTopicSaving.value = false; showNewTopic.value = false; newTopicName.value = ''
  await fetchSelectorTopics()
  if (data) selectorTopic.value = data.id
}

async function fetchSelectorQuestions() {
  selectorLoading.value = true
  const { data } = await client.GET('/questions', {
    params: { query: { search: selectorSearch.value || undefined, topic_id: selectorTopic.value || undefined, difficulty: (selectorDifficulty.value as 'easy' | 'medium' | 'hard') || undefined, limit: 100 } },
  })
  if (data) selectorQuestions.value = data.data
  selectorLoading.value = false
}

function toggleQ(id: string) {
  if (selectorSelectedIds.value.has(id)) { selectorSelectedIds.value.delete(id) }
  else { selectorSelectedIds.value.add(id) }
  selectorSelectedIds.value = new Set(selectorSelectedIds.value)
}

async function saveQSelection() {
  selectorError.value = ''; selectorSaving.value = true
  const { error } = await client.PUT('/tests/{testId}/questions', { params: { path: { testId: selectorTestId.value } }, body: { question_ids: [...selectorSelectedIds.value] } })
  selectorSaving.value = false
  if (error) { selectorError.value = 'Gagal menyimpan soal.'; return }
  showQSelector.value = false; await fetchTests()
}

const topicName = computed(() => {
  const map = Object.fromEntries(selectorTopics.value.map((t) => [t.id, t.name]))
  return (id: string) => map[id] ?? '—'
})

function getAttempts(t: Test): number { return (t as any).attempt_count ?? t.questions.length }

const top5 = computed(() =>
  [...tests.value].filter(t => t.status === 'published').sort((a, b) => getAttempts(b) - getAttempts(a)).slice(0, 5)
)

const diffColor: Record<string, string> = { easy: 'var(--success)', medium: 'var(--warning)', hard: 'var(--danger)' }

const auth = useAuthStore()

// Category edit
const catEditId = ref('')
const catEditName = ref('')
const catEditDesc = ref('')
const catEditSaving = ref(false)
const catEditError = ref('')
const catToast = ref('')
let catToastTimer: ReturnType<typeof setTimeout> | null = null

const ownedCategories = computed(() =>
  categories.value.filter((c) => (c as any).created_by === auth.user?.id)
)

function startEditCat(cat: Category) {
  catEditId.value = cat.id
  catEditName.value = cat.name
  catEditDesc.value = (cat as any).description ?? ''
  catEditError.value = ''
}

function cancelEditCat() {
  catEditId.value = ''
  catEditError.value = ''
}

function showToast(msg: string) {
  catToast.value = msg
  if (catToastTimer) clearTimeout(catToastTimer)
  catToastTimer = setTimeout(() => { catToast.value = '' }, 3500)
}

async function saveEditCat() {
  const name = catEditName.value.trim()
  if (!name) { catEditError.value = 'Nama kategori tidak boleh kosong.'; return }
  catEditSaving.value = true; catEditError.value = ''
  const result = await updateOwnedCategory(catEditId.value, name, catEditDesc.value.trim() || undefined)
  catEditSaving.value = false
  if (!result.ok) {
    if (result.status === 403) {
      showToast('Tidak diizinkan: bukan pemilik kategori.')
    } else {
      catEditError.value = result.message ?? 'Gagal menyimpan.'
    }
    return
  }
  // Update in-place without refetch
  const idx = categories.value.findIndex((c) => c.id === catEditId.value)
  if (idx !== -1 && result.data) {
    categories.value[idx] = result.data
  }
  cancelEditCat()
}

onMounted(async () => { await Promise.all([fetchCategories(), fetchTests()]); isLoading.value = false })
</script>

<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Pembuatan Tes</h1>
      <button class="btn-primary" @click="showCreate = true">+ Tes Baru</button>
    </div>
    <p v-if="actionError" class="error-msg">{{ actionError }}</p>
    <div v-if="isLoading" class="loading">Loading…</div>

    <div v-else class="test-layout">
      <div class="tests-col">
        <div class="section-label">Tes Saya <span class="section-badge">{{ tests.length }} tes</span></div>
        <div v-if="tests.length === 0" class="empty-state">Belum ada tes. Buat tes pertama Anda!</div>
        <div v-else class="test-list">
          <div v-for="t in tests" :key="t.id" class="test-card">
            <div class="tc-header">
              <span class="tc-category">{{ t.category_name }}</span>
              <span class="tc-status" :class="t.status">{{ t.status }}</span>
            </div>
            <h3 class="tc-title">{{ t.title }}</h3>
            <div class="tc-meta">
              <span :style="{ color: diffColor[t.difficulty] }">{{ t.difficulty }}</span>
              <span>{{ t.duration_minutes }} mnt</span>
              <span>{{ t.questions.length }} soal</span>
              <span v-if="t.status === 'published' && (t as any).attempt_count > 0">{{ (t as any).attempt_count }} siswa</span>
            </div>
            <div class="tc-actions">
              <button v-if="t.status === 'draft'" class="btn-sm btn-questions" @click="openQSelector(t)">Pilih Soal</button>
              <button v-if="t.status === 'draft'" class="btn-sm btn-publish" :disabled="t.questions.length === 0" @click="setStatus(t.id, true)">Terbitkan</button>
              <RouterLink v-if="t.status === 'published'" :to="'/contrib/tests/' + t.id + '/results'" class="btn-sm btn-results">Lihat Hasil</RouterLink>
              <button v-if="t.status === 'published'" class="btn-sm btn-unpublish" @click="setStatus(t.id, false)">Batalkan Terbit</button>
              <button v-if="t.status === 'draft'" class="btn-sm btn-delete" @click="deleteTest(t.id)">Hapus</button>
            </div>
          </div>
        </div>
      </div>

      <div class="top5-col">
        <div class="section-label warm">Terpopuler <span class="section-badge warm">dari tes saya</span></div>
        <div v-if="top5.length === 0" class="top5-empty">Belum ada tes yang diterbitkan</div>
        <div v-else class="top5-list">
          <div v-for="(t, i) in top5" :key="t.id" class="top5-item">
            <div class="top5-rank" :class="i < 3 ? `rank-${i+1}` : 'rank-n'">{{ i + 1 }}</div>
            <div class="top5-info"><div class="top5-title">{{ t.title }}</div><div class="top5-meta">{{ t.category_name }} · {{ t.duration_minutes }} mnt</div></div>
            <div class="top5-right"><div class="top5-bar-wrap"><div class="top5-bar" :style="{ width: top5[0] && getAttempts(top5[0]) > 0 ? (getAttempts(t) / getAttempts(top5[0]) * 100) + '%' : '0%' }"></div></div><span class="attempt-count">{{ getAttempts(t) }}</span><span class="attempt-label">soal</span></div>
          </div>
        </div>

        <!-- Category manager — only shows owned categories -->
        <template v-if="ownedCategories.length > 0">
          <div class="section-label cat-section-label">Kategori Saya</div>
          <div class="cat-list">
            <template v-for="cat in ownedCategories" :key="cat.id">
              <!-- Editing row -->
              <div v-if="catEditId === cat.id" class="cat-edit-row">
                <input v-model="catEditName" class="cat-edit-input" placeholder="Nama kategori" maxlength="100" @keyup.enter="saveEditCat" @keyup.escape="cancelEditCat" />
                <input v-model="catEditDesc" class="cat-edit-input cat-edit-desc" placeholder="Deskripsi (opsional)" maxlength="300" />
                <p v-if="catEditError" class="cat-err">{{ catEditError }}</p>
                <div class="cat-edit-actions">
                  <button class="cat-btn-cancel" @click="cancelEditCat">Batal</button>
                  <button class="cat-btn-save" :disabled="catEditSaving" @click="saveEditCat">{{ catEditSaving ? '…' : 'Simpan' }}</button>
                </div>
              </div>
              <!-- Normal row -->
              <div v-else class="cat-row">
                <span class="cat-name">{{ cat.name }}</span>
                <button class="cat-edit-btn" title="Edit" @click="startEditCat(cat)">✎</button>
              </div>
            </template>
          </div>
        </template>
      </div>
    </div>

    <!-- Category edit error toast -->
    <Teleport to="body">
      <div v-if="catToast" class="cat-toast">{{ catToast }}</div>
    </Teleport>

    <!-- Create modal -->
    <div v-if="showCreate" class="modal-backdrop" @click.self="showCreate = false">
      <div class="modal"><h3>Tes Baru</h3>
        <div class="field"><label>Judul</label><input v-model="createForm.title" class="form-input" placeholder="Judul tes…" /></div>
        <div class="field"><label>Deskripsi (opsional)</label><input v-model="createForm.description" class="form-input" placeholder="Deskripsi singkat…" /></div>
        <div class="form-row">
          <div class="field"><label>Kategori</label><select v-model="createForm.category_id" class="form-select" required><option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option></select></div>
          <div class="field"><label>Kesulitan</label><select v-model="createForm.difficulty" class="form-select"><option value="easy">Mudah</option><option value="medium">Sedang</option><option value="hard">Sulit</option></select></div>
        </div>
        <div class="field"><label>Durasi (menit)</label><input v-model.number="createForm.duration_minutes" type="number" class="form-input" min="5" max="300" /></div>
        <div class="field"><label>Jenjang</label><select v-model="createForm.education_level" class="form-select"><option value="">Semua Jenjang</option><option value="sd">SD</option><option value="smp">SMP</option><option value="sma">SMA</option><option value="smk">SMK</option></select></div>
        <div class="modal-actions"><button class="btn-cancel" @click="showCreate = false">Batal</button><button class="btn-primary" :disabled="!createForm.title || creating" @click="createTest">{{ creating ? 'Membuat…' : 'Buat Tes' }}</button></div>
      </div>
    </div>

    <!-- Question Selector modal -->
    <div v-if="showQSelector" class="modal-backdrop" @click.self="showQSelector = false">
      <div class="modal modal--wide"><h3>Pilih Soal — {{ selectorTestTitle }}</h3>
        <div class="qs-filters">
          <input v-model="selectorSearch" class="qs-search-input" placeholder="Cari soal…" @input="fetchSelectorQuestions" />
          <div class="qs-filter-with-add">
            <select v-model="selectorTopic" class="qs-filter-select" @change="fetchSelectorQuestions" :disabled="showNewTopic"><option value="">Semua Topik</option><option v-for="t in selectorTopics" :key="t.id" :value="t.id">{{ t.name }}</option></select>
            <button class="qs-add-topic-btn" @click="showNewTopic = !showNewTopic" :title="showNewTopic ? 'Batal' : 'Topik Baru'">{{ showNewTopic ? '×' : '+' }}</button>
          </div>
          <div v-if="showNewTopic" class="qs-new-topic">
            <input v-model="newTopicName" type="text" class="qs-new-topic-input" placeholder="Nama topik baru..." maxlength="100" @keyup.enter="createTopic" />
            <p v-if="newTopicError" class="qs-new-topic-err">{{ newTopicError }}</p>
            <button class="qs-new-topic-save" :disabled="newTopicSaving || !newTopicName.trim()" @click="createTopic">{{ newTopicSaving ? '...' : 'Simpan' }}</button>
          </div>
          <select v-model="selectorDifficulty" class="qs-filter-select" @change="fetchSelectorQuestions"><option value="">Semua Kesulitan</option><option value="easy">Mudah</option><option value="medium">Sedang</option><option value="hard">Sulit</option></select>
        </div>
        <div v-if="selectorLoading" class="qs-loading">Loading…</div>
        <div v-else class="qs-list">
          <div v-for="q in selectorQuestions" :key="q.id" class="qs-item" :class="{ selected: selectorSelectedIds.has(q.id) }" @click="toggleQ(q.id)">
            <div class="qs-item-left"><span class="qs-type">{{ {mcq:'PG',multi_correct:'PGK',true_false:'B/S'}[q.question_type] }}</span><span class="qs-item-topic">{{ topicName(q.topic_id) }}</span></div>
            <div class="qs-item-body">
              <div class="qs-item-text" v-html="q.text.substring(0, 120) + (q.text.length > 120 ? '…' : '')"></div>
              <div class="qs-item-meta">
                <span v-if="(q as any).contributor_name" class="qs-owner">oleh: {{ (q as any).contributor_name }}</span>
                <span v-if="(q as any).usage && (q as any).usage.own_test_count + (q as any).usage.other_test_count > 0" class="qs-usage">
                  {{ (q as any).usage.own_test_count + (q as any).usage.other_test_count }} tes
                </span>
              </div>
            </div>
            <div class="qs-check" :class="{ on: selectorSelectedIds.has(q.id) }">{{ selectorSelectedIds.has(q.id) ? '✓' : '' }}</div>
          </div>
        </div>
        <div class="modal-actions"><button class="btn-cancel" @click="showQSelector = false">Batal</button><button class="btn-primary" :disabled="selectorSaving" @click="saveQSelection">{{ selectorSaving ? 'Menyimpan…' : `Simpan (${selectorSelectedIds.size} soal)` }}</button></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
.page-title { margin: 0; font-size: 1.5rem; font-weight: 800; }
.loading { color: var(--text-muted); }
.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: var(--danger-bg); color: var(--danger-text); font-size: 0.825rem; margin-bottom: 1rem; }
.test-layout { display: flex; gap: 1.25rem; align-items: flex-start; }
.tests-col { flex: 1; min-width: 0; }
.top5-col { width: 240px; flex-shrink: 0; background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; padding: 1rem; display: flex; flex-direction: column; gap: 0.75rem; }
.section-label { font-size: 0.65rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-muted); display: flex; align-items: center; gap: 0.4rem; }
.section-label.warm { color: #f59e0b; }
.section-badge { font-size: 0.55rem; font-weight: 700; padding: 0.1rem 0.4rem; border-radius: 99px; background: var(--bg-input); color: var(--text-muted); letter-spacing: 0.03em; }
.section-badge.warm { background: var(--warm-dim); color: var(--warm); }
.tests-col .test-list { display: flex; flex-direction: column; gap: 0.625rem; }
.test-card { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; padding: 1.25rem; display: flex; flex-direction: column; gap: 0.5rem; transition: border-color 0.15s; }
.tc-header { display: flex; align-items: center; justify-content: space-between; }
.tc-category { font-size: 0.72rem; font-weight: 600; color: var(--accent); text-transform: uppercase; }
.tc-status { font-size: 0.62rem; font-weight: 700; text-transform: uppercase; padding: 0.18rem 0.55rem; border-radius: 99px; }
.tc-status.draft { background: var(--bg-input); color: var(--text-muted); }
.tc-status.published { background: rgba(34,197,94,0.1); color: var(--success); }
.tc-title { margin: 0; font-size: 0.95rem; font-weight: 700; color: var(--text-heading); }
.tc-meta { display: flex; gap: 0.75rem; font-size: 0.72rem; color: var(--text-muted); }
.tc-actions { display: flex; gap: 0.375rem; margin-top: 0.25rem; }
.btn-sm { padding: 0.35rem 0.7rem; border-radius: 6px; font-size: 0.72rem; font-weight: 700; cursor: pointer; border: none; transition: all 0.12s; }
.btn-sm:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-questions { background: var(--bg-input); color: var(--accent); }
.btn-questions:hover { background: var(--accent); color: #fff; }
.btn-publish { background: var(--success); color: #fff; }
.btn-publish:hover:not(:disabled) { opacity: 0.85; }
.btn-unpublish { background: transparent; border: 1px solid var(--warning); color: var(--warning); }
.btn-delete { background: transparent; border: 1px solid var(--border); color: var(--danger); }
.btn-delete:hover { background: var(--danger); color: #fff; border-color: var(--danger); }
.btn-results { background: var(--accent); color: #fff; text-decoration: none; display: inline-block; }
.btn-results:hover { opacity: 0.85; }
.empty-state { color: var(--text-muted); text-align: center; padding: 2rem; }
.top5-empty { font-size: 0.72rem; color: var(--text-muted); text-align: center; padding: 1rem 0; }
.top5-list { display: flex; flex-direction: column; gap: 0.5rem; }
.top5-item { display: flex; align-items: center; gap: 0.5rem; }
.top5-rank { width: 22px; height: 22px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 0.65rem; font-weight: 800; background: var(--bg-input); color: var(--text-muted); flex-shrink: 0; }
.rank-1 { background: #f59e0b; color: #fff; }
.rank-2 { background: #94a3b8; color: #fff; }
.rank-3 { background: #b45309; color: #fff; }
.top5-info { flex: 1; min-width: 0; }
.top5-title { font-size: 0.72rem; font-weight: 600; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.top5-meta { font-size: 0.62rem; color: var(--text-muted); margin-top: 1px; }
.top5-right { display: flex; align-items: center; gap: 0.3rem; flex-shrink: 0; }
.top5-bar-wrap { width: 40px; height: 4px; background: var(--bg-input); border-radius: 2px; overflow: hidden; }
.top5-bar { height: 100%; background: var(--accent); border-radius: 2px; }
.attempt-count { font-size: 0.7rem; font-weight: 700; color: var(--accent); font-family: monospace; }
.attempt-label { font-size: 0.58rem; color: var(--text-muted); }

.btn-primary { padding: 0.55rem 1.25rem; border-radius: 8px; border: none; background: var(--accent); color: var(--text-on-accent); font-size: 0.875rem; font-weight: 600; cursor: pointer; transition: background 0.15s; }
.btn-primary:hover { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-cancel { padding: 0.55rem 1.25rem; border-radius: 8px; border: 1px solid var(--border); background: transparent; color: var(--text-muted); cursor: pointer; font-size: 0.875rem; font-weight: 600; transition: all 0.15s; }
.btn-cancel:hover { border-color: var(--danger); color: var(--danger); }

.modal-backdrop { position: fixed; inset: 0; z-index: 200; background: rgba(0,0,0,0.55); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; padding: 1rem; }
.modal { background: var(--bg-surface); border-radius: 14px; padding: 1.5rem; width: 500px; max-width: 90vw; max-height: 88vh; overflow-y: auto; display: flex; flex-direction: column; gap: 1rem; }
.modal--wide { width: 680px; }
.modal h3 { margin: 0; font-size: 1rem; font-weight: 700; }
.field { display: flex; flex-direction: column; gap: 0.3rem; }
.field label { font-size: 0.7rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; }
.form-input, .form-select { padding: 0.5rem 0.7rem; border-radius: 8px; border: 1px solid var(--border); background: var(--bg-input); color: var(--text-primary); font-size: 0.85rem; outline: none; font-family: inherit; }
.form-input:focus, .form-select:focus { border-color: var(--accent); }
.form-row { display: flex; gap: 0.75rem; }
.form-row .field { flex: 1; }
.modal-actions { display: flex; gap: 0.5rem; justify-content: flex-end; margin-top: 0.5rem; }

.qs-filters { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.qs-search-input { flex: 1; min-width: 140px; padding: 0.45rem 0.6rem; border-radius: 6px; border: 1px solid var(--border); background: var(--bg-input); color: var(--text-primary); font-size: 0.8rem; outline: none; font-family: inherit; }
.qs-filter-select { padding: 0.45rem 0.6rem; border-radius: 6px; border: 1px solid var(--border); background: var(--bg-input); color: var(--text-primary); font-size: 0.8rem; cursor: pointer; font-family: inherit; }
.qs-loading { text-align: center; padding: 2rem; color: var(--text-muted); }
.qs-list { display: flex; flex-direction: column; gap: 0.35rem; max-height: 380px; overflow-y: auto; }
.qs-item { display: flex; align-items: center; gap: 0.75rem; padding: 0.55rem 0.75rem; border: 1px solid var(--border); border-radius: 8px; cursor: pointer; transition: all 0.12s; }
.qs-item.selected { border-color: var(--accent); background: color-mix(in srgb, var(--accent) 5%, transparent); }
.qs-item-left { display: flex; gap: 0.4rem; align-items: center; flex-shrink: 0; }
.qs-type { font-size: 0.6rem; font-weight: 800; padding: 0.12rem 0.35rem; border-radius: 3px; border: 1px solid var(--accent); color: var(--accent); }
.qs-item-topic { font-size: 0.65rem; color: var(--text-muted); }
.qs-item-body { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.15rem; }
.qs-item-text { font-size: 0.78rem; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.qs-item-meta { display: flex; gap: 0.5rem; align-items: center; }
.qs-owner { font-size: 0.62rem; color: var(--text-muted); font-style: italic; }
.qs-usage { font-size: 0.6rem; font-weight: 600; color: var(--success); background: rgba(0,210,160,0.08); padding: 0.05rem 0.35rem; border-radius: 3px; }
.qs-check { width: 22px; height: 22px; border-radius: 50%; border: 2px solid var(--border); display: flex; align-items: center; justify-content: center; font-size: 0.7rem; flex-shrink: 0; }
.qs-check.on { background: var(--accent); border-color: var(--accent); color: #fff; }

.qs-filter-with-add { display: flex; gap: 0.25rem; }
.qs-add-topic-btn { width: 28px; height: 28px; border-radius: 6px; border: 1px solid var(--border); background: transparent; font-size: 1rem; font-weight: 700; color: var(--text-muted); cursor: pointer; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.qs-add-topic-btn:hover { background: var(--accent); color: #fff; border-color: var(--accent); }
.qs-new-topic { display: flex; flex-direction: column; gap: 0.35rem; margin-top: 0.35rem; }
.qs-new-topic-input { padding: 0.35rem 0.5rem; border-radius: 6px; border: 1px solid var(--accent); background: var(--bg-input); color: var(--text-primary); font-size: 0.75rem; font-family: inherit; }
.qs-new-topic-err { margin: 0; font-size: 0.65rem; color: var(--danger); }
.qs-new-topic-save { padding: 0.3rem 0.6rem; border-radius: 6px; border: none; background: var(--accent); color: #fff; font-size: 0.7rem; font-weight: 700; cursor: pointer; align-self: flex-end; }
.qs-new-topic-save:disabled { opacity: 0.5; cursor: not-allowed; }

@media (max-width: 768px) { .modal { max-width: 100%; border-radius: 12px; } .form-row { flex-direction: column; } .test-layout { flex-direction: column; } .top5-col { width: 100%; } }

/* ── Category manager ── */
.cat-section-label { margin-top: 0.25rem; padding-top: 0.75rem; border-top: 1px solid var(--border); }
.cat-list { display: flex; flex-direction: column; gap: 0.25rem; }
.cat-row { display: flex; align-items: center; gap: 0.4rem; padding: 0.35rem 0.45rem; border-radius: 6px; background: var(--bg-input); }
.cat-name { flex: 1; font-size: 0.75rem; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cat-edit-btn { width: 22px; height: 22px; border-radius: 4px; border: 1px solid var(--border); background: transparent; color: var(--text-muted); font-size: 0.7rem; cursor: pointer; display: flex; align-items: center; justify-content: center; flex-shrink: 0; transition: all 0.12s; }
.cat-edit-btn:hover { border-color: var(--accent); color: var(--accent); }
.cat-edit-row { display: flex; flex-direction: column; gap: 0.3rem; padding: 0.5rem 0.45rem; border-radius: 6px; background: color-mix(in srgb, var(--accent) 5%, var(--bg-input)); border: 1px solid color-mix(in srgb, var(--accent) 25%, var(--border)); }
.cat-edit-input { padding: 0.3rem 0.45rem; border-radius: 5px; border: 1px solid var(--border); background: var(--bg-input); color: var(--text-primary); font-size: 0.75rem; font-family: inherit; outline: none; }
.cat-edit-input:focus { border-color: var(--accent); }
.cat-edit-desc { font-size: 0.7rem; }
.cat-err { margin: 0; font-size: 0.65rem; color: var(--danger); }
.cat-edit-actions { display: flex; gap: 0.25rem; justify-content: flex-end; }
.cat-btn-save, .cat-btn-cancel { padding: 0.2rem 0.5rem; border-radius: 4px; border: none; font-size: 0.65rem; font-weight: 700; cursor: pointer; font-family: inherit; }
.cat-btn-save { background: var(--accent); color: #fff; }
.cat-btn-save:disabled { opacity: 0.5; cursor: not-allowed; }
.cat-btn-cancel { background: transparent; border: 1px solid var(--border); color: var(--text-muted); }

/* ── Toast ── */
@keyframes toast-in { from { opacity: 0; transform: translateY(8px); } to { opacity: 1; transform: translateY(0); } }
.cat-toast {
  position: fixed; bottom: 1.5rem; left: 50%; transform: translateX(-50%);
  background: var(--danger); color: #fff;
  padding: 0.6rem 1.25rem; border-radius: 9px; font-size: 0.82rem; font-weight: 600;
  box-shadow: 0 4px 16px rgba(0,0,0,0.3); z-index: 9999;
  animation: toast-in 0.2s ease both;
  white-space: nowrap;
}
</style>
