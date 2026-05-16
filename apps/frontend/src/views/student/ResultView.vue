<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type ResultDetail = components['schemas']['ResultDetailResponse']

const route = useRoute()
const resultId = route.params.resultId as string

const result = ref<ResultDetail | null>(null)
const isLoading = ref(true)

onMounted(async () => {
  const { data } = await client.GET('/results/{resultId}', { params: { path: { resultId } } })
  if (data) result.value = data
  isLoading.value = false
})

const scoreColor = (pct: number) =>
  pct >= 75 ? 'var(--success)' : pct >= 50 ? 'var(--warning)' : 'var(--danger)'
</script>

<template>
  <div v-if="isLoading" class="loading">Loading result…</div>

  <div v-else-if="!result" class="error">Result not found.</div>

  <div v-else class="result-page">
    <div class="result-hero">
      <div class="test-title-label">{{ result.test_title }}</div>
      <div class="score-ring">
        <div class="score-value" :style="{ color: scoreColor(result.percentage) }">
          {{ result.total_score.toFixed(1) }}
        </div>
        <div class="score-label">Total Score</div>
      </div>
      <div class="score-pct" :style="{ color: scoreColor(result.percentage) }">
        {{ result.percentage.toFixed(1) }}%
      </div>
    </div>

    <div class="counts-row">
      <div class="count-card correct">
        <div class="count-val">{{ result.correct_count }}</div>
        <div class="count-label">Correct</div>
      </div>
      <div class="count-card wrong">
        <div class="count-val">{{ result.wrong_count }}</div>
        <div class="count-label">Wrong</div>
      </div>
      <div class="count-card blank">
        <div class="count-val">{{ result.blank_count }}</div>
        <div class="count-label">Blank</div>
      </div>
    </div>

    <div class="section">
      <h2>Per-Topic Breakdown</h2>
      <div class="breakdown-list">
        <div v-for="tb in result.topic_breakdown" :key="tb.topic_id" class="breakdown-row">
          <div class="tb-topic">{{ tb.topic_name }}</div>
          <div class="tb-bar-wrap">
            <div
              class="tb-bar"
              :style="{ width: `${tb.total ? (tb.correct_count / tb.total * 100) : 0}%` }"
            />
          </div>
          <div class="tb-counts">{{ tb.correct_count }}/{{ tb.total }}</div>
        </div>
      </div>
    </div>

    <div class="actions">
      <RouterLink :to="`/results/${result.id}/review`" class="btn-review">View Review →</RouterLink>
      <RouterLink to="/tests" class="btn-secondary">Take Another Test</RouterLink>
    </div>
  </div>
</template>

<style scoped>
.loading, .error { color: var(--text-muted); padding: 2rem; }

.result-page { max-width: 640px; margin: 0 auto; }

.result-hero {
  text-align: center;
  padding: 2rem 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.test-title-label { font-size: 1rem; font-weight: 700; color: var(--text-primary); margin-bottom: 0.5rem; }

.score-ring {
  width: 140px; height: 140px;
  border-radius: 50%;
  border: 6px solid var(--border);
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  background: var(--bg-surface);
}

.score-value { font-size: 2rem; font-weight: 900; line-height: 1; }
.score-label { font-size: 0.72rem; color: var(--text-muted); margin-top: 4px; }
.score-pct { font-size: 1.1rem; font-weight: 700; }

.counts-row { display: flex; gap: 1rem; margin-bottom: 2rem; }

.count-card {
  flex: 1; text-align: center;
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px;
  padding: 1rem;
}
.count-val { font-size: 1.75rem; font-weight: 800; }
.count-label { font-size: 0.75rem; color: var(--text-muted); margin-top: 4px; }
.count-card.correct .count-val { color: var(--success); }
.count-card.wrong .count-val { color: var(--danger); }
.count-card.blank .count-val { color: var(--text-muted); }

.section { margin-bottom: 2rem; }
.section h2 { font-size: 1rem; font-weight: 700; margin: 0 0 1rem; }

.breakdown-list { display: flex; flex-direction: column; gap: 0.625rem; }
.breakdown-row {
  display: grid; grid-template-columns: 180px 1fr 60px;
  align-items: center; gap: 0.75rem;
  padding: 0.625rem 0.875rem;
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 8px;
}
.tb-topic { font-size: 0.825rem; font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.tb-bar-wrap { background: var(--border); border-radius: 4px; height: 6px; overflow: hidden; }
.tb-bar { height: 100%; background: var(--accent); border-radius: 4px; transition: width 0.3s; }
.tb-counts { font-size: 0.78rem; color: var(--text-muted); text-align: right; }

.actions { display: flex; gap: 0.75rem; flex-wrap: wrap; }

.btn-review {
  padding: 0.65rem 1.5rem; border-radius: 8px;
  background: var(--accent); color: var(--text-on-accent);
  font-size: 0.875rem; font-weight: 600;
  text-decoration: none;
  transition: background 0.15s;
}
.btn-review:hover { background: var(--accent-hover); }

.btn-secondary {
  padding: 0.65rem 1.5rem; border-radius: 8px;
  border: 1px solid var(--border); color: var(--text-muted);
  font-size: 0.875rem;
  text-decoration: none;
  transition: all 0.15s;
}
.btn-secondary:hover { border-color: var(--accent); color: var(--text-primary); }
</style>
@media (max-width: 768px) { .stats-row { grid-template-columns: 1fr; } }
