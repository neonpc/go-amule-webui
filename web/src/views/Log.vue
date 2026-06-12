<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../lib/api'

const entries = ref<string[]>([])
const loading = ref(true)

onMounted(async () => {
  try { entries.value = await api.log() } catch {}
  loading.value = false
})
</script>

<template>
  <div>
    <h1>Log</h1>
    <p v-if="loading">Loading...</p>
    <pre v-else class="log-view">{{ entries.join('\n') }}</pre>
  </div>
</template>

<style scoped>
.log-view { background: var(--bg-card); padding: 16px; border-radius: 8px; font-size: 0.8rem; line-height: 1.4; overflow-x: auto; max-height: 70vh; margin-top: 12px; white-space: pre-wrap; word-break: break-all; }
</style>
