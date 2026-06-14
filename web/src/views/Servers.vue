<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, ServerEntry as SE } from '../lib/api'

const servers = ref<SE[]>([])
const loading = ref(true)
const newAddress = ref('')
const newName = ref('')
const adding = ref(false)
const err = ref('')
const msg = ref('')
const connecting = ref('')

const status = ref<{ ed2k_server: string; ed2k_connected: boolean; kad_connected: boolean } | null>(null)

onMounted(async () => {
  try { servers.value = await api.servers() } catch {}
  try { status.value = await api.status() } catch {}
  loading.value = false
})

async function refreshStatus() {
  try { status.value = await api.status() } catch {}
}

async function connectServer(address: string) {
  if (connecting.value) return
  connecting.value = address
  err.value = ''
  msg.value = ''
  try {
    await api.serverConnect(address)
    msg.value = 'Connecting...'
    await new Promise(r => setTimeout(r, 2000))
    await refreshStatus()
    msg.value = ''
  } catch (e: any) {
    err.value = `Failed to connect to ${address}: ${e.message}`
  }
  connecting.value = ''
}

async function addServer() {
  if (!newAddress.value.trim()) return
  adding.value = true
  err.value = ''
  try {
    await api.serverAdd(newAddress.value.trim(), newName.value.trim())
    newAddress.value = ''
    newName.value = ''
    servers.value = await api.servers()
  } catch (e: any) {
    err.value = e.message
  }
  adding.value = false
}

async function removeServer(address: string) {
  try {
    await api.serverRemove(address)
    servers.value = await api.servers()
  } catch (e: any) {
    err.value = e.message
  }
}
</script>

<template>
  <div>
    <h1>Servers</h1>
    <div v-if="status" class="conn-info">
      <span v-if="status.ed2k_connected" class="tag connected">Connected to {{ status.ed2k_server }}</span>
      <span v-else class="tag disconnected">Disconnected</span>
      <span v-if="status.kad_connected" class="tag connected">Kad Active</span>
    </div>
    <div class="add-box">
      <input v-model="newAddress" placeholder="host:port" @keyup.enter="addServer" />
      <input v-model="newName" placeholder="name (optional)" @keyup.enter="addServer" />
      <button @click="addServer" :disabled="adding">Add</button>
    </div>
    <p v-if="msg" class="msg">{{ msg }}</p>
    <p v-if="err" class="err">{{ err }}</p>
    <p v-if="loading">Loading...</p>
    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr><th>Name</th><th>Address</th><th>Users</th><th>Files</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="s in servers" :key="s.name + s.address">
            <td>{{ s.name }}</td><td>{{ s.address }}</td><td>{{ s.users }}</td><td>{{ s.files }}</td>
            <td>
              <button class="btn-connect" @click="connectServer(s.address)" :disabled="connecting === s.address">{{ connecting === s.address ? 'Connecting...' : 'Connect' }}</button>
              <button class="btn-remove" @click="removeServer(s.address)">Remove</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.conn-info { display: flex; gap: 12px; margin: 12px 0; align-items: center; }
.tag { padding: 4px 12px; border-radius: 12px; font-size: 0.85rem; font-weight: 600; }
.connected { background: #16a34a22; color: #16a34a; border: 1px solid #16a34a44; }
.disconnected { background: #ef444422; color: #ef4444; border: 1px solid #ef444444; }
.add-box { display: flex; gap: 8px; margin: 12px 0; }
.add-box input { flex: 1; padding: 10px 14px; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-card); color: var(--text); min-width: 0; }
.add-box button { padding: 10px 20px; background: var(--accent); color: white; border: none; border-radius: 8px; cursor: pointer; }

@media (max-width: 768px) {
  .add-box { flex-wrap: wrap; }
  .add-box input { flex: 1 1 100%; }
}
.add-box button:disabled { opacity: 0.5; }
.err { color: var(--err, #ef4444); font-size: 0.85rem; }
.msg { color: var(--accent, #3b82f6); font-size: 0.85rem; }
.btn-connect { padding: 4px 10px; background: var(--accent); color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 0.8rem; margin-right: 4px; }
.btn-remove { padding: 4px 10px; background: #ef4444; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 0.8rem; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; margin-top: 12px; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--border); font-size: 0.9rem; }
th { color: var(--text-muted); font-weight: 600; }
</style>
