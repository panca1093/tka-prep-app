<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
    client.GET('/leaderboard', { params: { query: { scope: scope.value as 'global' | 'tka' | 'smbt' | 'week' } } }),
    client.GET('/leaderboard/me', { params: { query: { scope: scope.value as 'global' | 'tka' | 'smbt' | 'week' } } }),
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
  if (expandedId.value === testId) {
    expandedId.value = ''
    return
  }
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

const medal = (rank: number) =>
  rank === 1 ? '🥇' : rank === 2 ? '🥈' : rank === 3 ? '🥉' : ''

const medalColor = (rank: number) =>
  rank === 1 ? '#FFD700' : rank === 2 ? '#C0C0C0' : rank === 3 ? '#CD7F32' : ''

const topScorer = (testId: string): { name: string; score: number } | null => {
  const board = testBoards.value[testId]
  if (!board || board.length === 0) return null
  return { name: board[0].student_name, score: board[0].total_score }
}
</script>

<template>
  <div>
    <h1 class="page-title">Leaderboard</h1>

    <!-- Global section -->
    <div class="section">
      <div class="section-header">
        <h2 class="section-title">Global Ranking</h2>
        <select v-model="scope" class="scope-select" @change="fetchGlobal">
          <option value="global">All Tests</option>
          <option value="tka">TKA</option>
          <option value="smbt">SMBT</option>
          <option value="week">This Week</option>
        </select>
      </div>

      <div v-if="myRank" class="my-rank-card">
        <span class="my-rank-label">Your rank</span>
        <span class="my-rank-val">#{{ myRank.rank }}</span>
        <span class="my-rank-score">{{ myRank.total_score.toFixed(1) }} pts · {{ myRank.test_count }} tests</span>
      </div>

      <div v-if="isGlobalLoading" class="loading">Loading…</div>

      <div v-else class="board">
        <div class="board-header">
          <span>Rank</span>
          <span>Student</span>
          <span>Tests</span>
          <span>Score</span>
        </div>
        <div v-for="e in entries" :key="String(e.student_id)" class="board-row">
          <span class="rank-cell" :style="{ color: medalColor(e.rank) }">
            {{ medal(e.rank) || `#${e.rank}` }}
          </span>
          <span class="name-cell">{{ e.student_name }}</span>
          <span class="meta-cell">{{ e.test_count }}</span>
          <span class="score-cell">{{ e.total_score.toFixed(1) }}</span>
        </div>
      </div>
    </div>

    <!-- Per-test cards -->
    <div class="section">
      <h2 class="section-title">Per Test</h2>

      <div v-if="tests.length === 0" class="empty-state">No tests available.</div>

      <div v-else class="test-cards">
        <div v-for="t in tests" :key="t.id" class="test-card" :class="{ expanded: expandedId === t.id }">
          <button class="card-summary" @click="toggleTest(t.id)">
            <div class="card-left">
              <div class="card-icon">{{ expandedId === t.id ? '▾' : '▸' }}</div>
              <div class="card-info">
                <div class="card-title">{{ t.title }}</div>
                <div class="card-meta">
                  {{ t.duration_minutes }} min · {{ t.questions.length }} questions
                  <span v-if="t.education_level" class="card-level">{{ t.education_level.toUpperCase() }}</span>
                </div>
              </div>
            </div>
            <div class="card-right">
              <template v-if="loadingBoard[t.id]">
                <span class="card-loading">Loading…</span>
              </template>
              <template v-else>
                <div class="card-top-scorer">
                  <template v-if="topScorer(t.id)">
                    <span class="top-medal">🥇</span>
                    <span class="top-name">{{ topScorer(t.id)!.name }}</span>
                    <span class="top-score">{{ topScorer(t.id)!.score.toFixed(1) }}</span>
                  </template>
                  <template v-else>
                    <span class="no-data">No results yet</span>
                  </template>
                </div>
              </template>
            </div>
          </button>

          <!-- Expanded leaderboard -->
          <div v-if="expandedId === t.id" class="card-expand">
            <div v-if="testBoards[t.id]?.length === 0" class="expand-empty">
              No one has completed this test yet.
            </div>
            <div v-else-if="testBoards[t.id]" class="expand-board">
              <div class="eb-header">
                <span>Rank</span>
                <span>Student</span>
                <span>Score</span>
                <span>Date</span>
              </div>
              <div v-for="e in testBoards[t.id]" :key="String(e.student_id)" class="eb-row">
                <span class="eb-rank" :style="{ color: medalColor(e.rank) }">
                  {{ medal(e.rank) || `#${e.rank}` }}
                </span>
                <span class="eb-name">{{ e.student_name }}</span>
                <span class="eb-score">{{ e.total_score.toFixed(1) }}</span>
                <span class="eb-date">{{ new Date(e.completed_at).toLocaleDateString('id-ID') }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 2rem; font-size: 1.5rem; font-weight: 800; }

.section { margin-bottom: 2.5rem; }

.section-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 1rem;
}
.section-title { margin: 0; font-size: 1rem; font-weight: 700; }

.scope-select {
  padding: 0.35rem 0.75rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-input);
  color: var(--text-primary); font-size: 0.8rem; outline: none;
}

.my-rank-card {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.75rem 1rem; margin-bottom: 1rem;
  background: color-mix(in srgb, var(--accent) 8%, transparent);
  border: 1px solid var(--bg-active-nav); border-radius: 10px;
}
.my-rank-label { font-size: 0.75rem; color: var(--text-muted); }
.my-rank-val { font-size: 1.15rem; font-weight: 800; color: var(--accent); }
.my-rank-score { margin-left: auto; font-size: 0.825rem; color: var(--text-muted); }

.loading { color: var(--text-muted); padding: 1rem 0; }
.empty-state { padding: 2rem; text-align: center; color: var(--text-muted); background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; font-size: 0.9rem; }

/* Global board */
.board {
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px;
  overflow: hidden;
}
.board-header, .board-row {
  display: grid;
  grid-template-columns: 60px 1fr 60px 80px;
  padding: 0.65rem 1rem;
  font-size: 0.85rem; align-items: center;
}
.board-header { color: var(--text-muted); font-size: 0.72rem; font-weight: 700; text-transform: uppercase; border-bottom: 1px solid var(--border); }
.board-row { border-bottom: 1px solid var(--bg-input); transition: background 0.1s; }
.board-row:last-child { border-bottom: none; }
.board-row:hover { background: var(--border); }
.rank-cell { font-weight: 700; }
.name-cell { font-weight: 500; }
.meta-cell { color: var(--text-muted); }
.score-cell { font-weight: 700; color: var(--accent); text-align: right; }

/* Test cards */
.test-cards { display: flex; flex-direction: column; gap: 0.5rem; }

.test-card {
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px;
  overflow: hidden; transition: border-color 0.15s;
}
.test-card:hover { border-color: color-mix(in srgb, var(--accent) 50%, var(--border)); }
.test-card.expanded { border-color: var(--accent); }

.card-summary {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.25rem; cursor: pointer; border: none; background: none;
  width: 100%; text-align: left; color: inherit; font: inherit;
  gap: 1rem;
}
.card-summary:hover { background: color-mix(in srgb, var(--accent) 3%, transparent); }

.card-left { display: flex; align-items: center; gap: 0.75rem; min-width: 0; flex: 1; }
.card-icon { font-size: 0.85rem; color: var(--text-muted); width: 16px; flex-shrink: 0; }
.card-info { min-width: 0; }
.card-title { font-weight: 600; font-size: 0.9rem; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.card-meta { font-size: 0.72rem; color: var(--text-muted); margin-top: 3px; display: flex; gap: 0.5rem; align-items: center; }
.card-level {
  background: color-mix(in srgb, var(--accent) 15%, transparent);
  color: var(--accent); padding: 1px 6px; border-radius: 3px;
  font-size: 0.65rem; font-weight: 700;
}

.card-right { flex-shrink: 0; min-width: 160px; text-align: right; }
.card-loading { font-size: 0.75rem; color: var(--text-muted); }
.card-top-scorer { display: flex; align-items: center; gap: 0.375rem; justify-content: flex-end; font-size: 0.8rem; }
.top-medal { font-size: 1rem; }
.top-name { color: var(--text-primary); font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 120px; }
.top-score { color: var(--accent); font-weight: 700; }
.no-data { color: var(--text-muted); font-style: italic; font-size: 0.75rem; }

/* Expanded inner board */
.card-expand { border-top: 1px solid var(--border); }
.expand-empty { padding: 1.5rem; text-align: center; color: var(--text-muted); font-size: 0.85rem; }

.expand-board { }
.eb-header, .eb-row {
  display: grid;
  grid-template-columns: 55px 1fr 70px 85px;
  padding: 0.55rem 1.25rem; font-size: 0.8rem; align-items: center;
}
.eb-header { color: var(--text-muted); font-size: 0.7rem; font-weight: 700; text-transform: uppercase; }
.eb-row { border-top: 1px solid var(--bg-input); }
.eb-rank { font-weight: 700; }
.eb-name { font-weight: 500; }
.eb-score { font-weight: 700; color: var(--accent); text-align: right; }
.eb-date { color: var(--text-muted); text-align: right; font-size: 0.72rem; }

@media (max-width: 768px) {
  .card-right { min-width: 120px; }
  .card-summary { padding: 0.875rem 1rem; }
  .eb-header, .eb-row { grid-template-columns: 40px 1fr 60px 75px; padding: 0.5rem 1rem; font-size: 0.75rem; }
  .board-header, .board-row { grid-template-columns: 50px 1fr 50px 70px; font-size: 0.78rem; padding: 0.5rem 0.75rem; }
}
</style>
