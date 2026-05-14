<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Result = components['schemas']['TestResultResponse']
type LeaderboardEntry = components['schemas']['LeaderboardEntryResponse']

const auth = useAuthStore()

const results = ref<Result[]>([])
const myRank = ref<LeaderboardEntry | null>(null)
const isLoading = ref(true)

onMounted(async () => {
  const [resResults, resRank] = await Promise.allSettled([
    client.GET('/results'),
    client.GET('/leaderboard/me', { params: { query: {} } }),
  ])
  if (resResults.status === 'fulfilled' && resResults.value.data) {
    results.value = resResults.value.data.data
  }
  if (resRank.status === 'fulfilled' && resRank.value.data) {
    myRank.value = resRank.value.data
  }
  isLoading.value = false
})

const bestScore = () => results.value.length
  ? Math.max(...results.value.map((r) => r.total_score))
  : null

const totalTests = () => results.value.length

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}
</script>

<template>
  <div>
    <h1 class="page-title">Dashboard</h1>
    <p class="greeting">Good to see you, {{ auth.user?.name }} 👋</p>

    <div v-if="isLoading" class="loading">Loading…</div>

    <template v-else>
      <div class="stats-row">
        <div class="stat-card">
          <div class="stat-value">{{ totalTests() }}</div>
          <div class="stat-label">Tests Taken</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ bestScore()?.toFixed(1) ?? '—' }}</div>
          <div class="stat-label">Best Score</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ myRank ? `#${myRank.rank}` : '—' }}</div>
          <div class="stat-label">Global Rank</div>
        </div>
      </div>

      <div class="section">
        <div class="section-header">
          <h2>Recent Results</h2>
          <RouterLink to="/tests" class="link-btn">Browse Tests →</RouterLink>
        </div>

        <div v-if="results.length === 0" class="empty-state">
          No tests taken yet.
          <RouterLink to="/tests">Start your first test</RouterLink>
        </div>

        <div v-else class="result-list">
          <div v-for="r in results.slice(0, 5)" :key="r.id" class="result-row">
            <div class="result-meta">
              <span class="result-score">{{ r.total_score.toFixed(1) }}</span>
              <div class="result-info">
                <span class="result-title">{{ r.test_title }}</span>
                <span class="result-counts">✓ {{ r.correct_count }} &nbsp;✗ {{ r.wrong_count }} &nbsp;— {{ r.blank_count }}</span>
              </div>
            </div>
            <div class="result-date">{{ formatDate(r.completed_at) }}</div>
            <RouterLink :to="`/results/${r.id}`" class="result-link">Details →</RouterLink>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 0.25rem; font-size: 1.5rem; font-weight: 800; }
.greeting { margin: 0 0 2rem; color: #94a3b8; font-size: 0.9rem; }

.loading { color: #94a3b8; }

.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: #141c2e;
  border: 1px solid #1e2a45;
  border-radius: 12px;
  padding: 1.25rem 1.5rem;
}

.stat-value { font-size: 2rem; font-weight: 800; color: #4f8ef7; }
.stat-label { font-size: 0.8rem; color: #94a3b8; margin-top: 0.25rem; }

.section { margin-bottom: 2rem; }

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
}

.section-header h2 { margin: 0; font-size: 1rem; font-weight: 700; }
.link-btn { font-size: 0.85rem; color: #4f8ef7; }

.empty-state {
  padding: 2rem;
  text-align: center;
  color: #94a3b8;
  background: #141c2e;
  border: 1px solid #1e2a45;
  border-radius: 12px;
  font-size: 0.9rem;
}

.result-list { display: flex; flex-direction: column; gap: 0.5rem; }

.result-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.875rem 1rem;
  background: #141c2e;
  border: 1px solid #1e2a45;
  border-radius: 10px;
}

.result-meta { flex: 1; display: flex; align-items: center; gap: 1rem; min-width: 0; }
.result-score { font-size: 1.25rem; font-weight: 700; color: #4f8ef7; min-width: 4rem; flex-shrink: 0; }
.result-info { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.result-title { font-size: 0.875rem; font-weight: 600; color: #e2e8f0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.result-counts { font-size: 0.8rem; color: #94a3b8; }
.result-date { font-size: 0.8rem; color: #64748b; }
.result-link { font-size: 0.8rem; color: #4f8ef7; white-space: nowrap; }
</style>
