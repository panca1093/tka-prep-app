<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import client from '@/api/client'
import { updateOwnedCategory } from '@/api/contributor'
import { useAuthStore } from '@/stores/auth'
import type { components } from '@tkaprep/shared-types'

type Category = components['schemas']['Category']

const auth = useAuthStore()

const categories = ref<Category[]>([])
const isLoading = ref(false)
const error = ref('')

// Create form
const showCreate = ref(false)
const createName = ref('')
const createDesc = ref('')
const createSaving = ref(false)
const createError = ref('')

// Edit state (one at a time)
const editId = ref<string | null>(null)
const editName = ref('')
const editDesc = ref('')
const editSaving = ref(false)
const editError = ref('')

const toast = ref('')
let toastTimer: ReturnType<typeof setTimeout> | null = null

const ownedCategories = computed(() =>
  categories.value.filter((c) => (c as any).created_by === auth.user?.id)
)
const systemCategories = computed(() =>
  categories.value.filter((c) => !(c as any).created_by)
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
  const { data, error: err, response } = await client.POST('/contributor/categories', {
    body: {
      name: createName.value.trim(),
      description: createDesc.value.trim() || undefined,
    },
  })
  createSaving.value = false
  if (err) {
    const msg = (err as any)?.error?.message ?? 'Gagal membuat kategori.'
    if (response?.status === 409) {
      createError.value = 'Nama kategori sudah digunakan.'
    } else {
      createError.value = msg
    }
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

async function submitEdit() {
  if (!editId.value || !editName.value.trim()) return
  editSaving.value = true
  editError.value = ''
  const result = await updateOwnedCategory(editId.value, editName.value.trim(), editDesc.value.trim() || undefined)
  editSaving.value = false
  if (!result.ok) {
    editError.value = result.status === 409 ? 'Nama kategori sudah digunakan.' : (result.message ?? 'Gagal menyimpan.')
    return
  }
  const idx = categories.value.findIndex((c) => (c.id as unknown as string) === editId.value)
  if (idx !== -1 && result.data) categories.value[idx] = result.data
  editId.value = null
  showToast('Kategori berhasil diperbarui.')
}

function showToast(msg: string) {
  toast.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toast.value = '' }, 3000)
}

onMounted(fetchCategories)
</script>

<template>
  <div class="cat-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">Kategori</h1>
        <p class="page-sub">Kelola kategori ujian yang Anda miliki.</p>
      </div>
      <button class="btn-create" @click="showCreate = !showCreate">
        <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16"><path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd"/></svg>
        Buat Kategori
      </button>
    </div>

    <!-- Create form -->
    <div v-if="showCreate" class="create-card">
      <h2 class="section-title">Kategori Baru</h2>
      <p v-if="createError" class="field-error">{{ createError }}</p>
      <div class="form-row">
        <div class="field">
          <label class="field-label">Nama <span class="req">*</span></label>
          <input v-model="createName" class="field-input" placeholder="contoh: TKA Saintek 2026" maxlength="100" />
        </div>
        <div class="field field--desc">
          <label class="field-label">Deskripsi</label>
          <input v-model="createDesc" class="field-input" placeholder="Opsional" />
        </div>
      </div>
      <div class="form-actions">
        <button class="btn-save" :disabled="createSaving || !createName.trim()" @click="submitCreate">
          {{ createSaving ? 'Menyimpan…' : 'Simpan' }}
        </button>
        <button class="btn-cancel" @click="showCreate = false; createError = ''">Batal</button>
      </div>
    </div>

    <p v-if="error" class="err-msg">{{ error }}</p>
    <div v-if="isLoading" class="loading">Memuat…</div>

    <template v-else>
      <!-- Owned categories -->
      <section class="cat-section">
        <h2 class="section-title">Kategori Saya ({{ ownedCategories.length }})</h2>
        <div v-if="ownedCategories.length === 0" class="empty">
          Belum ada kategori. Buat kategori baru di atas untuk mulai menggunakannya dalam ujian Anda.
        </div>
        <div v-else class="cat-grid">
          <div
            v-for="cat in ownedCategories"
            :key="cat.id as unknown as string"
            class="cat-card cat-card--owned"
          >
            <!-- View mode -->
            <template v-if="editId !== (cat.id as unknown as string)">
              <div class="cat-info">
                <span class="cat-name">{{ cat.name }}</span>
                <span v-if="(cat as any).description" class="cat-desc">{{ (cat as any).description }}</span>
                <span class="cat-meta">{{ cat.test_count }} ujian</span>
              </div>
              <button class="btn-edit" @click="openEdit(cat)">Edit</button>
            </template>

            <!-- Edit mode -->
            <template v-else>
              <div class="edit-form">
                <p v-if="editError" class="field-error">{{ editError }}</p>
                <input v-model="editName" class="field-input" placeholder="Nama kategori" maxlength="100" />
                <input v-model="editDesc" class="field-input" placeholder="Deskripsi (opsional)" style="margin-top:0.4rem" />
                <div class="form-actions" style="margin-top:0.5rem">
                  <button class="btn-save btn-save--sm" :disabled="editSaving || !editName.trim()" @click="submitEdit">
                    {{ editSaving ? '…' : 'Simpan' }}
                  </button>
                  <button class="btn-cancel btn-cancel--sm" @click="cancelEdit">Batal</button>
                </div>
              </div>
            </template>
          </div>
        </div>
      </section>

      <!-- System categories -->
      <section class="cat-section">
        <h2 class="section-title">Kategori Sistem</h2>
        <p class="section-sub">Kategori ini tersedia untuk semua kontributor dan tidak dapat diedit.</p>
        <div class="cat-grid">
          <div v-for="cat in systemCategories" :key="cat.id as unknown as string" class="cat-card cat-card--system">
            <div class="cat-info">
              <span class="cat-name">{{ cat.name }}</span>
              <span v-if="(cat as any).description" class="cat-desc">{{ (cat as any).description }}</span>
              <span class="cat-meta">{{ cat.test_count }} ujian</span>
            </div>
            <span class="badge-system">Sistem</span>
          </div>
        </div>
      </section>
    </template>
  </div>

  <!-- Toast -->
  <Teleport to="body">
    <div v-if="toast" class="toast">{{ toast }}</div>
  </Teleport>
</template>

<style scoped>
.cat-page { max-width: 780px; }

.page-header {
  display: flex; align-items: flex-start; justify-content: space-between;
  margin-bottom: 1.5rem; gap: 1rem; flex-wrap: wrap;
}
.page-title { margin: 0; font-size: 1.35rem; font-weight: 800; }
.page-sub { margin: 0.2rem 0 0; font-size: 0.8rem; color: var(--text-muted); }

.btn-create {
  display: flex; align-items: center; gap: 0.4rem;
  padding: 0.55rem 1rem; border-radius: 8px;
  background: var(--accent); color: #fff; font-size: 0.82rem; font-weight: 600;
  border: none; cursor: pointer; white-space: nowrap;
  transition: opacity 0.15s;
}
.btn-create:hover { opacity: 0.88; }

.create-card {
  background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: 12px; padding: 1.25rem; margin-bottom: 1.5rem;
}

.section-title { font-size: 0.82rem; font-weight: 700; color: var(--text-heading); margin: 0 0 0.75rem; text-transform: uppercase; letter-spacing: 0.04em; }
.section-sub { font-size: 0.77rem; color: var(--text-muted); margin: -0.5rem 0 0.75rem; }

.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
@media (max-width: 600px) { .form-row { grid-template-columns: 1fr; } }

.field { display: flex; flex-direction: column; gap: 0.3rem; }
.field-label { font-size: 0.72rem; font-weight: 600; color: var(--text-muted); }
.req { color: var(--danger-text, #ef4444); }
.field-input {
  padding: 0.5rem 0.7rem; border-radius: 7px;
  border: 1px solid var(--border); background: var(--bg-input);
  color: var(--text-primary); font-size: 0.82rem;
  transition: border-color 0.15s;
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

.cat-section { margin-bottom: 2rem; }
.cat-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 0.65rem; }

.cat-card {
  display: flex; align-items: center; justify-content: space-between;
  gap: 1rem; padding: 0.9rem 1rem;
  background: var(--bg-surface); border: 1px solid var(--border);
  border-radius: 10px;
}
.cat-card--owned { border-left: 3px solid var(--accent); }
.cat-card--system { opacity: 0.8; }

.cat-info { display: flex; flex-direction: column; gap: 0.15rem; min-width: 0; }
.cat-name { font-size: 0.88rem; font-weight: 700; color: var(--text-heading); }
.cat-desc { font-size: 0.72rem; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.cat-meta { font-size: 0.68rem; color: var(--text-muted); }

.btn-edit {
  padding: 0.3rem 0.7rem; border-radius: 6px;
  border: 1px solid var(--border); background: transparent;
  color: var(--accent); font-size: 0.75rem; font-weight: 600;
  cursor: pointer; flex-shrink: 0; transition: background 0.15s;
}
.btn-edit:hover { background: var(--accent-dim); }

.badge-system {
  font-size: 0.65rem; font-weight: 700; padding: 0.2rem 0.55rem;
  border-radius: 20px; background: var(--bg-input); color: var(--text-muted);
  flex-shrink: 0; text-transform: uppercase; letter-spacing: 0.04em;
}

.edit-form { flex: 1; }
.empty { font-size: 0.82rem; color: var(--text-muted); padding: 1.5rem; text-align: center; background: var(--bg-surface); border: 1px dashed var(--border); border-radius: 10px; }
.loading { color: var(--text-muted); }
.err-msg { font-size: 0.82rem; color: var(--danger-text, #ef4444); }

.toast {
  position: fixed; bottom: 1.5rem; left: 50%; transform: translateX(-50%);
  background: #1a1a2e; color: #fff; font-size: 0.82rem; font-weight: 600;
  padding: 0.6rem 1.25rem; border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0,0,0,0.25); z-index: 9999;
  white-space: nowrap; pointer-events: none;
}
</style>
