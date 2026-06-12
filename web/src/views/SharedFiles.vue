<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, SharedFile as SF } from '../lib/api'

const files = ref<SF[]>([])
const loading = ref(true)

onMounted(async () => {
  try { files.value = await api.shared() } catch {}
  loading.value = false
})

function fmtSize(bytes: number): string {
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1073741824) return `${(bytes / 1048576).toFixed(1)} MB`
  return `${(bytes / 1073741824).toFixed(2)} GB`
}
</script>

<template>
  <div>
    <h1>Shared Files</h1>
    <p v-if="loading">Loading...</p>
    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr><th>Name</th><th>Size</th><th>Requests</th><th>Transfers</th></tr>
        </thead>
        <tbody>
          <tr v-for="f in files" :key="f.hash">
            <td>{{ f.name }}</td><td>{{ fmtSize(f.size) }}</td><td>{{ f.requests }}</td><td>{{ f.transfers }}</td>
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
