<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, RouterLink, useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import MasukelasLogo from '@/components/MasukelasLogo.vue'
import ThemeToggle from '@/components/ThemeToggle.vue'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const isFullScreen = computed(() => !!route.meta.fullScreen)

interface NavItem { label: string; to: string; icon: string }

const nav = computed((): NavItem[] => {
  if (auth.role === 'student') {
    return [
      { label: 'Dashboard', to: '/dashboard', icon: '⊞' },
      { label: 'Tests', to: '/tests', icon: '📄' },
      { label: 'Leaderboard', to: '/leaderboard', icon: '🏆' },
      { label: 'Profile', to: '/profile', icon: '👤' },
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
          <MasukelasLogo />
        </div>
        <ul class="nav-list">
          <li v-for="item in nav" :key="item.to">
            <RouterLink :to="item.to" class="nav-item" active-class="nav-item--active">
              <span class="nav-icon">{{ item.icon }}</span>
              <span class="nav-label">{{ item.label }}</span>
            </RouterLink>
          </li>
        </ul>
      </div>
      <div class="sidebar-bottom">
        <ThemeToggle />
        <div class="user-info">
          <div class="user-name">{{ auth.user?.name }}</div>
          <div class="user-role">{{ auth.user?.role }}</div>
        </div>
        <button class="logout-btn" @click="handleLogout">Sign out</button>
      </div>
    </nav>
    <main :class="['content', { 'content--full': isFullScreen }]">
      <RouterView />
    </main>

    <!-- Mobile bottom nav (hidden during full-screen test sessions) -->
    <nav v-if="!isFullScreen" class="bottom-nav">
      <RouterLink
        v-for="item in nav"
        :key="item.to"
        :to="item.to"
        class="bottom-nav-item"
        active-class="bottom-nav-item--active"
      >
        <span class="bottom-nav-icon">{{ item.icon }}</span>
        <span class="bottom-nav-label">{{ item.label }}</span>
      </RouterLink>
    </nav>
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
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
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
  color: var(--text-muted);
  font-size: 0.875rem;
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
  text-decoration: none;
}

.nav-item:hover { background: var(--border); color: var(--text-primary); }
.nav-item--active { background: var(--bg-active-nav); color: var(--text-active-nav); }
.nav-icon { font-size: 1rem; width: 1.25rem; text-align: center; }

.user-info {
  padding: 0.75rem;
  border-top: 1px solid var(--border);
  margin-bottom: 0.5rem;
}

.user-name { font-size: 0.875rem; font-weight: 600; color: var(--text-primary); }
.user-role {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: capitalize;
  margin-top: 2px;
}

.logout-btn {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--text-muted);
  font-size: 0.8rem;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.logout-btn:hover { background: var(--border); color: var(--danger); }

.content {
  flex: 1;
  padding: 2rem;
  overflow-y: auto;
  max-width: 1100px;
}

.content--full {
  padding: 0;
  max-width: unset;
  overflow: hidden;
}

/* Bottom navigation (mobile) */
.bottom-nav {
  display: none;
}

.bottom-nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 0.5rem 0.25rem;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.7rem;
  font-weight: 500;
  border-radius: 8px;
  transition: color 0.15s;
}

.bottom-nav-item--active {
  color: var(--text-active-nav);
}

.bottom-nav-icon {
  font-size: 1.2rem;
}

.bottom-nav-label {
  font-size: 0.65rem;
}

/* Tablet: icon-only sidebar */
@media (min-width: 769px) and (max-width: 1024px) {
  .sidebar {
    width: 60px;
    padding: 1rem 0.5rem;
  }
  .nav-label { display: none; }
  .brand :deep(.logo-text) { display: none; }
  .user-info { display: none; }
}

/* Mobile: bottom nav, hide sidebar */
@media (max-width: 768px) {
  .sidebar { display: none; }
  .bottom-nav {
    display: flex;
    justify-content: space-around;
    align-items: center;
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    height: 60px;
    background: var(--bg-sidebar);
    border-top: 1px solid var(--border);
    z-index: 100;
  }
  .content {
    padding: 1rem;
    padding-bottom: 76px;
    max-width: 100%;
  }
}
</style>
