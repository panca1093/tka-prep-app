<script setup lang="ts">
import { ref } from 'vue'
import client from '@/api/client'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const educationLevel = ref<string>(auth.user?.education_level ?? '')
const isSaving = ref(false)
const message = ref('')
const isError = ref(false)

const levelOptions = [
  { value: '', label: 'All Levels (no filter)' },
  { value: 'sd', label: 'SD — Sekolah Dasar' },
  { value: 'smp', label: 'SMP — Sekolah Menengah Pertama' },
  { value: 'sma', label: 'SMA — Sekolah Menengah Atas' },
  { value: 'smk', label: 'SMK — Sekolah Menengah Kejuruan' },
]

async function save() {
  isSaving.value = true
  message.value = ''
  try {
    const body: Record<string, string | null> = {}
    body.education_level = educationLevel.value || null
    const { data, error } = await client.PATCH('/auth/me', { body })
    if (error) {
      isError.value = true
      message.value = 'Failed to update profile. Please try again.'
      return
    }
    if (data) {
      if (auth.user) {
        auth.user.education_level = data.education_level
      }
      isError.value = false
      message.value = 'Profile updated successfully.'
    }
  } catch {
    isError.value = true
    message.value = 'Failed to update profile. Please try again.'
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <div class="profile-page">
    <h1>Profile</h1>

    <div class="profile-card">
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
        <p class="help-text">
          Tests will be filtered to match your education level. Set to "All Levels" to see everything.
        </p>
        <select v-model="educationLevel" class="edu-select">
          <option v-for="opt in levelOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <p v-if="message" class="msg" :class="{ error: isError }">{{ message }}</p>

      <button class="btn-primary" :disabled="isSaving" @click="save">
        {{ isSaving ? 'Saving…' : 'Save' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  max-width: 480px;
  margin: 0 auto;
  padding: 2rem 1rem;
}
h1 {
  margin: 0 0 1.5rem;
  font-size: 1.5rem;
  font-weight: 700;
}
.profile-card {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.5rem;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--bg-input);
}
.field {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}
.field > label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-muted);
}
.readonly {
  padding: 0.65rem 0.875rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-muted);
  font-size: 0.9rem;
}
.help-text {
  margin: 0;
  font-size: 0.75rem;
  color: var(--text-muted);
}
.edu-select {
  padding: 0.65rem 0.875rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-input);
  color: var(--text-primary);
  font-size: 0.9rem;
  outline: none;
}
.edu-select:focus {
  border-color: var(--accent);
}
.msg {
  margin: 0;
  padding: 0.6rem 0.75rem;
  border-radius: 8px;
  font-size: 0.825rem;
  background: var(--success-bg);
  color: var(--success);
}
.msg.error {
  background: var(--danger-bg);
  color: var(--danger-text);
}
.btn-primary {
  margin-top: 0.5rem;
  padding: 0.7rem;
  border-radius: 8px;
  border: none;
  background: var(--accent);
  color: var(--text-on-accent);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-primary:hover { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
