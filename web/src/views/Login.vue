<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api, setToken } from '../lib/api'

const router = useRouter()
const password = ref('')
const error = ref('')
const loading = ref(false)

async function doLogin() {
  if (!password.value) return
  loading.value = true
  error.value = ''
  try {
    const res = await api.login(password.value)
    setToken(res.token)
    router.push('/')
  } catch (e: any) {
    error.value = e.message || 'Login failed'
  }
  loading.value = false
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1>aMule Web UI</h1>
      <p class="login-sub">Enter your password to continue</p>
      <form @submit.prevent="doLogin">
        <input
          v-model="password"
          type="password"
          placeholder="Password"
          autofocus
          :disabled="loading"
        />
        <p v-if="error" class="login-error">{{ error }}</p>
        <button type="submit" :disabled="loading">
          {{ loading ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--bg);
}

.login-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 40px;
  width: 100%;
  max-width: 360px;
  text-align: center;
}

.login-card h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--accent);
  margin-bottom: 4px;
}

.login-sub {
  color: var(--text-muted);
  font-size: 0.85rem;
  margin-bottom: 24px;
}

form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

input {
  padding: 12px 14px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bg);
  color: var(--text);
  font-size: 1rem;
  outline: none;
}

input:focus {
  border-color: var(--accent);
}

button {
  padding: 12px;
  background: var(--accent);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
}

button:disabled {
  opacity: 0.5;
}

.login-error {
  color: #ef4444;
  font-size: 0.85rem;
}
</style>
