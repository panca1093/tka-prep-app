<script setup lang="ts">
import { ref, onMounted } from 'vue'
import client from '@/api/client'

const status = ref<string>('checking…')
const timestamp = ref<string>('')
const error = ref<string>('')

onMounted(async () => {
  const { data, error: err } = await client.GET('/health')
  if (data) {
    status.value = data.status
    timestamp.value = data.timestamp
  } else {
    status.value = 'error'
    error.value = err ? String(err) : 'unknown error'
  }
})
</script>

<template>
  <main class="container">
    <h1>✦ TKAPrep</h1>
    <p class="subtitle">Working skeleton — backend wired, frontend wired, ready to build.</p>

    <section class="card">
      <h2>Backend health</h2>
      <div class="row">
        <span class="label">Status</span>
        <span :class="['value', `value--${status}`]">{{ status }}</span>
      </div>
      <div v-if="timestamp" class="row">
        <span class="label">Server time</span>
        <span class="value">{{ timestamp }}</span>
      </div>
      <div v-if="error" class="error">{{ error }}</div>
    </section>
  </main>
</template>

<style scoped>
.container {
  max-width: 600px;
  margin: 4rem auto;
  padding: 0 1.5rem;
}
h1 {
  font-size: 2rem;
  font-weight: 800;
  color: #4f8ef7;
  margin: 0 0 0.5rem;
}
.subtitle {
  color: #94a3b8;
  margin: 0 0 2rem;
}
.card {
  background: #141c2e;
  border: 1px solid #1e2a45;
  border-radius: 12px;
  padding: 1.5rem;
}
.card h2 {
  font-size: 1rem;
  margin: 0 0 1rem;
  color: #cbd5e1;
}
.row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  border-bottom: 1px solid #1e2a45;
}
.row:last-child {
  border-bottom: none;
}
.label {
  color: #94a3b8;
  font-size: 0.875rem;
}
.value {
  font-weight: 600;
  color: #f1f5f9;
}
.value--ok { color: #22c55e; }
.value--error { color: #ef4444; }
.value--checking\u2026 { color: #f59e0b; }
.error {
  margin-top: 1rem;
  padding: 0.75rem;
  background: #450a0a;
  border-radius: 6px;
  color: #fca5a5;
  font-size: 0.875rem;
}
</style>
