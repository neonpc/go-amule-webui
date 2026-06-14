<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const open = ref(false)

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

const currentLabel = computed(() => {
  const item = navItems.find(i => i.path === route.path)
  return item ? item.label : ''
})

function navigate(path: string) {
  router.push(path)
  open.value = false
}
</script>

<template>
  <header class="mobile-bar">
    <button class="hamburger" @click="open = !open" aria-label="Toggle menu">
      <span class="hamburger-line" :class="{ open }" />
      <span class="hamburger-line" :class="{ open }" />
      <span class="hamburger-line" :class="{ open }" />
    </button>
    <span class="mobile-title">{{ currentLabel }}</span>
  </header>

  <div v-if="open" class="overlay" @click="open = false" />

  <aside class="sidebar" :class="{ open }">
    <div class="sidebar-header">
      <h2>aMule</h2>
    </div>
    <nav class="sidebar-nav">
      <a
        v-for="item in navItems"
        :key="item.path"
        :class="{ active: route.path === item.path }"
        @click.prevent="navigate(item.path)"
        href="#"
      >
        <span class="nav-icon">{{ item.icon }}</span>
        {{ item.label }}
      </a>
    </nav>
  </aside>
</template>

<style scoped>
.hamburger {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  flex-direction: column;
  gap: 4px;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  padding: 8px;
  border-radius: 8px;
}

.hamburger:hover {
  background: var(--bg-hover);
}

.hamburger-line {
  display: block;
  width: 20px;
  height: 2px;
  background: var(--text);
  border-radius: 2px;
  transition: all 0.2s;
}

.hamburger-line.open:nth-child(1) {
  transform: translateY(6px) rotate(45deg);
}
.hamburger-line.open:nth-child(2) {
  opacity: 0;
}
.hamburger-line.open:nth-child(3) {
  transform: translateY(-6px) rotate(-45deg);
}

.mobile-bar {
  display: none;
}

.overlay {
  display: none;
}

.sidebar {
  width: 220px;
  background: var(--bg-card);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  transition: transform 0.25s ease;
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

@media (max-width: 768px) {
  .hamburger {
    display: flex;
  }

  .mobile-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: 50px;
    background: var(--bg-card);
    border-bottom: 1px solid var(--border);
    z-index: 200;
    padding: 0 8px;
  }

  .mobile-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--text);
  }

  .overlay {
    display: block;
    position: fixed;
    inset: 0;
    top: 50px;
    background: rgba(0,0,0,0.5);
    z-index: 99;
  }

  .sidebar {
    position: fixed;
    top: 50px;
    left: 0;
    height: calc(100vh - 50px);
    z-index: 100;
    transform: translateX(-100%);
    width: 280px;
    box-shadow: 4px 0 24px rgba(0,0,0,0.4);
  }

  .sidebar.open {
    transform: translateX(0);
  }

  .sidebar-header {
    padding: 14px 16px;
  }

  .sidebar-header h2 {
    font-size: 1.1rem;
  }

  .sidebar-nav a {
    padding: 14px 16px;
    font-size: 1rem;
    gap: 12px;
  }
}
</style>
