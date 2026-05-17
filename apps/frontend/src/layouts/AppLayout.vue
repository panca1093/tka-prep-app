<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, RouterLink, useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import ThemeToggle from '@/components/ThemeToggle.vue'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const isFullScreen = computed(() => !!route.meta.fullScreen)

interface NavItem { label: string; to: string; icon: string }

const nav = computed((): NavItem[] => {
  if (auth.role === 'student') {
    return [
      { label: 'Dashboard', to: '/dashboard', icon: 'home' },
      { label: 'Ujian', to: '/tests', icon: 'doc' },
      { label: 'Peringkat', to: '/leaderboard', icon: 'trophy' },
      { label: 'Profil', to: '/profile', icon: 'user' },
    ]
  }
  if (auth.role === 'contributor') {
    return [
      { label: 'Dashboard', to: '/contrib/dashboard', icon: 'home' },
      { label: 'Bank Soal', to: '/contrib/questions', icon: 'bank' },
      { label: 'Ujian', to: '/contrib/tests', icon: 'doc' },
    ]
  }
  if (auth.role === 'admin') {
    return [
      { label: 'Dashboard', to: '/admin/dashboard', icon: 'home' },
      { label: 'Pengguna', to: '/admin/users', icon: 'users' },
      { label: 'Ujian', to: '/admin/tests', icon: 'doc' },
    ]
  }
  return []
})

const initials = computed(() => {
  const name = auth.user?.name ?? ''
  return name.split(' ').map((w: string) => w[0]).join('').substring(0, 2).toUpperCase()
})

async function handleLogout() {
  await auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="shell">
    <nav v-if="!isFullScreen" class="sidebar">
      <div class="sidebar-top">
        <!-- Brand -->
        <div class="brand">
          <div class="brand-mark">
            <svg viewBox="0 0 24 24" fill="none" width="20" height="20">
              <path d="M12 3L1 9l11 6 9-4.91V17h2V9L12 3z" fill="currentColor"/>
              <path d="M5 13.18v4L12 21l7-3.82v-4L12 17l-7-3.82z" fill="currentColor" opacity=".85"/>
            </svg>
          </div>
          <div class="brand-text">
            <div class="brand-name">Masukelas</div>
          </div>
        </div>

        <!-- Nav items -->
        <ul class="nav-list">
          <li v-for="item in nav" :key="item.to">
            <RouterLink :to="item.to" class="nav-item" active-class="nav-item--active">
              <!-- Home icon -->
              <svg v-if="item.icon === 'home'" class="nav-icon" viewBox="0 0 20 20" fill="currentColor"><path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h4a1 1 0 001-1v-3h2v3a1 1 0 001 1h4a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z"/></svg>
              <!-- Doc icon -->
              <svg v-else-if="item.icon === 'doc'" class="nav-icon" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clip-rule="evenodd"/></svg>
              <!-- Trophy icon -->
              <svg v-else-if="item.icon === 'trophy'" class="nav-icon" viewBox="0 0 20 20" fill="currentColor"><path d="M2 4a1 1 0 011-1h1V2a1 1 0 012 0v1h8V2a1 1 0 112 0v1h1a1 1 0 011 1v2a4 4 0 01-3.16 3.906A6.01 6.01 0 0111 13.32V15h1a1 1 0 110 2H8a1 1 0 110-2h1v-1.68a6.01 6.01 0 01-3.84-3.414A4 4 0 012 6V4zm2 2v.354A2 2 0 006 8.236V5H4v1zm12 0v-.999h-2v3.236A2 2 0 0016 6.354V6z"/></svg>
              <!-- User icon -->
              <svg v-else-if="item.icon === 'user'" class="nav-icon" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd"/></svg>
              <!-- Bank icon -->
              <svg v-else-if="item.icon === 'bank'" class="nav-icon" viewBox="0 0 20 20" fill="currentColor"><path d="M7 3a1 1 0 000 2h6a1 1 0 100-2H7zM4 7a1 1 0 011-1h10a1 1 0 110 2H5a1 1 0 01-1-1zM2 11a2 2 0 012-2h12a2 2 0 012 2v4a2 2 0 01-2 2H4a2 2 0 01-2-2v-4z"/></svg>
              <!-- Users icon -->
              <svg v-else-if="item.icon === 'users'" class="nav-icon" viewBox="0 0 20 20" fill="currentColor"><path d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z"/></svg>
              <span class="nav-label">{{ item.label }}</span>
            </RouterLink>
          </li>
        </ul>
      </div>

      <!-- Sidebar bottom -->
      <div class="sidebar-bottom">
        <ThemeToggle />
        <div class="user-card">
          <div class="user-avatar">
                <img v-if="auth.user?.avatar_url" :src="auth.user.avatar_url" class="avatar-img" alt="" />
                <span v-else>{{ initials }}</span>
              </div>
          <div class="user-info">
            <div class="user-name">{{ auth.user?.name }}</div>
            <div class="user-role">{{ auth.user?.role }}</div>
          </div>
          <button class="logout-icon" title="Sign out" @click="handleLogout">
            <svg viewBox="0 0 20 20" fill="currentColor" width="16" height="16"><path fill-rule="evenodd" d="M3 3a1 1 0 00-1 1v12a1 1 0 102 0V4a1 1 0 00-1-1zm10.293 9.293a1 1 0 001.414 1.414l3-3a1 1 0 000-1.414l-3-3a1 1 0 10-1.414 1.414L14.586 9H7a1 1 0 100 2h7.586l-1.293 1.293z" clip-rule="evenodd"/></svg>
          </button>
        </div>
      </div>
    </nav>

    <main :class="['content', { 'content--full': isFullScreen }]">
      <RouterView />
    </main>

    <!-- Mobile bottom nav -->
    <nav v-if="!isFullScreen" class="bottom-nav">
      <RouterLink
        v-for="item in nav"
        :key="item.to"
        :to="item.to"
        class="bottom-nav-item"
        active-class="bottom-nav-item--active"
      >
        <svg v-if="item.icon === 'home'" class="bn-icon" viewBox="0 0 20 20" fill="currentColor"><path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h4a1 1 0 001-1v-3h2v3a1 1 0 001 1h4a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z"/></svg>
        <svg v-else-if="item.icon === 'doc'" class="bn-icon" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clip-rule="evenodd"/></svg>
        <svg v-else-if="item.icon === 'trophy'" class="bn-icon" viewBox="0 0 20 20" fill="currentColor"><path d="M2 4a1 1 0 011-1h1V2a1 1 0 012 0v1h8V2a1 1 0 112 0v1h1a1 1 0 011 1v2a4 4 0 01-3.16 3.906A6.01 6.01 0 0111 13.32V15h1a1 1 0 110 2H8a1 1 0 110-2h1v-1.68a6.01 6.01 0 01-3.84-3.414A4 4 0 012 6V4zm2 2v.354A2 2 0 006 8.236V5H4v1zm12 0v-.999h-2v3.236A2 2 0 0016 6.354V6z"/></svg>
        <svg v-else-if="item.icon === 'user'" class="bn-icon" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clip-rule="evenodd"/></svg>
        <span class="bn-label">{{ item.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<style scoped>
.shell { display: flex; min-height: 100vh; }

/* ─── Sidebar ─────────────────────────────────────────────────────────────── */
.sidebar {
  width: 230px;
  flex-shrink: 0;
  background: var(--bg-sidebar);
  display: flex;
  flex-direction: column;
  padding: 1.25rem 0.875rem;
  position: sticky;
  top: 0;
  height: 100vh;
  overflow: hidden;
}

.sidebar-top { flex: 1; display: flex; flex-direction: column; gap: 0; }
.sidebar-bottom { margin-top: auto; display: flex; flex-direction: column; gap: 0.75rem; padding-top: 0.75rem; border-top: 1px solid var(--border); }

/* Brand */
.brand {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  margin-bottom: 2rem;
  padding: 0 0.375rem;
}
.brand-mark {
  width: 32px; height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #4f8ef7, #3b7be8);
  display: flex; align-items: center; justify-content: center;
  font-size: 1rem; font-weight: 800; color: #fff;
  flex-shrink: 0;
}
.brand-name { font-size: 0.9rem; font-weight: 700; color: var(--text-heading); line-height: 1.1; }

/* Nav */
.nav-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 2px; }
.nav-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.6rem 0.75rem;
  border-radius: 8px;
  color: var(--text-muted);
  font-size: 0.85rem;
  font-weight: 500;
  transition: background 0.15s, color 0.15s;
  text-decoration: none;
  position: relative;
}
.nav-item:hover { background: var(--accent-dim); color: var(--text-primary); }
.nav-item--active {
  background: var(--bg-active-nav);
  color: var(--text-active-nav);
}
.nav-item--active::before {
  content: '';
  position: absolute;
  left: 0; top: 6px; bottom: 6px;
  width: 2.5px;
  background: var(--accent);
  border-radius: 0 2px 2px 0;
}
.nav-icon { width: 18px; height: 18px; flex-shrink: 0; }

/* User card */
.user-card {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.625rem 0.375rem;
}
.user-avatar {
  width: 36px; height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4f8ef7, #3b7be8);
  display: flex; align-items: center; justify-content: center;
  font-size: 0.72rem; font-weight: 700; color: #fff;
  flex-shrink: 0; overflow: hidden;
  border: 2px solid color-mix(in srgb, var(--accent) 25%, var(--border));
}
.avatar-img { width: 100%; height: 100%; object-fit: cover; }
.user-info { flex: 1; min-width: 0; }
.user-name { font-size: 0.78rem; font-weight: 600; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.user-role { font-size: 0.65rem; color: var(--text-muted); text-transform: capitalize; margin-top: 1px; }
.logout-icon {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 6px;
  display: flex;
  transition: color 0.15s, background 0.15s;
  flex-shrink: 0;
}
.logout-icon:hover { color: #ef4444; background: rgba(239,68,68,0.1); }

/* ─── Content ─────────────────────────────────────────────────────────────── */
.content {
  flex: 1;
  padding: 2rem 2.5rem;
  overflow-y: auto;
  min-width: 0;
}
.content--full { padding: 0; max-width: unset; overflow: hidden; }

/* ─── Bottom nav (mobile) ─────────────────────────────────────────────────── */
.bottom-nav { display: none; }

.bottom-nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 0.5rem 0.5rem;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 0.62rem;
  font-weight: 500;
  border-radius: 8px;
  transition: color 0.15s;
  flex: 1;
}
.bottom-nav-item--active { color: var(--text-active-nav); }
.bn-icon { width: 20px; height: 20px; }
.bn-label { font-size: 0.62rem; }

/* ─── Responsive ──────────────────────────────────────────────────────────── */
@media (min-width: 769px) and (max-width: 1024px) {
  .sidebar { width: 64px; padding: 1rem 0.625rem; }
  .brand-text, .nav-label, .user-info { display: none; }
  .brand { justify-content: center; margin-bottom: 1.5rem; padding: 0; }
  .nav-item { justify-content: center; padding: 0.6rem; }
  .user-card { justify-content: center; padding: 0.5rem 0; }
  .logout-icon { display: none; }
  .content { padding: 1.5rem; }
}

@media (max-width: 768px) {
  .sidebar { display: none; }
  .bottom-nav {
    display: flex;
    position: fixed;
    bottom: 0; left: 0; right: 0;
    height: 60px;
    background: var(--bg-sidebar);
    border-top: 1px solid var(--border);
    z-index: 100;
  }
  .content { padding: 1rem; padding-bottom: 76px; max-width: 100%; }
}
</style>
