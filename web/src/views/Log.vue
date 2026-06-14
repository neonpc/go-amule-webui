<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { api } from '../lib/api'

const entries = ref<string[]>([])
const loading = ref(true)
const autoRefresh = ref(false)

let timer: ReturnType<typeof setTimeout> | null = null

const reversed = computed(() => [...entries.value].reverse())

onMounted(async () => {
  await fetchLog()
})

onUnmounted(() => {
  stopAutoRefresh()
})

async function fetchLog() {
  try { entries.value = await api.log() } catch {}
  loading.value = false
}

function onToggleRefresh(val: boolean) {
  if (val) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

function startAutoRefresh() {
  stopAutoRefresh()
  timer = setTimeout(async () => {
    await fetchLog()
    if (autoRefresh.value) startAutoRefresh()
  }, 5000)
}

function stopAutoRefresh() {
  if (timer) {
    clearTimeout(timer)
    timer = null
  }
}
</script>

<template>
  <div>
    <h1>Log</h1>
    <label class="refresh-toggle">
      <input type="checkbox" v-model="autoRefresh" @change="onToggleRefresh(autoRefresh)" />
      <span>Auto-refresh (5s)</span>
    </label>
    <p v-if="loading">Loading...</p>
    <pre v-else class="log-view">{{ reversed.join('\n') }}</pre>
  </div>
</template>

<style scoped>
.refresh-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  font-size: 0.85rem;
  color: var(--text-muted);
  cursor: pointer;
}

.refresh-toggle input {
  accent-color: var(--accent);
}

.log-view {
  background: var(--bg-card);
  padding: 16px;
  border-radius: 8px;
  font-size: 0.8rem;
  line-height: 1.4;
  overflow-x: auto;
  max-height: 70vh;
  margin-top: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
