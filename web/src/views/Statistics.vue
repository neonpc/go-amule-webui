<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../lib/api'

const stats = ref<Record<string, number>>({})
const loading = ref(true)

onMounted(async () => {
  try { stats.value = await api.stats() } catch {}
  loading.value = false
})
</script>

<template>
  <div>
    <h1>Statistics</h1>
    <p v-if="loading">Loading...</p>
    <div v-else class="stats-grid">
      <div v-for="(v, k) in stats" :key="k" class="stat-item">
        <span class="stat-key">{{ k }}</span>
        <span class="stat-val">{{ v }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.stats-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 12px; margin-top: 12px; }
.stat-item { background: var(--bg-card); padding: 16px; border-radius: 8px; display: flex; flex-direction: column; gap: 4px; }
.stat-key { font-size: 0.8rem; color: var(--text-muted); text-transform: uppercase; }
.stat-val { font-size: 1.25rem; font-weight: 700; }
</style>
