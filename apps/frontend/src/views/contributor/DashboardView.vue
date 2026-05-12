<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'

const testCount = ref(0)
const questionCount = ref(0)
const isLoading = ref(true)

onMounted(async () => {
  const [resTests, resQuestions] = await Promise.allSettled([
    client.GET('/tests', { params: { query: { limit: 1 } } }),
    client.GET('/questions', { params: { query: { limit: 1 } } }),
  ])
  if (resTests.status === 'fulfilled' && resTests.value.data) testCount.value = resTests.value.data.total
  if (resQuestions.status === 'fulfilled' && resQuestions.value.data) questionCount.value = resQuestions.value.data.total
  isLoading.value = false
})
</script>

<template>
  <div>
    <h1 class="page-title">Contributor Dashboard</h1>

    <div v-if="isLoading" class="loading">Loading…</div>

    <template v-else>
      <div class="stats-row">
        <div class="stat-card">
          <div class="stat-value">{{ questionCount }}</div>
          <div class="stat-label">Questions</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ testCount }}</div>
          <div class="stat-label">Tests</div>
        </div>
      </div>

      <div class="quick-actions">
        <RouterLink to="/contrib/questions" class="action-card">
          <div class="action-icon">❓</div>
          <div class="action-label">Question Bank</div>
          <div class="action-sub">Create and manage questions</div>
        </RouterLink>
        <RouterLink to="/contrib/tests" class="action-card">
          <div class="action-icon">📄</div>
          <div class="action-label">Tests</div>
          <div class="action-sub">Build and publish tests</div>
        </RouterLink>
      </div>
    </template>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 2rem; font-size: 1.5rem; font-weight: 800; }
.loading { color: #94a3b8; }

.stats-row { display: flex; gap: 1rem; margin-bottom: 2rem; }
.stat-card {
  flex: 1; max-width: 160px;
  background: #141c2e; border: 1px solid #1e2a45; border-radius: 12px;
  padding: 1.25rem 1.5rem;
}
.stat-value { font-size: 2rem; font-weight: 800; color: #4f8ef7; }
.stat-label { font-size: 0.8rem; color: #94a3b8; margin-top: 0.25rem; }

.quick-actions { display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem; max-width: 480px; }
.action-card {
  background: #141c2e; border: 1px solid #1e2a45; border-radius: 12px;
  padding: 1.5rem; text-decoration: none; display: flex; flex-direction: column; gap: 0.375rem;
  transition: border-color 0.15s;
}
.action-card:hover { border-color: #4f8ef7; }
.action-icon { font-size: 1.5rem; margin-bottom: 0.25rem; }
.action-label { font-size: 0.95rem; font-weight: 700; color: #f1f5f9; }
.action-sub { font-size: 0.8rem; color: #94a3b8; }
</style>
