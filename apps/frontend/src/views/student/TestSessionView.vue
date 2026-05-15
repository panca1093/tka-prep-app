<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSessionStore } from '@/stores/session'
import RichTextViewer from '@/components/editor/RichTextViewer.vue'
import { resolveUrl } from '@/utils/url'

const route = useRoute()
const router = useRouter()
const store = useSessionStore()

const sessionId = route.params.sessionId as string
const showSubmitModal = ref(false)
const isSubmitting = ref(false)
const submitError = ref('')

// ─── Timer ────────────────────────────────────────────────────────────────────
const secondsLeft = ref(0)
let timerInterval: ReturnType<typeof setInterval> | null = null

function startTimer(seconds: number) {
  secondsLeft.value = seconds
  if (timerInterval) clearInterval(timerInterval)
  timerInterval = setInterval(() => {
    if (secondsLeft.value <= 0) {
      clearInterval(timerInterval!)
      autoSubmit()
    } else {
      secondsLeft.value--
    }
  }, 1000)
}

const timerDisplay = computed(() => {
  const s = secondsLeft.value
  const m = Math.floor(s / 60)
  const ss = s % 60
  return `${String(m).padStart(2, '0')}:${String(ss).padStart(2, '0')}`
})

const timerColor = computed(() =>
  secondsLeft.value <= 60 ? 'var(--danger)' : secondsLeft.value <= 300 ? 'var(--warning)' : 'var(--success)'
)

async function autoSubmit() {
  try {
    await doSubmit()
  } catch {
    submitError.value = 'Time is up! Auto-submit failed. Please click Submit now.'
    showSubmitModal.value = true
  }
}

// ─── Question navigation ──────────────────────────────────────────────────────
const currentQ = computed(() => store.questions[store.currentIndex])

function goTo(i: number) { store.currentIndex = i }

function questionStatus(i: number) {
  const q = store.questions[i]
  if (!q) return 'unanswered'
  if (store.flagged.has(q.id)) return 'flagged'
  if (store.isAnswered(q)) return 'answered'
  return 'unanswered'
}

// ─── MCQ ──────────────────────────────────────────────────────────────────────
async function selectMCQOption(optionId: string) {
  if (!currentQ.value) return
  const qid = currentQ.value.id
  const current = store.answers[qid]
  await store.saveMCQAnswer(qid, current === optionId ? null : optionId)
}

// ─── PGK (multi-correct) ──────────────────────────────────────────────────────
async function togglePGKOption(optionId: string) {
  if (!currentQ.value) return
  await store.togglePGKOption(currentQ.value.id, optionId)
}

function isPGKSelected(optionId: string): boolean {
  if (!currentQ.value) return false
  return store.pgkAnswers[currentQ.value.id]?.has(optionId) ?? false
}

// ─── B/S (true_false) ─────────────────────────────────────────────────────────
async function setBSAnswer(statementId: string, value: boolean) {
  if (!currentQ.value) return
  await store.saveBSStatement(currentQ.value.id, statementId, value)
}

function getBSAnswer(statementId: string): boolean | null {
  if (!currentQ.value) return null
  const map = store.bsAnswers[currentQ.value.id]
  if (!map || !(statementId in map)) return null
  return map[statementId]
}

// ─── Flag ─────────────────────────────────────────────────────────────────────
async function toggleFlag() {
  if (!currentQ.value) return
  await store.toggleFlag(currentQ.value.id)
}

// ─── Submit ───────────────────────────────────────────────────────────────────
function openSubmitModal() { showSubmitModal.value = true }

async function doSubmit() {
  isSubmitting.value = true
  submitError.value = ''
  try {
    const resultId = await store.submit()
    if (timerInterval) clearInterval(timerInterval)
    router.push({ name: 'result', params: { resultId } })
  } catch {
    submitError.value = 'Submission failed. Please try again.'
    isSubmitting.value = false
  }
}

// ─── Lifecycle ────────────────────────────────────────────────────────────────
onMounted(async () => {
  await store.load(sessionId)
  if (store.session) {
    startTimer(store.session.time_remaining_seconds)
  }
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
})

const answeredCount = computed(() => store.answeredCount())
const flaggedCount = computed(() => store.flaggedCount())
const blankCount = computed(() => store.questions.length - answeredCount.value)
</script>

<template>
  <div v-if="store.isLoading" class="loading-screen">Loading session…</div>

  <div v-else-if="store.error" class="error-screen">{{ store.error }}</div>

  <div v-else-if="store.session && store.questions.length" class="session-shell">
    <!-- Top bar -->
    <header class="session-header">
      <div class="test-name">{{ store.test?.title }}</div>
      <div class="timer" :style="{ color: timerColor }">⏱ {{ timerDisplay }}</div>
      <button class="btn-submit-header" @click="openSubmitModal">Submit</button>
    </header>

    <div class="session-body">
      <!-- Question area -->
      <div class="question-area">
        <div class="question-meta">
          <span>Question {{ store.currentIndex + 1 }} of {{ store.questions.length }}</span>
          <div class="meta-right">
            <span class="q-type-badge" :class="currentQ?.question_type">
              {{ currentQ?.question_type === 'mcq' ? 'Pilihan Ganda'
                : currentQ?.question_type === 'multi_correct' ? 'PGK'
                : 'Benar / Salah' }}
            </span>
            <button
              class="flag-btn"
              :class="{ flagged: store.flagged.has(currentQ?.id ?? '') }"
              @click="toggleFlag"
            >
              {{ store.flagged.has(currentQ?.id ?? '') ? '🚩 Flagged' : '⚑ Flag' }}
            </button>
          </div>
        </div>

        <div class="question-text">
          <RichTextViewer :html="currentQ?.text ?? ''" />
          <img v-if="currentQ?.image_url" :src="resolveUrl(currentQ.image_url)" class="q-img" alt="" />
        </div>

        <!-- MCQ: single select -->
        <div v-if="currentQ?.question_type === 'mcq'" class="options">
          <button
            v-for="opt in currentQ.options"
            :key="opt.id"
            class="option-btn"
            :class="{ selected: store.answers[currentQ.id] === opt.id }"
            @click="selectMCQOption(opt.id)"
          >
            <span class="opt-label">{{ opt.label }}</span>
            <div class="opt-body">
              <span class="opt-text"><RichTextViewer :html="opt.text" /></span>
              <img v-if="opt.image_url" :src="resolveUrl(opt.image_url)" class="opt-img" alt="" />
            </div>
          </button>
        </div>

        <!-- PGK: multi-select checkboxes -->
        <div v-else-if="currentQ?.question_type === 'multi_correct'" class="options">
          <div class="pgk-hint">Pilih semua jawaban yang benar.</div>
          <button
            v-for="opt in currentQ.options"
            :key="opt.id"
            class="option-btn"
            :class="{ selected: isPGKSelected(opt.id) }"
            @click="togglePGKOption(opt.id)"
          >
            <span class="opt-label pgk" :class="{ checked: isPGKSelected(opt.id) }">
              {{ isPGKSelected(opt.id) ? '✓' : opt.label }}
            </span>
            <div class="opt-body">
              <span class="opt-text"><RichTextViewer :html="opt.text" /></span>
              <img v-if="opt.image_url" :src="resolveUrl(opt.image_url)" class="opt-img" alt="" />
            </div>
          </button>
        </div>

        <!-- B/S: per-statement Benar/Salah toggles -->
        <div v-else-if="currentQ?.question_type === 'true_false'" class="statements">
          <div class="bs-hint">Tentukan apakah setiap pernyataan berikut Benar atau Salah.</div>
          <div
            v-for="(stmt, idx) in currentQ.statements"
            :key="stmt.id"
            class="stmt-row"
            :class="{
              'stmt-benar': getBSAnswer(stmt.id) === true,
              'stmt-salah': getBSAnswer(stmt.id) === false,
            }"
          >
            <div class="stmt-num">{{ idx + 1 }}</div>
            <div class="stmt-body">
              <div class="stmt-text"><RichTextViewer :html="stmt.text" /></div>
              <img v-if="stmt.image_url" :src="resolveUrl(stmt.image_url)" class="stmt-img" alt="" />
            </div>
            <div class="bs-btns">
              <button
                class="bs-btn benar"
                :class="{ active: getBSAnswer(stmt.id) === true }"
                @click="setBSAnswer(stmt.id, true)"
              >Benar</button>
              <button
                class="bs-btn salah"
                :class="{ active: getBSAnswer(stmt.id) === false }"
                @click="setBSAnswer(stmt.id, false)"
              >Salah</button>
            </div>
          </div>
        </div>

        <div class="nav-arrows">
          <button class="arrow-btn" :disabled="store.currentIndex === 0" @click="goTo(store.currentIndex - 1)">← Prev</button>
          <button class="arrow-btn" :disabled="store.currentIndex >= store.questions.length - 1" @click="goTo(store.currentIndex + 1)">Next →</button>
        </div>
      </div>

      <!-- Side panel -->
      <div class="side-panel">
        <div class="panel-title">Questions</div>
        <div class="q-grid">
          <button
            v-for="(q, i) in store.questions"
            :key="q.id"
            class="q-cell"
            :class="[questionStatus(i), { current: i === store.currentIndex }]"
            @click="goTo(i)"
          >
            {{ i + 1 }}
          </button>
        </div>
        <div class="legend">
          <span class="legend-item"><span class="dot answered" />Answered</span>
          <span class="legend-item"><span class="dot flagged" />Flagged</span>
          <span class="legend-item"><span class="dot unanswered" />Blank</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Submit modal -->
  <div v-if="showSubmitModal" class="modal-backdrop" @click.self="showSubmitModal = false">
    <div class="modal">
      <h3>Submit Test?</h3>
      <div class="modal-stats">
        <div class="modal-stat"><span class="ms-val answered">{{ answeredCount }}</span> answered</div>
        <div class="modal-stat"><span class="ms-val flagged">{{ flaggedCount }}</span> flagged</div>
        <div class="modal-stat"><span class="ms-val blank">{{ blankCount }}</span> blank</div>
      </div>
      <p v-if="blankCount > 0" class="modal-warn">You have {{ blankCount }} unanswered question(s).</p>
      <p v-if="submitError" class="error-msg">{{ submitError }}</p>
      <div class="modal-actions">
        <button class="btn-cancel" :disabled="isSubmitting" @click="showSubmitModal = false">Cancel</button>
        <button class="btn-confirm" :disabled="isSubmitting" @click="doSubmit">
          {{ isSubmitting ? 'Submitting…' : 'Submit Now' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.loading-screen, .error-screen {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  font-size: 1rem;
}
.error-screen { color: var(--danger); }

.session-shell { display: flex; flex-direction: column; height: 100vh; background: var(--bg-app); }

.session-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  background: var(--bg-input);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 10;
}

.test-name { font-weight: 700; font-size: 0.95rem; flex: 1; }
.timer { font-size: 1.25rem; font-weight: 800; font-variant-numeric: tabular-nums; padding: 0 1.5rem; }

.btn-submit-header {
  padding: 0.5rem 1.25rem;
  border-radius: 8px;
  border: none;
  background: var(--accent);
  color: var(--text-on-accent);
  font-weight: 600;
  cursor: pointer;
}
.btn-submit-header:hover { background: var(--accent-hover); }

.session-body {
  display: flex;
  flex: 1;
  gap: 0;
  overflow: hidden;
}

.question-area {
  flex: 1;
  padding: 2rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.question-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.825rem;
  color: var(--text-muted);
}

.meta-right { display: flex; align-items: center; gap: 0.625rem; }

.q-type-badge {
  padding: 0.2rem 0.6rem;
  border-radius: 20px;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  background: var(--border);
  color: var(--text-muted);
}
.q-type-badge.mcq { background: rgba(79,142,247,0.15); color: var(--accent); }
.q-type-badge.multi_correct { background: rgba(168,85,247,0.15); color: var(--accent); }
.q-type-badge.true_false { background: rgba(34,197,94,0.15); color: var(--success); }

.flag-btn {
  padding: 0.3rem 0.75rem;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.78rem;
  cursor: pointer;
  transition: all 0.15s;
}
.flag-btn.flagged { border-color: var(--warning); color: var(--warning); background: rgba(245, 158, 11, 0.1); }

.question-text {
  font-size: 1.05rem;
  line-height: 1.7;
  color: var(--text-primary);
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.5rem;
}

/* MCQ / PGK options */
.options { display: flex; flex-direction: column; gap: 0.625rem; }

.pgk-hint, .bs-hint {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 0.25rem;
}

.option-btn {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.875rem 1rem;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-primary);
  cursor: pointer;
  text-align: left;
  transition: all 0.15s;
  width: 100%;
}
.option-btn:hover { border-color: var(--accent); }
.option-btn.selected { border-color: var(--accent); background: rgba(79, 142, 247, 0.12); }

.opt-label {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 50%;
  background: var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 700;
  flex-shrink: 0;
}
.opt-label.pgk { border-radius: 5px; }
.opt-label.pgk.checked { background: var(--accent); }
.option-btn.selected .opt-label:not(.pgk) { background: var(--accent); }
.opt-body { display: flex; flex-direction: column; gap: 0.5rem; flex: 1; }
.opt-text { font-size: 0.9rem; line-height: 1.5; padding-top: 0.15rem; }
.opt-img { max-width: 100%; max-height: 160px; border-radius: 6px; border: 1px solid var(--border); }
.q-img { max-width: 100%; max-height: 240px; border-radius: 8px; margin-top: 0.75rem; border: 1px solid var(--border); }

/* B/S statements */
.statements { display: flex; flex-direction: column; gap: 0.75rem; }

.stmt-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  transition: all 0.15s;
}
.stmt-row.stmt-benar { border-color: var(--success); background: rgba(34,197,94,0.07); }
.stmt-row.stmt-salah { border-color: var(--danger); background: rgba(239,68,68,0.07); }

.stmt-num {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 50%;
  background: var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 700;
  flex-shrink: 0;
  color: var(--text-muted);
}

.stmt-body { flex: 1; display: flex; flex-direction: column; gap: 0.375rem; }
.stmt-text {
  font-size: 0.9rem;
  line-height: 1.5;
  color: var(--text-primary);
}
.stmt-img { max-width: 100%; max-height: 140px; border-radius: 6px; border: 1px solid var(--border); }

.bs-btns { display: flex; gap: 0.5rem; flex-shrink: 0; }

.bs-btn {
  padding: 0.35rem 0.875rem;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: transparent;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  color: var(--text-muted);
}
.bs-btn.benar:hover, .bs-btn.benar.active { border-color: var(--success); background: rgba(34,197,94,0.15); color: var(--success); }
.bs-btn.salah:hover, .bs-btn.salah.active { border-color: var(--danger); background: rgba(239,68,68,0.15); color: var(--danger); }

/* Nav */
.nav-arrows { display: flex; gap: 0.75rem; justify-content: space-between; }
.arrow-btn {
  padding: 0.5rem 1.25rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.15s;
}
.arrow-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--text-primary); }
.arrow-btn:disabled { opacity: 0.4; cursor: not-allowed; }

/* Side panel */
.side-panel {
  width: 220px;
  flex-shrink: 0;
  border-left: 1px solid var(--border);
  padding: 1.25rem;
  overflow-y: auto;
  background: var(--bg-input);
}

.panel-title { font-size: 0.75rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 0.875rem; }

.q-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 4px;
  margin-bottom: 1rem;
}

.q-cell {
  aspect-ratio: 1;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-muted);
  font-size: 0.7rem;
  cursor: pointer;
  transition: all 0.1s;
}
.q-cell.answered { background: rgba(34, 197, 94, 0.12); border-color: var(--success); color: var(--success); }
.q-cell.flagged { background: rgba(245, 158, 11, 0.12); border-color: var(--warning); color: var(--warning); }
.q-cell.current { outline: 2px solid var(--accent); outline-offset: 1px; }

.legend { display: flex; flex-direction: column; gap: 0.375rem; }
.legend-item { display: flex; align-items: center; gap: 0.5rem; font-size: 0.72rem; color: var(--text-muted); }
.dot { width: 10px; height: 10px; border-radius: 3px; }
.dot.answered { background: var(--success); }
.dot.flagged { background: var(--warning); }
.dot.unanswered { background: var(--border); }

/* Modal */
.modal-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.6);
  display: flex; align-items: center; justify-content: center;
  z-index: 100;
}
.modal {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 2rem;
  width: 360px;
  max-width: calc(100vw - 2rem);
}
.modal h3 { margin: 0 0 1.25rem; font-size: 1.1rem; }
.modal-stats { display: flex; gap: 1rem; margin-bottom: 1rem; }
.modal-stat { flex: 1; text-align: center; font-size: 0.8rem; color: var(--text-muted); }
.ms-val { display: block; font-size: 1.5rem; font-weight: 800; margin-bottom: 0.25rem; }
.ms-val.answered { color: var(--success); }
.ms-val.flagged { color: var(--warning); }
.ms-val.blank { color: var(--danger); }
.modal-warn { margin: 0 0 0.75rem; font-size: 0.825rem; color: var(--warning); }
.error-msg { padding: 0.5rem; border-radius: 6px; background: color-mix(in srgb, var(--danger) 15%, var(--bg-surface)); color: var(--danger); font-size: 0.8rem; margin-bottom: 0.75rem; }
.modal-actions { display: flex; gap: 0.75rem; }
.btn-cancel {
  flex: 1; padding: 0.65rem; border-radius: 8px;
  border: 1px solid var(--border); background: transparent; color: var(--text-muted);
  cursor: pointer; transition: all 0.15s;
}
.btn-cancel:hover { border-color: var(--danger); color: var(--danger); }
.btn-cancel:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-confirm {
  flex: 1; padding: 0.65rem; border-radius: 8px;
  border: none; background: var(--accent); color: var(--text-on-accent);
  font-weight: 600; cursor: pointer; transition: background 0.15s;
}
.btn-confirm:hover { background: var(--accent-hover); }
.btn-confirm:disabled { opacity: 0.6; cursor: not-allowed; }

@media (max-width: 768px) { .q-option { min-height: 52px; } .q-text { font-size: clamp(0.95rem, 2vw, 1.1rem); } }
</style>
