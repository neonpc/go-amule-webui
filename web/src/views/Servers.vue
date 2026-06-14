<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, ServerEntry as SE } from '../lib/api'

const servers = ref<SE[]>([])
const loading = ref(true)
const newAddress = ref('')
const newName = ref('')
const adding = ref(false)
const err = ref('')

onMounted(async () => {
  try { servers.value = await api.servers() } catch {}
  loading.value = false
})

async function connectServer(address: string) {
  try {
    await api.serverConnect(address)
  } catch (e: any) {
    err.value = e.message
  }
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
    <div class="add-box">
      <input v-model="newAddress" placeholder="host:port" @keyup.enter="addServer" />
      <input v-model="newName" placeholder="name (optional)" @keyup.enter="addServer" />
      <button @click="addServer" :disabled="adding">Add</button>
    </div>
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
              <button class="btn-connect" @click="connectServer(s.address)">Connect</button>
              <button class="btn-remove" @click="removeServer(s.address)">Remove</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.add-box { display: flex; gap: 8px; margin: 12px 0; }
.add-box input { flex: 1; padding: 10px 14px; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-card); color: var(--text); }
.add-box button { padding: 10px 20px; background: var(--accent); color: white; border: none; border-radius: 8px; cursor: pointer; }
.add-box button:disabled { opacity: 0.5; }
.err { color: var(--err, #ef4444); font-size: 0.85rem; }
.btn-connect { padding: 4px 10px; background: var(--accent); color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 0.8rem; margin-right: 4px; }
.btn-remove { padding: 4px 10px; background: #ef4444; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 0.8rem; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; margin-top: 12px; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--border); font-size: 0.9rem; }
th { color: var(--text-muted); font-weight: 600; }
</style>
