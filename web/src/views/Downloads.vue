<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { api, Download as D } from '../lib/api'

const downloads = ref<D[]>([])
const loading = ref(true)

let timer: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  await fetchDownloads()
  timer = setTimeout(poll, 3000)
})

onUnmounted(() => {
  if (timer) clearTimeout(timer)
})

async function poll() {
  await fetchDownloads()
  timer = setTimeout(poll, 3000)
}

async function fetchDownloads() {
  try {
    downloads.value = await api.downloads()
  } catch {}
  loading.value = false
}

async function doAction(hash: string, action: string) {
  try {
    await api.downloadAction(hash, action)
    await fetchDownloads()
  } catch (e: any) {
    alert(e.message)
  }
}

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

    <template v-else>
      <div class="table-wrap desktop-only">
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
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="d in downloads" :key="d.hash">
              <td>{{ d.name }}</td>
              <td>{{ fmtSize(d.size) }}</td>
              <td>{{ fmtSize(d.done) }}</td>
              <td>
                <div class="bar-wrap">
                  <div class="bar-fill" :style="{ width: Math.min(d.progress, 100).toFixed(1) + '%' }" />
                </div>
                {{ Math.min(d.progress, 100).toFixed(1) }}%
              </td>
              <td>{{ fmtSize(d.speed) }}/s</td>
              <td>{{ d.sources }}</td>
              <td>{{ d.status }}</td>
              <td>
                <button v-if="d.paused" class="btn-action" @click="doAction(d.hash, 'resume')">Resume</button>
                <button v-else class="btn-action btn-pause" @click="doAction(d.hash, 'pause')">Pause</button>
                <button class="btn-action btn-cancel" @click="doAction(d.hash, 'cancel')">Cancel</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="cards mobile-only">
        <div v-for="d in downloads" :key="d.hash" class="card">
          <div class="card-name">{{ d.name }}</div>
          <div class="card-body">
            <div class="card-row">
              <span class="card-label">Size</span>
              <span>{{ fmtSize(d.size) }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">Done</span>
              <span>{{ fmtSize(d.done) }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">Progress</span>
              <span class="card-progress">
                <div class="bar-wrap card-bar">
                  <div class="bar-fill" :style="{ width: Math.min(d.progress, 100).toFixed(1) + '%' }" />
                </div>
                {{ Math.min(d.progress, 100).toFixed(1) }}%
              </span>
            </div>
            <div class="card-row">
              <span class="card-label">Speed</span>
              <span>{{ fmtSize(d.speed) }}/s</span>
            </div>
            <div class="card-row">
              <span class="card-label">Sources</span>
              <span>{{ d.sources }}</span>
            </div>
            <div class="card-row">
              <span class="card-label">Status</span>
              <span>{{ d.status }}</span>
            </div>
          </div>
          <div class="card-actions">
            <button v-if="d.paused" class="btn-action" @click="doAction(d.hash, 'resume')">Resume</button>
            <button v-else class="btn-action btn-pause" @click="doAction(d.hash, 'pause')">Pause</button>
            <button class="btn-action btn-cancel" @click="doAction(d.hash, 'cancel')">Cancel</button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; margin-top: 12px; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--border); font-size: 0.9rem; }
th { color: var(--text-muted); font-weight: 600; }
.bar-wrap { width: 100px; height: 6px; background: var(--border); border-radius: 3px; display: inline-block; vertical-align: middle; margin-right: 8px; }
.bar-fill { height: 100%; background: var(--accent); border-radius: 3px; }
.btn-action { padding: 4px 10px; border: none; border-radius: 4px; cursor: pointer; font-size: 0.8rem; margin-right: 4px; background: var(--accent); color: white; }
.btn-pause { background: #f59e0b; }
.btn-cancel { background: #ef4444; }

.cards { display: flex; flex-direction: column; gap: 12px; margin-top: 12px; }
.card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 14px; }
.card-name { font-weight: 600; font-size: 0.95rem; margin-bottom: 10px; word-break: break-word; }
.card-body { display: flex; flex-direction: column; gap: 6px; }
.card-row { display: flex; justify-content: space-between; align-items: center; font-size: 0.85rem; }
.card-label { color: var(--text-muted); }
.card-progress { display: flex; align-items: center; gap: 6px; }
.card-bar { width: 80px; margin-right: 0; }
.card-actions { display: flex; gap: 6px; margin-top: 12px; padding-top: 10px; border-top: 1px solid var(--border); }
.card-actions .btn-action { flex: 1; padding: 10px; text-align: center; font-size: 0.85rem; }

.desktop-only { display: block; }
.mobile-only { display: none; }

@media (max-width: 768px) {
  .desktop-only { display: none; }
  .mobile-only { display: block; }
}
</style>
