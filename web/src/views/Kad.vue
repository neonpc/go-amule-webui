<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '../lib/api'

const connected = ref(false)
const firewalled = ref(false)
const loading = ref(true)
const busy = ref(false)
const err = ref('')

onMounted(async () => {
  try {
    const k = await api.kad()
    connected.value = k.connected
    firewalled.value = k.firewalled
  } catch (e: any) {
    err.value = e.message
  }
  loading.value = false
})

async function toggle() {
  const action = connected.value ? 'stop' : 'start'
  busy.value = true
  err.value = ''
  try {
    await api.kadAction(action)
    await new Promise(r => setTimeout(r, 2000))
    const k = await api.kad()
    connected.value = k.connected
    firewalled.value = k.firewalled
    if (action === 'start' && !k.connected) {
      err.value = 'Kad did not connect. It may take longer or aMule might need configuration.'
    }
  } catch (e: any) {
    err.value = e.message
  }
  busy.value = false
}
</script>

<template>
  <div>
    <h1>Kad Network</h1>
    <p v-if="loading">Loading...</p>
    <div v-else>
      <p v-if="err" class="err">{{ err }}</p>
      <p>Status: {{ connected ? 'Connected' : 'Disconnected' }}</p>
      <p v-if="firewalled">Firewalled</p>
      <button @click="toggle" :disabled="busy">{{ busy ? '...' : (connected ? 'Disconnect' : 'Connect') }}</button>
    </div>
  </div>
</template>

<style scoped>
.err { color: #ef4444; font-size: 0.85rem; margin-bottom: 8px; }
button { margin-top: 12px; padding: 10px 20px; background: var(--accent); color: white; border: none; border-radius: 8px; cursor: pointer; }
button:disabled { opacity: 0.5; }
</style>
