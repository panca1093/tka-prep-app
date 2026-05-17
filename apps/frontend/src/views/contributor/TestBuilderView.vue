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
  if (error) { actionError.value = 'Gagal mengubah status.'; return }
  await fetchTests()
}

async function deleteTest(id: string) {
  if (!confirm('Hapus tes ini?')) return
  const { error } = await client.DELETE('/tests/{testId}', { params: { path: { testId: id } } })
  if (error) { actionError.value = 'Tidak bisa menghapus tes yang sudah diterbitkan. Unpublish terlebih dahulu.'; return }
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
  if (error) { actionError.value = 'Gagal membuat tes.'; return }
  showCreate.value = false
  await fetchTests()
}

// ── Question selector ──────────────────────────────────────────────
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
  if (error) { selectorError.value = 'Gagal menyimpan soal.'; return }
  showQSelector.value = false
  await fetchTests()
}

const topicName = computed(() => {
  const map = Object.fromEntries(selectorTopics.value.map((t) => [t.id, t.name]))
  return (id: string) => map[id] ?? '—'
})

// TestDetailResponse does not have attempt_count; use questions.length as fallback
function getAttempts(t: Test): number {
  return (t as any).attempt_count ?? t.questions.length
}

const top5 = computed(() =>
  [...tests.value]
    .filter(t => t.status === 'published')
    .sort((a, b) => getAttempts(b) - getAttempts(a))
    .slice(0, 5)
)

const categoryLabel: Record<string, string> = { tka_saintek: 'TKA Saintek', tka_soshum: 'TKA Soshum', smbt: 'SMBT' }
const diffColor: Record<string, string> = { easy: 'var(--success)', medium: 'var(--warning)', hard: 'var(--danger)' }

onMounted(async () => { await fetchTests(); isLoading.value = false })
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
      <!-- Left column: test list -->
      <div class="tests-col">
        <div class="section-label">
          Tes Saya <span class="section-badge">{{ tests.length }} tes</span>
        </div>

        <div v-if="tests.length === 0" class="empty-state">Belum ada tes. Buat tes pertama Anda!</div>

        <div v-else class="test-list">
          <div v-for="t in tests" :key="t.id" class="test-card">
            <div class="tc-header">
              <span class="tc-category">{{ categoryLabel[t.category] }}</span>
              <span class="tc-status" :class="t.status">{{ t.status }}</span>
            </div>
            <h3 class="tc-title">{{ t.title }}</h3>
            <div class="tc-meta">
              <span :style="{ color: diffColor[t.difficulty] }">{{ t.difficulty }}</span>
              <span>{{ t.duration_minutes }} mnt</span>
              <span>{{ t.questions.length }} soal</span>
            </div>
            <div class="tc-actions">
              <button v-if="t.status === 'draft'" class="btn-sm btn-questions" @click="openQSelector(t)">
                Pilih Soal
              </button>
              <button v-if="t.status === 'draft'" class="btn-sm btn-publish" :disabled="t.questions.length === 0" @click="setStatus(t.id, true)">Terbitkan</button>
              <button v-else class="btn-sm btn-unpublish" @click="setStatus(t.id, false)">Batalkan Terbit</button>
              <button v-if="t.status === 'draft'" class="btn-sm btn-delete" @click="deleteTest(t.id)">Hapus</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Right column: top 5 -->
      <div class="top5-col">
        <div class="section-label warm">
          Terpopuler <span class="section-badge warm">dari tes saya</span>
        </div>
        <div v-if="top5.length === 0" class="top5-empty">Belum ada tes yang diterbitkan</div>
        <div v-else class="top5-list">
          <div v-for="(t, i) in top5" :key="t.id" class="top5-item">
            <div class="top5-rank" :class="i < 3 ? `rank-${i+1}` : 'rank-n'">{{ i + 1 }}</div>
            <div class="top5-info">
              <div class="top5-title">{{ t.title }}</div>
              <div class="top5-meta">{{ categoryLabel[t.category] }} · {{ t.duration_minutes }} mnt</div>
            </div>
            <div class="top5-right">
              <div class="top5-bar-wrap">
                <div
                  class="top5-bar"
                  :style="{ width: top5[0] && getAttempts(top5[0]) > 0 ? (getAttempts(t) / getAttempts(top5[0]) * 100) + '%' : '0%' }"
                ></div>
              </div>
              <span class="attempt-count">{{ getAttempts(t) }}</span>
              <span class="attempt-label">soal</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create modal -->
    <div v-if="showCreate" class="modal-backdrop" @click.self="showCreate = false">
      <div class="modal">
        <h3>Tes Baru</h3>
        <div class="field">
          <label>Judul</label>
          <input v-model="createForm.title" class="form-input" placeholder="Judul tes…" />
        </div>
        <div class="field">
          <label>Deskripsi (opsional)</label>
          <input v-model="createForm.description" class="form-input" placeholder="Deskripsi singkat…" />
        </div>
        <div class="form-row">
          <div class="field">
            <label>Kategori</label>
            <select v-model="createForm.category" class="form-select">
              <option value="tka_saintek">TKA Saintek</option>
              <option value="tka_soshum">TKA Soshum</option>
              <option value="smbt">SMBT</option>
            </select>
          </div>
          <div class="field">
            <label>Kesulitan</label>
            <select v-model="createForm.difficulty" class="form-select">
              <option value="easy">Mudah</option>
              <option value="medium">Sedang</option>
              <option value="hard">Sulit</option>
            </select>
          </div>
        </div>
        <div class="field">
          <label>Durasi (menit)</label>
          <input v-model.number="createForm.duration_minutes" type="number" class="form-input" min="5" max="300" />
        </div>
        <div class="field">
          <label>Jenjang</label>
          <select v-model="createForm.education_level" class="form-select">
            <option value="">Semua Jenjang</option>
            <option value="sd">SD</option>
            <option value="smp">SMP</option>
            <option value="sma">SMA</option>
            <option value="smk">SMK</option>
          </select>
        </div>
        <div class="modal-actions">
          <button class="btn-cancel" @click="showCreate = false">Batal</button>
          <button class="btn-primary" :disabled="!createForm.title || creating" @click="createTest">
            {{ creating ? 'Membuat…' : 'Buat' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Question selector panel -->
    <div v-if="showQSelector" class="qs-backdrop" @click.self="showQSelector = false">
      <div class="qs-panel">
        <div class="qs-header">
          <div>
            <h2 class="qs-title">Pilih Soal</h2>
            <p class="qs-sub">{{ selectorTestTitle }} · {{ selectorSelectedIds.size }} dipilih</p>
          </div>
          <button class="qs-close" @click="showQSelector = false">✕</button>
        </div>

        <div class="qs-filters">
          <input
            v-model="selectorSearch"
            class="qs-search-input"
            placeholder="Cari soal…"
            @input="fetchSelectorQuestions"
          />
          <select v-model="selectorTopic" class="qs-filter-select" @change="fetchSelectorQuestions">
            <option value="">Semua Topik</option>
            <option v-for="t in selectorTopics" :key="t.id" :value="t.id">{{ t.name }}</option>
          </select>
          <select v-model="selectorDifficulty" class="qs-filter-select" @change="fetchSelectorQuestions">
            <option value="">Semua Kesulitan</option>
            <option value="easy">Mudah</option>
            <option value="medium">Sedang</option>
            <option value="hard">Sulit</option>
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

          <div v-if="selectorQuestions.length === 0" class="qs-empty">Tidak ada soal ditemukan.</div>
        </div>

        <div class="qs-footer">
          <p v-if="selectorError" class="error-msg">{{ selectorError }}</p>
          <div class="qs-footer-actions">
            <button class="btn-cancel" @click="showQSelector = false">Batal</button>
            <button class="btn-primary" :disabled="selectorSaving" @click="saveQSelection">
              {{ selectorSaving ? 'Menyimpan…' : `Simpan (${selectorSelectedIds.size})` }}
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
.loading { color: var(--text-muted); }
.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: var(--danger-bg); color: var(--danger-text); font-size: 0.825rem; margin-bottom: 1rem; }
.empty-state { color: var(--text-muted); text-align: center; padding: 2rem; background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; }

/* ── Two-column layout ─────────────────────────────────────────── */
.test-layout { display: grid; grid-template-columns: 3fr 2fr; gap: 1.75rem; align-items: start; }
.tests-col { display: flex; flex-direction: column; }
.top5-col { position: sticky; top: 2rem; }

/* ── Section labels ────────────────────────────────────────────── */
.section-label { font-size: 0.68rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.07em; color: var(--text-muted); display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.875rem; }
.section-label::after { content: ''; flex: 1; height: 1px; background: var(--border); }
.section-label.warm { color: var(--warm); }
.section-badge { font-size: 0.62rem; font-weight: 700; padding: 0.15rem 0.45rem; border-radius: 4px; background: var(--accent-dim); color: var(--accent); text-transform: none; letter-spacing: 0; }
.section-badge.warm { background: var(--warm-dim); color: var(--warm); }

/* ── Test list (single column inside tests-col) ────────────────── */
.tests-col .test-list { display: flex; flex-direction: column; gap: 0.625rem; }
.test-card { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; padding: 1.25rem; display: flex; flex-direction: column; gap: 0.5rem; transition: border-color 0.15s; }
.test-card:hover { border-color: color-mix(in srgb, var(--accent) 30%, var(--border)); }
.tc-header { display: flex; align-items: center; justify-content: space-between; }
.tc-category { font-size: 0.72rem; font-weight: 600; color: var(--accent); text-transform: uppercase; }
.tc-status { font-size: 0.72rem; font-weight: 700; padding: 0.2rem 0.6rem; border-radius: 4px; text-transform: capitalize; }
.tc-status.published { background: rgba(34,197,94,0.15); color: var(--success); }
.tc-status.draft { background: rgba(148,163,184,0.15); color: var(--text-muted); }
.tc-title { margin: 0; font-size: 0.95rem; font-weight: 700; }
.tc-meta { display: flex; gap: 0.75rem; font-size: 0.78rem; color: var(--text-muted); }
.tc-actions { display: flex; gap: 0.5rem; margin-top: 0.25rem; flex-wrap: wrap; }
.btn-sm { padding: 0.3rem 0.75rem; border-radius: 6px; border: 1px solid; font-size: 0.78rem; cursor: pointer; transition: all 0.15s; background: transparent; }
.btn-questions { border-color: var(--accent); color: var(--accent); }
.btn-questions:hover { background: rgba(79,142,247,0.1); }
.btn-publish { border-color: var(--success); color: var(--success); }
.btn-publish:hover:not(:disabled) { background: rgba(34,197,94,0.1); }
.btn-publish:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-unpublish { border-color: var(--warning); color: var(--warning); }
.btn-unpublish:hover { background: rgba(245,158,11,0.1); }
.btn-delete { border-color: var(--danger); color: var(--danger); }
.btn-delete:hover { background: rgba(239,68,68,0.1); }

/* ── Top 5 sidebar ─────────────────────────────────────────────── */
.top5-list { display: flex; flex-direction: column; gap: 0.5rem; }
.top5-item { display: flex; align-items: center; gap: 0.75rem; padding: 0.875rem 1rem; background: var(--bg-surface); border: 1px solid var(--border); border-radius: 10px; transition: border-color 0.15s; }
.top5-item:hover { border-color: color-mix(in srgb, var(--accent) 35%, var(--border)); }
.top5-rank { width: 26px; height: 26px; border-radius: 50%; flex-shrink: 0; display: flex; align-items: center; justify-content: center; font-size: 0.75rem; font-weight: 800; }
.rank-1 { background: rgba(250,204,21,0.12); color: #fbbf24; border: 1.5px solid rgba(250,204,21,0.3); }
.rank-2 { background: rgba(148,163,184,0.1); color: #94a3b8; border: 1.5px solid rgba(148,163,184,0.25); }
.rank-3 { background: rgba(251,146,60,0.1); color: #fb923c; border: 1.5px solid rgba(251,146,60,0.25); }
.rank-n { background: rgba(255,255,255,0.03); color: var(--text-muted); border: 1.5px solid var(--border); }
.top5-info { flex: 1; min-width: 0; }
.top5-title { font-size: 0.82rem; font-weight: 600; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.top5-meta { font-size: 0.68rem; color: var(--text-muted); margin-top: 0.15rem; }
.top5-right { display: flex; flex-direction: column; align-items: flex-end; gap: 0.2rem; flex-shrink: 0; }
.top5-bar-wrap { width: 56px; height: 3px; background: var(--border); border-radius: 2px; overflow: hidden; }
.top5-bar { height: 100%; background: var(--accent); border-radius: 2px; opacity: 0.55; }
.attempt-count { font-size: 1rem; font-weight: 800; color: var(--accent); line-height: 1; }
.attempt-label { font-size: 0.62rem; color: var(--text-muted); }
.top5-empty { font-size: 0.82rem; color: var(--text-muted); padding: 1.5rem; text-align: center; background: var(--bg-surface); border: 1px solid var(--border); border-radius: 10px; }

/* ── Shared buttons ─────────────────────────────────────────────── */
.btn-primary { padding: 0.55rem 1.25rem; border-radius: 8px; border: none; background: var(--accent); color: var(--text-on-accent); font-size: 0.875rem; font-weight: 600; cursor: pointer; transition: background 0.15s; }
.btn-primary:hover { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

/* ── Create modal ───────────────────────────────────────────────── */
.modal-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.6); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 16px; padding: 2rem; width: 420px; max-width: calc(100vw - 2rem); display: flex; flex-direction: column; gap: 1rem; }
.modal h3 { margin: 0; font-size: 1.1rem; }
.field { display: flex; flex-direction: column; gap: 0.375rem; }
.field label { font-size: 0.78rem; font-weight: 600; color: var(--text-muted); }
.form-input, .form-select { padding: 0.6rem 0.875rem; border-radius: 8px; border: 1px solid var(--border); background: var(--bg-input); color: var(--text-primary); font-size: 0.875rem; outline: none; font-family: inherit; }
.form-input:focus, .form-select:focus { border-color: var(--accent); }
.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.modal-actions { display: flex; gap: 0.75rem; margin-top: 0.5rem; }
.btn-cancel { flex: 1; padding: 0.65rem; border-radius: 8px; border: 1px solid var(--border); background: transparent; color: var(--text-muted); cursor: pointer; transition: all 0.15s; font-family: inherit; }
.btn-cancel:hover { border-color: var(--danger); color: var(--danger); }

/* ── Question selector panel ────────────────────────────────────── */
.qs-backdrop { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 200; }
.qs-panel {
  position: fixed; right: 0; top: 0; bottom: 0; width: 560px; max-width: 100vw;
  background: var(--bg-input); border-left: 1px solid var(--border);
  display: flex; flex-direction: column;
}
.qs-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  padding: 1.5rem 1.5rem 1rem;
  border-bottom: 1px solid var(--border);
}
.qs-title { margin: 0; font-size: 1.05rem; font-weight: 700; }
.qs-sub { margin: 0.25rem 0 0; font-size: 0.8rem; color: var(--accent); }
.qs-close { background: transparent; border: none; color: var(--text-muted); font-size: 1.1rem; cursor: pointer; padding: 0.25rem; }
.qs-close:hover { color: var(--text-primary); }
.qs-filters {
  display: flex; gap: 0.625rem; padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--border); flex-wrap: wrap;
}
.qs-search-input {
  flex: 1; min-width: 160px;
  padding: 0.5rem 0.75rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-primary); font-size: 0.85rem; outline: none; font-family: inherit;
}
.qs-search-input:focus { border-color: var(--accent); }
.qs-filter-select {
  padding: 0.5rem 0.75rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-primary); font-size: 0.85rem;
  cursor: pointer; outline: none; font-family: inherit;
}
.qs-loading { padding: 2rem; color: var(--text-muted); text-align: center; }
.qs-list { flex: 1; overflow-y: auto; padding: 0.75rem 1.5rem; display: flex; flex-direction: column; gap: 0.5rem; }
.qs-item {
  display: flex; align-items: flex-start; gap: 0.875rem;
  padding: 0.875rem; border-radius: 10px; border: 1px solid var(--border);
  background: var(--bg-surface); cursor: pointer; transition: all 0.12s;
}
.qs-item:hover { border-color: var(--accent); }
.qs-item.selected { border-color: var(--accent); background: rgba(79,142,247,0.08); }
.qs-checkbox {
  width: 1.25rem; height: 1.25rem; flex-shrink: 0; border-radius: 4px;
  border: 2px solid var(--border); background: var(--bg-input);
  display: flex; align-items: center; justify-content: center;
  margin-top: 0.1rem;
}
.qs-item.selected .qs-checkbox { border-color: var(--accent); background: var(--accent); }
.check-mark { color: var(--text-on-accent); font-size: 0.7rem; font-weight: 900; }
.qs-content { flex: 1; min-width: 0; }
.qs-meta { display: flex; gap: 0.625rem; margin-bottom: 0.35rem; }
.qs-difficulty { font-size: 0.72rem; font-weight: 700; text-transform: capitalize; }
.qs-topic { font-size: 0.72rem; color: var(--text-muted); }
.qs-text { margin: 0; font-size: 0.875rem; color: var(--text-secondary); line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }
.qs-empty { padding: 2rem; text-align: center; color: var(--text-muted); font-size: 0.875rem; }
.qs-footer { padding: 1rem 1.5rem; border-top: 1px solid var(--border); }
.qs-footer-actions { display: flex; gap: 0.75rem; }
.qs-footer-actions .btn-cancel { flex: 0 0 auto; padding: 0.65rem 1.25rem; }
.qs-footer-actions .btn-primary { flex: 1; }

@media (max-width: 900px) {
  .test-layout { grid-template-columns: 1fr; }
  .top5-col { position: static; }
}
@media (max-width: 768px) {
  .modal { width: 100%; max-width: 100%; }
  .form-row { flex-direction: column; }
}
</style>
