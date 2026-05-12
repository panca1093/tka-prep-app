<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSessionStore } from '@/stores/session'

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
  secondsLeft.value <= 60 ? '#ef4444' : secondsLeft.value <= 300 ? '#f59e0b' : '#22c55e'
)

async function autoSubmit() {
  try { await doSubmit() } catch {}
}

// ─── Question navigation ──────────────────────────────────────────────────────
const currentQ = computed(() => store.questions[store.currentIndex])

function goTo(i: number) { store.currentIndex = i }

const questionStatus = (qid: string) => {
  const answered = store.answers[qid] != null
  const isFlagged = store.flagged.has(qid)
  if (isFlagged) return 'flagged'
  if (answered) return 'answered'
  return 'unanswered'
}

// ─── Answer + flag ────────────────────────────────────────────────────────────
async function selectOption(optionId: string) {
  if (!currentQ.value) return
  const qid = currentQ.value.id
  const current = store.answers[qid]
  // Toggle off if same option
  const newVal = current === optionId ? null : optionId
  await store.saveAnswer(qid, newVal)
}

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
          Question {{ store.currentIndex + 1 }} of {{ store.questions.length }}
          <button
            class="flag-btn"
            :class="{ flagged: store.flagged.has(currentQ?.id ?? '') }"
            @click="toggleFlag"
          >
            {{ store.flagged.has(currentQ?.id ?? '') ? '🚩 Flagged' : '⚑ Flag' }}
          </button>
        </div>

        <div class="question-text">{{ currentQ?.text }}</div>

        <div class="options">
          <button
            v-for="opt in currentQ?.options"
            :key="opt.id"
            class="option-btn"
            :class="{ selected: store.answers[currentQ?.id ?? ''] === opt.id }"
            @click="selectOption(opt.id)"
          >
            <span class="opt-label">{{ opt.label }}</span>
            <span class="opt-text">{{ opt.text }}</span>
          </button>
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
            :class="[questionStatus(q.id), { current: i === store.currentIndex }]"
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
  color: #94a3b8;
  font-size: 1rem;
}
.error-screen { color: #ef4444; }

.session-shell { display: flex; flex-direction: column; height: 100vh; background: #080c18; }

.session-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.5rem;
  background: #0d1424;
  border-bottom: 1px solid #1e2a45;
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
  background: #4f8ef7;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
}
.btn-submit-header:hover { background: #3b7be8; }

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
  color: #94a3b8;
}

.flag-btn {
  padding: 0.3rem 0.75rem;
  border-radius: 6px;
  border: 1px solid #1e2a45;
  background: transparent;
  color: #94a3b8;
  font-size: 0.78rem;
  cursor: pointer;
  transition: all 0.15s;
}
.flag-btn.flagged { border-color: #f59e0b; color: #f59e0b; background: rgba(245, 158, 11, 0.1); }

.question-text {
  font-size: 1.05rem;
  line-height: 1.7;
  color: #f1f5f9;
  background: #141c2e;
  border: 1px solid #1e2a45;
  border-radius: 12px;
  padding: 1.5rem;
}

.options { display: flex; flex-direction: column; gap: 0.625rem; }

.option-btn {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.875rem 1rem;
  border-radius: 10px;
  border: 1px solid #1e2a45;
  background: #141c2e;
  color: #f1f5f9;
  cursor: pointer;
  text-align: left;
  transition: all 0.15s;
  width: 100%;
}
.option-btn:hover { border-color: #4f8ef7; }
.option-btn.selected { border-color: #4f8ef7; background: rgba(79, 142, 247, 0.12); }

.opt-label {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 50%;
  background: #1e2a45;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 700;
  flex-shrink: 0;
}
.option-btn.selected .opt-label { background: #4f8ef7; }
.opt-text { font-size: 0.9rem; line-height: 1.5; padding-top: 0.15rem; }

.nav-arrows { display: flex; gap: 0.75rem; justify-content: space-between; }
.arrow-btn {
  padding: 0.5rem 1.25rem;
  border-radius: 8px;
  border: 1px solid #1e2a45;
  background: transparent;
  color: #94a3b8;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.15s;
}
.arrow-btn:hover:not(:disabled) { border-color: #4f8ef7; color: #f1f5f9; }
.arrow-btn:disabled { opacity: 0.4; cursor: not-allowed; }

.side-panel {
  width: 220px;
  flex-shrink: 0;
  border-left: 1px solid #1e2a45;
  padding: 1.25rem;
  overflow-y: auto;
  background: #0d1424;
}

.panel-title { font-size: 0.75rem; font-weight: 700; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 0.875rem; }

.q-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 4px;
  margin-bottom: 1rem;
}

.q-cell {
  aspect-ratio: 1;
  border-radius: 6px;
  border: 1px solid #1e2a45;
  background: #141c2e;
  color: #94a3b8;
  font-size: 0.7rem;
  cursor: pointer;
  transition: all 0.1s;
}
.q-cell.answered { background: #1a3a1a; border-color: #22c55e; color: #22c55e; }
.q-cell.flagged { background: #3a2a00; border-color: #f59e0b; color: #f59e0b; }
.q-cell.current { outline: 2px solid #4f8ef7; outline-offset: 1px; }

.legend { display: flex; flex-direction: column; gap: 0.375rem; }
.legend-item { display: flex; align-items: center; gap: 0.5rem; font-size: 0.72rem; color: #94a3b8; }
.dot { width: 10px; height: 10px; border-radius: 3px; }
.dot.answered { background: #22c55e; }
.dot.flagged { background: #f59e0b; }
.dot.unanswered { background: #1e2a45; }

/* Modal */
.modal-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.6);
  display: flex; align-items: center; justify-content: center;
  z-index: 100;
}
.modal {
  background: #141c2e;
  border: 1px solid #1e2a45;
  border-radius: 16px;
  padding: 2rem;
  width: 360px;
  max-width: calc(100vw - 2rem);
}
.modal h3 { margin: 0 0 1.25rem; font-size: 1.1rem; }
.modal-stats { display: flex; gap: 1rem; margin-bottom: 1rem; }
.modal-stat { flex: 1; text-align: center; font-size: 0.8rem; color: #94a3b8; }
.ms-val { display: block; font-size: 1.5rem; font-weight: 800; margin-bottom: 0.25rem; }
.ms-val.answered { color: #22c55e; }
.ms-val.flagged { color: #f59e0b; }
.ms-val.blank { color: #ef4444; }
.modal-warn { margin: 0 0 0.75rem; font-size: 0.825rem; color: #f59e0b; }
.error-msg { padding: 0.5rem; border-radius: 6px; background: #450a0a; color: #fca5a5; font-size: 0.8rem; margin-bottom: 0.75rem; }
.modal-actions { display: flex; gap: 0.75rem; }
.btn-cancel {
  flex: 1; padding: 0.65rem; border-radius: 8px;
  border: 1px solid #1e2a45; background: transparent; color: #94a3b8;
  cursor: pointer; transition: all 0.15s;
}
.btn-cancel:hover { border-color: #ef4444; color: #ef4444; }
.btn-cancel:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-confirm {
  flex: 1; padding: 0.65rem; border-radius: 8px;
  border: none; background: #4f8ef7; color: #fff;
  font-weight: 600; cursor: pointer; transition: background 0.15s;
}
.btn-confirm:hover { background: #3b7be8; }
.btn-confirm:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
