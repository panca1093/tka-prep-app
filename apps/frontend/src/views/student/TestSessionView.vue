<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSessionStore } from '@/stores/session'
import { useAuthStore } from '@/stores/auth'
import RichTextViewer from '@/components/editor/RichTextViewer.vue'
import { resolveUrl } from '@/utils/url'

const route = useRoute()
const router = useRouter()
const store = useSessionStore()
const auth = useAuthStore()

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

const timerUrgency = computed(() => {
  if (secondsLeft.value <= 60) return 'critical'
  if (secondsLeft.value <= 300) return 'warning'
  return 'normal'
})

async function autoSubmit() {
  try {
    await doSubmit()
  } catch {
    submitError.value = 'Waktu habis! Auto-submit gagal. Klik Submit sekarang.'
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
    submitError.value = 'Submission gagal. Coba lagi.'
    isSubmitting.value = false
  }
}

// ─── Tab-switch detection ──────────────────────────────────────────────────────
const tabSwitchWarning = ref(false)
let tabWarningTimer: ReturnType<typeof setTimeout> | null = null

function handleVisibilityChange() {
  if (document.visibilityState === 'visible' && store.session?.status === 'in_progress') {
    tabSwitchWarning.value = true
    if (tabWarningTimer) clearTimeout(tabWarningTimer)
    tabWarningTimer = setTimeout(() => { tabSwitchWarning.value = false }, 6000)
  }
}

// ─── Lifecycle ────────────────────────────────────────────────────────────────
onMounted(async () => {
  await store.load(sessionId)
  if (store.session) {
    startTimer(store.session.time_remaining_seconds)
  }
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
  if (tabWarningTimer) clearTimeout(tabWarningTimer)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

const answeredCount = computed(() => store.answeredCount())
const flaggedCount = computed(() => store.flaggedCount())
const blankCount = computed(() => store.questions.length - answeredCount.value)
const progressPct = computed(() =>
  store.questions.length ? (answeredCount.value / store.questions.length) * 100 : 0
)

const watermarkStyle = computed(() => {
  const label = auth.user?.name ?? sessionId
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="340" height="180">
    <text x="170" y="90" text-anchor="middle" dominant-baseline="middle"
      transform="rotate(-22, 170, 90)"
      fill="rgba(128,128,128,0.065)"
      font-size="14" font-family="system-ui,sans-serif" font-weight="500"
      letter-spacing="0.5">${label}</text>
  </svg>`
  const b64 = btoa(unescape(encodeURIComponent(svg)))
  return {
    backgroundImage: `url("data:image/svg+xml;base64,${b64}")`,
    backgroundSize: '340px 180px',
    backgroundRepeat: 'repeat',
  }
})
</script>

<template>
  <div v-if="store.isLoading" class="state-screen">
    <div class="loading-dots"><span/><span/><span/></div>
  </div>

  <div v-else-if="store.error" class="state-screen err">{{ store.error }}</div>

  <div v-else-if="store.session && store.questions.length" class="session-shell">
    <!-- Watermark -->
    <div class="session-watermark" :style="watermarkStyle" aria-hidden="true" />

    <!-- Tab-switch warning banner -->
    <Transition name="tab-warn">
      <div v-if="tabSwitchWarning" class="tab-warning">
        <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16">
          <path fill-rule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z" clip-rule="evenodd"/>
        </svg>
        Peringatan: Kamu beralih tab. Aktivitas ini tercatat oleh sistem.
      </div>
    </Transition>

    <!-- Top bar -->
    <header class="session-header">
      <div class="header-left">
        <p class="test-name">{{ store.test?.title }}</p>
        <p class="question-counter">Soal {{ store.currentIndex + 1 }} / {{ store.questions.length }}</p>
      </div>

      <div class="timer-wrap" :class="timerUrgency">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="15" height="15">
          <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
        </svg>
        <span class="timer-digits">{{ timerDisplay }}</span>
      </div>

      <button class="btn-submit-header" @click="openSubmitModal">
        <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14">
          <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"/>
        </svg>
        Selesai
      </button>
    </header>

    <!-- Progress bar -->
    <div class="progress-track">
      <div class="progress-fill" :style="{ width: progressPct + '%' }" />
    </div>

    <div class="session-body">
      <!-- Question area -->
      <div class="question-area">

        <!-- Question type badge + flag -->
        <div class="question-meta">
          <span class="q-type-badge" :class="currentQ?.question_type">
            {{ currentQ?.question_type === 'mcq' ? 'Pilihan Ganda'
              : currentQ?.question_type === 'multi_correct' ? 'PGK – Multi Jawaban'
              : 'Pernyataan Benar/Salah' }}
          </span>
          <button
            class="flag-btn"
            :class="{ flagged: store.flagged.has(currentQ?.id ?? '') }"
            @click="toggleFlag"
          >
            <svg v-if="!store.flagged.has(currentQ?.id ?? '')" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" width="13" height="13">
              <path d="M3 3h14l-2.5 5L17 13H3V3z"/>
            </svg>
            <svg v-else viewBox="0 0 20 20" fill="currentColor" width="13" height="13">
              <path d="M3 3h14l-2.5 5L17 13H3V3z"/>
            </svg>
            {{ store.flagged.has(currentQ?.id ?? '') ? 'Ditandai' : 'Tandai' }}
          </button>
        </div>

        <!-- Question text -->
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

        <!-- PGK: multi-select -->
        <div v-else-if="currentQ?.question_type === 'multi_correct'" class="options">
          <p class="type-hint">Pilih semua jawaban yang benar.</p>
          <button
            v-for="opt in currentQ.options"
            :key="opt.id"
            class="option-btn pgk-btn"
            :class="{ selected: isPGKSelected(opt.id) }"
            @click="togglePGKOption(opt.id)"
          >
            <span class="opt-label pgk" :class="{ checked: isPGKSelected(opt.id) }">
              <svg v-if="isPGKSelected(opt.id)" viewBox="0 0 12 12" fill="currentColor" width="10" height="10">
                <path d="M10 3L5 8.5 2 5.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" fill="none"/>
              </svg>
            </span>
            <div class="opt-body">
              <span class="opt-text"><RichTextViewer :html="opt.text" /></span>
              <img v-if="opt.image_url" :src="resolveUrl(opt.image_url)" class="opt-img" alt="" />
            </div>
          </button>
        </div>

        <!-- B/S: per-statement -->
        <div v-else-if="currentQ?.question_type === 'true_false'" class="statements">
          <p class="type-hint">Tentukan apakah setiap pernyataan berikut Benar atau Salah.</p>
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
              <button class="bs-btn benar" :class="{ active: getBSAnswer(stmt.id) === true }" @click="setBSAnswer(stmt.id, true)">Benar</button>
              <button class="bs-btn salah" :class="{ active: getBSAnswer(stmt.id) === false }" @click="setBSAnswer(stmt.id, false)">Salah</button>
            </div>
          </div>
        </div>

        <!-- Bottom navigation -->
        <div class="nav-bar">
          <button class="nav-btn prev" :disabled="store.currentIndex === 0" @click="goTo(store.currentIndex - 1)">
            <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16"><path fill-rule="evenodd" d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z" clip-rule="evenodd"/></svg>
            Sebelumnya
          </button>
          <span class="nav-counter">{{ store.currentIndex + 1 }} / {{ store.questions.length }}</span>
          <button class="nav-btn next" :disabled="store.currentIndex >= store.questions.length - 1" @click="goTo(store.currentIndex + 1)">
            Berikutnya
            <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16"><path fill-rule="evenodd" d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z" clip-rule="evenodd"/></svg>
          </button>
        </div>
      </div>

      <!-- Side panel -->
      <aside class="side-panel">
        <div class="panel-stats">
          <div class="pstat pstat--done">
            <span class="pstat-val">{{ answeredCount }}</span>
            <span class="pstat-lbl">Dijawab</span>
          </div>
          <div class="pstat pstat--flag">
            <span class="pstat-val">{{ flaggedCount }}</span>
            <span class="pstat-lbl">Ditandai</span>
          </div>
          <div class="pstat pstat--blank">
            <span class="pstat-val">{{ blankCount }}</span>
            <span class="pstat-lbl">Kosong</span>
          </div>
        </div>

        <p class="panel-section-label">Navigasi Soal</p>
        <div class="q-grid">
          <button
            v-for="(q, i) in store.questions"
            :key="q.id"
            class="q-cell"
            :class="[questionStatus(i), { current: i === store.currentIndex }]"
            @click="goTo(i)"
          >{{ i + 1 }}</button>
        </div>

        <div class="legend">
          <span class="legend-item"><span class="dot answered"/>Dijawab</span>
          <span class="legend-item"><span class="dot flagged"/>Ditandai</span>
          <span class="legend-item"><span class="dot unanswered"/>Kosong</span>
        </div>
      </aside>
    </div>
  </div>

  <!-- Submit modal -->
  <Transition name="modal">
    <div v-if="showSubmitModal" class="modal-backdrop" @click.self="showSubmitModal = false">
      <div class="modal">
        <div class="modal-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="28" height="28">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
          </svg>
        </div>
        <h3 class="modal-title">Selesaikan Ujian?</h3>
        <p class="modal-sub">Pastikan semua jawaban sudah benar sebelum submit.</p>
        <div class="modal-stats">
          <div class="modal-stat"><span class="ms-val answered">{{ answeredCount }}</span><span class="ms-lbl">Dijawab</span></div>
          <div class="modal-stat"><span class="ms-val flagged">{{ flaggedCount }}</span><span class="ms-lbl">Ditandai</span></div>
          <div class="modal-stat"><span class="ms-val blank">{{ blankCount }}</span><span class="ms-lbl">Kosong</span></div>
        </div>
        <p v-if="blankCount > 0" class="modal-warn">
          <svg viewBox="0 0 16 16" fill="currentColor" width="13" height="13"><path d="M8.982 1.566a1.13 1.13 0 00-1.96 0L.165 13.233c-.457.778.091 1.767.98 1.767h13.713c.889 0 1.438-.99.98-1.767L8.982 1.566zM8 5c.535 0 .954.462.9.995l-.35 3.507a.552.552 0 01-1.1 0L7.1 5.995A.905.905 0 018 5zm.002 6a1 1 0 110 2 1 1 0 010-2z"/></svg>
          {{ blankCount }} soal belum dijawab.
        </p>
        <p v-if="submitError" class="error-msg">{{ submitError }}</p>
        <div class="modal-actions">
          <button class="btn-cancel" :disabled="isSubmitting" @click="showSubmitModal = false">Batal</button>
          <button class="btn-confirm" :disabled="isSubmitting" @click="doSubmit">
            {{ isSubmitting ? 'Menyimpan…' : 'Submit Sekarang' }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
/* ─── States ──────────────────────────────────────────────────────────────── */
.state-screen {
  display: flex; align-items: center; justify-content: center;
  height: 100vh; color: var(--text-muted);
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

/* ─── Watermark ───────────────────────────────────────────────────────────── */
.session-watermark {
  position: fixed; inset: 0; pointer-events: none; z-index: 1;
  user-select: none;
}

/* ─── Tab warning ─────────────────────────────────────────────────────────── */
.tab-warning {
  display: flex; align-items: center; justify-content: center; gap: 0.5rem;
  background: var(--danger);
  color: #fff;
  padding: 0.625rem 1.25rem;
  font-size: 0.8rem;
  font-weight: 700;
  letter-spacing: 0.01em;
  position: sticky; top: 0; z-index: 20;
}
.tab-warn-enter-active, .tab-warn-leave-active { transition: all 0.3s ease; }
.tab-warn-enter-from { transform: translateY(-100%); opacity: 0; }
.tab-warn-leave-to { transform: translateY(-100%); opacity: 0; }

/* ─── Shell ───────────────────────────────────────────────────────────────── */
.session-shell {
  display: flex; flex-direction: column;
  height: 100vh;
  background: var(--bg-app);
  position: relative;
}

/* ─── Header ──────────────────────────────────────────────────────────────── */
.session-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.875rem 1.5rem;
  background: var(--bg-surface);
  border-bottom: 1px solid var(--border);
  position: sticky; top: 0; z-index: 10;
}

.header-left { flex: 1; min-width: 0; }
.test-name {
  margin: 0;
  font-family: 'Cormorant Garamond', serif;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--text-heading);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.question-counter {
  margin: 0;
  font-size: 0.72rem;
  color: var(--text-muted);
  margin-top: 1px;
}

.timer-wrap {
  display: flex; align-items: center; gap: 0.375rem;
  padding: 0.4rem 0.875rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  flex-shrink: 0;
}
.timer-wrap.warning { border-color: var(--warning); background: color-mix(in srgb, var(--warning) 10%, var(--bg-input)); color: var(--warning); }
.timer-wrap.critical { border-color: var(--danger); background: color-mix(in srgb, var(--danger) 10%, var(--bg-input)); color: var(--danger); animation: pulse-timer 0.8s ease infinite; }
@keyframes pulse-timer { 0%,100%{opacity:1} 50%{opacity:0.7} }

.timer-digits {
  font-family: 'JetBrains Mono', monospace;
  font-size: 1.05rem;
  font-weight: 700;
}

.btn-submit-header {
  display: flex; align-items: center; gap: 0.4rem;
  padding: 0.45rem 1rem;
  border-radius: 8px; border: none;
  background: var(--accent); color: #fff;
  font-size: 0.8rem; font-weight: 700;
  cursor: pointer; flex-shrink: 0;
  transition: background 0.15s, transform 0.15s;
}
.btn-submit-header:hover { background: var(--accent-hover); transform: translateY(-1px); }

/* ─── Progress bar ────────────────────────────────────────────────────────── */
.progress-track {
  height: 3px;
  background: var(--border);
  position: sticky; top: 0; z-index: 9;
}
.progress-fill {
  height: 100%;
  background: var(--accent);
  transition: width 0.6s ease;
}

/* ─── Body ────────────────────────────────────────────────────────────────── */
.session-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* ─── Question area ───────────────────────────────────────────────────────── */
.question-area {
  flex: 1;
  padding: 1.5rem 2rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.question-meta {
  display: flex; align-items: center; justify-content: space-between;
  gap: 0.75rem;
}

.q-type-badge {
  padding: 0.2rem 0.6rem;
  border-radius: 20px;
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: var(--bg-input);
  color: var(--text-muted);
}
.q-type-badge.mcq { background: var(--accent-dim); color: var(--accent); }
.q-type-badge.multi_correct { background: rgba(168,85,247,0.12); color: #a855f7; }
.q-type-badge.true_false { background: color-mix(in srgb, var(--success) 12%, transparent); color: var(--success); }

.flag-btn {
  display: flex; align-items: center; gap: 0.35rem;
  padding: 0.3rem 0.7rem;
  border-radius: 6px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.75rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s;
}
.flag-btn:hover { border-color: var(--warning); color: var(--warning); }
.flag-btn.flagged { border-color: var(--warning); color: var(--warning); background: color-mix(in srgb, var(--warning) 10%, transparent); }

/* ─── Question text ───────────────────────────────────────────────────────── */
.question-text {
  font-size: 1rem;
  line-height: 1.75;
  color: var(--text-primary);
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.5rem;
  user-select: none;
  -webkit-user-select: none;
}
.q-img { max-width: 100%; max-height: 260px; border-radius: 8px; margin-top: 1rem; border: 1px solid var(--border); }

/* ─── Type hint ───────────────────────────────────────────────────────────── */
.type-hint {
  margin: 0;
  font-size: 0.78rem;
  color: var(--text-muted);
  font-style: italic;
}

/* ─── Options ─────────────────────────────────────────────────────────────── */
.options { display: flex; flex-direction: column; gap: 0.625rem; }

.option-btn {
  display: flex; align-items: flex-start; gap: 0.875rem;
  padding: 0.875rem 1rem;
  border-radius: 10px;
  border: 1.5px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-primary);
  cursor: pointer; text-align: left;
  transition: border-color 0.15s, background 0.15s;
  width: 100%;
}
.option-btn:hover { border-color: var(--accent); background: var(--accent-dim); }
.option-btn.selected {
  border-color: var(--accent);
  background: var(--accent-dim);
}

.opt-label {
  width: 1.875rem; height: 1.875rem; border-radius: 50%;
  background: var(--bg-input);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.78rem; font-weight: 700;
  flex-shrink: 0;
  color: var(--text-muted);
  transition: background 0.15s, color 0.15s;
}
.option-btn.selected .opt-label:not(.pgk) {
  background: var(--accent); color: #fff;
}

.opt-label.pgk {
  border-radius: 5px;
  border: 2px solid var(--border);
  background: transparent;
}
.opt-label.pgk.checked { background: var(--accent); border-color: var(--accent); color: #fff; }

.opt-body { display: flex; flex-direction: column; gap: 0.5rem; flex: 1; }
.opt-text { font-size: 0.9rem; line-height: 1.55; padding-top: 0.2rem; }
.opt-img { max-width: 100%; max-height: 160px; border-radius: 6px; border: 1px solid var(--border); }

/* ─── Statements ──────────────────────────────────────────────────────────── */
.statements { display: flex; flex-direction: column; gap: 0.625rem; }

.stmt-row {
  display: flex; align-items: center; gap: 1rem;
  padding: 1rem 1.25rem;
  border-radius: 10px;
  border: 1.5px solid var(--border);
  background: var(--bg-surface);
  transition: all 0.15s;
}
.stmt-row.stmt-benar { border-color: var(--success); background: color-mix(in srgb, var(--success) 6%, var(--bg-surface)); }
.stmt-row.stmt-salah { border-color: var(--danger); background: color-mix(in srgb, var(--danger) 6%, var(--bg-surface)); }

.stmt-num {
  width: 1.875rem; height: 1.875rem; border-radius: 50%;
  background: var(--bg-input);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.78rem; font-weight: 700; flex-shrink: 0;
  color: var(--text-muted);
}
.stmt-body { flex: 1; display: flex; flex-direction: column; gap: 0.375rem; }
.stmt-text { font-size: 0.9rem; line-height: 1.55; color: var(--text-primary); }
.stmt-img { max-width: 100%; max-height: 140px; border-radius: 6px; border: 1px solid var(--border); }

.bs-btns { display: flex; gap: 0.5rem; flex-shrink: 0; }
.bs-btn {
  padding: 0.35rem 0.875rem; border-radius: 6px;
  border: 1.5px solid var(--border); background: transparent;
  font-size: 0.8rem; font-weight: 600; cursor: pointer;
  transition: all 0.15s; color: var(--text-muted);
}
.bs-btn.benar:hover, .bs-btn.benar.active { border-color: var(--success); background: color-mix(in srgb, var(--success) 12%, transparent); color: var(--success); }
.bs-btn.salah:hover, .bs-btn.salah.active { border-color: var(--danger); background: color-mix(in srgb, var(--danger) 12%, transparent); color: var(--danger); }

/* ─── Bottom nav ──────────────────────────────────────────────────────────── */
.nav-bar {
  display: flex; align-items: center; justify-content: space-between;
  gap: 1rem;
  margin-top: auto;
  padding-top: 0.5rem;
}
.nav-btn {
  display: flex; align-items: center; gap: 0.375rem;
  padding: 0.65rem 1.25rem;
  border-radius: 10px;
  border: 1.5px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-secondary);
  font-size: 0.875rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s;
}
.nav-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); background: var(--accent-dim); }
.nav-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.nav-counter {
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.8rem; color: var(--text-muted);
}

/* ─── Side panel ──────────────────────────────────────────────────────────── */
.side-panel {
  width: 224px; flex-shrink: 0;
  border-left: 1px solid var(--border);
  padding: 1.25rem;
  overflow-y: auto;
  background: var(--bg-surface);
  display: flex; flex-direction: column; gap: 1rem;
}

.panel-stats {
  display: grid; grid-template-columns: repeat(3,1fr);
  gap: 0.5rem;
}
.pstat {
  text-align: center; padding: 0.5rem 0.25rem;
  border-radius: 8px;
  border: 1px solid var(--border);
}
.pstat-val {
  display: block;
  font-family: 'JetBrains Mono', monospace;
  font-size: 1.2rem; font-weight: 700; line-height: 1;
}
.pstat-lbl { display: block; font-size: 0.62rem; color: var(--text-muted); margin-top: 2px; }
.pstat--done { border-color: color-mix(in srgb, var(--success) 25%, var(--border)); }
.pstat--done .pstat-val { color: var(--success); }
.pstat--flag { border-color: color-mix(in srgb, var(--warning) 25%, var(--border)); }
.pstat--flag .pstat-val { color: var(--warning); }
.pstat--blank .pstat-val { color: var(--text-muted); }

.panel-section-label {
  margin: 0;
  font-size: 0.68rem; font-weight: 700;
  text-transform: uppercase; letter-spacing: 0.07em;
  color: var(--text-muted);
}

.q-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 4px;
}
.q-cell {
  aspect-ratio: 1; border-radius: 6px;
  border: 1.5px solid var(--border);
  background: var(--bg-input);
  color: var(--text-muted);
  font-size: 0.68rem; font-weight: 600;
  cursor: pointer; transition: all 0.1s;
}
.q-cell:hover { border-color: var(--accent); color: var(--accent); }
.q-cell.answered { background: color-mix(in srgb, var(--success) 12%, transparent); border-color: var(--success); color: var(--success); }
.q-cell.flagged { background: color-mix(in srgb, var(--warning) 12%, transparent); border-color: var(--warning); color: var(--warning); }
.q-cell.current { outline: 2px solid var(--accent); outline-offset: 1px; }

.legend { display: flex; flex-direction: column; gap: 0.375rem; }
.legend-item { display: flex; align-items: center; gap: 0.4rem; font-size: 0.7rem; color: var(--text-muted); }
.dot { width: 9px; height: 9px; border-radius: 3px; }
.dot.answered { background: var(--success); }
.dot.flagged { background: var(--warning); }
.dot.unanswered { background: var(--border); }

/* ─── Modal ───────────────────────────────────────────────────────────────── */
.modal-backdrop {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.55);
  display: flex; align-items: center; justify-content: center;
  z-index: 100;
  backdrop-filter: blur(4px);
}
.modal {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 20px;
  padding: 2rem;
  width: 380px; max-width: calc(100vw - 2rem);
  display: flex; flex-direction: column; gap: 1rem;
}
.modal-enter-active, .modal-leave-active { transition: all 0.2s ease; }
.modal-enter-from { opacity: 0; transform: scale(0.95) translateY(8px); }
.modal-leave-to { opacity: 0; transform: scale(0.95) translateY(8px); }

.modal-icon {
  width: 52px; height: 52px; border-radius: 14px;
  background: var(--accent-dim);
  display: flex; align-items: center; justify-content: center;
  color: var(--accent); align-self: flex-start;
}
.modal-title { margin: 0; font-family: 'Cormorant Garamond', serif; font-size: 1.4rem; font-weight: 700; color: var(--text-heading); }
.modal-sub { margin: 0; font-size: 0.825rem; color: var(--text-muted); }

.modal-stats { display: flex; gap: 0.75rem; }
.modal-stat { flex: 1; text-align: center; padding: 0.875rem; border-radius: 10px; border: 1px solid var(--border); background: var(--bg-input); }
.ms-val { display: block; font-family: 'JetBrains Mono', monospace; font-size: 1.5rem; font-weight: 700; line-height: 1; }
.ms-lbl { display: block; font-size: 0.65rem; color: var(--text-muted); margin-top: 3px; }
.ms-val.answered { color: var(--success); }
.ms-val.flagged { color: var(--warning); }
.ms-val.blank { color: var(--danger); }

.modal-warn {
  margin: 0;
  display: flex; align-items: center; gap: 0.4rem;
  font-size: 0.8rem; color: var(--warning);
  padding: 0.6rem 0.875rem;
  background: color-mix(in srgb, var(--warning) 10%, transparent);
  border-radius: 8px;
}
.error-msg {
  margin: 0; padding: 0.6rem 0.875rem;
  border-radius: 8px; background: color-mix(in srgb, var(--danger) 10%, transparent);
  color: var(--danger); font-size: 0.8rem;
}

.modal-actions { display: flex; gap: 0.75rem; }
.btn-cancel {
  flex: 1; padding: 0.7rem; border-radius: 10px;
  border: 1.5px solid var(--border); background: transparent;
  color: var(--text-muted); font-size: 0.875rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s;
}
.btn-cancel:hover { border-color: var(--danger); color: var(--danger); }
.btn-cancel:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-confirm {
  flex: 1; padding: 0.7rem; border-radius: 10px;
  border: none; background: var(--accent); color: #fff;
  font-size: 0.875rem; font-weight: 700;
  cursor: pointer; transition: background 0.15s;
}
.btn-confirm:hover { background: var(--accent-hover); }
.btn-confirm:disabled { opacity: 0.6; cursor: not-allowed; }

/* ─── Mobile ──────────────────────────────────────────────────────────────── */
@media (max-width: 768px) {
  .session-shell { height: 100dvh; }

  .session-header { padding: 0.6rem 0.875rem; }
  .test-name { font-size: 0.9rem; }
  .question-counter { display: none; }
  .timer-digits { font-size: 0.9rem; }
  .btn-submit-header { font-size: 0.75rem; padding: 0.4rem 0.75rem; }

  .session-body { flex-direction: column; overflow-y: auto; }

  .question-area { padding: 0.875rem; gap: 0.875rem; overflow-y: visible; }
  .question-text { padding: 0.875rem; font-size: 0.95rem; }

  .option-btn { padding: 0.75rem; }
  .opt-text { font-size: 0.875rem; }

  .stmt-row { flex-wrap: wrap; padding: 0.75rem; }
  .bs-btns { width: 100%; margin-top: 0.5rem; }
  .bs-btn { flex: 1; min-height: 42px; }

  .side-panel {
    width: 100%; border-left: none; border-top: 1px solid var(--border);
    padding: 0.75rem 0.875rem; max-height: 210px;
  }
  .q-grid { grid-template-columns: repeat(8, 1fr); }
  .panel-section-label { display: none; }
  .legend { flex-direction: row; gap: 0.75rem; }

  .nav-bar { padding: 0.25rem 0; }
  .nav-btn { padding: 0.55rem 0.875rem; font-size: 0.8rem; }

  .modal { padding: 1.5rem; }
  .modal-stats { gap: 0.5rem; }
}
</style>
