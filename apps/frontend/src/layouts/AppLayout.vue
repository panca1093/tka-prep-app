<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

interface NavItem { label: string; to: string; icon: string }

const nav = computed((): NavItem[] => {
  if (auth.role === 'student') {
    return [
      { label: 'Dashboard', to: '/dashboard', icon: '⊞' },
      { label: 'Tests', to: '/tests', icon: '📄' },
      { label: 'Leaderboard', to: '/leaderboard', icon: '🏆' },
    ]
  }
  if (auth.role === 'contributor') {
    return [
      { label: 'Dashboard', to: '/contrib/dashboard', icon: '⊞' },
      { label: 'Question Bank', to: '/contrib/questions', icon: '❓' },
      { label: 'Tests', to: '/contrib/tests', icon: '📄' },
    ]
  }
  if (auth.role === 'admin') {
    return [
      { label: 'Dashboard', to: '/admin/dashboard', icon: '⊞' },
      { label: 'Users', to: '/admin/users', icon: '👥' },
      { label: 'Tests', to: '/admin/tests', icon: '📄' },
    ]
  }
  return []
})

async function handleLogout() {
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="shell">
    <nav class="sidebar">
      <div class="sidebar-top">
        <div class="brand">
          <span class="brand-mark">✦</span>
          <span class="brand-name">TKAPrep</span>
        </div>
        <ul class="nav-list">
          <li v-for="item in nav" :key="item.to">
            <RouterLink :to="item.to" class="nav-item" active-class="nav-item--active">
              <span class="nav-icon">{{ item.icon }}</span>
              {{ item.label }}
            </RouterLink>
          </li>
        </ul>
      </div>
      <div class="sidebar-bottom">
        <div class="user-info">
          <div class="user-name">{{ auth.user?.name }}</div>
          <div class="user-role">{{ auth.user?.role }}</div>
        </div>
        <button class="logout-btn" @click="handleLogout">Sign out</button>
      </div>
    </nav>
    <main class="content">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.shell {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 220px;
  flex-shrink: 0;
  background: #0d1424;
  border-right: 1px solid #1e2a45;
  display: flex;
  flex-direction: column;
  padding: 1.5rem 1rem;
  position: sticky;
  top: 0;
  height: 100vh;
}

.sidebar-top { flex: 1; }
.sidebar-bottom { margin-top: auto; }

.brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 2rem;
  padding: 0 0.5rem;
}

.brand-mark { font-size: 1.1rem; color: #4f8ef7; }
.brand-name { font-size: 1rem; font-weight: 800; color: #f1f5f9; }

.nav-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.6rem 0.75rem;
  border-radius: 8px;
  color: #94a3b8;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
  text-decoration: none;
}

.nav-item:hover { background: #1e2a45; color: #f1f5f9; }
.nav-item--active { background: #1a2f5c; color: #4f8ef7; }
.nav-icon { font-size: 1rem; width: 1.25rem; text-align: center; }

.user-info {
  padding: 0.75rem;
  border-top: 1px solid #1e2a45;
  margin-bottom: 0.5rem;
}

.user-name { font-size: 0.875rem; font-weight: 600; color: #f1f5f9; }
.user-role {
  font-size: 0.75rem;
  color: #94a3b8;
  text-transform: capitalize;
  margin-top: 2px;
}

.logout-btn {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  border: 1px solid #1e2a45;
  background: transparent;
  color: #94a3b8;
  font-size: 0.8rem;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.logout-btn:hover { background: #1e2a45; color: #ef4444; }

.content {
  flex: 1;
  padding: 2rem;
  overflow-y: auto;
  max-width: 1100px;
}
</style>
