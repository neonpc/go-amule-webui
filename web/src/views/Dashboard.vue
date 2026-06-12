<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAmuleStore } from '../stores/amule'
import { useSocket } from '../composables/useSocket'
import SpeedCard from '../components/SpeedCard.vue'
import StatusBar from '../components/StatusBar.vue'
import { api, Status } from '../lib/api'

const store = useAmuleStore()
const { connected } = useSocket()
const status = ref<Status | null>(null)

onMounted(async () => {
  try {
    status.value = await api.status()
  } catch {}
})
</script>

<template>
  <div class="dashboard">
    <h1>Dashboard</h1>
    <StatusBar :status="status" />
    <div style="display: flex; gap: 16px; flex-wrap: wrap; margin-top: 16px;">
      <SpeedCard label="Download" :speed="status?.dl_speed ?? 0" color="#22c55e" />
      <SpeedCard label="Upload" :speed="status?.ul_speed ?? 0" color="#3b82f6" />
    </div>
    <div class="ws-indicator">
      WebSocket: {{ connected ? '🟢 Connected' : '🔴 Disconnected' }}
    </div>
  </div>
</template>

<style scoped>
.ws-indicator {
  margin-top: 16px;
  font-size: 0.85rem;
  color: var(--text-muted);
}
</style>
