<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Stats = components['schemas']['PlatformStatsResponse']
type User = components['schemas']['UserResponse']

const stats = ref<Stats | null>(null)
const pendingUsers = ref<User[]>([])
const isLoading = ref(true)
const approveError = ref('')

async function fetchData() {
  const [resStats, resPending] = await Promise.allSettled([
    client.GET('/admin/stats'),
    client.GET('/admin/contributors/pending', { params: { query: { limit: 10 } } }),
  ])
  if (resStats.status === 'fulfilled' && resStats.value.data) stats.value = resStats.value.data
  if (resPending.status === 'fulfilled' && resPending.value.data) pendingUsers.value = resPending.value.data.data
  isLoading.value = false
}

async function approve(userId: string) {
  approveError.value = ''
  const { error } = await client.POST('/admin/contributors/{userId}/approve', { params: { path: { userId } } })
  if (error) { approveError.value = 'Failed to approve.'; return }
  await fetchData()
}

async function reject(userId: string) {
  approveError.value = ''
  const { error } = await client.POST('/admin/contributors/{userId}/reject', { params: { path: { userId } } })
  if (error) { approveError.value = 'Failed to reject.'; return }
  await fetchData()
}

onMounted(fetchData)
</script>

<template>
  <div>
    <h1 class="page-title">Admin Dashboard</h1>

    <div v-if="isLoading" class="loading">Loading…</div>

    <template v-else>
      <div class="stats-grid" v-if="stats">
        <div class="stat-card">
          <div class="stat-value">{{ stats.total_students }}</div>
          <div class="stat-label">Students</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ stats.total_contributors }}</div>
          <div class="stat-label">Contributors</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ stats.total_tests }}</div>
          <div class="stat-label">Tests</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ stats.total_questions }}</div>
          <div class="stat-label">Questions</div>
        </div>
        <div class="stat-card highlight">
          <div class="stat-value">{{ stats.pending_approvals }}</div>
          <div class="stat-label">Pending Approvals</div>
        </div>
      </div>

      <div class="section">
        <h2>Pending Contributor Requests</h2>
        <p v-if="approveError" class="error-msg">{{ approveError }}</p>
        <div v-if="pendingUsers.length === 0" class="empty-state">No pending requests.</div>
        <div v-else class="pending-list">
          <div v-for="u in pendingUsers" :key="u.id" class="pending-row">
            <div class="pending-info">
              <div class="pending-name">{{ u.name }}</div>
              <div class="pending-email">{{ u.email }}</div>
            </div>
            <div class="pending-actions">
              <button class="btn-approve" @click="approve(u.id)">Approve</button>
              <button class="btn-reject" @click="reject(u.id)">Reject</button>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 2rem; font-size: 1.5rem; font-weight: 800; }
.loading { color: var(--text-muted); }

.stats-grid { display: grid; grid-template-columns: repeat(5, 1fr); gap: 1rem; margin-bottom: 2.5rem; }
.stat-card { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; padding: 1.25rem 1.5rem; }
.stat-card.highlight { border-color: var(--warning); }
.stat-value { font-size: 2rem; font-weight: 800; color: var(--accent); }
.stat-card.highlight .stat-value { color: var(--warning); }
.stat-label { font-size: 0.8rem; color: var(--text-muted); margin-top: 0.25rem; }

.section { margin-bottom: 2rem; }
.section h2 { font-size: 1rem; font-weight: 700; margin: 0 0 1rem; }
.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: var(--danger-bg); color: var(--danger-text); font-size: 0.825rem; margin-bottom: 0.75rem; }
.empty-state { color: var(--text-muted); font-size: 0.875rem; padding: 1rem 0; }

.pending-list { display: flex; flex-direction: column; gap: 0.5rem; }
.pending-row {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0.875rem 1rem; background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: 10px;
}
.pending-name { font-weight: 600; font-size: 0.9rem; }
.pending-email { font-size: 0.8rem; color: var(--text-muted); margin-top: 2px; }
.pending-actions { display: flex; gap: 0.5rem; }

.btn-approve, .btn-reject {
  padding: 0.4rem 0.875rem; border-radius: 6px; font-size: 0.8rem; font-weight: 600;
  cursor: pointer; border: none; transition: opacity 0.15s;
}
.btn-approve { background: var(--success); color: #000; }
.btn-approve:hover { opacity: 0.85; }
.btn-reject { background: var(--border); color: var(--danger); border: 1px solid var(--danger); background: transparent; }
.btn-reject:hover { background: rgba(239,68,68,0.1); }
</style>
