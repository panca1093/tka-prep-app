<template>
  <div class="auth-shell">
    <div class="auth-card">
      <div class="auth-brand">
        <span class="brand-mark">✦</span>
        <span class="brand-name">TKAPrep</span>
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
  background: radial-gradient(ellipse at 50% 0%, #0d1a35 0%, #080c18 60%);
}

.auth-card {
  width: 100%;
  max-width: 420px;
  background: #141c2e;
  border: 1px solid #1e2a45;
  border-radius: 16px;
  padding: 2.5rem 2rem;
}

.auth-brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 2rem;
}

.brand-mark {
  font-size: 1.4rem;
  color: #4f8ef7;
}

.brand-name {
  font-size: 1.25rem;
  font-weight: 800;
  color: #f1f5f9;
}
</style>
