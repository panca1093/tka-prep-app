<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Entry = components['schemas']['LeaderboardEntryResponse']
type TestEntry = components['schemas']['TestLeaderboardEntry']
type Test = components['schemas']['TestDetailResponse']

const mode = ref<'global' | 'per-test'>('global')
const scope = ref<'global' | 'tka' | 'smbt' | 'week'>('global')
const entries = ref<Entry[]>([])
const myRank = ref<Entry | null>(null)
const isLoading = ref(true)

// Per-test state
const tests = ref<Test[]>([])
const selectedTestId = ref('')
const testEntries = ref<TestEntry[]>([])

onMounted(async () => {
  await fetchGlobal()
  await fetchTests()
})

watch(scope, fetchGlobal)

async function fetchTests() {
  const { data } = await client.GET('/tests', { params: { query: { limit: 50 } } })
  if (data) tests.value = data.data
}

async function fetchGlobal() {
  isLoading.value = true
  const [resBoard, resMe] = await Promise.allSettled([
    client.GET('/leaderboard', { params: { query: { scope: scope.value as 'global' | 'tka' | 'smbt' | 'week' } } }),
    client.GET('/leaderboard/me', { params: { query: { scope: scope.value as 'global' | 'tka' | 'smbt' | 'week' } } }),
  ])
  if (resBoard.status === 'fulfilled' && resBoard.value.data) entries.value = resBoard.value.data.data
  myRank.value = resMe.status === 'fulfilled' && resMe.value.data ? resMe.value.data : null
  isLoading.value = false
}

async function fetchPerTest() {
  if (!selectedTestId.value) return
  isLoading.value = true
  const { data, error } = await client.GET('/tests/{testId}/leaderboard', {
    params: { path: { testId: selectedTestId.value }, query: { limit: 100 } },
  })
  if (!error && data) testEntries.value = data.data ?? []
  isLoading.value = false
}

watch(selectedTestId, (id) => {
  if (id) fetchPerTest()
})

const medalColor = (rank: number) =>
  rank === 1 ? '#FFD700' : rank === 2 ? '#C0C0C0' : rank === 3 ? '#CD7F32' : 'var(--text-muted)'

const selectedTestTitle = () => tests.value.find(t => t.id === selectedTestId.value)?.title ?? ''
</script>

<template>
  <div>
    <h1 class="page-title">Leaderboard</h1>

    <!-- Mode tabs -->
    <div class="scope-tabs">
      <button class="scope-btn" :class="{ active: mode === 'global' }" @click="mode = 'global'">
        Global
      </button>
      <button class="scope-btn" :class="{ active: mode === 'per-test' }" @click="mode = 'per-test'">
        Per Test
      </button>
    </div>

    <!-- Global view -->
    <template v-if="mode === 'global'">
      <div class="scope-tabs">
        <button
          v-for="s in ['global', 'tka', 'smbt', 'week']"
          :key="s"
          class="scope-btn"
          :class="{ active: scope === s }"
          @click="scope = s as typeof scope"
        >
          {{ s === 'global' ? 'All' : s === 'tka' ? 'TKA' : s === 'smbt' ? 'SMBT' : 'Week' }}
        </button>
      </div>

      <div v-if="myRank" class="my-rank-card">
        <span class="my-rank-label">Your rank</span>
        <span class="my-rank-val">#{{ myRank.rank }}</span>
        <span class="my-rank-score">{{ myRank.total_score.toFixed(1) }} pts</span>
        <span class="my-rank-meta">{{ myRank.test_count }} tests</span>
      </div>

      <div v-if="isLoading" class="loading">Loading…</div>

      <div v-else class="board">
        <div class="board-header">
          <span>Rank</span>
          <span>Student</span>
          <span>Tests</span>
          <span>Score</span>
        </div>
        <div v-for="e in entries" :key="String(e.student_id)" class="board-row">
          <span class="rank-cell" :style="{ color: medalColor(e.rank) }">
            {{ e.rank <= 3 ? ['🥇','🥈','🥉'][e.rank - 1] : `#${e.rank}` }}
          </span>
          <span class="name-cell">{{ e.student_name }}</span>
          <span class="meta-cell">{{ e.test_count }}</span>
          <span class="score-cell">{{ e.total_score.toFixed(1) }}</span>
        </div>
      </div>
    </template>

    <!-- Per-Test view -->
    <template v-else>
      <div class="scope-tabs">
        <select v-model="selectedTestId" class="test-select">
          <option value="">— Select a test —</option>
          <option v-for="t in tests" :key="t.id" :value="t.id">
            {{ t.title }}
          </option>
        </select>
      </div>

      <div v-if="!selectedTestId" class="empty-state">
        Select a test to see per-test rankings.
      </div>

      <template v-else>
        <h2 class="test-title">{{ selectedTestTitle() }}</h2>

        <div v-if="isLoading" class="loading">Loading…</div>

        <div v-else-if="testEntries.length === 0" class="empty-state">
          No results for this test yet.
        </div>

        <div v-else class="board">
          <div class="board-header">
            <span>Rank</span>
            <span>Student</span>
            <span>Score</span>
            <span>Completed</span>
          </div>
          <div v-for="e in testEntries" :key="String(e.student_id)" class="board-row">
            <span class="rank-cell" :style="{ color: medalColor(e.rank) }">
              {{ e.rank <= 3 ? ['🥇','🥈','🥉'][Number(e.rank) - 1] : `#${e.rank}` }}
            </span>
            <span class="name-cell">{{ e.student_name }}</span>
            <span class="score-cell">{{ e.total_score.toFixed(1) }}</span>
            <span class="meta-cell">{{ new Date(e.completed_at).toLocaleDateString('id-ID') }}</span>
          </div>
        </div>
      </template>
    </template>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 1.5rem; font-size: 1.5rem; font-weight: 800; }
.test-title { margin: 0 0 1rem; font-size: 1rem; font-weight: 600; color: var(--text-primary); }

.scope-tabs { display: flex; gap: 0.5rem; margin-bottom: 1.25rem; flex-wrap: wrap; }
.scope-btn {
  padding: 0.4rem 1rem; border-radius: 20px;
  border: 1px solid var(--border); background: transparent;
  color: var(--text-muted); font-size: 0.8rem; cursor: pointer; transition: all 0.15s;
}
.scope-btn:hover { border-color: var(--accent); color: var(--text-primary); }
.scope-btn.active { background: var(--bg-active-nav); border-color: var(--accent); color: var(--accent); }

.test-select {
  padding: 0.5rem 0.875rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-input);
  color: var(--text-primary); font-size: 0.875rem; outline: none;
  width: 100%; max-width: 400px;
}
.test-select:focus { border-color: var(--accent); }

.my-rank-card {
  display: flex; align-items: center; gap: 1rem;
  padding: 0.875rem 1.25rem; margin-bottom: 1.25rem;
  background: color-mix(in srgb, var(--accent) 8%, transparent);
  border: 1px solid var(--bg-active-nav); border-radius: 10px;
}
.my-rank-label { font-size: 0.8rem; color: var(--text-muted); }
.my-rank-val { font-size: 1.25rem; font-weight: 800; color: var(--accent); }
.my-rank-score { margin-left: auto; font-size: 0.875rem; color: var(--text-muted); }
.my-rank-meta { font-size: 0.75rem; color: var(--text-muted); }

.loading { color: var(--text-muted); }
.empty-state { padding: 2rem; text-align: center; color: var(--text-muted); background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; font-size: 0.9rem; }

.board {
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px;
  overflow: hidden;
}

.board-header, .board-row {
  display: grid;
  grid-template-columns: 70px 1fr 90px 100px;
  padding: 0.75rem 1rem;
  font-size: 0.85rem;
}
.board-header { color: var(--text-muted); font-size: 0.75rem; font-weight: 700; text-transform: uppercase; border-bottom: 1px solid var(--border); }
.board-row { border-bottom: 1px solid var(--bg-input); transition: background 0.1s; }
.board-row:last-child { border-bottom: none; }
.board-row:hover { background: var(--border); }

.rank-cell { font-weight: 700; }
.name-cell { font-weight: 500; }
.meta-cell { color: var(--text-muted); }
.score-cell { font-weight: 700; color: var(--accent); text-align: right; }

@media (max-width: 768px) {
  .board-header, .board-row { grid-template-columns: 50px 1fr 70px 80px; font-size: 0.78rem; padding: 0.6rem 0.75rem; }
}
</style>
