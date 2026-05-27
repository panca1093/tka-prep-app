<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import MostUsedQuestions from '@/components/MostUsedQuestions.vue'
import client from '@/api/client'

const auth = useAuthStore()

const questionCount = ref(0)
const testCount = ref(0)
const totalAttempts = ref(0)
const displayQuestionCount = ref(0)
const displayTestCount = ref(0)
const displayAttempts = ref(0)

const firstName = computed(() => {
  const name = auth.user?.name ?? 'Kontributor'
  return name.split(' ')[0]
})

const dateString = computed(() => {
  const now = new Date()
  const days = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu']
  const months = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
  return days[now.getDay()] + ', ' + now.getDate() + ' ' + months[now.getMonth()] + ' ' + now.getFullYear()
})

function animateCount(target: number, displayRef: { value: number }, duration = 900) {
  const start = performance.now()
  const from = displayRef.value
  function tick(now: number) {
    const elapsed = now - start
    const progress = Math.min(elapsed / duration, 1)
    const eased = 1 - Math.pow(1 - progress, 3)
    displayRef.value = Math.round(from + (target - from) * eased)
    if (progress < 1) requestAnimationFrame(tick)
  }
  requestAnimationFrame(tick)
}

onMounted(async () => {
  try {
    const [resTests, resQuestions] = await Promise.allSettled([
      client.GET('/tests', { params: { query: { status: 'published', limit: 50 } } }),
      client.GET('/questions', { params: { query: { limit: 1 } } }),
    ])
    if (resTests.status === 'fulfilled' && resTests.value.data) {
      testCount.value = resTests.value.data.total
      let sum = 0
      for (const t of resTests.value.data.data) {
        if ((t as any).attempt_count) sum += (t as any).attempt_count
      }
      totalAttempts.value = sum
    }
    if (resQuestions.status === 'fulfilled' && resQuestions.value.data) {
      questionCount.value = resQuestions.value.data.total
    }
    animateCount(questionCount.value, displayQuestionCount)
    animateCount(testCount.value, displayTestCount)
    animateCount(totalAttempts.value, displayAttempts)
  } catch (_) {
    // keep defaults at 0
  }
})
</script>

<template>
  <div class="contributor-dashboard">
    <div class="dash-welcome">
      <div>
        <h1 class="dash-greeting">
          Selamat kembali, <span class="name">{{ firstName }}</span> 👋
        </h1>
        <p class="dash-sub">Kontributor · Masukelas Platform</p>
      </div>
      <div class="dash-date">{{ dateString }}</div>
    </div>

    <div class="stat-grid">
      <!-- Soal dalam Bank -->
      <div class="stat-card">
        <div class="stat-icon-row">
          <div class="stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="3" width="20" height="14" rx="2"/>
              <path d="M8 21h8M12 17v4"/>
            </svg>
          </div>
          <span class="stat-trend">↑ +12 bulan ini</span>
        </div>
        <div class="stat-value">{{ displayQuestionCount }}</div>
        <div class="stat-label">Soal dalam Bank</div>
      </div>

      <!-- Tes Diterbitkan -->
      <div class="stat-card">
        <div class="stat-icon-row">
          <div class="stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
              <polyline points="10 9 9 9 8 9"/>
            </svg>
          </div>
          <span class="stat-trend">↑ +2 bulan ini</span>
        </div>
        <div class="stat-value">{{ displayTestCount }}</div>
        <div class="stat-label">Tes Diterbitkan</div>
      </div>

      <!-- Percobaan Siswa -->
      <div class="stat-card">
        <div class="stat-icon-row">
          <div class="stat-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2z"/>
            </svg>
          </div>
        </div>
        <div class="stat-value">{{ displayAttempts }}</div>
        <!-- stat-value styles remain via stat-value class -->
        <div class="stat-label">Percobaan Siswa</div>
      </div>
    </div>

    <div class="section-label">Aksi Cepat</div>

    <div class="action-grid">
      <RouterLink to="/contrib/questions" class="action-card">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="16"/>
            <line x1="8" y1="12" x2="16" y2="12"/>
          </svg>
        </div>
        <div class="action-title">Buat Soal Baru</div>
        <div class="action-sub">Tambah soal ke bank — PG, PGK, atau Benar/Salah</div>
        <span class="action-arrow">→</span>
      </RouterLink>

      <RouterLink to="/contrib/tests" class="action-card">
        <div class="action-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="12" y1="18" x2="12" y2="12"/>
            <line x1="9" y1="15" x2="15" y2="15"/>
          </svg>
        </div>
        <div class="action-title">Buat Tes Baru</div>
        <div class="action-sub">Susun paket ujian dan pilih soal dari bank</div>
        <span class="action-arrow">→</span>
      </RouterLink>
    </div>

    <div class="section-label">Pertanyaan Teratas</div>
    <MostUsedQuestions />
  </div>
</template>

<style scoped>
.contributor-dashboard {
  padding: 0;
}

.dash-welcome {
  margin-bottom: 1.75rem;
  padding: 1.375rem 1.5rem;
  background: linear-gradient(135deg, rgba(79, 142, 247, 0.08), rgba(79, 142, 247, 0.02));
  border: 1px solid rgba(79, 142, 247, 0.18);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.dash-greeting {
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--text-heading);
  margin: 0;
}

.dash-greeting .name {
  color: var(--accent);
}

.dash-sub {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin: 0.2rem 0 0;
}

.dash-date {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 1.25rem 1.375rem;
  transition: all 0.18s;
}

.stat-card:hover {
  border-color: color-mix(in srgb, var(--accent) 35%, var(--border));
  transform: translateY(-2px);
  box-shadow: 0 6px 24px rgba(79, 142, 247, 0.08);
}

.stat-icon-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.875rem;
}

.stat-icon {
  width: 34px;
  height: 34px;
  border-radius: 8px;
  background: var(--accent-dim);
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-icon svg {
  width: 16px;
  height: 16px;
  color: var(--accent);
}

.stat-trend {
  font-size: 0.68rem;
  font-weight: 700;
  color: var(--success);
  background: var(--success-bg);
  padding: 0.15rem 0.45rem;
  border-radius: 4px;
}

.stat-value {
  font-size: 2rem;
  font-weight: 800;
  color: var(--text-heading);
  line-height: 1;
  margin-bottom: 0.25rem;
}

.stat-value--muted {
  color: var(--text-muted);
}

.stat-label {
  font-size: 0.78rem;
  color: var(--text-muted);
}

.section-label {
  font-size: 0.68rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.07em;
  color: var(--text-muted);
  margin-bottom: 0.875rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.section-label::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--border);
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  max-width: 520px;
}

.action-card {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 1.375rem;
  text-decoration: none;
  display: block;
  transition: all 0.18s;
  position: relative;
  overflow: hidden;
}

.action-card:hover {
  border-color: color-mix(in srgb, var(--accent) 40%, var(--border));
  transform: translateY(-2px);
  box-shadow: 0 6px 24px rgba(79, 142, 247, 0.1);
}

.action-icon {
  width: 36px;
  height: 36px;
  border-radius: 9px;
  background: var(--accent-dim);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.75rem;
}

.action-icon svg {
  width: 17px;
  height: 17px;
  color: var(--accent);
}

.action-title {
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 0.2rem;
}

.action-sub {
  font-size: 0.75rem;
  color: var(--text-muted);
  line-height: 1.5;
}

.action-arrow {
  position: absolute;
  right: 1.125rem;
  bottom: 1.125rem;
  color: var(--accent);
  opacity: 0;
  transform: translateX(-4px);
  transition: all 0.18s;
  font-size: 0.85rem;
}

.action-card:hover .action-arrow {
  opacity: 1;
  transform: translateX(0);
}

@media (max-width: 768px) {
  .stat-grid {
    grid-template-columns: 1fr;
  }

  .action-grid {
    grid-template-columns: 1fr;
    max-width: 100%;
  }

  .dash-welcome {
    flex-direction: column;
    gap: 0.5rem;
  }
}
</style>
