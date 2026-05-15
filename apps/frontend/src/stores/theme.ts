import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export type ThemePreference = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'masukelas-theme'

export const useThemeStore = defineStore('theme', () => {
  const preference = ref<ThemePreference>('system')

  // dark mode media query not supported → fallback to light
  let mq: MediaQueryList | undefined
  try {
    mq = window.matchMedia('(prefers-color-scheme: dark)')
  } catch {
    // noop — effectiveTheme will always be 'light'
  }

  const effectiveTheme = computed<'light' | 'dark'>(() => {
    if (preference.value === 'system') {
      return mq?.matches ? 'dark' : 'light'
    }
    return preference.value
  })

  let _mqHandler: (() => void) | null = null

  function _applyTheme() {
    document.documentElement.dataset.theme = effectiveTheme.value
  }

  function _watchSystem() {
    if (_mqHandler) {
      mq?.removeEventListener('change', _mqHandler)
    }
    if (preference.value === 'system' && mq) {
      _mqHandler = () => _applyTheme()
      mq.addEventListener('change', _mqHandler)
    }
  }

  function initTheme() {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark' || stored === 'system') {
      preference.value = stored
    }
    _applyTheme()
    _watchSystem()
  }

  function setTheme(pref: ThemePreference) {
    preference.value = pref
    localStorage.setItem(STORAGE_KEY, pref)
    _applyTheme()
    _watchSystem()
  }

  return { preference, effectiveTheme, initTheme, setTheme }
})
