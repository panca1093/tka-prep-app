import { ref, type Ref } from 'vue'

export interface AntiCheatState {
  fullscreenWarning: Ref<boolean>
  cleanup: () => void
}

const BLOCKED_KEYS = new Set([
  'PrintScreen',
  'F12',
])

const BLOCKED_COMBOS = [
  // Save
  { ctrl: true, key: 's' },
  { meta: true, key: 's' },
  // Print
  { ctrl: true, key: 'p' },
  { meta: true, key: 'p' },
  // DevTools
  { ctrl: true, shift: true, key: 'i' },
  { meta: true, shift: true, key: 'i' },
  { ctrl: true, shift: true, key: 'c' },
  { meta: true, shift: true, key: 'c' },
  { ctrl: true, shift: true, key: 'j' },
  { meta: true, shift: true, key: 'j' },
  { ctrl: true, key: 'u' },
  { meta: true, key: 'u' },
  // Copy / Cut / Select All
  { ctrl: true, key: 'c' },
  { meta: true, key: 'c' },
  { ctrl: true, key: 'x' },
  { meta: true, key: 'x' },
  { ctrl: true, key: 'a' },
  { meta: true, key: 'a' },
]

function matches(wanted: { ctrl?: boolean; shift?: boolean; key: string; meta?: boolean }, e: KeyboardEvent): boolean {
  if (e.key.toLowerCase() !== wanted.key) return false
  if (wanted.ctrl !== undefined && e.ctrlKey !== wanted.ctrl) return false
  if (wanted.meta !== undefined && e.metaKey !== wanted.meta) return false
  if (wanted.shift !== undefined && e.shiftKey !== wanted.shift) return false
  return true
}

export function useAntiCheat(): AntiCheatState {
  const fullscreenWarning = ref(false)
  let fullscreenReentryTimer: ReturnType<typeof setTimeout> | null = null

  // ── Fullscreen ──────────────────────────────────────────────────────────

  function requestFullscreen() {
    if (!document.fullscreenEnabled) return
    try {
      document.documentElement.requestFullscreen().catch(() => {
        // Browser may reject without user gesture — acceptable.
      })
    } catch {
      // Safari can throw on requestFullscreen
    }
  }

  function onFullscreenChange() {
    if (document.fullscreenElement) {
      fullscreenWarning.value = false
      if (fullscreenReentryTimer) {
        clearTimeout(fullscreenReentryTimer)
        fullscreenReentryTimer = null
      }
    } else {
      fullscreenWarning.value = true
      fullscreenReentryTimer = setTimeout(() => {
        fullscreenReentryTimer = null
        requestFullscreen()
      }, 3000)
    }
  }

  // ── Context menu ────────────────────────────────────────────────────────

  function onContextMenu(e: Event) {
    e.preventDefault()
  }

  // ── Keyboard shortcuts ──────────────────────────────────────────────────

  function onKeyDown(e: KeyboardEvent) {
    if (BLOCKED_KEYS.has(e.key)) {
      e.preventDefault()
      return
    }
    for (const combo of BLOCKED_COMBOS) {
      if (matches(combo, e)) {
        e.preventDefault()
        return
      }
    }
  }

  // ── Copy event — redundant with key block but catches Edit menu copy ────

  function onCopy(e: ClipboardEvent) {
    e.preventDefault()
  }

  // ── Mount / cleanup ─────────────────────────────────────────────────────

  requestFullscreen()
  document.addEventListener('fullscreenchange', onFullscreenChange)
  document.addEventListener('contextmenu', onContextMenu)
  document.addEventListener('keydown', onKeyDown)
  document.addEventListener('copy', onCopy)

  function cleanup() {
    document.removeEventListener('fullscreenchange', onFullscreenChange)
    document.removeEventListener('contextmenu', onContextMenu)
    document.removeEventListener('keydown', onKeyDown)
    document.removeEventListener('copy', onCopy)
    if (fullscreenReentryTimer) clearTimeout(fullscreenReentryTimer)
    // Exit fullscreen when leaving the session so the student isn't stuck.
    if (document.fullscreenElement) {
      try { document.exitFullscreen() } catch { /* ignore */ }
    }
  }

  return { fullscreenWarning, cleanup }
}
