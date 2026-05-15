<script setup lang="ts">
import { useThemeStore, type ThemePreference } from '@/stores/theme'

const theme = useThemeStore()

const options: { pref: ThemePreference; label: string; icon: string }[] = [
  { pref: 'light', label: 'Light', icon: '☀' },
  { pref: 'dark', label: 'Dark', icon: '☾' },
  { pref: 'system', label: 'System', icon: '◑' },
]
</script>

<template>
  <div class="theme-toggle" role="group" aria-label="Theme">
    <button
      v-for="opt in options"
      :key="opt.pref"
      class="toggle-btn"
      :class="{ active: theme.preference === opt.pref }"
      :aria-label="opt.label"
      :title="opt.label"
      @click="theme.setTheme(opt.pref)"
    >
      {{ opt.icon }}
    </button>
  </div>
</template>

<style scoped>
.theme-toggle {
  display: flex;
  gap: 4px;
  padding: 4px;
  border-radius: 8px;
  background: var(--bg-input);
  border: 1px solid var(--border);
}

.toggle-btn {
  flex: 1;
  padding: 0.4rem 0;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-muted);
  font-size: 0.9rem;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  text-align: center;
}

.toggle-btn:hover {
  color: var(--text-primary);
}

.toggle-btn.active {
  background: var(--accent);
  color: var(--text-on-accent);
}
</style>
