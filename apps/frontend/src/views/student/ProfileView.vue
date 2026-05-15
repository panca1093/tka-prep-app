<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const levelLabel: Record<string, string> = {
  sd: 'SD — Sekolah Dasar',
  smp: 'SMP — Sekolah Menengah Pertama',
  sma: 'SMA — Sekolah Menengah Atas',
  smk: 'SMK — Sekolah Menengah Kejuruan',
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
        <p class="help-text">Education level is set during registration and cannot be changed here.</p>
      </div>

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
  background: var(--bg-surface);
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
  color: var(--text-primary);
  font-size: 0.9rem;
}
.help-text {
  margin: 0.25rem 0 0;
  font-size: 0.75rem;
  color: var(--text-muted);
}
.btn-signout {
  margin-top: 1rem;
  padding: 0.7rem;
  border-radius: 8px;
  border: 1px solid var(--danger);
  background: transparent;
  color: var(--danger);
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
}
.btn-signout:hover { background: var(--danger); color: #fff; }
</style>
