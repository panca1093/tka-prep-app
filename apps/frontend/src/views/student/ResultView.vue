<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type ResultDetail = components['schemas']['ResultDetailResponse']

const route = useRoute()
const resultId = route.params.resultId as string
const result = ref<ResultDetail | null>(null)
const isLoading = ref(true)
const mounted = ref(false)

onMounted(async () => {
  const { data } = await client.GET('/results/{resultId}', { params: { path: { resultId } } })
  if (data) result.value = data
  isLoading.value = false
  requestAnimationFrame(() => { mounted.value = true })
})

onUnmounted(() => { mounted.value = false })

// SVG ring math (r=54, circumference≈339.3)
const R = 54
const C = computed(() => 2 * Math.PI * R)
const dashOffset = computed(() => {
  if (!result.value) return C.value
  const pct = Math.min(result.value.percentage / 100, 1)
  return mounted.value ? C.value * (1 - pct) : C.value
})

const ringColor = computed(() => {
  if (!result.value) return 'var(--accent)'
  const p = result.value.percentage
  return p >= 70 ? 'var(--success)' : p >= 40 ? 'var(--warning)' : 'var(--danger)'
})

const grade = computed(() => {
  if (!result.value) return ''
  const p = result.value.percentage
  if (p >= 85) return 'Luar Biasa!'
  if (p >= 70) return 'Bagus!'
  if (p >= 50) return 'Cukup Baik'
  return 'Perlu Latihan'
})

const gradeClass = computed(() => {
  if (!result.value) return ''
  const p = result.value.percentage
  return p >= 70 ? 'grade-good' : p >= 40 ? 'grade-mid' : 'grade-low'
})

function barWidth(tb: { correct_count: number; total: number }) {
  return tb.total ? (tb.correct_count / tb.total * 100) : 0
}
</script>

<template>
  <div v-if="isLoading" class="state-screen">
    <div class="loading-dots"><span/><span/><span/></div>
  </div>

  <div v-else-if="!result" class="state-screen err">Hasil tidak ditemukan.</div>

  <div v-else class="result-page">

    <!-- Hero score card -->
    <div class="score-card">
      <div class="score-card-left">
        <p class="test-label">{{ result.test_title }}</p>
        <div class="grade-badge" :class="gradeClass">{{ grade }}</div>
        <p class="score-pct" :style="{ color: ringColor }">
          {{ result.percentage.toFixed(1) }}%
        </p>
      </div>

      <div class="ring-wrap">
        <svg class="ring-svg" viewBox="0 0 120 120">
          <circle class="ring-bg" cx="60" cy="60" :r="R" />
          <circle
            class="ring-progress"
            cx="60" cy="60" :r="R"
            :stroke="ringColor"
            :stroke-dasharray="C"
            :stroke-dashoffset="dashOffset"
          />
        </svg>
        <div class="ring-inner">
          <div class="ring-score" :style="{ color: ringColor }">
            {{ result.total_score.toFixed(1) }}
          </div>
          <div class="ring-sub">poin</div>
        </div>
      </div>
    </div>

    <!-- Count pills -->
    <div class="count-row">
      <div class="count-pill count-pill--correct">
        <div class="cp-val">{{ result.correct_count }}</div>
        <div class="cp-label">Benar</div>
      </div>
      <div class="count-pill count-pill--wrong">
        <div class="cp-val">{{ result.wrong_count }}</div>
        <div class="cp-label">Salah</div>
      </div>
      <div class="count-pill count-pill--blank">
        <div class="cp-val">{{ result.blank_count }}</div>
        <div class="cp-label">Kosong</div>
      </div>
    </div>

    <!-- Per-topic breakdown -->
    <div v-if="result.topic_breakdown?.length" class="breakdown-card">
      <h3 class="breakdown-title">Hasil per Topik</h3>
      <div class="breakdown-list">
        <div v-for="tb in result.topic_breakdown" :key="tb.topic_id" class="breakdown-row">
          <div class="tb-name">{{ tb.topic_name }}</div>
          <div class="tb-bar-wrap">
            <div class="tb-bar" :style="{ width: barWidth(tb) + '%' }" />
          </div>
          <div class="tb-frac">{{ tb.correct_count }}<span>/{{ tb.total }}</span></div>
        </div>
      </div>
    </div>

    <!-- Actions -->
    <div class="actions">
      <RouterLink :to="`/results/${result.id}/review`" class="btn-primary">
        <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16"><path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z"/></svg>
        Review Jawaban
      </RouterLink>
      <RouterLink to="/tests" class="btn-secondary">
        <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16"><path fill-rule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clip-rule="evenodd"/></svg>
        Coba Ujian Lain
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
.state-screen {
  display: flex; align-items: center; justify-content: center;
  height: 60vh; color: var(--text-muted);
}
.err { color: var(--danger); }
.loading-dots { display: flex; gap: 6px; }
.loading-dots span {
  width: 8px; height: 8px; border-radius: 50%;
  background: var(--accent);
  animation: dot-bounce 1.4s infinite ease-in-out;
}
.loading-dots span:nth-child(2) { animation-delay: 0.2s; }
.loading-dots span:nth-child(3) { animation-delay: 0.4s; }
@keyframes dot-bounce { 0%,80%,100%{transform:scale(0.6);opacity:0.5}40%{transform:scale(1);opacity:1} }

.result-page {
  max-width: 600px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

/* ─── Score card ──────────────────────────────────────────────────────────── */
.score-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 20px;
  padding: 2rem 2rem 2rem 2.25rem;
}
.score-card-left { flex: 1; display: flex; flex-direction: column; gap: 0.5rem; }
.test-label {
  margin: 0;
  font-family: 'Cormorant Garamond', serif;
  font-size: 1.375rem;
  font-weight: 600;
  color: var(--text-heading);
  line-height: 1.25;
}
.grade-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 99px;
  font-size: 0.78rem;
  font-weight: 700;
}
.grade-good { background: var(--success-bg); color: var(--success); }
.grade-mid { background: var(--warning-bg); color: var(--warning); }
.grade-low { background: var(--danger-bg); color: var(--danger); }
.score-pct { margin: 0; font-size: 1rem; font-weight: 600; }

/* SVG ring */
.ring-wrap { position: relative; width: 120px; height: 120px; flex-shrink: 0; }
.ring-svg { width: 120px; height: 120px; transform: rotate(-90deg); }
.ring-bg { fill: none; stroke: var(--border); stroke-width: 8; }
.ring-progress {
  fill: none;
  stroke-width: 8;
  stroke-linecap: round;
  transition: stroke-dashoffset 1.2s cubic-bezier(0.16, 1, 0.3, 1);
}
.ring-inner {
  position: absolute; inset: 0;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
}
.ring-score {
  font-family: 'JetBrains Mono', monospace;
  font-size: 1.375rem;
  font-weight: 700;
  line-height: 1;
}
.ring-sub { font-size: 0.65rem; color: var(--text-muted); margin-top: 2px; }

/* ─── Count pills ─────────────────────────────────────────────────────────── */
.count-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.875rem; }
.count-pill {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1rem;
  text-align: center;
}
.cp-val { font-family: 'JetBrains Mono', monospace; font-size: 2rem; font-weight: 700; line-height: 1; }
.cp-label { font-size: 0.72rem; color: var(--text-muted); margin-top: 4px; font-weight: 500; }
.count-pill--correct { border-color: color-mix(in srgb, var(--success) 30%, var(--border)); }
.count-pill--correct .cp-val { color: var(--success); }
.count-pill--wrong { border-color: color-mix(in srgb, var(--danger) 30%, var(--border)); }
.count-pill--wrong .cp-val { color: var(--danger); }
.count-pill--blank .cp-val { color: var(--text-muted); }

/* ─── Breakdown ───────────────────────────────────────────────────────────── */
.breakdown-card {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 1.25rem 1.5rem;
}
.breakdown-title { margin: 0 0 1rem; font-size: 0.825rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: var(--text-muted); }
.breakdown-list { display: flex; flex-direction: column; gap: 0.75rem; }
.breakdown-row { display: grid; grid-template-columns: 1fr 2fr 2.5rem; align-items: center; gap: 0.75rem; }
.tb-name { font-size: 0.825rem; font-weight: 500; color: var(--text-primary); }
.tb-bar-wrap { background: var(--border); border-radius: 99px; height: 6px; overflow: hidden; }
.tb-bar { height: 100%; background: var(--accent); border-radius: 99px; transition: width 1s ease 0.3s; }
.tb-frac { font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; font-weight: 600; color: var(--text-muted); text-align: right; }
.tb-frac span { color: var(--border); }

/* ─── Actions ─────────────────────────────────────────────────────────────── */
.actions { display: flex; gap: 0.875rem; flex-wrap: wrap; }
.btn-primary {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.7rem 1.5rem; border-radius: 10px;
  background: var(--accent); color: #fff;
  font-size: 0.875rem; font-weight: 700;
  text-decoration: none;
  transition: background 0.15s, transform 0.15s;
}
.btn-primary:hover { background: var(--accent-hover); transform: translateY(-1px); }
.btn-secondary {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.7rem 1.5rem; border-radius: 10px;
  border: 1px solid var(--border);
  color: var(--text-muted);
  font-size: 0.875rem; font-weight: 600;
  text-decoration: none;
  transition: all 0.15s;
}
.btn-secondary:hover { border-color: var(--accent); color: var(--accent); }

@media (max-width: 640px) {
  .score-card { flex-direction: column; align-items: center; text-align: center; }
  .score-card-left { align-items: center; }
  .breakdown-row { grid-template-columns: 1fr 1.5fr 2.5rem; }
}
</style>
