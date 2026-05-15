<script setup lang="ts">
import MasukelasLogo from '@/components/MasukelasLogo.vue'
</script>

<template>
  <div class="auth-shell">
    <div class="auth-card">
      <div class="auth-brand">
        <MasukelasLogo size="lg" />
      </div>
      <slot />
    </div>
  </div>
</template>

<style scoped>
.auth-shell {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background: radial-gradient(ellipse at 50% 0%, var(--bg-sidebar) 0%, var(--bg-app) 60%);
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 2.5rem 2rem;
}

.auth-brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 2rem;
}

@media (max-width: 480px) {
  .auth-shell {
    padding: 0.75rem;
    align-items: flex-start;
    padding-top: 2rem;
  }
  .auth-card {
    padding: 1.5rem 1rem;
  }
}
</style>
