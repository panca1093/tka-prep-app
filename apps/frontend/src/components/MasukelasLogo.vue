<script setup lang="ts">
withDefaults(defineProps<{
  size?: 'sm' | 'md' | 'lg'
  showText?: boolean
}>(), {
  size: 'md',
  showText: true,
})
</script>

<template>
  <div class="logo" :class="`logo--${size}`">
    <svg
      class="logo-mark"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="1.8"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path d="M12 2L2 7l10 5 10-5-10-5z" />
      <path d="M2 17l10 5 10-5" />
      <path d="M2 12l10 5 10-5" />
    </svg>
    <span v-if="showText" class="logo-text">Masukelas</span>
  </div>
</template>

<style scoped>
.logo {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.logo-mark {
  color: var(--accent);
  flex-shrink: 0;
}

.logo--sm .logo-mark {
  width: 20px;
  height: 20px;
}

.logo--sm .logo-text {
  font-size: 0.9rem;
}

.logo--md .logo-mark {
  width: 24px;
  height: 24px;
}

.logo--md .logo-text {
  font-size: 1rem;
}

.logo--lg .logo-mark {
  width: 32px;
  height: 32px;
}

.logo--lg .logo-text {
  font-size: 1.25rem;
}

.logo-text {
  font-weight: 800;
  color: var(--text-primary);
  white-space: nowrap;
}
</style>
