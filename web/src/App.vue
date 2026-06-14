<script setup lang="ts">
import { useRoute } from 'vue-router'
import Sidebar from './components/Sidebar.vue'

const route = useRoute()
const isLogin = () => route.name === 'login'
</script>

<template>
  <div class="app-layout">
    <Sidebar v-if="!isLogin()" />
    <main class="main-content" :class="{ 'login-mode': isLogin() }">
      <router-view />
    </main>
  </div>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

:root {
  --bg: #0f172a;
  --bg-card: #1e293b;
  --bg-hover: #334155;
  --text: #e2e8f0;
  --text-muted: #94a3b8;
  --accent: #3b82f6;
  --accent-hover: #2563eb;
  --success: #22c55e;
  --warning: #eab308;
  --danger: #ef4444;
  --border: #334155;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: var(--bg);
  color: var(--text);
  line-height: 1.5;
}

.app-layout {
  display: flex;
  min-height: 100vh;
}

.main-content {
  flex: 1;
  padding: 24px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.main-content.login-mode {
  padding: 0;
}

@media (max-width: 768px) {
  .main-content {
    padding: 74px 16px 16px 16px;
  }
  .main-content.login-mode {
    padding: 0;
  }
}

a {
  color: var(--accent);
  text-decoration: none;
}

a:hover {
  color: var(--accent-hover);
}

@media (max-width: 768px) {
  button, .btn-action, .btn-connect, .btn-remove, .btn-dl {
    min-height: 38px;
  }
}
</style>
