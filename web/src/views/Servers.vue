<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, ServerEntry as SE } from '../lib/api'

const servers = ref<SE[]>([])
const loading = ref(true)

onMounted(async () => {
  try { servers.value = await api.servers() } catch {}
  loading.value = false
})
</script>

<template>
  <div>
    <h1>Servers</h1>
    <p v-if="loading">Loading...</p>
    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr><th>Name</th><th>Address</th><th>Users</th><th>Files</th></tr>
        </thead>
        <tbody>
          <tr v-for="s in servers" :key="s.name + s.address">
            <td>{{ s.name }}</td><td>{{ s.address }}</td><td>{{ s.users }}</td><td>{{ s.files }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; margin-top: 12px; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--border); font-size: 0.9rem; }
th { color: var(--text-muted); font-weight: 600; }
</style>
