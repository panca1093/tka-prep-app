<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Entry = components['schemas']['LeaderboardEntryResponse']
type TestEntry = components['schemas']['TestLeaderboardEntry']
type Test = components['schemas']['TestDetailResponse']

// ─── Global leaderboard ─────────────────────────────────────────────────────
const scope = ref<'global' | 'tka' | 'smbt' | 'week'>('global')
const entries = ref<Entry[]>([])
const myRank = ref<Entry | null>(null)
const isGlobalLoading = ref(true)

// ─── Per-test cards ─────────────────────────────────────────────────────────
const tests = ref<Test[]>([])
const expandedId = ref('')
const testBoards = ref<Record<string, TestEntry[]>>({})
const loadingBoard = ref<Record<string, boolean>>({})

onMounted(async () => {
  await Promise.all([fetchGlobal(), fetchTests()])
})

async function fetchGlobal() {
  isGlobalLoading.value = true
  const [resBoard, resMe] = await Promise.allSettled([
    client.GET('/leaderboard', { params: { query: { scope: scope.value } } }),
    client.GET('/leaderboard/me', { params: { query: { scope: scope.value } } }),
  ])
  if (resBoard.status === 'fulfilled' && resBoard.value.data) entries.value = resBoard.value.data.data
  myRank.value = resMe.status === 'fulfilled' && resMe.value.data ? resMe.value.data : null
  isGlobalLoading.value = false
}

async function fetchTests() {
  const { data } = await client.GET('/tests', { params: { query: { limit: 50, status: 'published' } } })
  if (data) tests.value = data.data
}

async function toggleTest(testId: string) {
  if (expandedId.value === testId) { expandedId.value = ''; return }
  expandedId.value = testId
  if (!testBoards.value[testId]) {
    loadingBoard.value[testId] = true
    const { data } = await client.GET('/tests/{testId}/leaderboard', {
      params: { path: { testId }, query: { limit: 100 } },
    })
    testBoards.value[testId] = data?.data ?? []
    loadingBoard.value[testId] = false
  }
}

const scopes = [
  { key: 'global', label: 'Semua' },
  { key: 'tka', label: 'TKA' },
  { key: 'smbt', label: 'SMBT' },
  { key: 'week', label: 'Minggu Ini' },
] as const

function setScope(s: typeof scope.value) {
  scope.value = s
  fetchGlobal()
}

const podium = computed(() => entries.value.slice(0, 3))
const rest = computed(() => entries.value.slice(3))

function initials(name: string) {
  return name.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase()
}

const avatarGradients = [
  'linear-gradient(135deg, #667eea, #764ba2)',
  'linear-gradient(135deg, #f093fb, #f5576c)',
  'linear-gradient(135deg, #4facfe, #00f2fe)',
  'linear-gradient(135deg, #43e97b, #38f9d7)',
  'linear-gradient(135deg, #fa709a, #fee140)',
  'linear-gradient(135deg, #a18cd1, #fbc2eb)',
  'linear-gradient(135deg, #ffecd2, #fcb69f)',
  'linear-gradient(135deg, #84fab0, #8fd3f4)',
  'linear-gradient(135deg, #d4fc79, #96e6a1)',
  'linear-gradient(135deg, #fddb92, #d1fdff)',
]

function avatarGradient(name: string) {
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return avatarGradients[Math.abs(hash) % avatarGradients.length]
}

const podiumOrder = computed(() => {
  // Display order for podium: 2nd left, 1st center, 3rd right
  if (podium.value.length === 0) return []
  if (podium.value.length === 1) return [podium.value[0]]
  if (podium.value.length === 2) return [podium.value[1], podium.value[0]]
  return [podium.value[1], podium.value[0], podium.value[2]]
})

function podiumHeight(rank: number) {
  return rank === 1 ? 90 : rank === 2 ? 65 : 45
}

function medalColor(rank: number) {
  return rank === 1 ? '#FFD700' : rank === 2 ? '#C0C0C0' : rank === 3 ? '#CD7F32' : ''
}

function topScorer(testId: string) {
  const board = testBoards.value[testId]
  if (!board || board.length === 0) return null
  return { name: board[0].student_name, score: board[0].total_score }
}
</script>

<template>
  <div class="lb-page">

    <!-- Page header -->
    <div class="page-hero">
      <h1 class="page-title">Papan Peringkat</h1>
      <p class="page-sub">Lihat posisi kamu di antara semua peserta ujian.</p>
    </div>

    <!-- Global section -->
    <section class="section">

      <!-- Scope tabs -->
      <div class="scope-tabs">
        <button
          v-for="s in scopes" :key="s.key"
          class="scope-tab"
          :class="{ active: scope === s.key }"
          @click="setScope(s.key)"
        >{{ s.label }}</button>
      </div>

      <!-- My rank banner -->
      <div v-if="myRank" class="my-rank-card">
        <div class="my-rank-avatar" :style="{ background: avatarGradient(myRank.student_name) }">
          {{ initials(myRank.student_name) }}
        </div>
        <div class="my-rank-info">
          <p class="my-rank-name">{{ myRank.student_name }}</p>
          <p class="my-rank-meta">{{ myRank.test_count }} ujian selesai</p>
        </div>
        <div class="my-rank-position">
          <div class="my-rank-num">#{{ myRank.rank }}</div>
          <div class="my-rank-pts">{{ myRank.total_score.toFixed(1) }} poin</div>
        </div>
      </div>

      <div v-if="isGlobalLoading" class="loading-wrap">
        <div class="loading-dots"><span/><span/><span/></div>
      </div>

      <template v-else>
        <!-- Podium (top 3) -->
        <div v-if="podium.length" class="podium">
          <div
            v-for="entry in podiumOrder"
            :key="String(entry.student_id)"
            class="podium-slot"
            :class="`rank-${entry.rank}`"
          >
            <div class="podium-avatar" :style="{ background: avatarGradient(entry.student_name) }">
              {{ initials(entry.student_name) }}
              <div class="podium-medal" :style="{ color: medalColor(entry.rank) }">
                {{ entry.rank === 1 ? '🥇' : entry.rank === 2 ? '🥈' : '🥉' }}
              </div>
            </div>
            <p class="podium-name">{{ entry.student_name }}</p>
            <p class="podium-score">{{ entry.total_score.toFixed(1) }}</p>
            <div class="podium-block" :style="{ height: podiumHeight(entry.rank) + 'px' }">
              <span class="podium-rank">#{{ entry.rank }}</span>
            </div>
          </div>
        </div>

        <!-- Rank 4+ table -->
        <div v-if="rest.length" class="board">
          <div class="board-header">
            <span>Peringkat</span>
            <span>Peserta</span>
            <span class="col-right">Ujian</span>
            <span class="col-right">Poin</span>
          </div>
          <div v-for="e in rest" :key="String(e.student_id)" class="board-row">
            <span class="rank-cell">#{{ e.rank }}</span>
            <span class="name-cell">
              <span class="mini-avatar" :style="{ background: avatarGradient(e.student_name) }">{{ initials(e.student_name) }}</span>
              {{ e.student_name }}
            </span>
            <span class="meta-cell col-right">{{ e.test_count }}</span>
            <span class="score-cell col-right">{{ e.total_score.toFixed(1) }}</span>
          </div>
        </div>

        <div v-if="!podium.length && !rest.length" class="empty-state">
          Belum ada data peringkat.
        </div>
      </template>
    </section>

    <!-- Per-test section -->
    <section class="section">
      <h2 class="section-title">
        <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16"><path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z"/></svg>
        Per Ujian
      </h2>

      <div v-if="tests.length === 0" class="empty-state">Belum ada ujian yang tersedia.</div>

      <div v-else class="test-cards">
        <div
          v-for="t in tests"
          :key="t.id"
          class="test-card"
          :class="{ expanded: expandedId === t.id }"
        >
          <button class="card-summary" @click="toggleTest(t.id)">
            <div class="card-chevron" :class="{ rotated: expandedId === t.id }">
              <svg viewBox="0 0 20 20" fill="currentColor" width="14" height="14"><path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd"/></svg>
            </div>
            <div class="card-info">
              <p class="card-title">{{ t.title }}</p>
              <p class="card-meta">
                {{ t.duration_minutes }} menit · {{ t.questions.length }} soal
                <span v-if="t.education_level" class="card-level">{{ t.education_level.toUpperCase() }}</span>
              </p>
            </div>
            <div class="card-right">
              <template v-if="loadingBoard[t.id]">
                <span class="card-loading">Memuat…</span>
              </template>
              <template v-else-if="expandedId !== t.id">
                <template v-if="topScorer(t.id)">
                  <span class="mini-avatar" :style="{ background: avatarGradient(topScorer(t.id)!.name) }">{{ initials(topScorer(t.id)!.name) }}</span>
                  <span class="top-name">{{ topScorer(t.id)!.name }}</span>
                  <span class="top-score">{{ topScorer(t.id)!.score.toFixed(1) }}</span>
                </template>
                <span v-else class="no-data">Belum ada</span>
              </template>
            </div>
          </button>

          <Transition name="expand">
            <div v-if="expandedId === t.id" class="card-expand">
              <div v-if="testBoards[t.id]?.length === 0" class="expand-empty">
                Belum ada yang menyelesaikan ujian ini.
              </div>
              <div v-else-if="testBoards[t.id]" class="expand-board">
                <div class="eb-header">
                  <span>Rank</span>
                  <span>Peserta</span>
                  <span class="col-right">Poin</span>
                  <span class="col-right">Tanggal</span>
                </div>
                <div v-for="e in testBoards[t.id]" :key="String(e.student_id)" class="eb-row" :class="{ 'eb-top': e.rank <= 3 }">
                  <span class="eb-rank" :style="e.rank <= 3 ? { color: medalColor(e.rank) } : {}">
                    {{ e.rank <= 3 ? (e.rank === 1 ? '🥇' : e.rank === 2 ? '🥈' : '🥉') : `#${e.rank}` }}
                  </span>
                  <span class="eb-name">
                    <span class="mini-avatar mini-avatar--sm" :style="{ background: avatarGradient(e.student_name) }">{{ initials(e.student_name) }}</span>
                    {{ e.student_name }}
                  </span>
                  <span class="eb-score col-right">{{ e.total_score.toFixed(1) }}</span>
                  <span class="eb-date col-right">{{ new Date(e.completed_at).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' }) }}</span>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* ─── Page ────────────────────────────────────────────────────────────────── */
.lb-page {
  max-width: 720px;
  margin: 0 auto;
  display: flex; flex-direction: column; gap: 2rem;
}

.page-hero { display: flex; flex-direction: column; gap: 0.25rem; }
.page-title {
  margin: 0;
  font-family: 'Cormorant Garamond', serif;
  font-size: 2rem; font-weight: 700;
  color: var(--text-heading);
}
.page-sub { margin: 0; font-size: 0.875rem; color: var(--text-muted); }

/* ─── Sections ────────────────────────────────────────────────────────────── */
.section { display: flex; flex-direction: column; gap: 1rem; }

.section-title {
  margin: 0;
  display: flex; align-items: center; gap: 0.5rem;
  font-size: 0.875rem; font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase; letter-spacing: 0.05em;
}

/* ─── Scope tabs ──────────────────────────────────────────────────────────── */
.scope-tabs {
  display: flex; gap: 0.375rem;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 0.3rem;
  width: fit-content;
}
.scope-tab {
  padding: 0.35rem 1rem;
  border-radius: 7px; border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.8rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s;
}
.scope-tab:hover { color: var(--text-primary); }
.scope-tab.active { background: var(--accent); color: #fff; }

/* ─── My rank ─────────────────────────────────────────────────────────────── */
.my-rank-card {
  display: flex; align-items: center; gap: 0.875rem;
  padding: 0.875rem 1.25rem;
  background: var(--accent-dim);
  border: 1px solid color-mix(in srgb, var(--accent) 30%, var(--border));
  border-radius: 12px;
}
.my-rank-avatar {
  width: 40px; height: 40px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.78rem; font-weight: 800; color: #fff;
  flex-shrink: 0;
}
.my-rank-info { flex: 1; }
.my-rank-name { margin: 0; font-size: 0.9rem; font-weight: 700; color: var(--text-heading); }
.my-rank-meta { margin: 0; font-size: 0.72rem; color: var(--text-muted); }
.my-rank-position { text-align: right; flex-shrink: 0; }
.my-rank-num {
  font-family: 'JetBrains Mono', monospace;
  font-size: 1.375rem; font-weight: 700;
  color: var(--accent); line-height: 1;
}
.my-rank-pts { font-size: 0.72rem; color: var(--text-muted); }

/* ─── Loading ─────────────────────────────────────────────────────────────── */
.loading-wrap { display: flex; justify-content: center; padding: 2rem; }
.loading-dots { display: flex; gap: 6px; }
.loading-dots span {
  width: 8px; height: 8px; border-radius: 50%;
  background: var(--accent);
  animation: dot-bounce 1.4s infinite ease-in-out;
}
.loading-dots span:nth-child(2) { animation-delay: 0.2s; }
.loading-dots span:nth-child(3) { animation-delay: 0.4s; }
@keyframes dot-bounce { 0%,80%,100%{transform:scale(0.6);opacity:0.5}40%{transform:scale(1);opacity:1} }

/* ─── Podium ──────────────────────────────────────────────────────────────── */
.podium {
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 0.75rem;
  padding: 1.5rem 1rem 0;
}

.podium-slot {
  display: flex; flex-direction: column; align-items: center;
  gap: 0.5rem; flex: 1; max-width: 160px;
}
.podium-slot.rank-1 .podium-avatar { width: 72px; height: 72px; font-size: 1.1rem; }

.podium-avatar {
  width: 56px; height: 56px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.875rem; font-weight: 800; color: #fff;
  position: relative;
  box-shadow: 0 4px 16px rgba(0,0,0,0.2);
}
.podium-medal {
  position: absolute; bottom: -4px; right: -4px;
  font-size: 1rem; line-height: 1;
}

.podium-name {
  margin: 0;
  font-size: 0.78rem; font-weight: 700;
  color: var(--text-primary);
  text-align: center;
  max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.podium-score {
  margin: 0;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.85rem; font-weight: 700;
  color: var(--accent);
}

.podium-block {
  width: 100%;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-bottom: none;
  border-radius: 10px 10px 0 0;
  display: flex; align-items: center; justify-content: center;
}
.podium-slot.rank-1 .podium-block {
  background: color-mix(in srgb, var(--accent) 8%, var(--bg-surface));
  border-color: color-mix(in srgb, var(--accent) 30%, var(--border));
}
.podium-rank {
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.75rem; font-weight: 700;
  color: var(--text-muted);
}

/* ─── Board (rank 4+) ─────────────────────────────────────────────────────── */
.board {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-top: none;
  border-radius: 0 0 12px 12px;
  overflow: hidden;
}
.board-header, .board-row {
  display: grid;
  grid-template-columns: 70px 1fr 60px 80px;
  padding: 0.6rem 1rem;
  font-size: 0.825rem; align-items: center;
  gap: 0.5rem;
}
.board-header {
  color: var(--text-muted); font-size: 0.7rem; font-weight: 700;
  text-transform: uppercase; letter-spacing: 0.04em;
  border-bottom: 1px solid var(--border);
  background: var(--bg-input);
}
.board-row { border-bottom: 1px solid var(--bg-input); transition: background 0.1s; }
.board-row:last-child { border-bottom: none; }
.board-row:hover { background: var(--bg-input); }
.rank-cell { font-family: 'JetBrains Mono', monospace; font-weight: 700; color: var(--text-muted); font-size: 0.78rem; }
.name-cell { display: flex; align-items: center; gap: 0.5rem; font-weight: 500; overflow: hidden; }
.col-right { text-align: right; }
.meta-cell { color: var(--text-muted); }
.score-cell { font-family: 'JetBrains Mono', monospace; font-weight: 700; color: var(--accent); }

/* ─── Mini avatars ────────────────────────────────────────────────────────── */
.mini-avatar {
  display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 50%;
  font-size: 0.65rem; font-weight: 800; color: #fff;
  flex-shrink: 0;
}
.mini-avatar--sm { width: 22px; height: 22px; font-size: 0.58rem; }

/* ─── Empty state ─────────────────────────────────────────────────────────── */
.empty-state {
  padding: 2rem; text-align: center;
  color: var(--text-muted); font-size: 0.875rem;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
}

/* ─── Test cards ──────────────────────────────────────────────────────────── */
.test-cards { display: flex; flex-direction: column; gap: 0.5rem; }

.test-card {
  background: var(--bg-surface);
  border: 1.5px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
  transition: border-color 0.15s;
}
.test-card:hover { border-color: color-mix(in srgb, var(--accent) 40%, var(--border)); }
.test-card.expanded { border-color: var(--accent); }

.card-summary {
  display: flex; align-items: center; gap: 0.875rem;
  padding: 1rem 1.25rem;
  cursor: pointer; border: none; background: none;
  width: 100%; text-align: left; color: inherit; font: inherit;
}
.card-summary:hover { background: color-mix(in srgb, var(--accent) 3%, transparent); }

.card-chevron {
  color: var(--text-muted); flex-shrink: 0;
  transition: transform 0.2s;
}
.card-chevron.rotated { transform: rotate(180deg); }

.card-info { flex: 1; min-width: 0; }
.card-title { margin: 0; font-weight: 600; font-size: 0.9rem; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.card-meta { margin: 0; font-size: 0.72rem; color: var(--text-muted); margin-top: 2px; display: flex; align-items: center; gap: 0.4rem; }
.card-level {
  background: var(--accent-dim); color: var(--accent);
  padding: 1px 6px; border-radius: 3px;
  font-size: 0.6rem; font-weight: 700;
}

.card-right {
  display: flex; align-items: center; gap: 0.375rem;
  flex-shrink: 0; min-width: 140px; justify-content: flex-end;
}
.card-loading { font-size: 0.75rem; color: var(--text-muted); }
.top-name { font-size: 0.8rem; font-weight: 600; color: var(--text-primary); max-width: 100px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.top-score { font-family: 'JetBrains Mono', monospace; font-size: 0.8rem; font-weight: 700; color: var(--accent); }
.no-data { font-size: 0.75rem; color: var(--text-muted); font-style: italic; }

/* ─── Expand ──────────────────────────────────────────────────────────────── */
.card-expand { border-top: 1px solid var(--border); }
.expand-enter-active, .expand-leave-active { transition: all 0.2s ease; overflow: hidden; }
.expand-enter-from, .expand-leave-to { max-height: 0; opacity: 0; }
.expand-enter-to, .expand-leave-from { max-height: 2000px; opacity: 1; }

.expand-empty { padding: 1.5rem; text-align: center; color: var(--text-muted); font-size: 0.825rem; }

.eb-header, .eb-row {
  display: grid;
  grid-template-columns: 50px 1fr 70px 80px;
  padding: 0.55rem 1.25rem;
  font-size: 0.8rem; align-items: center; gap: 0.5rem;
}
.eb-header {
  color: var(--text-muted); font-size: 0.68rem; font-weight: 700;
  text-transform: uppercase; letter-spacing: 0.04em;
  background: var(--bg-input);
}
.eb-row { border-top: 1px solid var(--bg-input); transition: background 0.1s; }
.eb-row.eb-top { background: color-mix(in srgb, var(--accent) 3%, transparent); }
.eb-row:hover { background: var(--bg-input); }
.eb-rank { font-weight: 700; }
.eb-name { display: flex; align-items: center; gap: 0.4rem; font-weight: 500; overflow: hidden; }
.eb-score { font-family: 'JetBrains Mono', monospace; font-weight: 700; color: var(--accent); text-align: right; }
.eb-date { color: var(--text-muted); text-align: right; font-size: 0.72rem; }

/* ─── Mobile ──────────────────────────────────────────────────────────────── */
@media (max-width: 640px) {
  .podium-slot.rank-1 .podium-avatar { width: 60px; height: 60px; }
  .podium-avatar { width: 48px; height: 48px; }
  .podium-name { font-size: 0.7rem; }

  .board-header, .board-row { grid-template-columns: 55px 1fr 50px 65px; font-size: 0.78rem; padding: 0.55rem 0.75rem; }

  .card-right { min-width: 100px; }
  .top-name { max-width: 60px; }

  .eb-header, .eb-row { grid-template-columns: 40px 1fr 60px 70px; padding: 0.5rem 1rem; font-size: 0.75rem; }
}
</style>
