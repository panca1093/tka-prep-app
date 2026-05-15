<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'
import RichTextViewer from '@/components/editor/RichTextViewer.vue'
import { resolveUrl } from '@/utils/url'

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

function isPGKSelected(item: ReviewItem, optId: string): boolean {
  return (item.selected_option_ids ?? []).some((id) => id === optId)
}

function isPGKCorrect(item: ReviewItem, optId: string): boolean {
  return (item.correct_option_ids ?? []).some((id) => id === optId)
}

const statusColor: Record<string, string> = {
  correct: 'var(--success)',
  wrong: 'var(--danger)',
  blank: 'var(--text-muted)',
}

const statusLabel: Record<string, string> = {
  correct: '✓ Correct',
  wrong: '✗ Wrong',
  blank: '— Blank',
}

const typeLabel: Record<string, string> = {
  mcq: 'PG',
  multi_correct: 'PGK',
  true_false: 'B/S',
}

const typeClass: Record<string, string> = {
  mcq: 'type-mcq',
  multi_correct: 'type-pgk',
  true_false: 'type-bs',
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
            <span class="q-type-badge" :class="typeClass[item.question_type]">
              {{ typeLabel[item.question_type] ?? item.question_type }}
            </span>
            <span class="q-status" :style="{ color: statusColor[item.status] }">
              {{ statusLabel[item.status] }}
            </span>
            <span class="q-topic">{{ item.topic_name }}</span>
          </div>

          <div class="q-text-block">
            <RichTextViewer :html="item.text" class="q-text" />
            <img v-if="item.image_url" :src="resolveUrl(item.image_url)" class="q-img" alt="" />
          </div>

          <!-- MCQ: single select review -->
          <div v-if="item.question_type === 'mcq'" class="options">
            <div
              v-for="opt in item.options"
              :key="opt.id"
              class="opt-row"
              :class="{
                correct: opt.is_correct,
                'wrong-selected': opt.id === item.selected_option_id && !opt.is_correct,
              }"
            >
              <span class="opt-label">{{ opt.label }}</span>
              <div class="opt-body">
                <span class="opt-text"><RichTextViewer :html="opt.text" /></span>
                <img v-if="opt.image_url" :src="resolveUrl(opt.image_url)" class="opt-img" alt="" />
              </div>
              <span v-if="opt.is_correct" class="opt-badge correct-badge">Jawaban Benar</span>
              <span v-if="opt.id === item.selected_option_id && !opt.is_correct" class="opt-badge wrong-badge">Jawaban Kamu</span>
              <span v-if="opt.id === item.selected_option_id && opt.is_correct" class="opt-badge selected-correct-badge">Jawaban Kamu ✓</span>
            </div>
          </div>

          <!-- PGK: multi-correct review -->
          <div v-else-if="item.question_type === 'multi_correct'" class="options">
            <div class="pgk-legend">
              <span class="pgk-legend-item correct-legend">Jawaban Benar</span>
              <span class="pgk-legend-item selected-legend">Pilihan Kamu</span>
            </div>
            <div
              v-for="opt in item.options"
              :key="opt.id"
              class="opt-row pgk-opt-row"
              :class="{
                correct: isPGKCorrect(item, opt.id),
                'wrong-selected': isPGKSelected(item, opt.id) && !isPGKCorrect(item, opt.id),
                'missed-correct': !isPGKSelected(item, opt.id) && isPGKCorrect(item, opt.id),
              }"
            >
              <span class="opt-label pgk-label" :class="{ checked: isPGKSelected(item, opt.id), 'correct-check': isPGKCorrect(item, opt.id) }">
                {{ isPGKSelected(item, opt.id) ? '✓' : opt.label }}
              </span>
              <div class="opt-body">
                <span class="opt-text"><RichTextViewer :html="opt.text" /></span>
                <img v-if="opt.image_url" :src="resolveUrl(opt.image_url)" class="opt-img" alt="" />
              </div>
              <div class="pgk-badges">
                <span v-if="isPGKCorrect(item, opt.id)" class="opt-badge correct-badge">Benar</span>
                <span v-if="isPGKSelected(item, opt.id)" class="opt-badge" :class="isPGKCorrect(item, opt.id) ? 'selected-correct-badge' : 'wrong-badge'">Dipilih</span>
              </div>
            </div>
          </div>

          <!-- B/S: per-statement review -->
          <div v-else-if="item.question_type === 'true_false'" class="statements">
            <div
              v-for="(stmt, idx) in item.statements"
              :key="stmt.id"
              class="stmt-review-row"
              :class="{
                'stmt-correct': stmt.student_answer != null && stmt.student_answer === stmt.is_correct,
                'stmt-wrong': stmt.student_answer != null && stmt.student_answer !== stmt.is_correct,
                'stmt-blank': stmt.student_answer == null,
              }"
            >
              <div class="stmt-num">{{ idx + 1 }}</div>
              <div class="stmt-body">
                <div class="stmt-text"><RichTextViewer :html="stmt.text" /></div>
                <img v-if="stmt.image_url" :src="resolveUrl(stmt.image_url)" class="stmt-img" alt="" />
              </div>
              <div class="stmt-answers">
                <div class="answer-row">
                  <span class="answer-label">Jawaban:</span>
                  <span
                    class="answer-val"
                    :class="stmt.is_correct ? 'val-benar' : 'val-salah'"
                  >{{ stmt.is_correct ? 'Benar' : 'Salah' }}</span>
                </div>
                <div class="answer-row">
                  <span class="answer-label">Kamu:</span>
                  <span
                    v-if="stmt.student_answer != null"
                    class="answer-val"
                    :class="stmt.student_answer ? 'val-benar' : 'val-salah'"
                  >{{ stmt.student_answer ? 'Benar' : 'Salah' }}</span>
                  <span v-else class="answer-val val-blank">—</span>
                </div>
              </div>
            </div>
          </div>

          <div v-if="item.explanation" class="explanation">
            <span class="exp-label">Explanation:</span> <RichTextViewer :html="item.explanation ?? ''" />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 1.5rem; }
.page-title { margin: 0; font-size: 1.5rem; font-weight: 800; }
.back-link { font-size: 0.875rem; color: var(--accent); }
.loading { color: var(--text-muted); }

.filters { display: flex; gap: 0.5rem; margin-bottom: 1.5rem; flex-wrap: wrap; }
.filter-btn {
  padding: 0.4rem 1rem; border-radius: 20px;
  border: 1px solid var(--border); background: transparent;
  color: var(--text-muted); font-size: 0.8rem; cursor: pointer; transition: all 0.15s;
}
.filter-btn:hover { border-color: var(--accent); color: var(--text-primary); }
.filter-btn.active { background: var(--bg-active-nav); border-color: var(--accent); color: var(--accent); }

.review-list { display: flex; flex-direction: column; gap: 1rem; }

.review-card {
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px;
  padding: 1.25rem; display: flex; flex-direction: column; gap: 0.875rem;
}

.review-header { display: flex; align-items: center; gap: 0.625rem; flex-wrap: wrap; }
.q-num { font-weight: 700; font-size: 0.85rem; color: var(--accent); }
.q-type-badge {
  font-size: 0.68rem; font-weight: 800; padding: 0.15rem 0.45rem;
  border-radius: 4px; border: 1px solid; text-transform: uppercase; letter-spacing: 0.03em;
}
.type-mcq { color: var(--accent); border-color: color-mix(in srgb, var(--accent) 27%, transparent); background: color-mix(in srgb, var(--accent) 9%, transparent); }
.type-pgk { color: var(--accent); border-color: color-mix(in srgb, var(--accent) 27%, transparent); background: color-mix(in srgb, var(--accent) 9%, transparent); }
.type-bs { color: var(--success); border-color: color-mix(in srgb, var(--success) 27%, transparent); background: color-mix(in srgb, var(--success) 9%, transparent); }
.q-status { font-size: 0.8rem; font-weight: 600; }
.q-topic { margin-left: auto; font-size: 0.75rem; color: var(--text-muted); }

.q-text { margin: 0; font-size: 0.95rem; line-height: 1.65; color: var(--text-primary); }

/* MCQ / PGK options */
.options { display: flex; flex-direction: column; gap: 0.5rem; }

.opt-row {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.625rem 0.875rem;
  border-radius: 8px; border: 1px solid var(--border); background: var(--bg-input);
  font-size: 0.875rem;
}
.opt-row.correct { border-color: var(--success); background: rgba(34, 197, 94, 0.06); }
.opt-row.wrong-selected { border-color: var(--danger); background: rgba(239, 68, 68, 0.06); }
.opt-row.missed-correct { border-color: var(--success); border-style: dashed; background: rgba(34, 197, 94, 0.03); }

.opt-label {
  width: 1.5rem; height: 1.5rem; border-radius: 50%;
  background: var(--border); display: flex; align-items: center; justify-content: center;
  font-size: 0.75rem; font-weight: 700; flex-shrink: 0;
}
.opt-row.correct .opt-label:not(.pgk-label) { background: var(--success); color: #000; }
.opt-row.wrong-selected .opt-label:not(.pgk-label) { background: var(--danger); }

/* PGK label override */
.pgk-opt-row .pgk-label { border-radius: 4px; }
.pgk-opt-row .pgk-label.checked { background: var(--text-muted); }
.pgk-opt-row.correct .pgk-label.correct-check { background: var(--success); color: #000; }
.pgk-opt-row.wrong-selected .pgk-label.checked:not(.correct-check) { background: var(--danger); }

.opt-body { flex: 1; display: flex; flex-direction: column; gap: 0.375rem; }
.opt-text { }
.opt-img { max-width: 100%; max-height: 140px; border-radius: 6px; border: 1px solid var(--border); }
.q-text-block { display: flex; flex-direction: column; gap: 0; }
.q-img { max-width: 100%; max-height: 220px; border-radius: 8px; margin-top: 0.625rem; border: 1px solid var(--border); }
.opt-badge { font-size: 0.7rem; font-weight: 700; padding: 0.15rem 0.5rem; border-radius: 4px; flex-shrink: 0; }
.correct-badge { background: rgba(34,197,94,0.15); color: var(--success); }
.wrong-badge { background: rgba(239,68,68,0.15); color: var(--danger); }
.selected-correct-badge { background: rgba(34,197,94,0.25); color: var(--success); border: 1px solid rgba(34,197,94,0.27); }

.pgk-legend { display: flex; gap: 1rem; margin-bottom: 0.25rem; }
.pgk-legend-item { font-size: 0.72rem; color: var(--text-muted); }
.correct-legend::before { content: '◼ '; color: var(--success); }
.selected-legend::before { content: '◼ '; color: var(--danger); }

.pgk-badges { display: flex; gap: 0.375rem; flex-shrink: 0; }

/* B/S statements */
.statements { display: flex; flex-direction: column; gap: 0.625rem; }

.stmt-review-row {
  display: flex; align-items: flex-start; gap: 0.875rem;
  padding: 0.875rem 1rem;
  border-radius: 10px; border: 1px solid var(--border); background: var(--bg-input);
}
.stmt-review-row.stmt-correct { border-color: var(--success); background: rgba(34,197,94,0.05); }
.stmt-review-row.stmt-wrong { border-color: var(--danger); background: rgba(239,68,68,0.05); }
.stmt-review-row.stmt-blank { border-color: var(--border); border-style: dashed; }

.stmt-num {
  width: 1.5rem; height: 1.5rem; border-radius: 50%; background: var(--border);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.72rem; font-weight: 700; color: var(--text-muted); flex-shrink: 0;
  margin-top: 0.1rem;
}

.stmt-body { flex: 1; display: flex; flex-direction: column; gap: 0.375rem; }
.stmt-text { font-size: 0.875rem; line-height: 1.5; color: var(--text-primary); }
.stmt-img { max-width: 100%; max-height: 120px; border-radius: 6px; border: 1px solid var(--border); }

.stmt-answers { display: flex; flex-direction: column; gap: 0.25rem; flex-shrink: 0; min-width: 110px; }
.answer-row { display: flex; align-items: center; gap: 0.375rem; font-size: 0.75rem; }
.answer-label { color: var(--text-muted); width: 3.5rem; flex-shrink: 0; }
.answer-val { font-weight: 700; padding: 0.1rem 0.5rem; border-radius: 4px; font-size: 0.72rem; }
.val-benar { color: var(--success); background: rgba(34,197,94,0.12); }
.val-salah { color: var(--danger); background: rgba(239,68,68,0.12); }
.val-blank { color: var(--text-muted); background: var(--border); }

.explanation {
  font-size: 0.825rem; line-height: 1.6; color: var(--text-muted);
  background: var(--bg-input); border-radius: 8px; padding: 0.75rem;
}
.exp-label { font-weight: 600; color: var(--text-muted); }
</style>
