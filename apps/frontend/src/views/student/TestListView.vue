<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useTestStore } from '@/stores/test'
import { useSessionStore } from '@/stores/session'

const testStore = useTestStore()
const sessionStore = useSessionStore()
const router = useRouter()

const categoryFilter = ref<'all' | 'tka_saintek' | 'tka_soshum' | 'smbt'>('all')
const startingId = ref<string | null>(null)
const startError = ref('')

onMounted(() => testStore.fetchPublished())

async function handleStart(testId: string) {
  startingId.value = testId
  startError.value = ''
  try {
    const sessionId = await sessionStore.startOrResume(testId)
    router.push({ name: 'test-session', params: { sessionId } })
  } catch {
    startError.value = 'Could not start session. Try again.'
  } finally {
    startingId.value = null
  }
}

const filtered = () =>
  categoryFilter.value === 'all'
    ? testStore.tests
    : testStore.tests.filter((t) => t.category === categoryFilter.value)

const categoryLabel: Record<string, string> = {
  tka_saintek: 'TKA Saintek',
  tka_soshum: 'TKA Soshum',
  smbt: 'SMBT',
}

const difficultyColor: Record<string, string> = {
  easy: '#22c55e',
  medium: '#f59e0b',
  hard: '#ef4444',
}

const levelLabel: Record<string, string> = {
  sd: 'SD',
  smp: 'SMP',
  sma: 'SMA',
  smk: 'SMK',
}
</script>

<template>
  <div>
    <h1 class="page-title">Available Tests</h1>

    <div class="filters">
      <button
        v-for="cat in ['all', 'tka_saintek', 'tka_soshum', 'smbt']"
        :key="cat"
        class="filter-btn"
        :class="{ active: categoryFilter === cat }"
        @click="categoryFilter = cat as typeof categoryFilter"
      >
        {{ cat === 'all' ? 'All' : categoryLabel[cat] }}
      </button>
    </div>

    <p v-if="startError" class="error-msg">{{ startError }}</p>

    <div v-if="testStore.isLoading" class="loading">Loading tests…</div>

    <div v-else-if="filtered().length === 0" class="empty-state">
      No tests available in this category.
    </div>

    <div v-else class="test-grid">
      <div v-for="t in filtered()" :key="t.id" class="test-card">
        <div class="test-header">
          <span class="test-category">{{ categoryLabel[t.category] }}</span>
          <span v-if="t.education_level" class="test-level">{{ levelLabel[t.education_level] ?? t.education_level }}</span>
          <span class="test-difficulty" :style="{ color: difficultyColor[t.difficulty] }">
            {{ t.difficulty }}
          </span>
        </div>
        <h3 class="test-title">{{ t.title }}</h3>
        <p v-if="t.description" class="test-desc">{{ t.description }}</p>
        <div class="test-footer">
          <span class="test-meta">{{ t.duration_minutes }} min &nbsp;·&nbsp; {{ t.questions.length }} questions</span>
          <button
            class="btn-start"
            :disabled="startingId === t.id"
            @click="handleStart(t.id)"
          >
            {{ startingId === t.id ? 'Starting…' : 'Start' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 1.5rem; font-size: 1.5rem; font-weight: 800; }

.filters { display: flex; gap: 0.5rem; margin-bottom: 1.5rem; flex-wrap: wrap; }

.filter-btn {
  padding: 0.4rem 1rem;
  border-radius: 20px;
  border: 1px solid #1e2a45;
  background: transparent;
  color: #94a3b8;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.15s;
}
.filter-btn:hover { border-color: #4f8ef7; color: #f1f5f9; }
.filter-btn.active { background: #1a2f5c; border-color: #4f8ef7; color: #4f8ef7; }

.error-msg {
  padding: 0.6rem 0.75rem;
  border-radius: 8px;
  background: #450a0a;
  color: #fca5a5;
  font-size: 0.825rem;
  margin-bottom: 1rem;
}

.loading, .empty-state {
  color: #94a3b8;
  padding: 2rem;
  text-align: center;
  background: #141c2e;
  border: 1px solid #1e2a45;
  border-radius: 12px;
}

.test-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.test-card {
  background: #141c2e;
  border: 1px solid #1e2a45;
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  transition: border-color 0.15s;
}
.test-card:hover { border-color: #2d3f64; }

.test-header { display: flex; align-items: center; justify-content: space-between; }
.test-category { font-size: 0.72rem; font-weight: 600; color: #4f8ef7; text-transform: uppercase; letter-spacing: 0.05em; }
.test-difficulty { font-size: 0.75rem; font-weight: 600; text-transform: capitalize; }
.test-level {
  font-size: 0.7rem;
  font-weight: 600;
  color: #a78bfa;
  background: #2e1a5c;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
}

.test-title { margin: 0; font-size: 1rem; font-weight: 700; color: #f1f5f9; line-height: 1.4; }
.test-desc { margin: 0; font-size: 0.825rem; color: #94a3b8; line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.test-footer { display: flex; align-items: center; justify-content: space-between; margin-top: auto; padding-top: 0.75rem; border-top: 1px solid #1e2a45; }
.test-meta { font-size: 0.78rem; color: #64748b; }

.btn-start {
  padding: 0.45rem 1.1rem;
  border-radius: 8px;
  border: none;
  background: #4f8ef7;
  color: #fff;
  font-size: 0.825rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, opacity 0.15s;
}
.btn-start:hover { background: #3b7be8; }
.btn-start:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
