<script setup lang="ts">
import { onMounted, ref, onUnmounted } from 'vue'
import { useAmuleStore } from '../stores/amule'
import { useSocket } from '../composables/useSocket'
import SpeedCard from '../components/SpeedCard.vue'
import StatusBar from '../components/StatusBar.vue'
import { api, Status, Download, Upload } from '../lib/api'

const store = useAmuleStore()
const { connected } = useSocket()
const status = ref<Status | null>(null)
const downloads = ref<Download[]>([])
const uploads = ref<Upload[]>([])

let pollTimer: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  await refresh()
  pollTimer = setTimeout(poll, 5000)
})

onUnmounted(() => {
  if (pollTimer) clearTimeout(pollTimer)
})

async function poll() {
  await refresh()
  pollTimer = setTimeout(poll, 5000)
}

async function refresh() {
  try {
    const [s, d, u] = await Promise.all([
      api.status(),
      api.downloads(),
      api.uploads(),
    ])
    status.value = s
    downloads.value = d
    uploads.value = u
  } catch {}
}

async function ed2kConnect() {
  try {
    await api.ed2kAction('connect')
    status.value = await api.status()
  } catch {}
}

async function ed2kDisconnect() {
  try {
    await api.ed2kAction('disconnect')
    status.value = await api.status()
  } catch {}
}

function fmtSize(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1073741824) return `${(bytes / 1048576).toFixed(1)} MB`
  return `${(bytes / 1073741824).toFixed(2)} GB`
}

function fmtSpeed(bytes: number): string {
  if (!bytes) return '0 B/s'
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB/s`
  return `${(bytes / 1048576).toFixed(1)} MB/s`
}

function fmtPct(pct: number): string {
  return (pct * 100).toFixed(1) + '%'
}
</script>

<template>
  <div class="dashboard">
    <h1>Dashboard</h1>
    <StatusBar :status="status" />

    <div class="kpi-row">
      <SpeedCard label="Download" :speed="status?.dl_speed ?? 0" color="#22c55e" />
      <SpeedCard label="Upload" :speed="status?.ul_speed ?? 0" color="#3b82f6" />
      <div class="kpi-card">
        <div class="kpi-label">Queue</div>
        <div class="kpi-value">{{ status?.queue_count ?? 0 }}</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">Sources</div>
        <div class="kpi-value">{{ status?.source_count ?? 0 }}</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">Downloads</div>
        <div class="kpi-value">{{ downloads.length }}</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-label">Uploads</div>
        <div class="kpi-value">{{ uploads.length }}</div>
      </div>
    </div>

    <div class="actions" v-if="status">
      <button v-if="!status.ed2k_connected" class="btn-connect" @click="ed2kConnect">Connect eD2K</button>
      <button v-else class="btn-disconnect" @click="ed2kDisconnect">Disconnect eD2K</button>
      <span class="tag-server" v-if="status.ed2k_server">Server: {{ status.ed2k_server }}</span>
    </div>

    <div class="ws-indicator">
      WebSocket: {{ connected ? '🟢 Connected' : '🔴 Disconnected' }}
    </div>

    <div v-if="downloads.length" class="section">
      <h2>Active Downloads</h2>
      <div class="table-wrap">
        <table>
          <thead>
            <tr><th>Name</th><th>Size</th><th>Progress</th><th>Speed</th><th>Sources</th></tr>
          </thead>
          <tbody>
            <tr v-for="d in downloads.slice(0, 10)" :key="d.hash">
              <td class="cell-name">{{ d.name }}</td>
              <td>{{ fmtSize(d.size) }}</td>
              <td>
                <div class="pbar-wrap">
                  <div class="pbar" :style="{ width: fmtPct(d.progress) }"></div>
                  <span>{{ fmtPct(d.progress) }}</span>
                </div>
              </td>
              <td>{{ fmtSpeed(d.speed) }}</td>
              <td>{{ d.sources }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="uploads.length" class="section">
      <h2>Active Uploads</h2>
      <div class="table-wrap">
        <table>
          <thead>
            <tr><th>File</th><th>Client</th><th>Speed</th><th>Uploaded</th></tr>
          </thead>
          <tbody>
            <tr v-for="u in uploads.slice(0, 10)" :key="u.client + u.name">
              <td class="cell-name">{{ u.name }}</td>
              <td>{{ u.client }}</td>
              <td>{{ fmtSpeed(u.speed) }}</td>
              <td>{{ fmtSize(u.uploaded) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.kpi-row { display: flex; gap: 12px; flex-wrap: wrap; margin-top: 16px; }
.kpi-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: 12px; padding: 16px 20px; min-width: 100px; text-align: center; }
.kpi-label { font-size: 0.75rem; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; }
.kpi-value { font-size: 1.5rem; font-weight: 700; margin-top: 4px; }
.actions { margin-top: 16px; display: flex; gap: 8px; align-items: center; }
.btn-connect { padding: 10px 20px; background: var(--accent); color: white; border: none; border-radius: 8px; cursor: pointer; }
.btn-disconnect { padding: 10px 20px; background: #ef4444; color: white; border: none; border-radius: 8px; cursor: pointer; }
.tag-server { padding: 4px 12px; border-radius: 12px; font-size: 0.85rem; font-weight: 600; background: #16a34a22; color: #16a34a; border: 1px solid #16a34a44; }
.ws-indicator { margin-top: 16px; font-size: 0.85rem; color: var(--text-muted); }
.section { margin-top: 24px; }
.section h2 { font-size: 1rem; margin-bottom: 8px; color: var(--text-muted); }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; }
th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid var(--border); font-size: 0.85rem; }
th { color: var(--text-muted); font-weight: 600; white-space: nowrap; }
.cell-name { max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pbar-wrap { display: flex; align-items: center; gap: 6px; }
.pbar { height: 6px; border-radius: 3px; background: var(--accent); min-width: 40px; max-width: 100px; }
.pbar-wrap span { font-size: 0.78rem; color: var(--text-muted); white-space: nowrap; }
</style>
