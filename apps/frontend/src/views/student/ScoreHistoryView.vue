<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Result = components['schemas']['TestResultResponse']

const results    = ref<Result[]>([])
const isLoading  = ref(true)
const filter     = ref<'all' | 'correct' | 'wrong' | 'blank'>('all')

onMounted(async () => {
  const { data } = await client.GET('/results')
  if (data) results.value = data.data
  isLoading.value = false
})

const filtered = computed(() => {
  if (filter.value === 'all') return results.value
  return results.value.filter((r) => {
    if (filter.value === 'correct') return r.correct_count > r.wrong_count
    if (filter.value === 'wrong')   return r.wrong_count > r.correct_count
    if (filter.value === 'blank')   return r.blank_count > 0
    return true
  })
})

function scoreClass(s: number) {
  return s >= 70 ? 'good' : s >= 40 ? 'mid' : 'low'
}

function pct(r: Result) {
  const total = r.correct_count + r.wrong_count + r.blank_count
  return total > 0 ? Math.round((r.correct_count / total) * 100) : 0
}

function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString('id-ID', {
    weekday: 'short', day: 'numeric', month: 'short', year: 'numeric',
  })
}

function fmtTime(iso: string) {
  return new Date(iso).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="history-page">
    <div class="page-header">
      <h1>Riwayat Nilai</h1>
      <p class="sub">Semua hasil ujian yang pernah kamu kerjakan</p>
    </div>

    <!-- Stats row -->
    <div v-if="results.length > 0" class="stats-row">
      <div class="stat">
        <div class="stat-num">{{ results.length }}</div>
        <div class="stat-label">Ujian</div>
      </div>
      <div class="stat">
        <div class="stat-num">
          {{ Math.max(...results.map(r => r.total_score)).toFixed(1) }}
        </div>
        <div class="stat-label">Nilai Tertinggi</div>
      </div>
      <div class="stat">
        <div class="stat-num">
          {{ (results.reduce((s, r) => s + r.total_score, 0) / results.length).toFixed(1) }}
        </div>
        <div class="stat-label">Rata-rata</div>
      </div>
    </div>

    <!-- Filter -->
    <div v-if="results.length > 0" class="filter-bar">
      <button
        v-for="f in ([
          { key: 'all', label: 'Semua' },
          { key: 'correct', label: 'Lulus' },
          { key: 'wrong', label: 'Kurang' },
          { key: 'blank', label: 'Ada Kosong' },
        ] as const)"
        :key="f.key"
        :class="{ active: filter === f.key }"
        class="filter-btn"
        @click="filter = f.key"
      >{{ f.label }}</button>
    </div>

    <!-- Loading -->
    <div v-if="isLoading" class="loading">Memuat...</div>

    <!-- Empty -->
    <div v-else-if="results.length === 0" class="empty">
      <div class="empty-icon">📝</div>
      <h3>Belum ada riwayat</h3>
      <p>Kamu belum mengerjakan ujian apapun. Yuk mulai!</p>
      <RouterLink to="/tests" class="btn-primary">Lihat Ujian</RouterLink>
    </div>

    <!-- Empty after filter -->
    <div v-else-if="filtered.length === 0" class="empty">
      <p>Tidak ada hasil dengan filter "{{ filter }}".</p>
      <button class="btn-ghost" @click="filter = 'all'">Tampilkan semua</button>
    </div>

    <!-- List -->
    <div v-else class="result-list">
      <RouterLink
        v-for="r in filtered"
        :key="r.id"
        :to="`/results/${r.id}`"
        class="result-card"
      >
        <div class="card-top">
          <h3 class="card-title">{{ r.test_title }}</h3>
          <span class="card-score" :class="scoreClass(r.total_score)">
            {{ r.total_score.toFixed(1) }}
          </span>
        </div>
        <div class="card-bar">
          <div
            class="bar-fill"
            :class="scoreClass(r.total_score)"
            :style="`width:${Math.min(pct(r), 100)}%`"
          />
        </div>
        <div class="card-meta">
          <span>{{ fmtDate(r.completed_at) }} · {{ fmtTime(r.completed_at) }}</span>
          <span class="card-counts">
            <span class="c-correct">{{ r.correct_count }} benar</span>
            <span v-if="r.wrong_count > 0" class="c-wrong">{{ r.wrong_count }} salah</span>
            <span v-if="r.blank_count > 0" class="c-blank">{{ r.blank_count }} kosong</span>
          </span>
        </div>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
.history-page {
  max-width: 720px;
  margin: 0 auto;
  padding: 24px;
}

.page-header h1 { font-size: 1.5rem; font-weight: 700; margin: 0; color: var(--text-primary); }
.page-header .sub { margin: 4px 0 0; color: var(--text-muted); font-size: 0.9rem; }

.stats-row { display: flex; gap: 16px; margin: 24px 0 20px; }
.stat {
  flex: 1; background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 12px; padding: 16px; text-align: center;
}
.stat-num { font-size: 1.6rem; font-weight: 700; color: var(--accent); }
.stat-label { font-size: 0.8rem; color: var(--text-muted); margin-top: 2px; }

.filter-bar { display: flex; gap: 8px; margin-bottom: 20px; }
.filter-btn {
  padding: 6px 14px; border-radius: 20px; border: 1px solid var(--border);
  background: var(--bg-card); color: var(--text-muted); font-size: 0.82rem;
  cursor: pointer; transition: all .15s;
}
.filter-btn:hover { border-color: var(--accent); color: var(--accent); }
.filter-btn.active { background: var(--accent); border-color: var(--accent); color: #fff; }

.loading, .empty { text-align: center; padding: 48px 0; color: var(--text-muted); }
.empty-icon { font-size: 3rem; margin-bottom: 8px; }
.empty h3 { margin: 0 0 4px; color: var(--text-primary); }
.empty p { margin: 0 0 16px; }

.result-list { display: flex; flex-direction: column; gap: 12px; }
.result-card {
  display: block; background: var(--bg-card); border: 1px solid var(--border);
  border-radius: 12px; padding: 16px; text-decoration: none;
  transition: border-color .15s, box-shadow .15s;
}
.result-card:hover { border-color: var(--accent); box-shadow: 0 2px 12px rgba(0,0,0,.06); }

.card-top { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; }
.card-title { margin: 0; font-size: 1rem; font-weight: 600; color: var(--text-primary); }
.card-score { font-size: 1.3rem; font-weight: 700; white-space: nowrap; }
.card-score.good { color: var(--success); }
.card-score.mid  { color: var(--warning); }
.card-score.low  { color: var(--danger); }

.card-bar {
  height: 4px; background: var(--bg-input); border-radius: 2px;
  margin: 10px 0 8px; overflow: hidden;
}
.bar-fill { height: 100%; border-radius: 2px; transition: width .3s; }
.bar-fill.good { background: var(--success); }
.bar-fill.mid  { background: var(--warning); }
.bar-fill.low  { background: var(--danger); }

.card-meta { display: flex; justify-content: space-between; font-size: 0.78rem; color: var(--text-muted); }
.card-counts { display: flex; gap: 8px; }
.c-correct { color: var(--success); }
.c-wrong   { color: var(--danger); }
.c-blank   { color: var(--text-muted); }

.btn-primary {
  display: inline-block; padding: 8px 20px; background: var(--accent);
  color: #fff; border-radius: 8px; text-decoration: none; font-weight: 600;
}
.btn-ghost {
  padding: 6px 16px; border: 1px solid var(--border); border-radius: 8px;
  background: transparent; color: var(--accent); cursor: pointer;
}
</style>
