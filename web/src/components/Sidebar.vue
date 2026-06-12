<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

const navItems = [
  { path: '/', label: 'Dashboard', icon: '📊' },
  { path: '/downloads', label: 'Downloads', icon: '⬇' },
  { path: '/uploads', label: 'Uploads', icon: '⬆' },
  { path: '/shared', label: 'Shared', icon: '📁' },
  { path: '/search', label: 'Search', icon: '🔍' },
  { path: '/servers', label: 'Servers', icon: '🖥' },
  { path: '/kad', label: 'Kad', icon: '🌐' },
  { path: '/stats', label: 'Statistics', icon: '📈' },
  { path: '/log', label: 'Log', icon: '📝' },
]
</script>

<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <h2>aMule</h2>
    </div>
    <nav class="sidebar-nav">
      <a
        v-for="item in navItems"
        :key="item.path"
        :class="{ active: route.path === item.path }"
        @click.prevent="router.push(item.path)"
        href="#"
      >
        <span class="nav-icon">{{ item.icon }}</span>
        {{ item.label }}
      </a>
    </nav>
  </aside>
</template>

<style scoped>
.sidebar {
  width: 220px;
  background: var(--bg-card);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid var(--border);
}

.sidebar-header h2 {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--accent);
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  padding: 8px;
  gap: 2px;
}

.sidebar-nav a {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  color: var(--text-muted);
  font-size: 0.9rem;
  transition: all 0.15s;
  cursor: pointer;
}

.sidebar-nav a:hover {
  background: var(--bg-hover);
  color: var(--text);
}

.sidebar-nav a.active {
  background: var(--accent);
  color: white;
}

.nav-icon {
  font-size: 1.1rem;
  width: 20px;
  text-align: center;
}
</style>
