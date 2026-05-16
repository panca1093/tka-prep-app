<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSessionStore } from '@/stores/session'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Result     = components['schemas']['TestResultResponse']
type Entry      = components['schemas']['LeaderboardEntryResponse']
type TestDetail = components['schemas']['TestDetailResponse']

const auth         = useAuthStore()
const sessionStore = useSessionStore()
const router       = useRouter()

const results      = ref<Result[]>([])
const topEntries   = ref<Entry[]>([])
const myRank       = ref<Entry | null>(null)
const ongoingTests = ref<TestDetail[]>([])
const isLoading    = ref(true)
const ready        = ref(false)
const resumingId   = ref<string | null>(null)

onMounted(async () => {
  const [rr, rb, rm, rt] = await Promise.allSettled([
    client.GET('/results'),
    client.GET('/leaderboard', { params: { query: { scope: 'global' } } }),
    client.GET('/leaderboard/me', { params: { query: {} } }),
    client.GET('/tests', { params: { query: { limit: 50 } } }),
  ])
  if (rr.status === 'fulfilled' && rr.value.data) results.value = rr.value.data.data
  if (rb.status === 'fulfilled' && rb.value.data) topEntries.value = rb.value.data.data.slice(0, 5)
  if (rm.status === 'fulfilled' && rm.value.data) myRank.value = rm.value.data
  if (rt.status === 'fulfilled' && rt.value.data) {
    ongoingTests.value = rt.value.data.data.filter(t => t.student_status === 'in_progress')
  }
  isLoading.value = false
  requestAnimationFrame(() => { ready.value = true })
})

const bestScore = computed(() =>
  results.value.length ? Math.max(...results.value.map(r => r.total_score)) : null)


const greeting = computed(() => {
  const h = new Date().getHours()
  return h < 11 ? 'Selamat pagi' : h < 15 ? 'Selamat siang' : h < 18 ? 'Selamat sore' : 'Selamat malam'
})

const firstName = computed(() => auth.user?.name?.split(' ')[0] ?? '')

const todayStr = computed(() =>
  new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }))

function timeAgo(iso: string) {
  const diff = Date.now() - new Date(iso).getTime()
  const mins  = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days  = Math.floor(diff / 86400000)
  if (mins < 60)  return `${mins} menit lalu`
  if (hours < 24) return `${hours} jam lalu`
  if (days === 1) return 'Kemarin'
  return `${days} hari lalu`
}

function scoreClass(s: number) {
  return s >= 70 ? 'good' : s >= 40 ? 'mid' : 'low'
}

function categoryLabel(cat: string) {
  const map: Record<string, string> = { tka_saintek: 'TKA Saintek', tka_soshum: 'TKA Soshum', smbt: 'SMBT' }
  return map[cat] ?? cat
}

async function handleResume(testId: string) {
  resumingId.value = testId
  try {
    const sessionId = await sessionStore.startOrResume(testId)
    router.push({ name: 'test-session', params: { sessionId } })
  } finally {
    resumingId.value = null
  }
}

/* Leaderboard avatars */
const gradients = [
  '#f59e0b,#d97706', '#a78bfa,#7c3aed', '#fb923c,#ea580c',
  '#34d399,#059669', '#60a5fa,#2563eb',
]
const avatarStyle = (rank: number) => {
  const [f, t] = gradients[(rank - 1) % gradients.length].split(',')
  return `background: linear-gradient(135deg, ${f}, ${t})`
}
const initials = (name: string) =>
  name.split(' ').map((w: string) => w[0]).join('').substring(0, 2).toUpperCase()
const medal = (r: number) => r === 1 ? '🥇' : r === 2 ? '🥈' : r === 3 ? '🥉' : ''
</script>

<template>
  <div class="dash" :class="{ ready }">

    <!-- ── Hero ─────────────────────────────────────────────────────────────── -->
    <div class="hero">
      <div class="hero-dots" aria-hidden="true" />
      <div class="hero-body">
        <p class="hero-greet">{{ greeting }},</p>
        <h1 class="hero-name">{{ firstName || auth.user?.name }}</h1>
        <p class="hero-date">{{ todayStr }}</p>
        <RouterLink to="/tests" class="hero-cta">
          <svg viewBox="0 0 20 20" fill="currentColor" width="15" height="15">
            <path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z"/>
          </svg>
          Mulai Ujian
        </RouterLink>
      </div>
    </div>

    <div v-if="isLoading" class="loading">
      <div class="loading-dots"><span /><span /><span /></div>
    </div>

    <template v-else>

      <!-- ── Stats ───────────────────────────────────────────────────────────── -->
      <div class="stats-row">
        <div class="stat stat--blue">
          <div class="stat-accent" />
          <div class="stat-num">{{ results.length }}</div>
          <div class="stat-label">Ujian Selesai</div>
          <div class="stat-bar" style="--pct: 100%"><div class="stat-fill" /></div>
        </div>

        <div class="stat stat--score" :class="bestScore !== null ? `stat--${scoreClass(bestScore)}` : ''">
          <div class="stat-accent" />
          <div class="stat-num" :class="bestScore !== null ? scoreClass(bestScore) : ''">
            {{ bestScore?.toFixed(1) ?? '—' }}
          </div>
          <div class="stat-label">Skor Terbaik</div>
          <div v-if="bestScore !== null" class="stat-bar" :style="`--pct: ${Math.min(bestScore,100)}%`">
            <div class="stat-fill" />
          </div>
        </div>

        <div class="stat stat--warm">
          <div class="stat-accent" />
          <div class="stat-num rank-num">{{ myRank ? `#${myRank.rank}` : '—' }}</div>
          <div class="stat-label">Peringkat Global</div>
          <div class="stat-bar" style="--pct: 0%"><div class="stat-fill stat-fill--warm" /></div>
        </div>
      </div>

      <!-- ── Ongoing session(s) ──────────────────────────────────────────────── -->
      <div v-if="ongoingTests.length > 0" class="ongoing-section">
        <button
          v-for="t in ongoingTests"
          :key="t.id"
          class="ongoing-card"
          :disabled="resumingId === t.id"
          @click="handleResume(t.id)"
        >
          <div class="ongoing-pulse" aria-hidden="true" />
          <div class="ongoing-icon">
            <svg viewBox="0 0 20 20" fill="currentColor" width="18" height="18">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd"/>
            </svg>
          </div>
          <div class="ongoing-body">
            <p class="ongoing-label">Ujian Berlangsung</p>
            <p class="ongoing-title">{{ t.title }}</p>
            <p class="ongoing-meta">{{ categoryLabel(t.category) }} · {{ t.duration_minutes }} menit</p>
          </div>
          <span class="ongoing-cta">{{ resumingId === t.id ? 'Membuka...' : 'Lanjutkan →' }}</span>
        </button>
      </div>

      <!-- ── Two columns ────────────────────────────────────────────────────── -->
      <div class="columns">

        <!-- Hasil Terbaru -->
        <section class="col-main">
          <div class="section-head">
            <h2 class="section-title">Hasil Terbaru</h2>
            <RouterLink to="/results" class="section-link">Lihat semua →</RouterLink>
          </div>

          <div v-if="results.length === 0" class="empty">
            <div class="empty-icon">📚</div>
            <p class="empty-msg">Belum ada ujian yang dikerjakan.</p>
            <RouterLink to="/tests" class="btn-primary">Mulai sekarang</RouterLink>
          </div>

          <div v-else class="ujian-list">
            <RouterLink
              v-for="(r, i) in results.slice(0, 5)"
              :key="r.id"
              :to="`/results/${r.id}`"
              class="ujian-card ujian-card--done"
              :class="scoreClass(r.total_score)"
              :style="`--i: ${i}`"
            >
              <div class="ujian-bar" :class="scoreClass(r.total_score)" />
              <div class="ujian-content">
                <h3 class="ujian-title">{{ r.test_title }}</h3>
                <div class="ujian-score-section">
                  <div class="ujian-score-row">
                    <span class="ujian-score-label">Skor</span>
                    <span class="ujian-score-val" :class="scoreClass(r.total_score)">
                      {{ r.total_score.toFixed(1) }}
                    </span>
                  </div>
                  <div class="ujian-progress-wrap">
                    <div
                      class="ujian-progress-bar"
                      :class="scoreClass(r.total_score)"
                      :style="`width: ${ready ? Math.min(Math.max(r.total_score, 0), 100) : 0}%`"
                    />
                  </div>
                </div>
                <div class="ujian-footer">
                  <span class="ujian-status ujian-status--done">
                    <span class="status-dot" />
                    Selesai
                  </span>
                  <span class="ujian-time">{{ timeAgo(r.completed_at) }}</span>
                </div>
              </div>
            </RouterLink>
          </div>
        </section>

        <!-- Peringkat Teratas -->
        <section class="col-side panel">
          <div class="panel-head">
            <h2 class="panel-title">Peringkat Teratas</h2>
            <RouterLink to="/leaderboard" class="section-link">Lihat semua →</RouterLink>
          </div>

          <div v-if="topEntries.length === 0" class="empty">
            <p class="empty-msg">Belum ada data peringkat.</p>
          </div>

          <div v-else class="lb-list">
            <div
              v-for="(e, i) in topEntries"
              :key="String(e.student_id)"
              class="lb-row"
              :class="{ 'lb-row--me': myRank && e.student_id === myRank.student_id }"
              :style="`--i: ${i}`"
            >
              <span class="lb-pos">{{ medal(e.rank) || `#${e.rank}` }}</span>
              <div class="lb-avatar" :style="avatarStyle(e.rank)">{{ initials(e.student_name) }}</div>
              <span class="lb-name">{{ e.student_name }}</span>
              <span class="lb-score">{{ e.total_score.toFixed(1) }}</span>
            </div>

            <template v-if="myRank && !topEntries.some(e => e.student_id === myRank!.student_id)">
              <div class="lb-sep">···</div>
              <div class="lb-row lb-row--me">
                <span class="lb-pos">#{{ myRank.rank }}</span>
                <div class="lb-avatar" style="background: linear-gradient(135deg,#4f8ef7,#3b7be8)">
                  {{ initials(auth.user?.name ?? 'Me') }}
                </div>
                <span class="lb-name lb-name--me">Kamu</span>
                <span class="lb-score">{{ myRank.total_score.toFixed(1) }}</span>
              </div>
            </template>
          </div>
        </section>

      </div>
    </template>
  </div>
</template>

<style scoped>
/* ─── Animations ────────────────────────────────────────────────────────────── */
@keyframes slide-up {
  from { opacity: 0; transform: translateY(14px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes dot-bounce {
  0%,80%,100% { transform: scale(0.6); opacity: 0.4; }
  40%          { transform: scale(1);   opacity: 1; }
}

.dash { display: flex; flex-direction: column; gap: 1.25rem; }

.dash.ready .hero           { animation: slide-up 0.45s ease both; }
.dash.ready .stats-row      { animation: slide-up 0.45s 0.08s ease both; }
.dash.ready .ongoing-section { animation: slide-up 0.45s 0.14s ease both; }
.dash.ready .columns        { animation: slide-up 0.45s 0.20s ease both; }
.dash.ready .ujian-card     { animation: slide-up 0.35s calc(var(--i) * 55ms + 280ms) ease both; }

/* ─── Hero ──────────────────────────────────────────────────────────────────── */
.hero {
  position: relative;
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 2rem 2.25rem;
  border-radius: 18px;
  overflow: hidden;
  background: linear-gradient(135deg,
    color-mix(in srgb, var(--accent) 14%, var(--bg-surface)) 0%,
    var(--bg-surface) 60%,
    color-mix(in srgb, var(--accent) 6%, var(--bg-surface)) 100%);
  border: 1px solid color-mix(in srgb, var(--accent) 22%, var(--border));
  box-shadow: 0 1px 3px rgba(0,0,0,.04), inset 0 1px 0 rgba(255,255,255,.6);
}
.hero-dots {
  position: absolute;
  inset: 0;
  background-image: radial-gradient(circle, color-mix(in srgb, var(--accent) 25%, transparent) 1px, transparent 1px);
  background-size: 22px 22px;
  opacity: .35;
  pointer-events: none;
}
.hero-body { position: relative; z-index: 1; }
.hero-greet {
  margin: 0 0 2px;
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--accent);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}
.hero-name {
  margin: 0 0 4px;
  font-size: 2.1rem;
  font-weight: 800;
  color: var(--text-heading);
  line-height: 1.1;
  letter-spacing: -0.5px;
}
.hero-date { margin: 0 0 1.25rem; font-size: 0.75rem; color: var(--text-muted); }
.hero-cta {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  padding: 0.6rem 1.25rem;
  border-radius: 10px;
  background: var(--accent);
  color: #fff;
  font-size: 0.825rem;
  font-weight: 700;
  text-decoration: none;
  transition: background 0.15s, transform 0.15s, box-shadow 0.15s;
  box-shadow: 0 2px 8px color-mix(in srgb, var(--accent) 40%, transparent);
}
.hero-cta:hover {
  background: var(--accent-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 14px color-mix(in srgb, var(--accent) 45%, transparent);
}

/* ─── Stats ─────────────────────────────────────────────────────────────────── */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.875rem;
}
.stat {
  position: relative;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 1.125rem 1.25rem 0.875rem;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  transition: border-color 0.15s, box-shadow 0.15s, transform 0.15s;
}
.stat:hover { transform: translateY(-2px); box-shadow: 0 4px 16px rgba(0,0,0,.07); }
.stat-accent {
  position: absolute;
  top: 0; left: 0; right: 0;
  height: 3px;
  border-radius: 14px 14px 0 0;
}
.stat--blue .stat-accent  { background: var(--accent); }
.stat--good .stat-accent  { background: var(--success); }
.stat--mid .stat-accent   { background: var(--warning); }
.stat--low .stat-accent   { background: var(--danger); }
.stat--warm .stat-accent  { background: var(--warm); }
.stat--score .stat-accent { background: var(--accent); }
.stat-num {
  font-family: monospace;
  font-size: 1.875rem;
  font-weight: 800;
  color: var(--accent);
  line-height: 1;
  margin-top: 2px;
  font-variant-numeric: tabular-nums;
}
.stat-num.good { color: var(--success); }
.stat-num.mid  { color: var(--warning); }
.stat-num.low  { color: var(--danger); }
.rank-num      { color: var(--warm); }
.stat-label {
  font-size: 0.7rem;
  color: var(--text-muted);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 0.5rem;
}
.stat-bar {
  height: 3px;
  background: var(--border);
  border-radius: 99px;
  overflow: hidden;
  margin-top: auto;
}
.stat-fill {
  height: 100%;
  width: var(--pct);
  background: var(--accent);
  border-radius: 99px;
  transition: width 0.9s cubic-bezier(.25,.8,.25,1);
}
.stat-fill--warm { background: var(--warm); }

/* ─── Two-column layout ─────────────────────────────────────────────────────── */
.columns {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 1.25rem;
  align-items: start;
}
.col-main { display: flex; flex-direction: column; gap: 0.875rem; }

/* ─── Section header ────────────────────────────────────────────────────────── */
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.section-title {
  margin: 0;
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--text-primary);
  text-transform: uppercase;
  letter-spacing: 0.07em;
}
.section-link { font-size: 0.72rem; color: var(--accent); font-weight: 600; text-decoration: none; }
.section-link:hover { text-decoration: underline; }

/* ─── Empty state ───────────────────────────────────────────────────────────── */
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.625rem;
  padding: 2.5rem 1.5rem;
  text-align: center;
}
.empty-icon { font-size: 2.25rem; }
.empty-msg  { margin: 0; font-size: 0.875rem; color: var(--text-muted); }
.btn-primary {
  margin-top: 0.25rem;
  padding: 0.5rem 1.25rem;
  border-radius: 8px;
  background: var(--accent);
  color: #fff;
  font-size: 0.825rem;
  font-weight: 600;
  text-decoration: none;
  transition: background 0.15s;
}
.btn-primary:hover { background: var(--accent-hover); }

/* ─── Ujian list ────────────────────────────────────────────────────────────── */
.ujian-list { display: flex; flex-direction: column; gap: 0.75rem; }

/* ─── Ongoing card ──────────────────────────────────────────────────────────── */
@keyframes pulse-ring {
  0%   { opacity: 0.6; transform: scale(1); }
  70%  { opacity: 0;   transform: scale(1.8); }
  100% { opacity: 0;   transform: scale(1.8); }
}

.ongoing-section { display: flex; flex-direction: column; gap: 0.625rem; }

.ongoing-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  border-radius: 14px;
  border: 1px solid color-mix(in srgb, var(--warning) 35%, var(--border));
  background: color-mix(in srgb, var(--warning) 6%, var(--bg-surface));
  cursor: pointer;
  text-align: left;
  width: 100%;
  transition: background 0.15s, transform 0.15s, box-shadow 0.15s;
  overflow: hidden;
}
.ongoing-card:hover:not(:disabled) {
  background: color-mix(in srgb, var(--warning) 10%, var(--bg-surface));
  transform: translateY(-1px);
  box-shadow: 0 4px 16px color-mix(in srgb, var(--warning) 20%, transparent);
}
.ongoing-card:disabled { opacity: 0.7; cursor: wait; }

.ongoing-pulse {
  position: absolute;
  left: 1.25rem;
  top: 50%;
  transform: translateY(-50%);
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--warning) 30%, transparent);
  animation: pulse-ring 1.8s ease-out infinite;
  pointer-events: none;
}
.ongoing-icon {
  position: relative;
  z-index: 1;
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--warning) 18%, var(--bg-surface));
  border: 1.5px solid color-mix(in srgb, var(--warning) 40%, var(--border));
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--warning);
}
.ongoing-body { flex: 1; min-width: 0; }
.ongoing-label {
  margin: 0 0 1px;
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--warning);
  text-transform: uppercase;
  letter-spacing: 0.07em;
}
.ongoing-title {
  margin: 0 0 2px;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-heading);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.ongoing-meta { margin: 0; font-size: 0.72rem; color: var(--text-muted); }
.ongoing-cta  { flex-shrink: 0; font-size: 0.78rem; font-weight: 700; color: var(--warning); }

/* ─── Ujian cards ────────────────────────────────────────────────────────────── */
.ujian-card {
  display: flex;
  align-items: stretch;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
  text-decoration: none;
  color: inherit;
  cursor: pointer;
  width: 100%;
  text-align: left;
  transition: transform 0.15s, box-shadow 0.15s, border-color 0.15s;
}
.ujian-card:hover { transform: translateY(-2px); box-shadow: 0 4px 16px rgba(0,0,0,.07); }
.ujian-card:disabled { opacity: 0.7; cursor: wait; }

/* colored left border strip */
.ujian-bar {
  width: 4px;
  flex-shrink: 0;
}
.ujian-bar.good        { background: var(--success); }
.ujian-bar.mid         { background: var(--warning); }
.ujian-bar.low         { background: var(--danger); }
.ujian-bar--ongoing    { background: var(--warning); }

.ujian-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem 1.25rem;
}

.ujian-category {
  margin: 0;
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--accent);
  text-transform: uppercase;
  letter-spacing: 0.07em;
}

.ujian-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 700;
  color: var(--text-heading);
  line-height: 1.3;
}

/* score section (completed only) */
.ujian-score-section { display: flex; flex-direction: column; gap: 0.375rem; }
.ujian-score-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}
.ujian-score-label {
  font-size: 0.72rem;
  color: var(--text-muted);
}
.ujian-score-val {
  font-family: monospace;
  font-size: 1rem;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}
.ujian-score-val.good { color: var(--success); }
.ujian-score-val.mid  { color: var(--warning); }
.ujian-score-val.low  { color: var(--danger); }

.ujian-progress-wrap {
  height: 4px;
  background: var(--border);
  border-radius: 99px;
  overflow: hidden;
}
.ujian-progress-bar {
  height: 100%;
  border-radius: 99px;
  transition: width 0.9s cubic-bezier(.25,.8,.25,1);
}
.ujian-progress-bar.good { background: var(--success); }
.ujian-progress-bar.mid  { background: var(--warning); }
.ujian-progress-bar.low  { background: var(--danger); }

/* footer row */
.ujian-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.ujian-status {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.72rem;
  font-weight: 600;
}
.ujian-status--done    { color: var(--success); }
.ujian-status--ongoing { color: var(--warning); }

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
  flex-shrink: 0;
}
.ujian-time {
  font-size: 0.68rem;
  color: var(--text-muted);
}
.ujian-time--ongoing {
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--warning);
}

/* ─── Leaderboard panel ─────────────────────────────────────────────────────── */
.col-side {}
.panel {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
  border-bottom: 1px solid var(--border);
}
.panel-title {
  margin: 0;
  font-size: 0.72rem;
  font-weight: 700;
  color: var(--text-primary);
  text-transform: uppercase;
  letter-spacing: 0.07em;
}

.lb-list { display: flex; flex-direction: column; }
.lb-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.7rem 1.125rem;
  border-bottom: 1px solid var(--bg-input);
  transition: background 0.12s;
  animation: slide-up 0.35s calc(var(--i) * 50ms + 240ms) ease both;
}
.lb-row:last-child { border-bottom: none; }
.lb-row:hover { background: var(--bg-input); }
.lb-row--me { background: var(--accent-dim) !important; }
.lb-sep {
  text-align: center;
  padding: 0.25rem;
  font-size: 0.8rem;
  color: var(--text-muted);
  border-bottom: 1px solid var(--bg-input);
}
.lb-pos {
  width: 2rem;
  text-align: center;
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-muted);
  flex-shrink: 0;
}
.lb-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.6rem;
  font-weight: 800;
  color: #fff;
  flex-shrink: 0;
}
.lb-name {
  flex: 1;
  font-size: 0.825rem;
  font-weight: 500;
  color: var(--text-primary);
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.lb-name--me { color: var(--accent); font-weight: 700; }
.lb-score {
  font-family: monospace;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--accent);
  flex-shrink: 0;
}

/* ─── Loading ────────────────────────────────────────────────────────────────── */
.loading { display: flex; justify-content: center; padding: 3.5rem; }
.loading-dots { display: flex; gap: 6px; align-items: center; }
.loading-dots span {
  width: 8px; height: 8px;
  border-radius: 50%;
  background: var(--accent);
  animation: dot-bounce 1.4s infinite ease-in-out;
}
.loading-dots span:nth-child(2) { animation-delay: 0.2s; }
.loading-dots span:nth-child(3) { animation-delay: 0.4s; }

/* ─── Responsive ─────────────────────────────────────────────────────────────── */
@media (max-width: 1100px) {
  .columns { grid-template-columns: 1fr 280px; }
}
@media (max-width: 768px) {
  .hero { padding: 1.5rem; }
  .hero-name { font-size: 1.625rem; }
  .stats-row { grid-template-columns: repeat(2, 1fr); gap: 0.625rem; }
  .columns { grid-template-columns: 1fr; }
}
</style>
