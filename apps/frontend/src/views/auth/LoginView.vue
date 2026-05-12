<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AuthLayout from '@/layouts/AuthLayout.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const isLoading = ref(false)
const errorMsg = ref('')

async function handleSubmit() {
  errorMsg.value = ''
  isLoading.value = true
  try {
    const user = await auth.login(email.value, password.value)
    const redirect = route.query.redirect as string | undefined
    if (redirect) { router.push(redirect); return }
    if (user.role === 'student') router.push({ name: 'student-dashboard' })
    else if (user.role === 'contributor') router.push({ name: 'contrib-dashboard' })
    else if (user.role === 'admin') router.push({ name: 'admin-dashboard' })
  } catch (e) {
    const code = e instanceof Error ? e.message : ''
    if (code === 'PENDING_APPROVAL') errorMsg.value = 'Your contributor account is pending admin approval.'
    else if (code === 'SUSPENDED') errorMsg.value = 'Your account has been suspended. Contact support.'
    else errorMsg.value = 'Invalid email or password.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <AuthLayout>
    <h1 class="page-title">Welcome back</h1>
    <p class="page-sub">Sign in to your account</p>

    <form @submit.prevent="handleSubmit" class="form">
      <div class="field">
        <label>Email</label>
        <input v-model="email" type="email" autocomplete="email" placeholder="you@example.com" required />
      </div>
      <div class="field">
        <label>Password</label>
        <input v-model="password" type="password" autocomplete="current-password" placeholder="••••••••" required />
      </div>

      <p v-if="errorMsg" class="error-msg">{{ errorMsg }}</p>

      <button type="submit" class="btn-primary" :disabled="isLoading">
        {{ isLoading ? 'Signing in…' : 'Sign in' }}
      </button>
    </form>

    <p class="footer-link">
      Don't have an account?
      <RouterLink to="/register">Register</RouterLink>
    </p>
  </AuthLayout>
</template>

<style scoped>
.page-title { margin: 0 0 0.25rem; font-size: 1.5rem; font-weight: 700; }
.page-sub { margin: 0 0 1.75rem; color: #94a3b8; font-size: 0.875rem; }

.form { display: flex; flex-direction: column; gap: 1rem; }

.field { display: flex; flex-direction: column; gap: 0.375rem; }
.field label { font-size: 0.8rem; font-weight: 600; color: #94a3b8; }
.field input {
  padding: 0.65rem 0.875rem;
  border-radius: 8px;
  border: 1px solid #1e2a45;
  background: #0d1424;
  color: #f1f5f9;
  font-size: 0.9rem;
  outline: none;
  transition: border-color 0.15s;
}
.field input:focus { border-color: #4f8ef7; }

.error-msg {
  margin: 0;
  padding: 0.6rem 0.75rem;
  border-radius: 8px;
  background: #450a0a;
  color: #fca5a5;
  font-size: 0.825rem;
}

.btn-primary {
  margin-top: 0.5rem;
  padding: 0.7rem;
  border-radius: 8px;
  border: none;
  background: #4f8ef7;
  color: #fff;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, opacity 0.15s;
}
.btn-primary:hover { background: #3b7be8; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.footer-link { margin: 1.5rem 0 0; text-align: center; font-size: 0.85rem; color: #94a3b8; }
</style>
