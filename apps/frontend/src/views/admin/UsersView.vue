<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type User = components['schemas']['UserResponse']

const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const isLoading = ref(false)
const search = ref('')
const roleFilter = ref('')
const statusFilter = ref('')
const actionError = ref('')

async function fetchUsers() {
  isLoading.value = true
  const { data } = await client.GET('/admin/users', {
    params: {
      query: {
        search: search.value || undefined,
        role: (roleFilter.value as 'student' | 'contributor' | 'admin') || undefined,
        status: (statusFilter.value as 'active' | 'pending' | 'suspended') || undefined,
        page: page.value,
        limit: 20,
      },
    },
  })
  if (data) { users.value = data.data; total.value = data.total }
  isLoading.value = false
}

async function updateStatus(userId: string, status: 'active' | 'suspended') {
  actionError.value = ''
  const { error } = await client.PATCH('/admin/users/{userId}/status', {
    params: { path: { userId } },
    body: { status },
  })
  if (error) { actionError.value = 'Failed to update status.'; return }
  await fetchUsers()
}

onMounted(fetchUsers)

const statusColor: Record<string, string> = {
  active: 'var(--success)', pending: 'var(--warning)', suspended: 'var(--danger)', inactive: 'var(--text-muted)',
}
</script>

<template>
  <div>
    <h1 class="page-title">Users</h1>

    <div class="filters">
      <input v-model="search" placeholder="Search name or email…" class="search-input" @input="fetchUsers" />
      <select v-model="roleFilter" class="filter-select" @change="fetchUsers">
        <option value="">All Roles</option>
        <option value="student">Student</option>
        <option value="contributor">Contributor</option>
        <option value="admin">Admin</option>
      </select>
      <select v-model="statusFilter" class="filter-select" @change="fetchUsers">
        <option value="">All Statuses</option>
        <option value="active">Active</option>
        <option value="pending">Pending</option>
        <option value="suspended">Suspended</option>
      </select>
    </div>

    <p v-if="actionError" class="error-msg">{{ actionError }}</p>

    <div class="total-line">{{ total }} user(s) found</div>

    <div v-if="isLoading" class="loading">Loading…</div>

    <div v-else class="table-wrap">
      <table class="user-table">
        <thead>
          <tr>
            <th>Name</th><th>Email</th><th>Role</th><th>Status</th><th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.name }}</td>
            <td class="email-cell">{{ u.email }}</td>
            <td class="role-cell">{{ u.role }}</td>
            <td>
              <span class="status-badge" :style="{ color: statusColor[u.status] }">{{ u.status }}</span>
            </td>
            <td class="action-cell">
              <button v-if="u.status !== 'suspended' && u.role !== 'admin'" class="btn-action suspend" @click="updateStatus(u.id, 'suspended')">Suspend</button>
              <button v-if="u.status === 'suspended' && u.role !== 'admin'" class="btn-action activate" @click="updateStatus(u.id, 'active')">Activate</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="total > pageSize" class="pagination">
      <button class="page-btn" :disabled="page === 1" @click="page--; fetchUsers()">← Prev</button>
      <span class="page-info">Page {{ page }} of {{ Math.ceil(total / pageSize) }}</span>
      <button class="page-btn" :disabled="page >= Math.ceil(total / pageSize)" @click="page++; fetchUsers()">Next →</button>
    </div>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 1.5rem; font-size: 1.5rem; font-weight: 800; }
.filters { display: flex; gap: 0.75rem; margin-bottom: 1rem; flex-wrap: wrap; }
.search-input { flex: 1; min-width: 200px; padding: 0.55rem 0.875rem; border-radius: 8px; border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-primary); font-size: 0.875rem; outline: none; }
.search-input:focus { border-color: var(--accent); }
.filter-select { padding: 0.55rem 0.875rem; border-radius: 8px; border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-primary); font-size: 0.875rem; cursor: pointer; outline: none; }
.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: var(--danger-bg); color: var(--danger-text); font-size: 0.825rem; margin-bottom: 0.75rem; }
.total-line { font-size: 0.8rem; color: var(--text-muted); margin-bottom: 0.75rem; }
.loading { color: var(--text-muted); }

.table-wrap { background: var(--bg-surface); border: 1px solid var(--border); border-radius: 12px; overflow: auto; }
.user-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
.user-table th { padding: 0.75rem 1rem; text-align: left; font-size: 0.75rem; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.04em; border-bottom: 1px solid var(--border); }
.user-table td { padding: 0.75rem 1rem; border-bottom: 1px solid var(--bg-input); vertical-align: middle; }
.user-table tbody tr:last-child td { border-bottom: none; }
.user-table tbody tr:hover td { background: var(--border); }
.email-cell { color: var(--text-muted); font-size: 0.825rem; }
.role-cell { text-transform: capitalize; color: var(--text-muted); }
.status-badge { font-size: 0.78rem; font-weight: 700; text-transform: capitalize; }
.action-cell { white-space: nowrap; }
.btn-action { padding: 0.3rem 0.75rem; border-radius: 6px; font-size: 0.78rem; font-weight: 600; cursor: pointer; border: none; transition: opacity 0.15s; }
.btn-action.suspend { background: transparent; border: 1px solid var(--danger); color: var(--danger); }
.btn-action.suspend:hover { background: rgba(239,68,68,0.1); }
.btn-action.activate { background: transparent; border: 1px solid var(--success); color: var(--success); }
.btn-action.activate:hover { background: rgba(34,197,94,0.1); }

.pagination { display: flex; align-items: center; gap: 1rem; margin-top: 1rem; justify-content: flex-end; }
.page-btn { padding: 0.4rem 0.875rem; border-radius: 6px; border: 1px solid var(--border); background: var(--bg-surface); color: var(--text-muted); font-size: 0.8rem; cursor: pointer; transition: all 0.15s; }
.page-btn:hover:not(:disabled) { border-color: var(--accent); color: var(--accent); }
.page-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.page-info { font-size: 0.8rem; color: var(--text-muted); }
</style>
@media (max-width: 768px) { .table-wrap { overflow-x: auto; } }
