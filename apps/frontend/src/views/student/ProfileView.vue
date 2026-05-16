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
  sd: 'SD — Sekolah Dasar',
  smp: 'SMP — Sekolah Menengah Pertama',
  sma: 'SMA — Sekolah Menengah Atas',
  smk: 'SMK — Sekolah Menengah Kejuruan',
}

const genderOptions = [
  { value: '', label: 'Prefer not to say' },
  { value: 'male', label: 'Male' },
  { value: 'female', label: 'Female' },
  { value: 'other', label: 'Other' },
]

async function handleUpload(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  const { data, error } = await client.POST('/upload', {
    body: formData as any,
  })
  if (error) {
    isError.value = true
    message.value = 'Failed to upload image.'
    return null
  }
  return (data as any)?.url ?? null
}

async function onAvatarChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  if (file.size > 2 * 1024 * 1024) {
    isError.value = true
    message.value = 'Image must be under 2 MB.'
    return
  }
  if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
    isError.value = true
    message.value = 'Only JPEG, PNG, or WebP images are allowed.'
    return
  }

  const url = await handleUpload(file)
  if (url) {
    avatarUrl.value = url
    isError.value = false
    message.value = 'Avatar uploaded. Save to apply.'
  }
}

async function save() {
  isSaving.value = true
  message.value = ''
  try {
    const body: Record<string, any> = {}
    if (gender.value) body.gender = gender.value
    else body.gender = null
    body.phone = phone.value || null
    body.avatar_url = avatarUrl.value || null

    const { data, error } = await client.PATCH('/auth/me', { body })
    if (error) {
      isError.value = true
      message.value = 'Failed to update profile. Please try again.'
      return
    }
    if (data && auth.user) {
      auth.user.gender = data.gender
      auth.user.phone = data.phone
      auth.user.avatar_url = data.avatar_url
    }
    isError.value = false
    message.value = 'Profile updated successfully.'
  } catch {
    isError.value = true
    message.value = 'Failed to update profile. Please try again.'
  } finally {
    isSaving.value = false
  }
}

async function handleLogout() {
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="profile-page">
    <h1>Profile</h1>

    <div class="profile-card">
      <!-- Avatar -->
      <div class="avatar-section">
        <div class="avatar-wrap">
          <img v-if="avatarUrl" :src="avatarUrl" class="avatar-img" alt="Profile picture" />
          <div v-else class="avatar-placeholder">
            {{ auth.user?.name?.split(' ').map(w => w[0]).join('').slice(0, 2).toUpperCase() }}
          </div>
        </div>
        <label class="avatar-upload-btn">
          Change Photo
          <input type="file" accept="image/jpeg,image/png,image/webp" class="avatar-input" @change="onAvatarChange" />
        </label>
      </div>

      <!-- Read-only fields -->
      <div class="field">
        <label>Name</label>
        <div class="readonly">{{ auth.user?.name }}</div>
      </div>
      <div class="field">
        <label>Email</label>
        <div class="readonly">{{ auth.user?.email }}</div>
      </div>
      <div class="field">
        <label>Education Level</label>
        <div class="readonly">
          {{ auth.user?.education_level ? levelLabel[auth.user.education_level] ?? auth.user.education_level : 'All Levels (not set)' }}
        </div>
        <p class="help-text">Set during registration. Contact support to change.</p>
      </div>

      <!-- Editable fields -->
      <div class="field">
        <label>Gender</label>
        <select v-model="gender" class="profile-select">
          <option v-for="opt in genderOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div class="field">
        <label>Phone Number</label>
        <input v-model="phone" type="tel" class="profile-input" placeholder="08xxxxxxxxxx" maxlength="20" />
      </div>

      <!-- Message -->
      <p v-if="message" class="msg" :class="{ error: isError }">{{ message }}</p>

      <button class="btn-primary" :disabled="isSaving" @click="save">
        {{ isSaving ? 'Saving…' : 'Save Changes' }}
      </button>

      <button class="btn-signout" @click="handleLogout">Sign out</button>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  max-width: 480px;
  margin: 0 auto;
  padding: 2rem 1rem;
}
h1 { margin: 0 0 1.5rem; font-size: 1.5rem; font-weight: 700; }

.profile-card {
  display: flex; flex-direction: column; gap: 1rem;
  padding: 1.5rem; border-radius: 12px;
  border: 1px solid var(--border); background: var(--bg-surface);
}

/* Avatar */
.avatar-section {
  display: flex; flex-direction: column; align-items: center; gap: 0.75rem;
  margin-bottom: 0.5rem;
}
.avatar-wrap { position: relative; }
.avatar-img {
  width: 96px; height: 96px; border-radius: 50%; object-fit: cover;
  border: 3px solid var(--accent);
}
.avatar-placeholder {
  width: 96px; height: 96px; border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), #764ba2);
  display: flex; align-items: center; justify-content: center;
  color: #fff; font-size: 1.5rem; font-weight: 700;
}
.avatar-upload-btn {
  font-size: 0.8rem; color: var(--accent); cursor: pointer; position: relative;
}
.avatar-upload-btn:hover { text-decoration: underline; }
.avatar-input { display: none; }

/* Fields */
.field { display: flex; flex-direction: column; gap: 0.375rem; }
.field > label { font-size: 0.8rem; font-weight: 600; color: var(--text-muted); }
.readonly {
  padding: 0.65rem 0.875rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-input);
  color: var(--text-primary); font-size: 0.9rem;
}
.help-text { margin: 0.25rem 0 0; font-size: 0.75rem; color: var(--text-muted); }

.profile-select, .profile-input {
  padding: 0.65rem 0.875rem; border-radius: 8px;
  border: 1px solid var(--border); background: var(--bg-input);
  color: var(--text-primary); font-size: 0.9rem; outline: none;
}
.profile-select:focus, .profile-input:focus { border-color: var(--accent); }

/* Messages */
.msg { margin: 0; padding: 0.6rem 0.75rem; border-radius: 8px; font-size: 0.825rem;
  background: color-mix(in srgb, var(--success) 15%, transparent); color: var(--success); }
.msg.error { background: var(--danger-bg); color: var(--danger-text); }

/* Buttons */
.btn-primary {
  margin-top: 0.5rem; padding: 0.7rem; border-radius: 8px; border: none;
  background: var(--accent); color: #fff; font-size: 0.9rem; font-weight: 600;
  cursor: pointer; transition: background 0.15s;
}
.btn-primary:hover { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-signout {
  margin-top: 0.5rem; padding: 0.7rem; border-radius: 8px;
  border: 1px solid var(--danger); background: transparent;
  color: var(--danger); font-size: 0.9rem; font-weight: 600;
  cursor: pointer; transition: background 0.15s;
}
.btn-signout:hover { background: var(--danger); color: #fff; }

@media (max-width: 768px) {
  .profile-page { padding: 1rem 0.75rem; }
  .profile-card { padding: 1.25rem 1rem; }
}
</style>
