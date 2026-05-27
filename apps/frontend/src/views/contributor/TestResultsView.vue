<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type ContributorResultEntry = components['schemas']['ContributorResultEntry']
type TestAnalyticsResponse = components['schemas']['TestAnalyticsResponse']
type TestDetailResponse = components['schemas']['TestDetailResponse']

const route = useRoute()
const testId = route.params.testId as string

const results = ref<ContributorResultEntry[]>([])
const analytics = ref<TestAnalyticsResponse | null>(null)
const test = ref<TestDetailResponse | null>(null)
const total = ref(0)
const page = ref(1)
const limit = 20
const isLoading = ref(false)
const error = ref('')

async function fetchResults() {
  isLoading.value = true
  error.value = ''
  const { data, error: err } = await client.GET('/tests/{testId}/results', {
    params: { path: { testId }, query: { page: page.value, limit } },
  })
  if (err) {
    error.value = (err as any)?.error?.message ?? 'Gagal memuat hasil.'
    isLoading.value = false
    return
  }
  if (data) {
    results.value = data.data
    total.value = data.total
    analytics.value = data.analytics ?? null
  }
  isLoading.value = false
}

async function fetchTest() {
  const { data } = await client.GET('/tests/{testId}', { params: { path: { testId } } })
  if (data) test.value = data
}

onMounted(async () => {
  await fetchTest()
  await fetchResults()
})

function prevPage() {
  if (page.value > 1) { page.value--; fetchResults() }
}
function nextPage() {
  if (page.value * limit < total.value) { page.value++; fetchResults() }
}

function fmtScore(v: number | null | undefined): string {
  if (v == null) return '—'
  return Number(v).toFixed(2)
}

function fmtAttempts(n: number | undefined): string {
  return n != null ? String(n) : '0'
}
</script>

<template>
  <div class="tr-page">
    <div class="page-header">
      <div>
        <RouterLink to="/contrib/tests" class="back-link">← Kembali ke Tes Saya</RouterLink>
        <h1 class="page-title">{{ test?.title ?? 'Hasil Tes' }}</h1>
      </div>
    </div>

    <!-- Analytics Summary -->
    <div v-if="analytics" class="analytics-bar">
      <div class="stat-item">
        <span class="stat-val">{{ fmtAttempts(analytics.total_attempts) }}</span>
        <span class="stat-lbl">Percobaan</span>
      </div>
      <div class="stat-item">
        <span class="stat-val">{{ fmtScore(analytics.avg_score) }}</span>
        <span class="stat-lbl">Rata-rata</span>
      </div>
      <div class="stat-item">
        <span class="stat-val">{{ fmtScore(analytics.max_score) }}</span>
        <span class="stat-lbl">Tertinggi</span>
      </div>
      <div class="stat-item">
        <span class="stat-val">{{ fmtScore(analytics.min_score) }}</span>
        <span class="stat-lbl">Terendah</span>
      </div>
    </div>

    <!-- Per-Topic Breakdown -->
    <div v-if="analytics?.per_topic && analytics.per_topic.length > 0" class="topic-section">
      <h2 class="section-title">Per-Topik</h2>
      <div class="topic-grid">
        <div v-for="tp in analytics.per_topic" :key="tp.topic_id" class="topic-card">
          <span class="topic-name">{{ tp.topic_name }}</span>
          <span class="topic-stat">{{ tp.question_count }} soal</span>
          <span class="topic-stat">Rata-rata benar: {{ tp.avg_correct_count != null ? Number(tp.avg_correct_count).toFixed(1) : '—' }}</span>
        </div>
      </div>
    </div>

    <p v-if="error" class="error-msg">{{ error }}</p>
    <div v-if="isLoading" class="loading">Memuat…</div>

    <div v-else-if="results.length === 0" class="empty-state">
      Belum ada siswa yang mengerjakan tes ini.
    </div>

    <div v-else>
      <h2 class="section-title">Hasil Siswa ({{ total }})</h2>
      <table class="results-table">
        <thead>
          <tr>
            <th>Nama</th>
            <th>Email</th>
            <th>Skor</th>
            <th>Benar</th>
            <th>Salah</th>
            <th>Kosong</th>
            <th>Selesai</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in results" :key="r.id">
            <td class="cell-name">{{ r.student_name }}</td>
            <td class="cell-email">{{ r.student_email }}</td>
            <td class="cell-score">{{ fmtScore(r.total_score) }}</td>
            <td>{{ r.correct_count }}</td>
            <td>{{ r.wrong_count }}</td>
            <td>{{ r.blank_count }}</td>
            <td class="cell-date">{{ new Date(r.completed_at).toLocaleDateString('id-ID') }}</td>
          </tr>
        </tbody>
      </table>

      <div class="pagination">
        <button :disabled="page <= 1" @click="prevPage">← Sebelumnya</button>
        <span class="page-num">Halaman {{ page }}</span>
        <button :disabled="page * limit >= total" @click="nextPage">Berikutnya →</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tr-page { max-width: 960px; }
.page-header { margin-bottom: 1.5rem; }
.back-link { font-size: 0.78rem; color: var(--accent); text-decoration: none; font-weight: 600; display: inline-block; margin-bottom: 0.35rem; }
.back-link:hover { text-decoration: underline; }
.page-title { margin: 0; font-size: 1.35rem; font-weight: 800; }

.analytics-bar {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 0.75rem;
  margin-bottom: 1.5rem;
}
.stat-item {
  background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: 10px; padding: 0.9rem 1rem; text-align: center;
}
.stat-val { font-size: 1.25rem; font-weight: 800; color: var(--text-heading); display: block; }
.stat-lbl { font-size: 0.68rem; color: var(--text-muted); margin-top: 0.15rem; display: block; }

.section-title { font-size: 0.85rem; font-weight: 700; margin: 1.25rem 0 0.75rem; color: var(--text-heading); }

.topic-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 0.5rem; margin-bottom: 1.5rem; }
.topic-card {
  background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: 8px; padding: 0.65rem 0.85rem;
  display: flex; flex-direction: column; gap: 0.15rem;
}
.topic-name { font-size: 0.78rem; font-weight: 700; color: var(--text-primary); }
.topic-stat { font-size: 0.68rem; color: var(--text-muted); }

.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: var(--danger-bg); color: var(--danger-text); font-size: 0.825rem; margin-bottom: 1rem; }
.loading { color: var(--text-muted); }
.empty-state { color: var(--text-muted); text-align: center; padding: 2rem; background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; }

.results-table { width: 100%; border-collapse: collapse; font-size: 0.82rem; }
.results-table th { text-align: left; padding: 0.55rem 0.65rem; font-size: 0.65rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); border-bottom: 1px solid var(--border); }
.results-table td { padding: 0.55rem 0.65rem; border-bottom: 1px solid var(--bg-input); color: var(--text-primary); }
.cell-name { font-weight: 600; }
.cell-email { font-size: 0.72rem; color: var(--text-muted); }
.cell-score { font-weight: 700; color: var(--accent); }
.cell-date { font-size: 0.72rem; color: var(--text-muted); white-space: nowrap; }

.pagination { display: flex; align-items: center; justify-content: center; gap: 1rem; margin-top: 1rem; }
.pagination button {
  padding: 0.4rem 0.8rem; border-radius: 6px; border: 1px solid var(--border);
  background: var(--bg-surface); color: var(--text-primary); font-size: 0.78rem; font-weight: 600;
  cursor: pointer; transition: all 0.12s;
}
.pagination button:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
.pagination button:disabled { opacity: 0.4; cursor: not-allowed; }
.page-num { font-size: 0.78rem; color: var(--text-muted); }
</style>
