<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import client from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type Category = components['schemas']['Category']

const categories = ref<Category[]>([])
const isLoading = ref(false)
const error = ref('')

// Create form
const showCreate = ref(false)
const createName = ref('')
const createDesc = ref('')
const createSaving = ref(false)
const createError = ref('')

// Edit state
const editId = ref<string | null>(null)
const editName = ref('')
const editDesc = ref('')
const editSaving = ref(false)
const editError = ref('')

// Delete confirmation
const deleteTarget = ref<Category | null>(null)
const deleteBusy = ref(false)
const deleteError = ref('')

const toast = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

const adminCategories = computed(() =>
  categories.value.filter((c) => !(c as any).created_by)
)
const contributorCategories = computed(() =>
  categories.value.filter((c) => !!(c as any).created_by)
)

async function fetchCategories() {
  isLoading.value = true
  error.value = ''
  const { data, error: err } = await client.GET('/admin/categories')
  isLoading.value = false
  if (err) { error.value = 'Gagal memuat kategori.'; return }
  if (data) categories.value = data.data
}

async function submitCreate() {
  if (!createName.value.trim()) return
  createSaving.value = true
  createError.value = ''
  const { data, error: err, response } = await client.POST('/admin/categories', {
    body: { name: createName.value.trim() },
  })
  createSaving.value = false
  if (err) {
    const msg = (err as any)?.error?.message ?? 'Gagal membuat kategori.'
    if (response?.status === 400) createError.value = msg
    else createError.value = 'Gagal membuat kategori.'
    return
  }
  if (data) {
    categories.value.unshift(data as Category)
    createName.value = ''
    createDesc.value = ''
    showCreate.value = false
    showToast('Kategori berhasil dibuat.')
  }
}

function openEdit(cat: Category) {
  editId.value = cat.id as unknown as string
  editName.value = cat.name
  editDesc.value = (cat as any).description ?? ''
  editError.value = ''
}

function cancelEdit() {
  editId.value = null
  editError.value = ''
}

async function submitEdit(cat: Category) {
  if (!editId.value || !editName.value.trim()) return
  editSaving.value = true
  editError.value = ''
  const { data, error: err, response } = await client.PATCH('/admin/categories/{id}', {
    params: { path: { id: cat.id as unknown as string } },
    body: { name: editName.value.trim() },
  })
  editSaving.value = false
  if (err) {
    if (response?.status === 404) editError.value = 'Kategori tidak ditemukan.'
    else if (response?.status === 400) editError.value = (err as any)?.error?.message ?? 'Gagal menyimpan.'
    else editError.value = 'Gagal menyimpan.'
    return
  }
  const idx = categories.value.findIndex((c) => (c.id as unknown as string) === editId.value)
  if (idx !== -1 && data) categories.value[idx] = data as Category
  editId.value = null
  showToast('Kategori berhasil diperbarui.')
}

function openDelete(cat: Category) {
  deleteTarget.value = cat
  deleteError.value = ''
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleteBusy.value = true
  deleteError.value = ''
  const { error: err, response } = await client.DELETE('/admin/categories/{id}', {
    params: { path: { id: deleteTarget.value.id as unknown as string } },
  })
  deleteBusy.value = false
  if (err) {
    if (response?.status === 409) deleteError.value = 'Kategori sedang digunakan oleh ujian — tidak dapat dihapus.'
    else deleteError.value = 'Gagal menghapus kategori.'
    return
  }
  categories.value = categories.value.filter((c) => c.id !== deleteTarget.value!.id)
  deleteTarget.value = null
  showToast('Kategori dihapus.')
}

function showToast(msg: string) {
  toast.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toast.value = '' }, 3000)
}

onMounted(fetchCategories)
</script>

<template>
  <div class="uv" :class="{ ready: !isLoading }">

    <!-- ── Header ───────────────────────────────────────────────────────────── -->
    <div class="uv-head">
      <div>
        <h1 class="uv-title">Kategori</h1>
        <p class="uv-subtitle">{{ categories.length }} kategori</p>
      </div>
      <button class="btn-add" @click="showCreate = !showCreate">
        <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd"/></svg>
        Tambah Kategori
      </button>
    </div>

    <!-- ── Create card ──────────────────────────────────────────────────────── -->
    <div v-if="showCreate" class="create-card">
      <h2 class="section-title">Kategori Baru</h2>
      <p v-if="createError" class="field-error">{{ createError }}</p>
      <div class="form-row">
        <div class="field">
          <label class="field-label">Nama <span class="req">*</span></label>
          <input v-model="createName" class="field-input" placeholder="contoh: TKA Saintek 2026" maxlength="100" @keyup.enter="submitCreate" />
        </div>
      </div>
      <div class="form-actions">
        <button class="btn-save" :disabled="createSaving || !createName.trim()" @click="submitCreate">
          {{ createSaving ? 'Menyimpan…' : 'Simpan' }}
        </button>
        <button class="btn-cancel" @click="showCreate = false; createError = ''">Batal</button>
      </div>
    </div>

    <p v-if="error" class="action-error">{{ error }}</p>
    <div v-if="isLoading" class="loading">Memuat…</div>

    <template v-else>
      <!-- Admin-managed categories (no owner) -->
      <section class="cat-section">
        <h2 class="section-title">Kategori Sistem ({{ adminCategories.length }})</h2>
        <p class="section-sub">Dikelola admin — tersedia untuk semua kontributor.</p>
        <div v-if="adminCategories.length === 0" class="empty">Belum ada kategori sistem.</div>
        <div v-else class="cat-grid">
          <div v-for="cat in adminCategories" :key="cat.id as unknown as string" class="cat-card cat-card--system">
            <template v-if="editId !== (cat.id as unknown as string)">
              <div class="cat-info">
                <span class="cat-name">{{ cat.name }}</span>
                <span v-if="(cat as any).description" class="cat-desc">{{ (cat as any).description }}</span>
                <span class="cat-meta">{{ cat.test_count }} ujian</span>
              </div>
              <div class="cat-actions">
                <button class="btn-edit" @click="openEdit(cat)">Edit</button>
                <button class="btn-delete" @click="openDelete(cat)">Hapus</button>
              </div>
            </template>
            <template v-else>
              <div class="edit-form">
                <p v-if="editError" class="field-error">{{ editError }}</p>
                <input v-model="editName" class="field-input" placeholder="Nama kategori" maxlength="100" @keyup.enter="submitEdit(cat)" />
                <div class="form-actions" style="margin-top:0.5rem">
                  <button class="btn-save btn-save--sm" :disabled="editSaving || !editName.trim()" @click="submitEdit(cat)">
                    {{ editSaving ? '…' : 'Simpan' }}
                  </button>
                  <button class="btn-cancel btn-cancel--sm" @click="cancelEdit">Batal</button>
                </div>
              </div>
            </template>
          </div>
        </div>
      </section>

      <!-- Contributor-owned categories -->
      <section class="cat-section">
        <h2 class="section-title">Kategori Kontributor ({{ contributorCategories.length }})</h2>
        <p class="section-sub">Dibuat oleh kontributor — admin dapat menghapus jika bermasalah.</p>
        <div v-if="contributorCategories.length === 0" class="empty">Belum ada kategori kontributor.</div>
        <div v-else class="cat-grid">
          <div v-for="cat in contributorCategories" :key="cat.id as unknown as string" class="cat-card cat-card--owned">
            <template v-if="editId !== (cat.id as unknown as string)">
              <div class="cat-info">
                <span class="cat-name">{{ cat.name }}</span>
                <span v-if="(cat as any).description" class="cat-desc">{{ (cat as any).description }}</span>
                <span class="cat-meta">{{ cat.test_count }} ujian</span>
              </div>
              <div class="cat-actions">
                <button class="btn-edit" @click="openEdit(cat)">Edit</button>
                <button class="btn-delete" @click="openDelete(cat)">Hapus</button>
              </div>
            </template>
            <template v-else>
              <div class="edit-form">
                <p v-if="editError" class="field-error">{{ editError }}</p>
                <input v-model="editName" class="field-input" placeholder="Nama kategori" maxlength="100" @keyup.enter="submitEdit(cat)" />
                <div class="form-actions" style="margin-top:0.5rem">
                  <button class="btn-save btn-save--sm" :disabled="editSaving || !editName.trim()" @click="submitEdit(cat)">
                    {{ editSaving ? '…' : 'Simpan' }}
                  </button>
                  <button class="btn-cancel btn-cancel--sm" @click="cancelEdit">Batal</button>
                </div>
              </div>
            </template>
          </div>
        </div>
      </section>
    </template>

    <!-- ── Delete confirmation modal ─────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="deleteTarget" class="modal-backdrop" @click.self="deleteTarget = null">
        <div class="modal-dialog" style="max-width: 420px;">
          <div class="modal-header">
            <h2 class="modal-title">Hapus Kategori</h2>
            <button class="modal-close" @click="deleteTarget = null">×</button>
          </div>
          <div class="modal-body">
            <p v-if="deleteError" class="field-error">{{ deleteError }}</p>
            <template v-if="deleteTarget.test_count > 0">
              <div class="del-icon del-icon--warn">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="22" height="22">
                  <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
                </svg>
              </div>
              <h3 style="margin:0;font-size:0.95rem;text-align:center;">Tidak Dapat Menghapus</h3>
              <p style="margin:0;font-size:0.78rem;color:var(--text-muted);text-align:center;">
                Kategori <strong>{{ deleteTarget.name }}</strong> sedang digunakan oleh <strong>{{ deleteTarget.test_count }} ujian</strong>.
              </p>
              <button class="btn-primary" style="width:100%;" @click="deleteTarget = null">Mengerti</button>
            </template>
            <template v-else>
              <div class="del-icon del-icon--danger">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="22" height="22">
                  <path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14z"/>
                </svg>
              </div>
              <h3 style="margin:0;font-size:0.95rem;text-align:center;">Hapus Kategori</h3>
              <p style="margin:0;font-size:0.78rem;color:var(--text-muted);text-align:center;">
                Anda akan menghapus kategori <strong>{{ deleteTarget.name }}</strong>.
              </p>
              <div class="del-safe-badge">Aman untuk dihapus — tidak ada ujian terkait</div>
            </template>
          </div>
          <div v-if="deleteTarget.test_count === 0" class="modal-footer">
            <button class="btn-cancel" @click="deleteTarget = null">Batal</button>
            <button class="btn-primary" style="background: var(--danger);" :disabled="deleteBusy" @click="confirmDelete">
              {{ deleteBusy ? 'Menghapus…' : 'Ya, Hapus' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Toast -->
    <Teleport to="body">
      <div v-if="toast" class="toast">{{ toast }}</div>
    </Teleport>
  </div>
</template>

<style scoped>
@keyframes fade-up {
  from { opacity: 0; transform: translateY(6px); }
  to   { opacity: 1; transform: translateY(0); }
}

.uv { display: flex; flex-direction: column; gap: 1.5rem; }

/* ── Header ──────────────────────────────────────────────────────────────────── */
.uv-head {
  display: flex; align-items: flex-start; justify-content: space-between;
  gap: 1rem; animation: fade-up 0.3s ease both;
}
.uv-title {
  margin: 0; font-size: 1.875rem; font-weight: 800;
  color: var(--text-heading); letter-spacing: -0.03em; line-height: 1;
}
.uv-subtitle {
  margin: 0.3rem 0 0; font-size: 0.72rem;
  color: var(--text-muted); font-variant-numeric: tabular-nums;
}

.btn-add {
  display: flex; align-items: center; gap: 0.4rem;
  padding: 0.55rem 1rem; border-radius: 8px;
  border: none; background: var(--accent); color: #fff;
  font-size: 0.8rem; font-weight: 700; cursor: pointer;
  white-space: nowrap; transition: background 0.15s;
}
.btn-add:hover { background: var(--accent-hover); }

/* ── Action error ────────────────────────────────────────────────────────────── */
.action-error {
  padding: 0.7rem 1rem;
  background: var(--danger-bg); color: var(--danger-text);
  border-radius: 8px; font-size: 0.825rem;
  border-left: 3px solid var(--danger);
}

/* ── Create card ─────────────────────────────────────────────────────────────── */
.create-card {
  background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: 14px; padding: 1.25rem; animation: fade-up 0.3s 0.05s ease both;
}

.section-title { font-size: 0.82rem; font-weight: 700; color: var(--text-heading); margin: 0 0 0.75rem; text-transform: uppercase; letter-spacing: 0.04em; }
.section-sub { font-size: 0.77rem; color: var(--text-muted); margin: -0.5rem 0 0.75rem; }

.form-row { display: grid; grid-template-columns: 1fr; gap: 0.75rem; }

.field { display: flex; flex-direction: column; gap: 0.3rem; }
.field-label { font-size: 0.72rem; font-weight: 600; color: var(--text-muted); }
.req { color: var(--danger-text, #ef4444); }
.field-input {
  padding: 0.5rem 0.7rem; border-radius: 7px;
  border: 1px solid var(--border); background: var(--bg-input);
  color: var(--text-primary); font-size: 0.82rem;
  transition: border-color 0.15s; font-family: inherit;
}
.field-input:focus { outline: none; border-color: var(--accent); }
.field-error { font-size: 0.75rem; color: var(--danger-text, #ef4444); margin: 0 0 0.5rem; }

.form-actions { display: flex; gap: 0.5rem; }
.btn-save {
  padding: 0.45rem 1rem; border-radius: 7px; border: none;
  background: var(--accent); color: #fff; font-size: 0.8rem; font-weight: 600;
  cursor: pointer; transition: opacity 0.15s;
}
.btn-save:disabled { opacity: 0.45; cursor: not-allowed; }
.btn-save--sm { padding: 0.35rem 0.75rem; font-size: 0.75rem; }
.btn-cancel {
  padding: 0.45rem 0.85rem; border-radius: 7px;
  border: 1px solid var(--border); background: transparent;
  color: var(--text-muted); font-size: 0.8rem; font-weight: 600;
  cursor: pointer; transition: border-color 0.15s, color 0.15s;
}
.btn-cancel:hover { border-color: var(--text-muted); color: var(--text-primary); }
.btn-cancel--sm { padding: 0.35rem 0.65rem; font-size: 0.75rem; }

/* ── Category sections ────────────────────────────────────────────────────────── */
.cat-section { margin-bottom: 2rem; animation: fade-up 0.3s 0.1s ease both; }
.cat-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 0.65rem; }

.cat-card {
  display: flex; align-items: center; justify-content: space-between;
  gap: 0.75rem; padding: 0.9rem 1rem;
  background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: 10px;
}
.cat-card--system { border-left: 3px solid var(--border); }
.cat-card--owned { border-left: 3px solid var(--accent); }

.cat-info { display: flex; flex-direction: column; gap: 0.15rem; min-width: 0; flex: 1; }
.cat-name { font-size: 0.88rem; font-weight: 700; color: var(--text-heading); }
.cat-desc { font-size: 0.72rem; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cat-meta { font-size: 0.68rem; color: var(--text-muted); }

.cat-actions { display: flex; gap: 0.3rem; flex-shrink: 0; }
.btn-edit {
  padding: 0.3rem 0.7rem; border-radius: 6px;
  border: 1px solid var(--border); background: transparent;
  color: var(--accent); font-size: 0.75rem; font-weight: 600;
  cursor: pointer; transition: background 0.15s;
}
.btn-edit:hover { background: var(--accent-dim); }
.btn-delete {
  padding: 0.3rem 0.7rem; border-radius: 6px;
  border: 1px solid color-mix(in srgb, var(--danger) 40%, transparent);
  background: transparent; color: var(--danger);
  font-size: 0.75rem; font-weight: 600; cursor: pointer; transition: background 0.15s;
}
.btn-delete:hover { background: var(--danger-bg); }

.edit-form { flex: 1; }

.empty {
  font-size: 0.82rem; color: var(--text-muted); padding: 1.5rem; text-align: center;
  background: var(--bg-surface); border: 1px dashed var(--border); border-radius: 10px;
}
.loading { color: var(--text-muted); }

/* ── Modal ──────────────────────────────────────────────────────────────────── */
.modal-backdrop {
  position: fixed; inset: 0; background: rgba(0,0,0,0.45);
  display: flex; align-items: center; justify-content: center;
  z-index: 200; animation: fade-up 0.15s ease both;
}
.modal-dialog {
  background: var(--bg-surface); border-radius: 14px;
  width: 90%; box-shadow: 0 16px 48px rgba(0,0,0,0.2); overflow: hidden;
}
.modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 1rem 1.25rem; border-bottom: 1px solid var(--border);
}
.modal-title { margin: 0; font-size: 0.95rem; font-weight: 700; color: var(--text-heading); }
.modal-close {
  background: none; border: none; font-size: 1.5rem;
  color: var(--text-muted); cursor: pointer; line-height: 1; padding: 0;
}
.modal-body { padding: 1.25rem; display: flex; flex-direction: column; gap: 0.75rem; }
.modal-footer {
  display: flex; gap: 0.5rem; justify-content: flex-end;
  padding: 1rem 1.25rem; border-top: 1px solid var(--border);
}

.btn-primary {
  padding: 0.55rem 1.25rem; border-radius: 8px; border: none;
  background: var(--accent); color: #fff; font-size: 0.875rem; font-weight: 600;
  cursor: pointer; transition: background 0.15s;
}
.btn-primary:hover { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

/* ── Delete confirmation ──────────────────────────────────────────────────────── */
.del-icon {
  width: 44px; height: 44px; border-radius: 50%; margin: 0 auto;
  display: flex; align-items: center; justify-content: center;
}
.del-icon--warn { background: rgba(245,166,35,0.12); color: var(--warning); }
.del-icon--danger { background: rgba(255,82,82,0.1); color: var(--danger); }
.del-safe-badge {
  text-align: center; padding: 0.35rem 0.5rem; border-radius: 6px;
  font-size: 0.7rem; color: var(--success); background: rgba(0,210,160,0.06);
  border: 1px solid rgba(0,210,160,0.12);
}

.toast {
  position: fixed; bottom: 1.5rem; left: 50%; transform: translateX(-50%);
  background: #1a1a2e; color: #fff; font-size: 0.82rem; font-weight: 600;
  padding: 0.6rem 1.25rem; border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.25); z-index: 9999;
  white-space: nowrap; pointer-events: none;
}
</style>
