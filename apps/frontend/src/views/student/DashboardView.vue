<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Result = components['schemas']['TestResultResponse']
type Entry = components['schemas']['LeaderboardEntryResponse']

const auth = useAuthStore()
const results = ref<Result[]>([])
const topEntries = ref<Entry[]>([])
const myRank = ref<Entry | null>(null)
const isLoading = ref(true)

onMounted(async () => {
  const [rr, rb, rm] = await Promise.allSettled([
    client.GET('/results'),
    client.GET('/leaderboard', { params: { query: { scope: 'global' } } }),
    client.GET('/leaderboard/me', { params: { query: {} } }),
  ])
  if (rr.status === 'fulfilled' && rr.value.data) results.value = rr.value.data.data
  if (rb.status === 'fulfilled' && rb.value.data) topEntries.value = rb.value.data.data.slice(0, 5)
  if (rm.status === 'fulfilled' && rm.value.data) myRank.value = rm.value.data
  isLoading.value = false
})

const bestScore = computed(() =>
  results.value.length ? Math.max(...results.value.map(r => r.total_score)) : null)

const greeting = computed(() => {
  const h = new Date().getHours()
  if (h < 11) return 'Selamat pagi'
  if (h < 15) return 'Selamat siang'
  if (h < 18) return 'Selamat sore'
  return 'Selamat malam'
})

const todayStr = computed(() =>
  new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }))

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
}

function scoreClass(s: number) {
  return s >= 70 ? 'good' : s >= 40 ? 'mid' : 'low'
}

const gradients = ['#f59e0b,#d97706','#a78bfa,#7c3aed','#fb923c,#ea580c','#34d399,#059669','#60a5fa,#2563eb']
const avatarStyle = (rank: number) => {
  const [f, t] = gradients[(rank - 1) % gradients.length].split(',')
  return `background: linear-gradient(135deg, ${f}, ${t})`
}
const initials = (name: string) => name.split(' ').map((w: string) => w[0]).join('').substring(0, 2).toUpperCase()
const medal = (r: number) => r === 1 ? '🥇' : r === 2 ? '🥈' : r === 3 ? '🥉' : ''
</script>

<template>
  <div class="dash">
    <!-- Hero -->
    <div class="hero">
      <div class="hero-left">
        <p class="hero-greet">{{ greeting }},</p>
        <h1 class="hero-name">{{ auth.user?.name }}</h1>
        <p class="hero-date">{{ todayStr }}</p>
      </div>
      <RouterLink to="/tests" class="hero-cta">
        <svg viewBox="0 0 20 20" fill="currentColor" width="18" height="18"><path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z"/></svg>
        Mulai Belajar
      </RouterLink>
    </div>

    <div v-if="isLoading" class="loading">
      <div class="loading-dots"><span/><span/><span/></div>
    </div>

    <template v-else>
      <!-- Stats row (3 cards — no rata-rata skor) -->
      <div class="stats-row">
        <div class="stat">
          <div class="stat-num">{{ results.length }}</div>
          <div class="stat-label">Ujian Selesai</div>
          <div class="stat-bar" style="--pct: 100%"><div class="stat-fill" /></div>
        </div>
        <div class="stat">
          <div class="stat-num" :class="bestScore !== null ? scoreClass(bestScore) : ''">
            {{ bestScore?.toFixed(1) ?? '—' }}
          </div>
          <div class="stat-label">Skor Terbaik</div>
          <div v-if="bestScore !== null" class="stat-bar" :style="`--pct: ${Math.min(bestScore,100)}%`">
            <div class="stat-fill" />
          </div>
        </div>
        <div class="stat stat--rank">
          <div class="stat-num rank-num">{{ myRank ? `#${myRank.rank}` : '—' }}</div>
          <div class="stat-label">Peringkat Global</div>
          <div class="stat-bar" style="--pct: 0%"><div class="stat-fill" /></div>
        </div>
      </div>

      <!-- Two column panels -->
      <div class="panels">
        <!-- Recent results -->
        <section class="panel">
          <div class="panel-head">
            <h2 class="panel-title">Hasil Terbaru</h2>
            <RouterLink to="/tests" class="panel-link">Lihat Semua →</RouterLink>
          </div>

          <div v-if="results.length === 0" class="empty">
            <div class="empty-icon">📚</div>
            <p class="empty-msg">Belum ada ujian yang dikerjakan.</p>
            <RouterLink to="/tests" class="btn-first">Mulai ujian pertamamu</RouterLink>
          </div>

          <div v-else class="result-list">
            <RouterLink
              v-for="r in results.slice(0, 6)"
              :key="r.id"
              :to="`/results/${r.id}`"
              class="result-row"
            >
              <div class="result-score" :class="scoreClass(r.total_score)">
                {{ r.total_score.toFixed(0) }}
              </div>
              <div class="result-body">
                <div class="result-title">{{ r.test_title }}</div>
                <div class="result-meta">
                  <span class="tag-correct">✓ {{ r.correct_count }}</span>
                  <span class="tag-wrong">✗ {{ r.wrong_count }}</span>
                  <span class="tag-blank">— {{ r.blank_count }}</span>
                  <span class="result-date">{{ formatDate(r.completed_at) }}</span>
                </div>
              </div>
              <svg class="result-chevron" viewBox="0 0 20 20" fill="currentColor" width="16" height="16"><path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd"/></svg>
            </RouterLink>
          </div>
        </section>

        <!-- Leaderboard widget -->
        <section class="panel">
          <div class="panel-head">
            <h2 class="panel-title">Peringkat Teratas</h2>
            <RouterLink to="/leaderboard" class="panel-link">Lihat Semua →</RouterLink>
          </div>

          <div v-if="topEntries.length === 0" class="empty">
            <p class="empty-msg">Belum ada data peringkat.</p>
          </div>
          <div v-else class="lb-list">
            <div
              v-for="e in topEntries"
              :key="String(e.student_id)"
              class="lb-row"
              :class="{ 'lb-row--me': myRank && e.student_id === myRank.student_id }"
            >
              <span class="lb-pos">{{ medal(e.rank) || `#${e.rank}` }}</span>
              <div class="lb-avatar" :style="avatarStyle(e.rank)">{{ initials(e.student_name) }}</div>
              <span class="lb-name">{{ e.student_name }}</span>
              <span class="lb-score">{{ e.total_score.toFixed(1) }}</span>
            </div>
            <template v-if="myRank && !topEntries.some(e => e.student_id === myRank!.student_id)">
              <div class="lb-ellipsis">···</div>
              <div class="lb-row lb-row--me">
                <span class="lb-pos">#{{ myRank.rank }}</span>
                <div class="lb-avatar" style="background: linear-gradient(135deg, #4f8ef7, #3b7be8)">
                  {{ initials(auth.user?.name ?? 'Me') }}
                </div>
                <span class="lb-name">Kamu</span>
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
.dash { display: flex; flex-direction: column; gap: 1.5rem; }

/* ─── Hero ─────────────────────────────────────────────────────────────────── */
.hero {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.75rem 2rem;
  border-radius: 16px;
  background: linear-gradient(135deg,
    color-mix(in srgb, var(--accent) 12%, var(--bg-surface)),
    var(--bg-surface));
  border: 1px solid color-mix(in srgb, var(--accent) 20%, var(--border));
}
.hero-greet { margin: 0; font-size: 0.8rem; color: var(--text-muted); }
.hero-name {
  margin: 0.1rem 0 0.25rem;
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-heading);
  line-height: 1.1;
}
.hero-date { margin: 0; font-size: 0.75rem; color: var(--text-muted); }
.hero-cta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.7rem 1.5rem;
  border-radius: 10px;
  background: var(--accent);
  color: #fff;
  font-size: 0.875rem;
  font-weight: 700;
  text-decoration: none;
  white-space: nowrap;
  transition: background 0.15s, transform 0.15s;
  flex-shrink: 0;
}
.hero-cta:hover { background: var(--accent-hover); transform: translateY(-1px); }

/* ─── Stats ────────────────────────────────────────────────────────────────── */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.875rem;
}
.stat {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.125rem 1.25rem 0.875rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  transition: border-color 0.15s;
}
.stat:hover { border-color: color-mix(in srgb, var(--accent) 35%, var(--border)); }
.stat-num {
  font-family: monospace;
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--accent);
  line-height: 1;
}
.stat-num.good { color: var(--success); }
.stat-num.mid { color: var(--warning); }
.stat-num.low { color: var(--danger); }
.rank-num { color: var(--warm); }
.stat-label { font-size: 0.72rem; color: var(--text-muted); font-weight: 500; margin-bottom: 0.5rem; }
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
  transition: width 0.8s ease;
}

/* ─── Two column panels ────────────────────────────────────────────────────── */
.panels { display: grid; grid-template-columns: 1fr 1fr; gap: 1.125rem; }

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
  font-size: 0.825rem;
  font-weight: 700;
  color: var(--text-primary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.panel-link { font-size: 0.75rem; color: var(--accent); text-decoration: none; }
.panel-link:hover { text-decoration: underline; }

/* ─── Empty ─────────────────────────────────────────────────────────────────── */
.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.625rem;
  padding: 2.5rem 1.5rem;
  text-align: center;
}
.empty-icon { font-size: 2.25rem; }
.empty-msg { margin: 0; font-size: 0.875rem; color: var(--text-muted); }
.btn-first {
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
.btn-first:hover { background: var(--accent-hover); }

/* ─── Result list ──────────────────────────────────────────────────────────── */
.result-list { display: flex; flex-direction: column; }
.result-row {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid var(--bg-input);
  text-decoration: none;
  color: inherit;
  transition: background 0.1s;
}
.result-row:last-child { border-bottom: none; }
.result-row:hover { background: var(--bg-input); }
.result-score {
  font-family: monospace;
  font-size: 1.125rem;
  font-weight: 700;
  min-width: 2.75rem;
  text-align: right;
  flex-shrink: 0;
}
.result-score.good { color: var(--success); }
.result-score.mid { color: var(--warning); }
.result-score.low { color: var(--danger); }
.result-body { flex: 1; min-width: 0; }
.result-title { font-size: 0.85rem; font-weight: 600; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.result-meta { display: flex; gap: 0.5rem; align-items: center; margin-top: 2px; font-size: 0.7rem; }
.tag-correct { color: var(--success); font-weight: 600; }
.tag-wrong { color: var(--danger); font-weight: 600; }
.tag-blank { color: var(--text-muted); }
.result-date { color: var(--text-muted); margin-left: auto; }
.result-chevron { color: var(--text-muted); flex-shrink: 0; }

/* ─── Leaderboard list ─────────────────────────────────────────────────────── */
.lb-list { display: flex; flex-direction: column; }
.lb-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.7rem 1.25rem;
  border-bottom: 1px solid var(--bg-input);
  transition: background 0.1s;
}
.lb-row:last-child { border-bottom: none; }
.lb-row--me { background: var(--accent-dim); }
.lb-ellipsis { text-align: center; padding: 0.25rem; color: var(--text-muted); font-size: 0.875rem; border-bottom: 1px solid var(--bg-input); }
.lb-pos { width: 2rem; text-align: center; font-size: 0.875rem; font-weight: 700; color: var(--text-muted); flex-shrink: 0; }
.lb-avatar {
  width: 28px; height: 28px;
  border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.6rem; font-weight: 700; color: #fff;
  flex-shrink: 0;
}
.lb-name { flex: 1; font-size: 0.85rem; font-weight: 500; color: var(--text-primary); min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.lb-score { font-family: monospace; font-size: 0.875rem; font-weight: 700; color: var(--accent); }

/* ─── Loading ───────────────────────────────────────────────────────────────── */
.loading { display: flex; justify-content: center; padding: 3rem; }
.loading-dots { display: flex; gap: 6px; align-items: center; }
.loading-dots span {
  width: 8px; height: 8px;
  border-radius: 50%;
  background: var(--accent);
  animation: dot-bounce 1.4s infinite ease-in-out;
}
.loading-dots span:nth-child(2) { animation-delay: 0.2s; }
.loading-dots span:nth-child(3) { animation-delay: 0.4s; }
@keyframes dot-bounce { 0%,80%,100% { transform: scale(0.6); opacity: 0.5; } 40% { transform: scale(1); opacity: 1; } }

/* ─── Responsive ──────────────────────────────────────────────────────────── */
@media (max-width: 1024px) {
  .stats-row { grid-template-columns: repeat(3, 1fr); }
}
@media (max-width: 768px) {
  .hero { flex-direction: column; align-items: flex-start; gap: 1rem; padding: 1.25rem; }
  .hero-name { font-size: 1.5rem; }
  .stats-row { grid-template-columns: repeat(2, 1fr); gap: 0.625rem; }
  .panels { grid-template-columns: 1fr; }
}
</style>
