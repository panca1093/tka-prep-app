<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type User = components['schemas']['UserResponse']

const users = ref<User[]>([])
const total = ref(0)
const page = ref(1)
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
  active: '#22c55e', pending: '#f59e0b', suspended: '#ef4444', inactive: '#94a3b8',
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
              <button v-if="u.status !== 'suspended'" class="btn-action suspend" @click="updateStatus(u.id, 'suspended')">Suspend</button>
              <button v-if="u.status === 'suspended'" class="btn-action activate" @click="updateStatus(u.id, 'active')">Activate</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.page-title { margin: 0 0 1.5rem; font-size: 1.5rem; font-weight: 800; }
.filters { display: flex; gap: 0.75rem; margin-bottom: 1rem; flex-wrap: wrap; }
.search-input { flex: 1; min-width: 200px; padding: 0.55rem 0.875rem; border-radius: 8px; border: 1px solid #1e2a45; background: #141c2e; color: #f1f5f9; font-size: 0.875rem; outline: none; }
.search-input:focus { border-color: #4f8ef7; }
.filter-select { padding: 0.55rem 0.875rem; border-radius: 8px; border: 1px solid #1e2a45; background: #141c2e; color: #f1f5f9; font-size: 0.875rem; cursor: pointer; outline: none; }
.error-msg { padding: 0.6rem 0.75rem; border-radius: 8px; background: #450a0a; color: #fca5a5; font-size: 0.825rem; margin-bottom: 0.75rem; }
.total-line { font-size: 0.8rem; color: #64748b; margin-bottom: 0.75rem; }
.loading { color: #94a3b8; }

.table-wrap { background: #141c2e; border: 1px solid #1e2a45; border-radius: 12px; overflow: auto; }
.user-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
.user-table th { padding: 0.75rem 1rem; text-align: left; font-size: 0.75rem; font-weight: 700; color: #64748b; text-transform: uppercase; letter-spacing: 0.04em; border-bottom: 1px solid #1e2a45; }
.user-table td { padding: 0.75rem 1rem; border-bottom: 1px solid #0d1424; vertical-align: middle; }
.user-table tbody tr:last-child td { border-bottom: none; }
.user-table tbody tr:hover td { background: #1a2535; }
.email-cell { color: #94a3b8; font-size: 0.825rem; }
.role-cell { text-transform: capitalize; color: #94a3b8; }
.status-badge { font-size: 0.78rem; font-weight: 700; text-transform: capitalize; }
.action-cell { white-space: nowrap; }
.btn-action { padding: 0.3rem 0.75rem; border-radius: 6px; font-size: 0.78rem; font-weight: 600; cursor: pointer; border: none; transition: opacity 0.15s; }
.btn-action.suspend { background: transparent; border: 1px solid #ef4444; color: #ef4444; }
.btn-action.suspend:hover { background: rgba(239,68,68,0.1); }
.btn-action.activate { background: transparent; border: 1px solid #22c55e; color: #22c55e; }
.btn-action.activate:hover { background: rgba(34,197,94,0.1); }
</style>
