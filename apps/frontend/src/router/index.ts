import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // ─── Public auth ──────────────────────────────────────────────────────────
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/auth/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/auth/RegisterView.vue'),
      meta: { public: true },
    },

    // ─── Protected (all under AppLayout) ─────────────────────────────────────
    {
      path: '/',
      component: () => import('@/layouts/AppLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        { path: '', redirect: () => { const auth = useAuthStore(); return roleHome(auth.role) } },

        // Student
        { path: 'dashboard', name: 'student-dashboard', component: () => import('@/views/student/DashboardView.vue'), meta: { role: 'student' } },
        { path: 'tests', name: 'test-list', component: () => import('@/views/student/TestListView.vue'), meta: { role: 'student' } },
        { path: 'sessions/:sessionId', name: 'test-session', component: () => import('@/views/student/TestSessionView.vue'), meta: { role: 'student', fullScreen: true } },
        { path: 'results/:resultId', name: 'result', component: () => import('@/views/student/ResultView.vue'), meta: { role: 'student' } },
        { path: 'results/:resultId/review', name: 'review', component: () => import('@/views/student/ReviewView.vue'), meta: { role: 'student' } },
        { path: 'leaderboard', name: 'leaderboard', component: () => import('@/views/student/LeaderboardView.vue'), meta: { role: 'student' } },
        { path: 'profile', name: 'student-profile', component: () => import('@/views/student/ProfileView.vue'), meta: { role: 'student' } },

        // Contributor
        { path: 'contrib', redirect: { name: 'contrib-dashboard' } },
        { path: 'contrib/dashboard', name: 'contrib-dashboard', component: () => import('@/views/contributor/DashboardView.vue'), meta: { role: 'contributor' } },
        { path: 'contrib/questions', name: 'question-bank', component: () => import('@/views/contributor/QuestionBankView.vue'), meta: { role: 'contributor' } },
        { path: 'contrib/tests', name: 'test-builder', component: () => import('@/views/contributor/TestBuilderView.vue'), meta: { role: 'contributor' } },

        // Admin
        { path: 'admin', redirect: { name: 'admin-dashboard' } },
        { path: 'admin/dashboard', name: 'admin-dashboard', component: () => import('@/views/admin/DashboardView.vue'), meta: { role: 'admin' } },
        { path: 'admin/users', name: 'admin-users', component: () => import('@/views/admin/UsersView.vue'), meta: { role: 'admin' } },
        { path: 'admin/tests', name: 'admin-tests', component: () => import('@/views/admin/AdminTestsView.vue'), meta: { role: 'admin' } },
        
      ],
    },

    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('@/views/NotFoundView.vue') },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.isInitialized) await auth.init()

  if (to.meta.public) {
    if (auth.isAuthenticated) return roleHome(auth.role)
    return true
  }

  if (!auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  const requiredRole = to.meta.role as string | undefined
  if (requiredRole && auth.role !== requiredRole) {
    return roleHome(auth.role)
  }

  // Student must have education_level to access tests & leaderboard
  if (
    auth.role === 'student' &&
    !auth.user?.education_level &&
    ['test-list', 'test-session', 'leaderboard'].includes(to.name as string)
  ) {
    return { name: 'student-profile' }
  }

  return true
})

function roleHome(role: string | null) {
  if (role === 'student') return { name: 'student-dashboard' }
  if (role === 'contributor') return { name: 'contrib-dashboard' }
  if (role === 'admin') return { name: 'admin-dashboard' }
  return { name: 'login' }
}

export default router
