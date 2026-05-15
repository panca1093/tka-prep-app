<template>
  <div class="not-found">
    <div class="code">404</div>
    <p class="msg">Page not found.</p>
    <RouterLink to="/" class="home-link">Go home →</RouterLink>
  </div>
</template>

<style scoped>
.not-found {
  min-height: 100vh; display: flex; flex-direction: column;
  align-items: center; justify-content: center; gap: 0.75rem;
  background: var(--bg-app);
}
.code { font-size: 5rem; font-weight: 900; color: var(--border); line-height: 1; }
.msg { color: var(--text-muted); font-size: 1rem; margin: 0; }
.home-link { color: var(--accent); font-size: 0.9rem; }
</style>
