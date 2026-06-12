<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, Download as D } from '../lib/api'

const downloads = ref<D[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    downloads.value = await api.downloads()
  } catch {}
  loading.value = false
})

function fmtSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1073741824) return `${(bytes / 1048576).toFixed(1)} MB`
  return `${(bytes / 1073741824).toFixed(2)} GB`
}
</script>

<template>
  <div>
    <h1>Downloads</h1>
    <p v-if="loading">Loading...</p>
    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Size</th>
            <th>Done</th>
            <th>Progress</th>
            <th>Speed</th>
            <th>Sources</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="d in downloads" :key="d.hash">
            <td>{{ d.name }}</td>
            <td>{{ fmtSize(d.size) }}</td>
            <td>{{ fmtSize(d.done) }}</td>
            <td>
              <div class="bar-wrap">
                <div class="bar-fill" :style="{ width: (d.progress * 100).toFixed(1) + '%' }" />
              </div>
              {{ (d.progress * 100).toFixed(1) }}%
            </td>
            <td>{{ fmtSize(d.speed) }}/s</td>
            <td>{{ d.sources }}</td>
            <td>{{ d.status }}</td>
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
.bar-wrap { width: 100px; height: 6px; background: var(--border); border-radius: 3px; display: inline-block; vertical-align: middle; margin-right: 8px; }
.bar-fill { height: 100%; background: var(--accent); border-radius: 3px; }
</style>
