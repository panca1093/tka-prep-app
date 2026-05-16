<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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
    startError.value = 'Tidak dapat memulai sesi. Coba lagi.'
  } finally {
    startingId.value = null
  }
}

const categoryLabel: Record<string, string> = {
  tka_saintek: 'TKA Saintek',
  tka_soshum: 'TKA Soshum',
  smbt: 'SMBT',
}

const cats = [
  { key: 'all', label: 'Semua', icon: '📋' },
  { key: 'tka_saintek', label: 'TKA Saintek', icon: '🔬' },
  { key: 'tka_soshum', label: 'TKA Soshum', icon: '📚' },
  { key: 'smbt', label: 'SMBT', icon: '🎓' },
]

const difficultyColor: Record<string, string> = {
  easy: 'var(--success)',
  medium: 'var(--warning)',
  hard: 'var(--danger)',
}

const difficultyLabel: Record<string, string> = {
  easy: 'Mudah',
  medium: 'Sedang',
  hard: 'Sulit',
}

const levelLabel: Record<string, string> = {
  sd: 'SD', smp: 'SMP', sma: 'SMA', smk: 'SMK',
}

const filtered = computed(() =>
  categoryFilter.value === 'all'
    ? testStore.tests
    : testStore.tests.filter((t) => t.category === categoryFilter.value))

const inProgress = computed(() => filtered.value.filter(t => t.student_status === 'in_progress'))
const notStarted = computed(() => filtered.value.filter(t => t.student_status !== 'in_progress' && t.student_status !== 'completed'))
const completed = computed(() => filtered.value.filter(t => t.student_status === 'completed'))
</script>

<template>
  <div class="tests-page">
    <div class="page-header">
      <h1 class="page-title">Ujian Tersedia</h1>
      <p class="page-sub">Pilih paket ujian dan mulai berlatih</p>
    </div>

    <!-- Category tabs -->
    <div class="cat-tabs">
      <button
        v-for="c in cats"
        :key="c.key"
        class="cat-tab"
        :class="{ 'cat-tab--active': categoryFilter === c.key }"
        @click="categoryFilter = c.key as typeof categoryFilter"
      >
        <span class="cat-icon">{{ c.icon }}</span>
        <span>{{ c.label }}</span>
      </button>
    </div>

    <div v-if="startError" class="error-banner">{{ startError }}</div>
    <div v-if="testStore.isLoading" class="loading-state">
      <div class="loading-dots"><span/><span/><span/></div>
    </div>

    <template v-else-if="filtered.length === 0">
      <div class="empty-state">
        <div class="empty-icon">🔍</div>
        <p>Tidak ada ujian dalam kategori ini.</p>
      </div>
    </template>

    <template v-else>
      <!-- In-progress strip -->
      <div v-if="inProgress.length" class="section">
        <div class="section-label">
          <span class="section-dot section-dot--progress" />
          Sedang Dikerjakan
        </div>
        <div class="test-grid">
          <div v-for="t in inProgress" :key="t.id" class="test-card test-card--progress">
            <div class="card-strip card-strip--progress" />
            <div class="card-body">
              <div class="card-meta">
                <span class="cat-badge">{{ categoryLabel[t.category] }}</span>
                <span v-if="t.education_level" class="level-badge">{{ levelLabel[t.education_level] ?? t.education_level }}</span>
                <span class="diff-badge" :style="{ color: difficultyColor[t.difficulty] }">{{ difficultyLabel[t.difficulty] }}</span>
              </div>
              <h3 class="card-title">{{ t.title }}</h3>
              <p v-if="t.description" class="card-desc">{{ t.description }}</p>
              <div class="card-footer">
                <div class="card-info">
                  <span class="info-item">
                    <svg viewBox="0 0 16 16" fill="currentColor" width="12" height="12"><path fill-rule="evenodd" d="M8 15A7 7 0 1 0 8 1a7 7 0 0 0 0 14zm.5-9.5a.5.5 0 0 0-1 0V8a.5.5 0 0 0 .146.354l3 3a.5.5 0 1 0 .708-.708L8.5 7.793V5.5z"/></svg>
                    {{ t.duration_minutes }} menit
                  </span>
                  <span class="info-item">{{ t.questions.length }} soal</span>
                </div>
                <button class="btn-resume" :disabled="startingId === t.id" @click="handleStart(t.id)">
                  {{ startingId === t.id ? 'Melanjutkan…' : 'Lanjutkan' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Not started -->
      <div v-if="notStarted.length" class="section">
        <div v-if="inProgress.length || completed.length" class="section-label">
          <span class="section-dot" />
          Belum Dikerjakan
        </div>
        <div class="test-grid">
          <div v-for="t in notStarted" :key="t.id" class="test-card">
            <div class="card-strip" />
            <div class="card-body">
              <div class="card-meta">
                <span class="cat-badge">{{ categoryLabel[t.category] }}</span>
                <span v-if="t.education_level" class="level-badge">{{ levelLabel[t.education_level] ?? t.education_level }}</span>
                <span class="diff-badge" :style="{ color: difficultyColor[t.difficulty] }">{{ difficultyLabel[t.difficulty] }}</span>
              </div>
              <h3 class="card-title">{{ t.title }}</h3>
              <p v-if="t.description" class="card-desc">{{ t.description }}</p>
              <div class="card-footer">
                <div class="card-info">
                  <span class="info-item">
                    <svg viewBox="0 0 16 16" fill="currentColor" width="12" height="12"><path fill-rule="evenodd" d="M8 15A7 7 0 1 0 8 1a7 7 0 0 0 0 14zm.5-9.5a.5.5 0 0 0-1 0V8a.5.5 0 0 0 .146.354l3 3a.5.5 0 1 0 .708-.708L8.5 7.793V5.5z"/></svg>
                    {{ t.duration_minutes }} menit
                  </span>
                  <span class="info-item">{{ t.questions.length }} soal</span>
                </div>
                <button class="btn-start" :disabled="startingId === t.id" @click="handleStart(t.id)">
                  {{ startingId === t.id ? 'Memulai…' : 'Mulai' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Completed -->
      <div v-if="completed.length" class="section">
        <div class="section-label">
          <span class="section-dot section-dot--done" />
          Sudah Dikerjakan
        </div>
        <div class="test-grid">
          <div v-for="t in completed" :key="t.id" class="test-card test-card--done">
            <div class="card-strip card-strip--done" />
            <div class="card-body">
              <div class="card-meta">
                <span class="cat-badge">{{ categoryLabel[t.category] }}</span>
                <span v-if="t.education_level" class="level-badge">{{ levelLabel[t.education_level] ?? t.education_level }}</span>
                <span class="done-badge">✓ Selesai</span>
              </div>
              <h3 class="card-title">{{ t.title }}</h3>
              <p v-if="t.description" class="card-desc">{{ t.description }}</p>
              <div class="card-footer">
                <div class="card-info">
                  <span class="info-item">{{ t.duration_minutes }} menit</span>
                  <span class="info-item">{{ t.questions.length }} soal</span>
                </div>
                <RouterLink v-if="t.result_id" :to="'/results/' + t.result_id" class="btn-review">
                  Lihat Hasil
                </RouterLink>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.tests-page { display: flex; flex-direction: column; gap: 1.5rem; }

.page-header { }
.page-title {
  margin: 0 0 0.25rem;
  
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-heading);
}
.page-sub { margin: 0; font-size: 0.85rem; color: var(--text-muted); }

/* ─── Category tabs ────────────────────────────────────────────────────────── */
.cat-tabs { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.cat-tab {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.825rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s;
}
.cat-tab:hover { border-color: var(--accent); color: var(--text-primary); }
.cat-tab--active {
  background: var(--accent-dim);
  border-color: var(--accent);
  color: var(--accent);
  font-weight: 700;
}
.cat-icon { font-size: 1rem; }

/* ─── Error/loading/empty ─────────────────────────────────────────────────── */
.error-banner {
  padding: 0.75rem 1rem;
  border-radius: 10px;
  background: var(--danger-bg);
  color: var(--danger-text);
  font-size: 0.825rem;
}
.loading-state { display: flex; justify-content: center; padding: 3rem; }
.loading-dots { display: flex; gap: 6px; align-items: center; }
.loading-dots span {
  width: 8px; height: 8px; border-radius: 50%;
  background: var(--accent);
  animation: dot-bounce 1.4s infinite ease-in-out;
}
.loading-dots span:nth-child(2) { animation-delay: 0.2s; }
.loading-dots span:nth-child(3) { animation-delay: 0.4s; }
@keyframes dot-bounce { 0%,80%,100%{transform:scale(0.6);opacity:0.5}40%{transform:scale(1);opacity:1} }
.empty-state {
  display: flex; flex-direction: column; align-items: center;
  gap: 0.75rem; padding: 3rem; text-align: center;
  background: var(--bg-surface); border: 1px solid var(--border); border-radius: 14px;
}
.empty-icon { font-size: 2.5rem; }

/* ─── Section ─────────────────────────────────────────────────────────────── */
.section { display: flex; flex-direction: column; gap: 0.875rem; }
.section-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.72rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-muted);
}
.section-dot {
  width: 8px; height: 8px; border-radius: 50%; background: var(--border);
}
.section-dot--progress { background: var(--warm); }
.section-dot--done { background: var(--success); }

/* ─── Test grid ───────────────────────────────────────────────────────────── */
.test-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.875rem;
}

/* ─── Test card ───────────────────────────────────────────────────────────── */
.test-card {
  display: flex;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  overflow: hidden;
  transition: border-color 0.15s, transform 0.15s;
}
.test-card:hover { border-color: color-mix(in srgb, var(--accent) 40%, var(--border)); transform: translateY(-2px); }
.test-card--progress { border-color: color-mix(in srgb, var(--warm) 30%, var(--border)); }
.test-card--progress:hover { border-color: var(--warm); }
.test-card--done { opacity: 0.82; }
.test-card--done:hover { opacity: 1; }

.card-strip { width: 4px; flex-shrink: 0; background: var(--border); }
.card-strip--progress { background: var(--warm); }
.card-strip--done { background: var(--success); }

.card-body { flex: 1; padding: 1rem 1.125rem; display: flex; flex-direction: column; gap: 0.5rem; min-width: 0; }

.card-meta { display: flex; align-items: center; gap: 0.375rem; flex-wrap: wrap; }
.cat-badge { font-size: 0.65rem; font-weight: 700; color: var(--accent); text-transform: uppercase; letter-spacing: 0.04em; }
.level-badge {
  font-size: 0.6rem; font-weight: 700; padding: 1px 6px;
  border-radius: 3px;
  background: var(--accent-dim); color: var(--accent);
}
.diff-badge { font-size: 0.7rem; font-weight: 600; margin-left: auto; }
.done-badge { font-size: 0.65rem; font-weight: 700; color: var(--success); margin-left: auto; }

.card-title {
  margin: 0;
  font-size: 0.925rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.card-desc {
  margin: 0;
  font-size: 0.78rem;
  color: var(--text-muted);
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 0.625rem;
  border-top: 1px solid var(--border);
}
.card-info { display: flex; gap: 0.75rem; }
.info-item {
  display: flex; align-items: center; gap: 3px;
  font-size: 0.72rem; color: var(--text-muted);
}

/* Buttons */
.btn-start, .btn-resume, .btn-review {
  padding: 0.4rem 0.875rem;
  border-radius: 8px;
  font-size: 0.78rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s;
  text-decoration: none;
  border: none;
  white-space: nowrap;
}
.btn-start { background: var(--accent); color: #fff; }
.btn-start:hover:not(:disabled) { background: var(--accent-hover); }
.btn-start:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-resume { background: var(--warm-dim); color: var(--warm); border: 1px solid color-mix(in srgb, var(--warm) 30%, transparent); }
.btn-resume:hover:not(:disabled) { background: var(--warm); color: #fff; }
.btn-resume:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-review { background: transparent; border: 1px solid var(--border); color: var(--text-muted); }
.btn-review:hover { border-color: var(--accent); color: var(--accent); }

/* ─── Responsive ──────────────────────────────────────────────────────────── */
@media (min-width: 769px) and (max-width: 1024px) {
  .test-grid { grid-template-columns: repeat(2, 1fr); }
}
@media (max-width: 768px) {
  .page-title { font-size: 1.5rem; }
  .test-grid { grid-template-columns: 1fr; }
  .cat-tabs { gap: 0.375rem; }
  .cat-tab { font-size: 0.78rem; padding: 0.4rem 0.75rem; }
}
</style>
