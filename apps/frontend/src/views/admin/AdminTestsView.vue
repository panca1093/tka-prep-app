<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type AdminTest = components['schemas']['AdminTestResponse']
type Category = components['schemas']['Category']

const tests      = ref<AdminTest[]>([])
const total      = ref(0)
const page       = ref(1)
const pageSize   = 20
const isLoading  = ref(false)
const actionError = ref('')
const confirmUnpublishId = ref<string | null>(null)
const busyId     = ref<string | null>(null)

// Category sidebar
const categories = ref<Category[]>([])
const selectedCategoryId = ref<string>('')
const showAddCat = ref(false)
const newCatName = ref('')
const catError = ref('')
const catSaving = ref(false)

// Category CRUD
async function fetchCategories() {
  const { data } = await client.GET('/admin/categories')
  if (data) categories.value = data.data
}

async function createCategory() {
  const name = newCatName.value.trim()
  if (!name) return
  catSaving.value = true; catError.value = ''
  const { error } = await client.POST('/admin/categories', { body: { name } })
  if (error) { catError.value = 'Gagal atau nama sudah ada.'; catSaving.value = false; return }
  catSaving.value = false; showAddCat.value = false; newCatName.value = ''
  await fetchCategories()
}

function selectCategory(id: string) {
  selectedCategoryId.value = selectedCategoryId.value === id ? '' : id
  page.value = 1
  fetchTests()
}

async function fetchTests() {
  isLoading.value = true
  const { data } = await client.GET('/admin/tests', {
    params: {
      query: {
        page: page.value,
        limit: pageSize,
        ...(selectedCategoryId.value ? { category_id: selectedCategoryId.value } : {}),
      },
    },
  })
  if (data) { tests.value = data.data; total.value = data.total }
  isLoading.value = false
}

async function unpublish(testId: string) {
  busyId.value = testId; confirmUnpublishId.value = null; actionError.value = ''
  const { error } = await client.POST('/tests/{testId}/unpublish', { params: { path: { testId } } })
  if (error) { actionError.value = 'Gagal membatalkan publikasi.'; busyId.value = null; return }
  busyId.value = null; await fetchTests()
}

const totalPages = computed(() => Math.ceil(total.value / pageSize))
const maxAttempts = computed(() => Math.max(...tests.value.map(t => t.attempt_count), 1))

const diffMeta: Record<string, { label: string; color: string }> = {
  easy: { label: 'Mudah', color: 'var(--success)' },
  medium: { label: 'Sedang', color: 'var(--warning)' },
  hard: { label: 'Sulit', color: 'var(--danger)' },
}

onMounted(async () => { await fetchCategories(); await fetchTests() })
</script>

<template>
  <div class="tv" :class="{ ready: !isLoading }">

    <!-- ── Header ── -->
    <div class="tv-head">
      <div>
        <h1 class="tv-title">Semua Ujian</h1>
        <p class="tv-subtitle">{{ total.toLocaleString('id-ID') }} ujian terdaftar</p>
      </div>
    </div>

    <!-- ── Error ── -->
    <div v-if="actionError" class="action-error">{{ actionError }}</div>

    <div class="tv-layout">
      <!-- ═══════════════ SIDEBAR: Categories ═══════════════ -->
      <aside class="tv-sidebar">
        <div class="sb-section">
          <div class="sb-head">
            <span class="sb-label">Kategori</span>
            <button class="sb-add-btn" @click="showAddCat = !showAddCat" :title="showAddCat ? 'Batal' : 'Tambah'">
              {{ showAddCat ? '×' : '+' }}
            </button>
          </div>

          <!-- inline create -->
          <div v-if="showAddCat" class="sb-create">
            <input
              v-model="newCatName"
              type="text"
              placeholder="Nama kategori..."
              class="sb-input"
              maxlength="100"
              @keyup.enter="createCategory"
            />
            <p v-if="catError" class="sb-err">{{ catError }}</p>
            <button class="sb-save" :disabled="catSaving || !newCatName.trim()" @click="createCategory">
              {{ catSaving ? '...' : 'Simpan' }}
            </button>
          </div>

          <!-- category list -->
          <div class="sb-list">
            <button
              class="sb-item"
              :class="{ 'sb-item--active': selectedCategoryId === '' }"
              @click="selectCategory('')"
            >
              <span class="sb-item-name">Semua</span>
              <span class="sb-item-count">{{ categories.reduce((s,c) => s + c.test_count, 0) }}</span>
            </button>
            <button
              v-for="cat in categories"
              :key="cat.id"
              class="sb-item"
              :class="{ 'sb-item--active': selectedCategoryId === cat.id }"
              @click="selectCategory(cat.id)"
            >
              <span class="sb-item-name">{{ cat.name }}</span>
              <span class="sb-item-count">{{ cat.test_count }}</span>
            </button>
          </div>
        </div>
      </aside>

      <!-- ═══════════════ MAIN: Test table ═══════════════ -->
      <main class="tv-main">
        <!-- filter indicator -->
        <div v-if="selectedCategoryId" class="filter-bar">
          <span>
            Difilter: <strong>{{ categories.find(c => c.id === selectedCategoryId)?.name }}</strong>
          </span>
          <button class="filter-clear" @click="selectCategory('')">Hapus Filter</button>
        </div>

        <!-- Loading -->
        <div v-if="isLoading" class="table-wrap">
          <div v-for="i in 8" :key="i" class="row-skeleton">
            <div class="sk-text"><div class="sk-block sk-title" /></div>
            <div class="sk-block sk-badge" />
            <div class="sk-block sk-badge" />
            <div class="sk-block sk-badge" />
          </div>
        </div>

        <!-- Empty -->
        <div v-else-if="tests.length === 0" class="table-wrap">
          <div class="empty-state">
            <svg viewBox="0 0 20 20" fill="currentColor" width="28" height="28"><path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z" clip-rule="evenodd"/></svg>
            <p>{{ selectedCategoryId ? 'Tidak ada ujian dalam kategori ini.' : 'Belum ada ujian.' }}</p>
          </div>
        </div>

        <!-- Table -->
        <div v-else class="table-wrap">
          <table class="test-table">
            <thead>
              <tr>
                <th class="th-title">Judul</th>
                <th>Kategori</th>
                <th>Kesulitan</th>
                <th>Status</th>
                <th>Percobaan</th>
                <th class="th-actions">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="t in tests" :key="t.id" class="test-row">
                <td class="td-title"><span class="test-title">{{ t.title }}</span></td>
                <td><span class="cat-badge">{{ t.category_name }}</span></td>
                <td>
                  <span class="diff-badge" :style="`color: ${diffMeta[t.difficulty]?.color ?? 'inherit'}`">
                    {{ diffMeta[t.difficulty]?.label ?? t.difficulty }}
                  </span>
                </td>
                <td>
                  <span class="status-badge" :class="`status-badge--${t.status}`">
                    <span class="status-dot" />{{ t.status === 'published' ? 'Dipublikasi' : 'Draft' }}
                  </span>
                </td>
                <td class="td-attempts">
                  <div class="attempts-cell">
                    <span class="attempts-num">{{ t.attempt_count.toLocaleString('id-ID') }}</span>
                    <div class="attempts-track"><div class="attempts-fill" :style="`width: ${(t.attempt_count / maxAttempts) * 100}%`" /></div>
                  </div>
                </td>
                <td class="td-actions">
                  <template v-if="t.status === 'published'">
                    <template v-if="confirmUnpublishId === t.id">
                      <span class="confirm-label">Batalkan publikasi?</span>
                      <button class="btn-sm btn-sm--warning" @click="unpublish(t.id)" :disabled="busyId === t.id">{{ busyId === t.id ? '···' : 'Ya' }}</button>
                      <button class="btn-sm btn-sm--ghost" @click="confirmUnpublishId = null">Batal</button>
                    </template>
                    <button v-else class="btn-sm btn-sm--outline-warning" @click="confirmUnpublishId = t.id" :disabled="!!busyId">Batalkan Publikasi</button>
                  </template>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="pagination">
          <button class="pg-btn" :disabled="page === 1" @click="page--; fetchTests()">
            <svg viewBox="0 0 16 16" fill="currentColor" width="14" height="14"><path fill-rule="evenodd" d="M9.78 11.22a.75.75 0 01-1.06 0L5.47 7.97a.75.75 0 010-1.06l3.25-3.25a.75.75 0 011.06 1.06L7.06 7.44l2.72 2.72a.75.75 0 010 1.06z" clip-rule="evenodd"/></svg>
          </button>
          <button v-for="p in totalPages" :key="p" class="pg-btn" :class="{ 'pg-btn--active': p === page }" @click="page = p; fetchTests()">{{ p }}</button>
          <button class="pg-btn" :disabled="page >= totalPages" @click="page++; fetchTests()">
            <svg viewBox="0 0 16 16" fill="currentColor" width="14" height="14"><path fill-rule="evenodd" d="M6.22 4.78a.75.75 0 011.06 0l3.25 3.25a.75.75 0 010 1.06L7.28 12.34a.75.75 0 01-1.06-1.06L9.94 8.5 7.22 5.78a.75.75 0 010-1.06z" clip-rule="evenodd"/></svg>
          </button>
        </div>
      </main>
    </div>

  </div>
</template>

<style scoped>
@keyframes fade-up {
  from { opacity: 0; transform: translateY(6px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.tv { display: flex; flex-direction: column; gap: 1.5rem; }

/* ── Header ── */
.tv-head { animation: fade-up 0.3s ease both; }
.tv-title { margin: 0; font-size: 1.875rem; font-weight: 800; color: var(--text-heading); letter-spacing: -0.03em; line-height: 1; }
.tv-subtitle { margin: 0.3rem 0 0; font-size: 0.72rem; color: var(--text-muted); font-variant-numeric: tabular-nums; }

/* ── Error ── */
.action-error { padding: 0.7rem 1rem; background: var(--danger-bg); color: var(--danger-text); border-radius: 8px; font-size: 0.825rem; border-left: 3px solid var(--danger); }

/* ── Two-column layout ── */
.tv-layout { display: flex; gap: 1.5rem; align-items: flex-start; }
.tv-sidebar { width: 220px; flex-shrink: 0; }
.tv-main    { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 0.75rem; }

/* ── Sidebar ── */
.sb-section { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; overflow: hidden; }
.sb-head { display: flex; align-items: center; justify-content: space-between; padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); }
.sb-label { font-size: 0.62rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-muted); }
.sb-add-btn { width: 24px; height: 24px; border-radius: 6px; border: 1px solid var(--border); background: transparent; font-size: 1rem; font-weight: 700; color: var(--text-muted); cursor: pointer; display: flex; align-items: center; justify-content: center; line-height: 1; }
.sb-add-btn:hover { background: var(--accent); color: #fff; border-color: var(--accent); }

/* inline create */
.sb-create { display: flex; flex-direction: column; gap: 0.4rem; padding: 0.625rem 1rem; border-bottom: 1px solid var(--border); }
.sb-input { padding: 0.4rem 0.5rem; border-radius: 6px; border: 1px solid var(--border); background: var(--bg-input); color: var(--text-primary); font-size: 0.75rem; }
.sb-err { margin: 0; font-size: 0.65rem; color: var(--danger); }
.sb-save { padding: 0.3rem 0.6rem; border-radius: 6px; border: none; background: var(--accent); color: #fff; font-size: 0.7rem; font-weight: 700; cursor: pointer; align-self: flex-end; }
.sb-save:disabled { opacity: 0.5; cursor: not-allowed; }

/* category list */
.sb-list { display: flex; flex-direction: column; }
.sb-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.55rem 1rem; border: none; border-bottom: 1px solid var(--bg-input);
  background: transparent; cursor: pointer; font-size: 0.78rem; color: var(--text-primary);
  transition: background 0.1s; width: 100%; text-align: left;
}
.sb-item:last-child { border-bottom: none; }
.sb-item:hover { background: color-mix(in srgb, var(--accent) 3%, transparent); }
.sb-item--active { background: color-mix(in srgb, var(--accent) 8%, transparent); font-weight: 700; color: var(--accent); }
.sb-item--active:hover { background: color-mix(in srgb, var(--accent) 10%, transparent); }
.sb-item-name { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.sb-item-count { font-size: 0.65rem; font-weight: 700; color: var(--text-muted); background: var(--bg-input); padding: 0.1rem 0.4rem; border-radius: 99px; min-width: 20px; text-align: center; }

/* ── Filter bar ── */
.filter-bar { display: flex; align-items: center; justify-content: space-between; padding: 0.4rem 0.75rem; background: color-mix(in srgb, var(--accent) 6%, transparent); border-radius: 8px; font-size: 0.72rem; color: var(--text-muted); }
.filter-clear { padding: 0.2rem 0.5rem; border-radius: 4px; border: 1px solid var(--border); background: transparent; font-size: 0.65rem; cursor: pointer; color: var(--text-muted); }
.filter-clear:hover { color: var(--danger); border-color: var(--danger); }

/* ── Table wrap ── */
.table-wrap { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 14px; overflow: hidden; animation: fade-up 0.3s 0.1s ease both; }

/* ── Skeleton ── */
.row-skeleton { display: flex; align-items: center; gap: 1rem; padding: 0.875rem 1.25rem; border-bottom: 1px solid var(--border); }
.row-skeleton:last-child { border-bottom: none; }
.sk-text { flex: 1; }
.sk-block { border-radius: 4px; background: linear-gradient(90deg, var(--border) 25%, var(--bg-input) 50%, var(--border) 75%); background-size: 200% 100%; animation: shimmer 1.4s infinite; }
.sk-title { height: 0.875rem; width: 55%; }
.sk-badge { height: 1.6rem; width: 70px; border-radius: 6px; }

/* ── Empty ── */
.empty-state { display: flex; flex-direction: column; align-items: center; gap: 0.625rem; padding: 3.5rem 2rem; color: var(--text-muted); font-size: 0.875rem; text-align: center; }
.empty-state p { margin: 0; }

/* ── Test table ── */
.test-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
.test-table th { padding: 0.75rem 1.25rem; text-align: left; font-size: 0.67rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.07em; border-bottom: 1px solid var(--border); white-space: nowrap; }
.th-title { width: 30%; }
.th-actions { text-align: right; }
.test-table td { padding: 0.75rem 1.25rem; border-bottom: 1px solid var(--bg-input); vertical-align: middle; }
.test-table tbody tr:last-child td { border-bottom: none; }
.test-row { transition: background 0.12s; }
.test-row:hover td { background: color-mix(in srgb, var(--accent) 4%, transparent); }
.td-title { max-width: 0; }
.test-title { display: block; font-weight: 600; color: var(--text-heading); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 280px; }
.cat-badge { display: inline-block; padding: 0.22rem 0.65rem; border-radius: 6px; font-size: 0.7rem; font-weight: 700; background: color-mix(in srgb, var(--accent) 12%, transparent); color: var(--accent); white-space: nowrap; }
.diff-badge { font-size: 0.78rem; font-weight: 700; text-transform: capitalize; }
.status-badge { display: inline-flex; align-items: center; gap: 0.375rem; font-size: 0.78rem; font-weight: 600; }
.status-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.status-badge--published .status-dot { background: var(--success); box-shadow: 0 0 0 2px color-mix(in srgb, var(--success) 25%, transparent); }
.status-badge--draft .status-dot { background: var(--text-muted); }
.status-badge--published { color: var(--success); }
.status-badge--draft { color: var(--text-muted); }
.td-attempts { min-width: 110px; }
.attempts-cell { display: flex; align-items: center; gap: 0.625rem; }
.attempts-num { font-family: monospace; font-size: 0.875rem; font-weight: 700; color: var(--accent); font-variant-numeric: tabular-nums; min-width: 2rem; text-align: right; }
.attempts-track { flex: 1; height: 4px; background: var(--border); border-radius: 99px; overflow: hidden; min-width: 40px; }
.attempts-fill { height: 100%; background: var(--accent); border-radius: 99px; opacity: 0.6; transition: width 0.9s cubic-bezier(0.25, 0.8, 0.25, 1); }
.td-actions { text-align: right; white-space: nowrap; }
.confirm-label { font-size: 0.75rem; color: var(--warning); font-weight: 600; margin-right: 0.25rem; }
.btn-sm { padding: 0.32rem 0.7rem; border-radius: 6px; font-size: 0.75rem; font-weight: 700; cursor: pointer; border: none; transition: all 0.12s; margin-left: 0.25rem; }
.btn-sm:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-sm--warning { background: var(--warning); color: #000; }
.btn-sm--ghost { background: transparent; border: 1px solid var(--border); color: var(--text-muted); }
.btn-sm--outline-warning { background: transparent; border: 1px solid var(--warning); color: var(--warning); }
.btn-sm--outline-warning:hover { background: color-mix(in srgb, var(--warning) 10%, transparent); }

/* ── Pagination ── */
.pagination { display: flex; align-items: center; gap: 0.25rem; justify-content: center; }
.pg-btn { width: 32px; height: 32px; border-radius: 6px; border: 1px solid var(--border); background: var(--bg-surface); font-size: 0.75rem; font-weight: 700; cursor: pointer; display: flex; align-items: center; justify-content: center; color: var(--text-muted); }
.pg-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.pg-btn--active { background: var(--accent); color: #fff; border-color: var(--accent); }

@media (max-width: 900px) {
  .tv-layout { flex-direction: column; }
  .tv-sidebar { width: 100%; }
}
</style>
