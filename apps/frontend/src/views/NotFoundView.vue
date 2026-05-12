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
  background: #080c18;
}
.code { font-size: 5rem; font-weight: 900; color: #1e2a45; line-height: 1; }
.msg { color: #94a3b8; font-size: 1rem; margin: 0; }
.home-link { color: #4f8ef7; font-size: 0.9rem; }
</style>
