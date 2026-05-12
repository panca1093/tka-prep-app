<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type ReviewItem = components['schemas']['ReviewItemResponse']
type ResultDetail = components['schemas']['ResultDetailResponse']

const route = useRoute()
const resultId = route.params.resultId as string

const items = ref<ReviewItem[]>([])
const result = ref<ResultDetail | null>(null)
const filter = ref<'all' | 'correct' | 'wrong' | 'blank'>('all')
const isLoading = ref(true)

onMounted(async () => {
  const [resDetail, resReview] = await Promise.all([
    client.GET('/results/{resultId}', { params: { path: { resultId } } }),
    client.GET('/results/{resultId}/review', { params: { path: { resultId } } }),
  ])
  if (resDetail.data) result.value = resDetail.data
  if (resReview.data) items.value = resReview.data.data
  isLoading.value = false
})

const filtered = computed(() =>
  filter.value === 'all' ? items.value : items.value.filter((i) => i.status === filter.value)
)

const statusColor: Record<string, string> = {
  correct: '#22c55e',
  wrong: '#ef4444',
  blank: '#94a3b8',
}

const statusLabel: Record<string, string> = {
  correct: '✓ Correct',
  wrong: '✗ Wrong',
  blank: '— Blank',
}
</script>

<template>
  <div>
    <div class="page-header">
      <h1 class="page-title">Answer Review</h1>
      <RouterLink :to="`/results/${resultId}`" class="back-link">← Back to Result</RouterLink>
    </div>

    <div v-if="isLoading" class="loading">Loading…</div>

    <template v-else>
      <div class="filters">
        <button
          v-for="f in ['all', 'correct', 'wrong', 'blank']"
          :key="f"
          class="filter-btn"
          :class="{ active: filter === f }"
          @click="filter = f as typeof filter"
        >
          {{ f === 'all' ? `All (${items.length})` : `${f.charAt(0).toUpperCase() + f.slice(1)} (${items.filter(i => i.status === f).length})` }}
        </button>
      </div>

      <div class="review-list">
        <div v-for="item in filtered" :key="item.question_id" class="review-card">
          <div class="review-header">
            <span class="q-num">Q{{ item.order_index + 1 }}</span>
            <span class="q-status" :style="{ color: statusColor[item.status] }">
              {{ statusLabel[item.status] }}
            </span>
            <span class="q-topic">{{ item.topic_name }}</span>
          </div>

          <p class="q-text">{{ item.text }}</p>

          <div class="options">
            <div
              v-for="opt in item.options"
              :key="opt.id"
              class="opt-row"
              :class="{
                correct: opt.is_correct,
                selected: opt.id === item.selected_option_id && !opt.is_correct,
              }"
            >
              <span class="opt-label">{{ opt.label }}</span>
              <span class="opt-text">{{ opt.text }}</span>
              <span v-if="opt.is_correct" class="opt-badge correct-badge">Correct</span>
              <span v-if="opt.id === item.selected_option_id && !opt.is_correct" class="opt-badge wrong-badge">Your answer</span>
            </div>
          </div>

          <div v-if="item.explanation" class="explanation">
            <span class="exp-label">Explanation:</span> {{ item.explanation }}
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
.page-title { margin: 0; font-size: 1.5rem; font-weight: 800; }
.back-link { font-size: 0.875rem; color: #4f8ef7; }
.loading { color: #94a3b8; }

.filters { display: flex; gap: 0.5rem; margin-bottom: 1.5rem; flex-wrap: wrap; }
.filter-btn {
  padding: 0.4rem 1rem; border-radius: 20px;
  border: 1px solid #1e2a45; background: transparent;
  color: #94a3b8; font-size: 0.8rem; cursor: pointer; transition: all 0.15s;
}
.filter-btn:hover { border-color: #4f8ef7; color: #f1f5f9; }
.filter-btn.active { background: #1a2f5c; border-color: #4f8ef7; color: #4f8ef7; }

.review-list { display: flex; flex-direction: column; gap: 1rem; }

.review-card {
  background: #141c2e; border: 1px solid #1e2a45; border-radius: 12px;
  padding: 1.25rem; display: flex; flex-direction: column; gap: 0.875rem;
}

.review-header { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; }
.q-num { font-weight: 700; font-size: 0.85rem; color: #4f8ef7; }
.q-status { font-size: 0.8rem; font-weight: 600; }
.q-topic { margin-left: auto; font-size: 0.75rem; color: #64748b; }

.q-text { margin: 0; font-size: 0.95rem; line-height: 1.65; color: #e2e8f0; }

.options { display: flex; flex-direction: column; gap: 0.5rem; }

.opt-row {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.625rem 0.875rem;
  border-radius: 8px; border: 1px solid #1e2a45; background: #0d1424;
  font-size: 0.875rem;
}
.opt-row.correct { border-color: #22c55e; background: rgba(34, 197, 94, 0.06); }
.opt-row.selected { border-color: #ef4444; background: rgba(239, 68, 68, 0.06); }

.opt-label {
  width: 1.5rem; height: 1.5rem; border-radius: 50%;
  background: #1e2a45; display: flex; align-items: center; justify-content: center;
  font-size: 0.75rem; font-weight: 700; flex-shrink: 0;
}
.opt-row.correct .opt-label { background: #22c55e; color: #000; }
.opt-row.selected .opt-label { background: #ef4444; }
.opt-text { flex: 1; }
.opt-badge { font-size: 0.7rem; font-weight: 700; padding: 0.15rem 0.5rem; border-radius: 4px; flex-shrink: 0; }
.correct-badge { background: rgba(34,197,94,0.15); color: #22c55e; }
.wrong-badge { background: rgba(239,68,68,0.15); color: #ef4444; }

.explanation {
  font-size: 0.825rem; line-height: 1.6; color: #94a3b8;
  background: #0d1424; border-radius: 8px; padding: 0.75rem;
}
.exp-label { font-weight: 600; color: #64748b; }
</style>
