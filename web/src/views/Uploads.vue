<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, Upload as U } from '../lib/api'

const uploads = ref<U[]>([])
const loading = ref(true)

onMounted(async () => {
  try { uploads.value = await api.uploads() } catch {}
  loading.value = false
})

function fmt(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1048576).toFixed(2)} MB`
}
</script>

<template>
  <div>
    <h1>Uploads</h1>
    <p v-if="loading">Loading...</p>
    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr><th>File</th><th>Client</th><th>Speed</th><th>Uploaded</th></tr>
        </thead>
        <tbody>
          <tr v-for="u in uploads" :key="u.name + u.client">
            <td>{{ u.name }}</td><td>{{ u.client }}</td>
            <td>{{ fmt(u.speed) }}/s</td><td>{{ fmt(u.uploaded) }}</td>
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
