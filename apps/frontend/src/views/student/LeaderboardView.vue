<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Entry = components['schemas']['LeaderboardEntryResponse']

const scope = ref<'global' | 'tka' | 'smbt' | 'week'>('global')
const entries = ref<Entry[]>([])
const myRank = ref<Entry | null>(null)
const isLoading = ref(true)

async function fetchData() {
  isLoading.value = true
  const [resBoard, resMe] = await Promise.allSettled([
    client.GET('/leaderboard', { params: { query: { scope: scope.value as 'global' | 'tka' | 'smbt' | 'week' } } }),
    client.GET('/leaderboard/me', { params: { query: { scope: scope.value as 'global' | 'tka' | 'smbt' | 'week' } } }),
  ])
  if (resBoard.status === 'fulfilled' && resBoard.value.data) entries.value = resBoard.value.data.data
  myRank.value = resMe.status === 'fulfilled' && resMe.value.data ? resMe.value.data : null
  isLoading.value = false
}

onMounted(fetchData)
watch(scope, fetchData)

const medalColor = (rank: number) =>
  rank === 1 ? '#FFD700' : rank === 2 ? '#C0C0C0' : rank === 3 ? '#CD7F32' : '#94a3b8'
</script>

<template>
  <div>
    <h1 class="page-title">Leaderboard</h1>

    <div class="scope-tabs">
      <button
        v-for="s in ['global', 'tka', 'smbt', 'week']"
        :key="s"
        class="scope-btn"
        :class="{ active: scope === s }"
        @click="scope = s as typeof scope"
      >
        {{ s === 'global' ? 'Global' : s === 'tka' ? 'TKA' : s === 'smbt' ? 'SMBT' : 'This Week' }}
      </button>
    </div>

    <div v-if="myRank" class="my-rank-card">
      <span class="my-rank-label">Your rank</span>
      <span class="my-rank-val">#{{ myRank.rank }}</span>
      <span class="my-rank-score">{{ myRank.total_score.toFixed(1) }} pts</span>
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
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 1.5rem; font-size: 1.5rem; font-weight: 800; }

.scope-tabs { display: flex; gap: 0.5rem; margin-bottom: 1.5rem; }
.scope-btn {
  padding: 0.4rem 1rem; border-radius: 20px;
  border: 1px solid #1e2a45; background: transparent;
  color: #94a3b8; font-size: 0.8rem; cursor: pointer; transition: all 0.15s;
}
.scope-btn:hover { border-color: #4f8ef7; color: #f1f5f9; }
.scope-btn.active { background: #1a2f5c; border-color: #4f8ef7; color: #4f8ef7; }

.my-rank-card {
  display: flex; align-items: center; gap: 1rem;
  padding: 0.875rem 1.25rem; margin-bottom: 1.25rem;
  background: rgba(79,142,247,0.08); border: 1px solid #1a2f5c; border-radius: 10px;
}
.my-rank-label { font-size: 0.8rem; color: #94a3b8; }
.my-rank-val { font-size: 1.25rem; font-weight: 800; color: #4f8ef7; }
.my-rank-score { margin-left: auto; font-size: 0.875rem; color: #94a3b8; }

.loading { color: #94a3b8; }

.board {
  background: #141c2e; border: 1px solid #1e2a45; border-radius: 12px;
  overflow: hidden;
}

.board-header, .board-row {
  display: grid;
  grid-template-columns: 70px 1fr 80px 100px;
  padding: 0.75rem 1rem;
  font-size: 0.85rem;
}
.board-header { color: #64748b; font-size: 0.75rem; font-weight: 700; text-transform: uppercase; border-bottom: 1px solid #1e2a45; }
.board-row { border-bottom: 1px solid #0d1424; transition: background 0.1s; }
.board-row:last-child { border-bottom: none; }
.board-row:hover { background: #1a2535; }

.rank-cell { font-weight: 700; }
.name-cell { font-weight: 500; }
.meta-cell { color: #94a3b8; }
.score-cell { font-weight: 700; color: #4f8ef7; text-align: right; }
</style>
