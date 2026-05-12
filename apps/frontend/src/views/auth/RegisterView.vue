<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthLayout from '@/layouts/AuthLayout.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const name = ref('')
const email = ref('')
const password = ref('')
const role = ref<'student' | 'contributor'>('student')
const isLoading = ref(false)
const errorMsg = ref('')
const isPending = ref(false)

async function handleSubmit() {
  errorMsg.value = ''
  isLoading.value = true
  try {
    const result = await auth.register({ name: name.value, email: email.value, password: password.value, role: role.value })
    if (result.isPending) {
      isPending.value = true
    } else {
      router.push({ name: 'student-dashboard' })
    }
  } catch (e) {
    const code = e instanceof Error ? e.message : ''
    if (code === 'EMAIL_CONFLICT') errorMsg.value = 'An account with this email already exists.'
    else errorMsg.value = 'Registration failed. Please try again.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <AuthLayout>
    <template v-if="isPending">
      <div class="pending-card">
        <div class="pending-icon">⏳</div>
        <h2>Application Submitted</h2>
        <p>Your contributor account is pending admin approval. You'll be able to log in once approved.</p>
        <RouterLink to="/login" class="btn-primary" style="display:block;text-align:center;margin-top:1.5rem;">
          Back to Login
        </RouterLink>
      </div>
    </template>

    <template v-else>
      <h1 class="page-title">Create account</h1>
      <p class="page-sub">Join thousands of students preparing for TKA &amp; SMBT</p>

      <form @submit.prevent="handleSubmit" class="form">
        <div class="field">
          <label>Full Name</label>
          <input v-model="name" type="text" autocomplete="name" placeholder="Budi Santoso" required minlength="2" maxlength="100" />
        </div>
        <div class="field">
          <label>Email</label>
          <input v-model="email" type="email" autocomplete="email" placeholder="you@example.com" required />
        </div>
        <div class="field">
          <label>Password</label>
          <input v-model="password" type="password" autocomplete="new-password" placeholder="min. 8 characters" required minlength="8" maxlength="72" />
        </div>
        <div class="field">
          <label>I am a…</label>
          <div class="role-group">
            <label class="role-option" :class="{ active: role === 'student' }">
              <input v-model="role" type="radio" value="student" />
              <span>Student</span>
              <span class="role-sub">Takes tests, tracks progress</span>
            </label>
            <label class="role-option" :class="{ active: role === 'contributor' }">
              <input v-model="role" type="radio" value="contributor" />
              <span>Contributor</span>
              <span class="role-sub">Authors questions &amp; tests</span>
            </label>
          </div>
        </div>

        <p v-if="errorMsg" class="error-msg">{{ errorMsg }}</p>

        <button type="submit" class="btn-primary" :disabled="isLoading">
          {{ isLoading ? 'Creating account…' : 'Create account' }}
        </button>
      </form>

      <p class="footer-link">
        Already have an account?
        <RouterLink to="/login">Sign in</RouterLink>
      </p>
    </template>
  </AuthLayout>
</template>

<style scoped>
.page-title { margin: 0 0 0.25rem; font-size: 1.5rem; font-weight: 700; }
.page-sub { margin: 0 0 1.75rem; color: #94a3b8; font-size: 0.875rem; }

.form { display: flex; flex-direction: column; gap: 1rem; }

.field { display: flex; flex-direction: column; gap: 0.375rem; }
.field > label { font-size: 0.8rem; font-weight: 600; color: #94a3b8; }
.field input[type="text"],
.field input[type="email"],
.field input[type="password"] {
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

.role-group { display: flex; gap: 0.75rem; }
.role-option {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 0.75rem;
  border-radius: 8px;
  border: 1px solid #1e2a45;
  background: #0d1424;
  cursor: pointer;
  transition: border-color 0.15s;
  font-size: 0.875rem;
  font-weight: 600;
  color: #94a3b8;
}
.role-option input { display: none; }
.role-option.active { border-color: #4f8ef7; color: #4f8ef7; }
.role-sub { font-size: 0.72rem; font-weight: 400; color: #64748b; margin-top: 2px; }

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
  text-decoration: none;
}
.btn-primary:hover { background: #3b7be8; }
.btn-primary:disabled { opacity: 0.6; cursor: not-allowed; }

.footer-link { margin: 1.5rem 0 0; text-align: center; font-size: 0.85rem; color: #94a3b8; }

.pending-card { text-align: center; }
.pending-icon { font-size: 2.5rem; margin-bottom: 1rem; }
.pending-card h2 { margin: 0 0 0.75rem; font-size: 1.25rem; }
.pending-card p { color: #94a3b8; font-size: 0.875rem; line-height: 1.6; margin: 0; }
</style>
