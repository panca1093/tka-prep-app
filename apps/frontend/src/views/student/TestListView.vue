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
  easy: 'var(--success)',
  medium: 'var(--warning)',
  hard: 'var(--danger)',
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
          <template v-if="t.student_status === 'completed'">
            <span class="completed-badge">Completed</span>
            <RouterLink v-if="t.result_id" :to="'/results/' + t.result_id" class="btn-start btn-view-result">View Result</RouterLink>
          </template>
          <template v-else>
            <button
              class="btn-start"
              :disabled="startingId === t.id"
              @click="handleStart(t.id)"
            >
              {{ startingId === t.id ? (t.student_status === 'in_progress' ? 'Resuming…' : 'Starting…') : (t.student_status === 'in_progress' ? 'Resume' : 'Start') }}
            </button>
          </template>
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
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.15s;
}
.filter-btn:hover { border-color: var(--accent); color: var(--text-primary); }
.filter-btn.active { background: var(--bg-active-nav); border-color: var(--accent); color: var(--accent); }

.error-msg {
  padding: 0.6rem 0.75rem;
  border-radius: 8px;
  background: color-mix(in srgb, var(--danger) 15%, var(--bg-surface));
  color: var(--danger);
  font-size: 0.825rem;
  margin-bottom: 1rem;
}

.loading, .empty-state {
  color: var(--text-muted);
  padding: 2rem;
  text-align: center;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
}

.test-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

.test-card {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  transition: border-color 0.15s;
}
.test-card:hover { border-color: var(--border); }

.test-header { display: flex; align-items: center; justify-content: space-between; }
.test-category { font-size: 0.72rem; font-weight: 600; color: var(--accent); text-transform: uppercase; letter-spacing: 0.05em; }
.test-difficulty { font-size: 0.75rem; font-weight: 600; text-transform: capitalize; }
.test-level {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--accent);
  background: color-mix(in srgb, var(--accent) 12%, var(--bg-surface));
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
}

.test-title { margin: 0; font-size: 1rem; font-weight: 700; color: var(--text-primary); line-height: 1.4; }
.test-desc { margin: 0; font-size: 0.825rem; color: var(--text-muted); line-height: 1.5; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; }

.test-footer { display: flex; align-items: center; justify-content: space-between; margin-top: auto; padding-top: 0.75rem; border-top: 1px solid var(--border); }
.test-meta { font-size: 0.78rem; color: var(--text-muted); }

.btn-start {
  padding: 0.45rem 1.1rem;
  border-radius: 8px;
  border: none;
  background: var(--accent);
  color: var(--text-on-accent);
  font-size: 0.825rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, opacity 0.15s;
}
.btn-start:hover { background: var(--accent-hover); }
.btn-start:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-view-result {
  background: transparent;
  border: 1px solid var(--accent);
  color: var(--accent);
  text-decoration: none;
}
.btn-view-result:hover { background: var(--accent); color: #fff; }

.completed-badge {
  font-size: 0.72rem;
  font-weight: 700;
  padding: 0.35rem 0.65rem;
  border-radius: 20px;
  background: color-mix(in srgb, var(--success) 15%, transparent);
  color: var(--success);
}

@media (min-width: 769px) and (max-width: 1024px) { .test-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) { .test-grid { grid-template-columns: 1fr; } .btn-start { min-height: 44px; } }
</style>
