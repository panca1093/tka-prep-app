<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import client from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const gender = ref<string>(auth.user?.gender ?? '')
const phone = ref<string>(auth.user?.phone ?? '')
const avatarUrl = ref<string>(auth.user?.avatar_url ?? '')
const isSaving = ref(false)
const message = ref('')
const isError = ref(false)

const levelLabel: Record<string, string> = {
  sd: 'SD', smp: 'SMP', sma: 'SMA', smk: 'SMK',
}

const genderOptions = [
  { value: '', label: 'Prefer not to say' },
  { value: 'male', label: 'Male' },
  { value: 'female', label: 'Female' },
  { value: 'other', label: 'Other' },
]

async function uploadAvatar(file: File): Promise<string | null> {
  const formData = new FormData()
  formData.append('file', file)
  const { data, error } = await client.POST('/upload', { body: formData as any })
  if (error) return null
  return (data as any)?.url ?? null
}

async function onAvatarChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  if (file.size > 2 * 1024 * 1024) { message.value = 'Image must be under 2 MB.'; isError.value = true; return }
  if (!['image/jpeg','image/png','image/webp'].includes(file.type)) { message.value = 'Only JPEG, PNG, or WebP allowed.'; isError.value = true; return }
  const url = await uploadAvatar(file)
  if (url) { avatarUrl.value = url; message.value = 'Avatar ready. Save to apply.'; isError.value = false }
  else { message.value = 'Upload failed.'; isError.value = true }
}

async function save() {
  isSaving.value = true; message.value = ''
  try {
    const body: Record<string, any> = {}
    body.gender = gender.value || null
    body.phone = phone.value || null
    body.avatar_url = avatarUrl.value || null
    const { data, error } = await client.PATCH('/auth/me', { body })
    if (error) { isError.value = true; message.value = 'Failed to save.'; return }
    if (data && auth.user) {
      auth.user.gender = data.gender; auth.user.phone = data.phone; auth.user.avatar_url = data.avatar_url
    }
    isError.value = false; message.value = 'Profile updated.'
  } catch { isError.value = true; message.value = 'Failed to save.' }
  finally { isSaving.value = false }
}

async function handleLogout() { await auth.logout(); router.push({ name: 'login' }) }
</script>

<template>
  <div class="profile-page">
    <h1>Profile</h1>
    <div class="profile-card">
      <div class="avatar-section">
        <div class="avatar-wrap">
          <img v-if="avatarUrl" :src="avatarUrl" class="avatar-img" alt="" />
          <div v-else class="avatar-placeholder">{{ auth.user?.name?.split(' ').map(w=>w[0]).join('').slice(0,2).toUpperCase() }}</div>
        </div>
        <label class="avatar-upload-btn">Change Photo<input type="file" accept="image/*" class="avatar-input" @change="onAvatarChange" /></label>
      </div>
      <div class="field"><label>Name</label><div class="readonly">{{ auth.user?.name }}</div></div>
      <div class="field"><label>Email</label><div class="readonly">{{ auth.user?.email }}</div></div>
      <div class="field"><label>Education Level</label><div class="readonly">{{ auth.user?.education_level ? levelLabel[auth.user.education_level] ?? auth.user.education_level : 'Not set' }}</div><p class="help-text">Set during registration.</p></div>
      <div class="field"><label>Gender</label><select v-model="gender" class="profile-select"><option v-for="o in genderOptions" :key="o.value" :value="o.value">{{ o.label }}</option></select></div>
      <div class="field"><label>Phone</label><input v-model="phone" type="tel" class="profile-input" placeholder="08xxxxxxxxxx" maxlength="20" /></div>
      <p v-if="message" class="msg" :class="{ error: isError }">{{ message }}</p>
      <button class="btn-primary" :disabled="isSaving" @click="save">{{ isSaving ? 'Saving…' : 'Save Changes' }}</button>
      <button class="btn-signout" @click="handleLogout">Sign out</button>
    </div>
  </div>
</template>

<style scoped>
.profile-page { max-width:480px;margin:0 auto;padding:2rem 1rem }
h1 { margin:0 0 1.5rem;font-size:1.5rem;font-weight:700 }
.profile-card { display:flex;flex-direction:column;gap:1rem;padding:1.5rem;border-radius:12px;border:1px solid var(--border);background:var(--bg-surface) }
.avatar-section { display:flex;flex-direction:column;align-items:center;gap:.75rem;margin-bottom:.5rem }
.avatar-img,.avatar-placeholder { width:96px;height:96px;border-radius:50% }
.avatar-img { object-fit:cover;border:3px solid var(--accent) }
.avatar-placeholder { background:linear-gradient(135deg,var(--accent),#764ba2);display:flex;align-items:center;justify-content:center;color:#fff;font-size:1.5rem;font-weight:700 }
.avatar-upload-btn { font-size:.8rem;color:var(--accent);cursor:pointer;position:relative }
.avatar-upload-btn:hover { text-decoration:underline }
.avatar-input { display:none }
.field { display:flex;flex-direction:column;gap:.375rem }
.field>label { font-size:.8rem;font-weight:600;color:var(--text-muted) }
.readonly { padding:.65rem .875rem;border-radius:8px;border:1px solid var(--border);background:var(--bg-input);color:var(--text-primary);font-size:.9rem }
.help-text { margin:.25rem 0 0;font-size:.75rem;color:var(--text-muted) }
.profile-select,.profile-input { padding:.65rem .875rem;border-radius:8px;border:1px solid var(--border);background:var(--bg-input);color:var(--text-primary);font-size:.9rem;outline:none }
.profile-select:focus,.profile-input:focus { border-color:var(--accent) }
.msg { margin:0;padding:.6rem .75rem;border-radius:8px;font-size:.825rem;background:color-mix(in srgb,var(--success) 15%,transparent);color:var(--success) }
.msg.error { background:var(--danger-bg);color:var(--danger-text) }
.btn-primary { margin-top:.5rem;padding:.7rem;border-radius:8px;border:none;background:var(--accent);color:#fff;font-size:.9rem;font-weight:600;cursor:pointer;transition:background .15s }
.btn-primary:hover { background:var(--accent-hover) }
.btn-primary:disabled { opacity:.6;cursor:not-allowed }
.btn-signout { margin-top:.5rem;padding:.7rem;border-radius:8px;border:1px solid var(--danger);background:transparent;color:var(--danger);font-size:.9rem;font-weight:600;cursor:pointer;transition:background .15s }
.btn-signout:hover { background:var(--danger);color:#fff }
@media(max-width:768px) { .profile-page { padding:1rem .75rem } .profile-card { padding:1.25rem 1rem } }
</style>
