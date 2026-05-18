<script setup lang="ts">
import { ref, onMounted, computed, reactive } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type User = components['schemas']['UserResponse']

const users        = ref<User[]>([])
const total        = ref(0)
const page         = ref(1)
const pageSize     = 20
const isLoading    = ref(false)
const search       = ref('')
const roleFilter   = ref('')
const statusFilter = ref('')
const actionError  = ref('')
const confirmSuspendId = ref<string | null>(null)
const busyId       = ref<string | null>(null)
const showAddModal = ref(false)
const addForm      = reactive({ name: '', email: '', password: '' })
const addError     = ref('')
const isAdding     = ref(false)

const activeFilterCount = computed(() => [roleFilter.value, statusFilter.value].filter(Boolean).length)

function clearFilters() { roleFilter.value = ''; statusFilter.value = ''; page.value = 1; fetchUsers() }

const roles = [
  { value: '', label: 'Semua Role' },
  { value: 'student', label: 'Siswa' },
  { value: 'contributor', label: 'Kontributor' },
  { value: 'admin', label: 'Admin' },
]
const statuses = [
  { value: '', label: 'Semua Status' },
  { value: 'active', label: 'Aktif' },
  { value: 'pending', label: 'Menunggu' },
  { value: 'suspended', label: 'Ditangguhkan' },
]

async function fetchUsers() {
  isLoading.value = true
  const { data } = await client.GET('/admin/users', {
    params: {
      query: {
        search: search.value || undefined,
        role: (roleFilter.value as 'student' | 'contributor' | 'admin') || undefined,
        status: (statusFilter.value as 'active' | 'pending' | 'suspended') || undefined,
        page: page.value,
        limit: pageSize,
      },
    },
  })
  if (data) { users.value = data.data; total.value = data.total }
  isLoading.value = false
}

async function updateStatus(userId: string, status: 'active' | 'suspended') {
  busyId.value = userId
  confirmSuspendId.value = null
  actionError.value = ''
  const { error } = await client.PATCH('/admin/users/{userId}/status', {
    params: { path: { userId } },
    body: { status },
  })
  if (error) { actionError.value = 'Gagal memperbarui status.'; busyId.value = null; return }
  busyId.value = null
  await fetchUsers()
}

function setRole(v: string) { roleFilter.value = v; page.value = 1; fetchUsers() }
function setStatus(v: string) { statusFilter.value = v; page.value = 1; fetchUsers() }

async function createContributor() {
  addError.value = ''
  if (!addForm.name.trim() || !addForm.email.trim() || addForm.password.length < 8) {
    addError.value = 'Semua kolom wajib diisi. Password minimal 8 karakter.'
    return
  }
  isAdding.value = true
  const { error, response } = await client.POST('/admin/contributors', { body: addForm })
  if (error) {
    addError.value = response.status === 409 ? 'Email sudah terdaftar.' : 'Gagal membuat kontributor.'
    isAdding.value = false
    return
  }
  isAdding.value = false
  showAddModal.value = false
  addForm.name = ''; addForm.email = ''; addForm.password = ''
  await fetchUsers()
}

function openAddModal() {
  addForm.name = ''; addForm.email = ''; addForm.password = ''; addError.value = ''
  showAddModal.value = true
}

const totalPages = computed(() => Math.ceil(total.value / pageSize))

function initials(name: string) {
  return name.split(' ').map(w => w[0]).join('').substring(0, 2).toUpperCase()
}

function avatarColor(name: string): string {
  const hues = [210, 145, 25, 270, 340, 185, 60]
  const i = name.charCodeAt(0) % hues.length
  return `hsl(${hues[i]}, 65%, 48%)`
}

onMounted(fetchUsers)
</script>

<template>
  <div class="uv" :class="{ ready: !isLoading }">

    <!-- ── Header ───────────────────────────────────────────────────────────── -->
    <div class="uv-head">
      <div>
        <h1 class="uv-title">Pengguna</h1>
        <p class="uv-subtitle">{{ total.toLocaleString('id-ID') }} pengguna ditemukan</p>
      </div>
      <button class="btn-add" @click="openAddModal">+ Tambah Kontributor</button>
    </div>

    <!-- ── Toolbar ──────────────────────────────────────────────────────────── -->
    <div class="toolbar">
      <div class="search-wrap">
        <svg class="search-icon" viewBox="0 0 20 20" fill="currentColor" width="16" height="16">
          <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd"/>
        </svg>
        <input
          v-model="search"
          class="search-input"
          placeholder="Cari nama atau email…"
          @input="page = 1; fetchUsers()"
        />
      </div>

      <div class="pill-group">
        <button
          v-for="r in roles" :key="r.value"
          class="pill" :class="{ 'pill--active': roleFilter === r.value }"
          @click="setRole(r.value)"
        >{{ r.label }}</button>
      </div>

      <div class="pill-group">
        <button
          v-for="s in statuses" :key="s.value"
          class="pill" :class="{ 'pill--active': statusFilter === s.value, [`pill--status-${s.value}`]: s.value && statusFilter === s.value }"
          @click="setStatus(s.value)"
        >{{ s.label }}</button>
      </div>

      <span v-if="activeFilterCount > 0" class="filter-badge">{{ activeFilterCount }} aktif</span>
      <button v-if="activeFilterCount > 0" class="btn-clear" @click="clearFilters">Hapus Semua Filter</button>
    </div>

    <!-- ── Error ────────────────────────────────────────────────────────────── -->
    <div v-if="actionError" class="action-error">{{ actionError }}</div>

    <!-- ── Loading ──────────────────────────────────────────────────────────── -->
    <div v-if="isLoading" class="table-wrap">
      <div v-for="i in 6" :key="i" class="row-skeleton" :style="`--i:${i}`">
        <div class="sk-avatar" />
        <div class="sk-text">
          <div class="sk-block sk-name" />
          <div class="sk-block sk-email" />
        </div>
        <div class="sk-block sk-badge" />
        <div class="sk-block sk-badge" />
      </div>
    </div>

    <!-- ── Table ────────────────────────────────────────────────────────────── -->
    <div v-else class="table-wrap">
      <div v-if="users.length === 0" class="empty-state">
        <svg viewBox="0 0 20 20" fill="currentColor" width="28" height="28">
          <path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd"/>
        </svg>
        <p>Tidak ada pengguna yang cocok.</p>
      </div>

      <table v-else class="user-table">
        <thead>
          <tr>
            <th>Pengguna</th>
            <th>Role</th>
            <th>Status</th>
            <th class="th-actions">Aksi</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(u, i) in users" :key="u.id" :style="`--i:${i}`" class="user-row">
            <td>
              <div class="user-cell">
                <div class="u-avatar" :style="`background: ${avatarColor(u.name)}`">
                  {{ initials(u.name) }}
                </div>
                <div class="u-info">
                  <div class="u-name">{{ u.name }}</div>
                  <div class="u-email">{{ u.email }}</div>
                </div>
              </div>
            </td>
            <td>
              <span class="role-badge" :class="`role-badge--${u.role}`">{{ u.role }}</span>
            </td>
            <td>
              <span class="status-badge" :class="`status-badge--${u.status}`">
                <span class="status-dot" />
                {{ u.status === 'active' ? 'Aktif' : u.status === 'pending' ? 'Menunggu' : u.status === 'suspended' ? 'Ditangguhkan' : u.status }}
              </span>
            </td>
            <td class="td-actions">
              <template v-if="u.role !== 'admin'">
                <template v-if="confirmSuspendId === u.id">
                  <span class="confirm-label">Tangguhkan?</span>
                  <button class="btn-sm btn-sm--danger" @click="updateStatus(u.id, 'suspended')" :disabled="busyId === u.id">
                    {{ busyId === u.id ? '···' : 'Ya' }}
                  </button>
                  <button class="btn-sm btn-sm--ghost" @click="confirmSuspendId = null">Batal</button>
                </template>
                <template v-else>
                  <button
                    v-if="u.status !== 'suspended'"
                    class="btn-sm btn-sm--outline-danger"
                    @click="confirmSuspendId = u.id"
                    :disabled="!!busyId"
                  >Tangguhkan</button>
                  <button
                    v-else
                    class="btn-sm btn-sm--outline-success"
                    @click="updateStatus(u.id, 'active')"
                    :disabled="!!busyId"
                  >{{ busyId === u.id ? '···' : 'Aktifkan' }}</button>
                </template>
              </template>
              <span v-else class="admin-lock">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- ── Pagination ───────────────────────────────────────────────────────── -->
    <div v-if="totalPages > 1" class="pagination">
      <button class="pg-btn" :disabled="page === 1" @click="page--; fetchUsers()">
        <svg viewBox="0 0 16 16" fill="currentColor" width="14" height="14"><path fill-rule="evenodd" d="M9.78 11.22a.75.75 0 01-1.06 0L5.47 7.97a.75.75 0 010-1.06l3.25-3.25a.75.75 0 011.06 1.06L7.06 7.44l2.72 2.72a.75.75 0 010 1.06z" clip-rule="evenodd"/></svg>
      </button>
      <button
        v-for="p in totalPages" :key="p"
        class="pg-btn" :class="{ 'pg-btn--active': p === page }"
        @click="page = p; fetchUsers()"
      >{{ p }}</button>
      <button class="pg-btn" :disabled="page >= totalPages" @click="page++; fetchUsers()">
        <svg viewBox="0 0 16 16" fill="currentColor" width="14" height="14"><path fill-rule="evenodd" d="M6.22 4.78a.75.75 0 011.06 0l3.25 3.25a.75.75 0 010 1.06L7.28 12.34a.75.75 0 01-1.06-1.06L9.94 8.5 7.22 5.78a.75.75 0 010-1.06z" clip-rule="evenodd"/></svg>
      </button>
    </div>

    <!-- ── Add Contributor Modal ─────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showAddModal" class="modal-backdrop" @click.self="showAddModal = false">
        <div class="modal-dialog">
          <div class="modal-header">
            <h3>Tambah Kontributor</h3>
            <button class="modal-close" @click="showAddModal = false">&times;</button>
          </div>
          <form @submit.prevent="createContributor" class="modal-body">
            <div class="field">
              <label>Nama Lengkap</label>
              <input v-model="addForm.name" type="text" required placeholder="Nama kontributor" maxlength="255" />
            </div>
            <div class="field">
              <label>Email</label>
              <input v-model="addForm.email" type="email" required placeholder="email@contoh.com" />
            </div>
            <div class="field">
              <label>Password</label>
              <input v-model="addForm.password" type="password" required placeholder="Min. 8 karakter" minlength="8" maxlength="72" />
            </div>
            <p v-if="addError" class="modal-error">{{ addError }}</p>
            <div class="modal-actions">
              <button type="button" class="btn-cancel" @click="showAddModal = false" :disabled="isAdding">Batal</button>
              <button type="submit" class="btn-submit" :disabled="isAdding">{{ isAdding ? 'Membuat…' : 'Buat Kontributor' }}</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<style scoped>
@keyframes fade-up {
  from { opacity: 0; transform: translateY(6px); }
  to   { opacity: 1; transform: translateY(0); }
}
@keyframes shimmer {
  0%   { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.uv { display: flex; flex-direction: column; gap: 1.5rem; }

/* ── Header ──────────────────────────────────────────────────────────────────── */
.uv-head { animation: fade-up 0.3s ease both; }
.uv-title {
  margin: 0;
  font-size: 1.875rem;
  font-weight: 800;
  color: var(--text-heading);
  letter-spacing: -0.03em;
  line-height: 1;
}
.uv-subtitle {
  margin: 0.3rem 0 0;
  font-size: 0.72rem;
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
}

/* ── Toolbar ─────────────────────────────────────────────────────────────────── */
.toolbar {
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  align-items: center;
  animation: fade-up 0.3s 0.05s ease both;
}
.search-wrap {
  position: relative;
  flex: 1;
  min-width: 200px;
}
.search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
  pointer-events: none;
}
.search-input {
  width: 100%;
  padding: 0.6rem 0.875rem 0.6rem 2.25rem;
  border-radius: 9px;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-primary);
  font-size: 0.875rem;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s;
  font-family: inherit;
}
.search-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent) 15%, transparent);
}

.pill-group { display: flex; gap: 0.3rem; flex-wrap: wrap; }
.pill {
  padding: 0.4rem 0.75rem;
  border-radius: 99px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  white-space: nowrap;
}
.pill:hover { border-color: var(--accent); color: var(--accent); }
.pill--active {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}
.pill--status-active.pill--active  { background: var(--success); border-color: var(--success); }
.pill--status-pending.pill--active { background: var(--warning); border-color: var(--warning); color: #000; }
.pill--status-suspended.pill--active { background: var(--danger); border-color: var(--danger); }

/* ── Error ───────────────────────────────────────────────────────────────────── */
.action-error {
  padding: 0.7rem 1rem;
  background: var(--danger-bg);
  color: var(--danger-text);
  border-radius: 8px;
  font-size: 0.825rem;
  border-left: 3px solid var(--danger);
}

/* ── Table wrap ──────────────────────────────────────────────────────────────── */
.table-wrap {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 14px;
  overflow: hidden;
  animation: fade-up 0.3s 0.1s ease both;
}

/* ── Skeletons ───────────────────────────────────────────────────────────────── */
.row-skeleton {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.875rem 1.25rem;
  border-bottom: 1px solid var(--border);
}
.row-skeleton:last-child { border-bottom: none; }
.sk-avatar {
  width: 36px; height: 36px;
  border-radius: 50%;
  background: linear-gradient(90deg, var(--border) 25%, var(--bg-input) 50%, var(--border) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
  flex-shrink: 0;
}
.sk-text { flex: 1; display: flex; flex-direction: column; gap: 0.4rem; }
.sk-block {
  border-radius: 4px;
  background: linear-gradient(90deg, var(--border) 25%, var(--bg-input) 50%, var(--border) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.4s infinite;
}
.sk-name  { height: 0.8rem; width: 45%; }
.sk-email { height: 0.65rem; width: 60%; }
.sk-badge { height: 1.4rem; width: 60px; border-radius: 99px; }

/* ── Empty ───────────────────────────────────────────────────────────────────── */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.625rem;
  padding: 3.5rem 2rem;
  color: var(--text-muted);
  font-size: 0.875rem;
  text-align: center;
}
.empty-state p { margin: 0; }

/* ── User table ──────────────────────────────────────────────────────────────── */
.user-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
}
.user-table th {
  padding: 0.75rem 1.25rem;
  text-align: left;
  font-size: 0.67rem;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.07em;
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
}
.th-actions { text-align: right; }
.user-table td { padding: 0.75rem 1.25rem; border-bottom: 1px solid var(--bg-input); vertical-align: middle; }
.user-table tbody tr:last-child td { border-bottom: none; }

.user-row {
  transition: background 0.12s;
}
.user-row:hover td { background: color-mix(in srgb, var(--accent) 4%, transparent); }

.user-cell { display: flex; align-items: center; gap: 0.75rem; }
.u-avatar {
  width: 36px; height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.68rem;
  font-weight: 800;
  color: #fff;
  flex-shrink: 0;
  letter-spacing: 0.02em;
}
.u-name  { font-weight: 600; color: var(--text-heading); font-size: 0.875rem; }
.u-email { font-size: 0.75rem; color: var(--text-muted); font-family: monospace; margin-top: 1px; }

/* role badge */
.role-badge {
  display: inline-block;
  padding: 0.2rem 0.6rem;
  border-radius: 6px;
  font-size: 0.7rem;
  font-weight: 700;
  text-transform: capitalize;
  letter-spacing: 0.03em;
}
.role-badge--student     { background: color-mix(in srgb, var(--accent) 12%, transparent); color: var(--accent); }
.role-badge--contributor { background: color-mix(in srgb, var(--success) 12%, transparent); color: var(--success); }
.role-badge--admin       { background: color-mix(in srgb, var(--warm) 12%, transparent); color: var(--warm); }

/* status badge */
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.78rem;
  font-weight: 600;
}
.status-dot {
  width: 7px; height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}
.status-badge--active    .status-dot { background: var(--success); box-shadow: 0 0 0 2px color-mix(in srgb, var(--success) 25%, transparent); }
.status-badge--pending   .status-dot { background: var(--warning); box-shadow: 0 0 0 2px color-mix(in srgb, var(--warning) 25%, transparent); }
.status-badge--suspended .status-dot { background: var(--danger);  box-shadow: 0 0 0 2px color-mix(in srgb, var(--danger) 25%, transparent); }
.status-badge--active    { color: var(--success); }
.status-badge--pending   { color: var(--warning); }
.status-badge--suspended { color: var(--danger);  }

/* actions cell */
.td-actions { text-align: right; white-space: nowrap; }
.confirm-label { font-size: 0.75rem; color: var(--danger); font-weight: 600; margin-right: 0.25rem; }
.admin-lock { color: var(--border); font-size: 0.875rem; }

.btn-sm {
  padding: 0.3rem 0.7rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
  border: none;
  transition: all 0.12s;
  margin-left: 0.25rem;
}
.btn-sm:disabled { opacity: 0.4; cursor: not-allowed; }
.btn-sm--outline-danger {
  background: transparent;
  border: 1px solid color-mix(in srgb, var(--danger) 40%, transparent);
  color: var(--danger);
}
.btn-sm--outline-danger:hover:not(:disabled) {
  background: color-mix(in srgb, var(--danger) 10%, transparent);
  border-color: var(--danger);
}
.btn-sm--outline-success {
  background: transparent;
  border: 1px solid color-mix(in srgb, var(--success) 40%, transparent);
  color: var(--success);
}
.btn-sm--outline-success:hover:not(:disabled) {
  background: color-mix(in srgb, var(--success) 10%, transparent);
  border-color: var(--success);
}
.btn-sm--danger { background: var(--danger); color: #fff; }
.btn-sm--ghost  { background: transparent; border: 1px solid var(--border); color: var(--text-muted); }

/* ── Pagination ──────────────────────────────────────────────────────────────── */
.pagination {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.25rem;
  animation: fade-up 0.3s 0.15s ease both;
}
.pg-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 34px;
  height: 34px;
  padding: 0 0.5rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  color: var(--text-muted);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.12s;
}
.pg-btn:hover:not(:disabled):not(.pg-btn--active) { border-color: var(--accent); color: var(--accent); }
.pg-btn:disabled { opacity: 0.35; cursor: not-allowed; }
.pg-btn--active { background: var(--accent); border-color: var(--accent); color: #fff; }

/* ── Header actions ──────────────────────────────────────────────────────────── */
.uv-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}
.btn-add {
  padding: 0.55rem 1rem;
  border-radius: 8px;
  border: none;
  background: var(--accent);
  color: #fff;
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.15s;
}
.btn-add:hover { background: var(--accent-hover); }

/* ── Filter badge & clear ────────────────────────────────────────────────────── */
.filter-badge {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 0.25rem 0.6rem;
  border-radius: 99px;
  background: color-mix(in srgb, var(--accent) 12%, transparent);
  color: var(--accent);
  white-space: nowrap;
}
.btn-clear {
  padding: 0.4rem 0.75rem;
  border-radius: 6px;
  border: 1px solid var(--danger);
  background: transparent;
  color: var(--danger);
  font-size: 0.72rem;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.12s;
}
.btn-clear:hover { background: var(--danger-bg); }

/* ── Modal ──────────────────────────────────────────────────────────────────── */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  animation: fade-up 0.15s ease both;
}
.modal-dialog {
  background: var(--bg-surface);
  border-radius: 14px;
  width: 90%;
  max-width: 420px;
  box-shadow: 0 16px 48px rgba(0,0,0,0.2);
  overflow: hidden;
}
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border);
}
.modal-header h3 { margin: 0; font-size: 0.95rem; font-weight: 700; color: var(--text-heading); }
.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: var(--text-muted);
  cursor: pointer;
  line-height: 1;
  padding: 0;
}
.modal-body { padding: 1.25rem; display: flex; flex-direction: column; gap: 0.875rem; }
.modal-body .field { display: flex; flex-direction: column; gap: 0.375rem; }
.modal-body .field > label { font-size: 0.75rem; font-weight: 600; color: var(--text-muted); }
.modal-body .field input {
  padding: 0.55rem 0.75rem;
  border-radius: 7px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: 0.85rem;
  outline: none;
  font-family: inherit;
}
.modal-body .field input:focus { border-color: var(--accent); }
.modal-error {
  margin: 0;
  padding: 0.5rem 0.75rem;
  border-radius: 6px;
  background: var(--danger-bg);
  color: var(--danger-text);
  font-size: 0.8rem;
}
.modal-actions {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 0.25rem;
}
.btn-cancel {
  padding: 0.5rem 1rem;
  border-radius: 7px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
}
.btn-submit {
  padding: 0.5rem 1.25rem;
  border-radius: 7px;
  border: none;
  background: var(--accent);
  color: #fff;
  font-size: 0.8rem;
  font-weight: 700;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-submit:hover { background: var(--accent-hover); }
.btn-submit:disabled { opacity: 0.5; cursor: not-allowed; }

/* ── Responsive ──────────────────────────────────────────────────────────────── */
@media (max-width: 768px) {
  .table-wrap { overflow-x: auto; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .search-wrap { min-width: unset; }
}
</style>
