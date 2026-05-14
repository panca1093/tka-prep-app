import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import client, { setAccessToken, clearAccessToken } from '@/api/client'
import type { components } from '@tkaprep/shared-types'

type User = components['schemas']['UserResponse']

const RT_KEY = 'tkaprep_rt'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const isInitialized = ref(false)
  let _refreshTimer: ReturnType<typeof setTimeout> | null = null

  const isAuthenticated = computed(() => user.value !== null)
  const role = computed(() => user.value?.role ?? null)

  function _setSession(at: string, rt: string, u: User) {
    setAccessToken(at)
    user.value = u
    localStorage.setItem(RT_KEY, rt)
    _scheduleRefresh()
  }

  function _scheduleRefresh() {
    if (_refreshTimer) clearTimeout(_refreshTimer)
    // access token TTL is 15 min; refresh 1 min before expiry
    _refreshTimer = setTimeout(() => refresh().catch(() => clearSession()), 14 * 60 * 1000)
  }

  function clearSession() {
    clearAccessToken()
    user.value = null
    localStorage.removeItem(RT_KEY)
    if (_refreshTimer) clearTimeout(_refreshTimer)
  }

  async function init() {
    const rt = localStorage.getItem(RT_KEY)
    if (rt) {
      try { await _refresh(rt) } catch { clearSession() }
    }
    isInitialized.value = true
  }

  async function login(email: string, password: string): Promise<User> {
    const { data, error } = await client.POST('/auth/login', { body: { email, password } })
    if (error) {
      const code = (error as Record<string, any>)?.error?.code as string | undefined
      throw new Error(code ?? 'LOGIN_FAILED')
    }
    _setSession(data.access_token, data.refresh_token, data.user)
    return data.user
  }

  async function register(payload: {
    name: string
    email: string
    password: string
    role: 'student' | 'contributor'
    education_level?: 'sd' | 'smp' | 'sma' | 'smk'
  }): Promise<{ user: User; isPending: boolean }> {
    const { data, error } = await client.POST('/auth/register', { body: payload })
    if (error) {
      const code = (error as Record<string, any>)?.error?.code as string | undefined
      throw new Error(code ?? 'REGISTER_FAILED')
    }
    const isPending = !data.access_token
    if (!isPending) {
      _setSession(data.access_token, data.refresh_token, data.user)
    }
    return { user: data.user, isPending }
  }

  async function refresh() {
    const rt = localStorage.getItem(RT_KEY)
    if (!rt) throw new Error('NO_REFRESH_TOKEN')
    await _refresh(rt)
  }

  async function _refresh(rt: string) {
    const { data, error } = await client.POST('/auth/refresh', { body: { refresh_token: rt } })
    if (error) { clearSession(); throw error }
    _setSession(data.access_token, data.refresh_token, data.user)
  }

  async function logout() {
    const rt = localStorage.getItem(RT_KEY)
    clearSession()
    if (rt) {
      await client.POST('/auth/logout', { body: { refresh_token: rt } }).catch(() => {})
    }
  }

  return { user, isInitialized, isAuthenticated, role, init, login, register, logout, clearSession }
})
