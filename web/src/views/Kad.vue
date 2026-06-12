<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../lib/api'

const connected = ref(false)
const firewalled = ref(false)
const loading = ref(true)

onMounted(async () => {
  try {
    const k = await api.kad()
    connected.value = k.connected
    firewalled.value = k.firewalled
  } catch {}
  loading.value = false
})

async function toggle() {
  const action = connected.value ? 'stop' : 'start'
  try {
    await api.kadAction(action)
    connected.value = !connected.value
  } catch {}
}
</script>

<template>
  <div>
    <h1>Kad Network</h1>
    <p v-if="loading">Loading...</p>
    <div v-else>
      <p>Status: {{ connected ? 'Connected' : 'Disconnected' }}</p>
      <p v-if="firewalled">Firewalled</p>
      <button @click="toggle">{{ connected ? 'Disconnect' : 'Connect' }}</button>
    </div>
  </div>
</template>

<style scoped>
button { margin-top: 12px; padding: 10px 20px; background: var(--accent); color: white; border: none; border-radius: 8px; cursor: pointer; }
</style>
