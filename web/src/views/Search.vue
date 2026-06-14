<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { api, SearchResult as SR } from '../lib/api'

const query = ref('')
const searchType = ref('global')
const extFilter = ref('')
const results = ref<SR[]>([])
const searching = ref(false)
const sortBy = ref<'name' | 'size' | 'sources'>('sources')
const sortDir = ref<'asc' | 'desc'>('desc')
const page = ref(0)
const perPage = 100

const searchTypes = [
  { value: 'local', label: 'Local' },
  { value: 'global', label: 'Global' },
  { value: 'kad', label: 'Kad' },
]

let pollTimer: ReturnType<typeof setTimeout> | null = null
let prevCount = 0
let stalePolls = 0

onUnmounted(() => {
  if (pollTimer) clearTimeout(pollTimer)
})

const filtered = computed(() => {
  if (!extFilter.value) return results.value
  const ext = extFilter.value.toLowerCase().replace(/^\./, '')
  return results.value.filter(r => {
    const name = r.name.toLowerCase()
    return name.endsWith(`.${ext}`)
  })
})

function sortByColumn(col: 'name' | 'size' | 'sources') {
  if (sortBy.value === col) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = col
    sortDir.value = col === 'name' ? 'asc' : 'desc'
  }
  page.value = 0
}

const sorted = computed(() => {
  const arr = [...filtered.value]
  const dir = sortDir.value === 'asc' ? 1 : -1
  arr.sort((a, b) => {
    if (sortBy.value === 'name') return a.name.localeCompare(b.name) * dir
    if (sortBy.value === 'size') return (Number(a.size) - Number(b.size)) * dir
    return (a.sources - b.sources) * dir
  })
  return arr
})

const pages = computed(() => Math.max(1, Math.ceil(sorted.value.length / perPage)))
const paged = computed(() => sorted.value.slice(page.value * perPage, (page.value + 1) * perPage))

async function pollResults() {
  try {
    const r = await api.searchResults()
    if (r) {
      results.value = r
      if (r.length === prevCount) {
        stalePolls++
      } else {
        stalePolls = 0
        prevCount = r.length
      }
      if (stalePolls >= 5) {
        searching.value = false
        return
      }
    }
  } catch {}
  if (searching.value) {
    pollTimer = setTimeout(pollResults, 2000)
  }
}

async function doSearch() {
  if (!query.value.trim()) return
  searching.value = true
  results.value = []
  prevCount = 0
  stalePolls = 0
  page.value = 0
  try {
    await api.search(query.value, searchType.value)
  } catch (e) {
    console.error('search failed', e)
    searching.value = false
    return
  }
  pollResults()
}

function cancelSearch() {
  if (pollTimer) clearTimeout(pollTimer)
  searching.value = false
  api.searchStop().catch(() => {})
}

async function download(hash: string) {
  try {
    await api.searchDownload(hash)
  } catch (e: any) {
    alert(e.message)
  }
}

function fmtSize(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1073741824) return `${(bytes / 1048576).toFixed(1)} MB`
  return `${(bytes / 1073741824).toFixed(2)} GB`
}

function sortIcon(col: string): string {
  if (sortBy.value !== col) return '⇅'
  return sortDir.value === 'asc' ? '↑' : '↓'
}
</script>

<template>
  <div>
    <h1>Search</h1>
    <div class="search-bar">
      <input v-model="query" @keyup.enter="doSearch" placeholder="Search files..." :disabled="searching" />
      <input v-model="extFilter" placeholder="Filter by extension (e.g. mp4)" class="ext-input" :disabled="searching" />
      <select v-model="searchType" class="search-type" :disabled="searching">
        <option v-for="t in searchTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
      </select>
      <button v-if="!searching" @click="doSearch">Search</button>
      <button v-else class="btn-cancel" @click="cancelSearch">Stop</button>
    </div>
    <div v-if="searching" class="search-status">Searching... ({{ results.length }} results found)</div>
    <div v-if="extFilter && results.length" class="search-status">Filtered: {{ filtered.length }} / {{ results.length }}</div>
    <template v-if="filtered.length > 0">
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th class="sortable" @click="sortByColumn('name')">Name {{ sortIcon('name') }}</th>
              <th class="sortable" @click="sortByColumn('size')">Size {{ sortIcon('size') }}</th>
              <th class="sortable" @click="sortByColumn('sources')">Sources {{ sortIcon('sources') }}</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in paged" :key="r.hash">
              <td class="cell-name">{{ r.name }}</td>
              <td>{{ fmtSize(r.size) }}</td>
              <td>{{ r.sources }}</td>
              <td><button class="btn-dl" @click="download(r.hash)">Download</button></td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="pager" v-if="pages > 1">
        <button @click="page = Math.max(0, page - 1)" :disabled="page === 0">← Prev</button>
        <span>Page {{ page + 1 }} / {{ pages }} ({{ filtered.length }} total)</span>
        <button @click="page = Math.min(pages - 1, page + 1)" :disabled="page >= pages - 1">Next →</button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.search-bar { position: sticky; top: 0; z-index: 10; background: var(--bg); display: flex; gap: 8px; margin: 12px 0; padding: 8px 0; }
input { flex: 1; padding: 10px 14px; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-card); color: var(--text); }
.ext-input { flex: 0 0 160px; }
select { padding: 10px 14px; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-card); color: var(--text); cursor: pointer; }
button { padding: 10px 20px; background: var(--accent); color: white; border: none; border-radius: 8px; cursor: pointer; white-space: nowrap; }
button:disabled { opacity: 0.5; cursor: default; }
.btn-cancel { background: #ef4444; }
.search-status { color: var(--text-muted); font-size: 0.85rem; padding: 8px 0; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; margin-top: 4px; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--border); font-size: 0.9rem; }
th { position: sticky; top: 56px; background: var(--bg); color: var(--text-muted); font-weight: 600; white-space: nowrap; }
.sortable { cursor: pointer; user-select: none; }
.sortable:hover { color: var(--text); }
.cell-name { max-width: 400px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.btn-dl { padding: 4px 10px; background: var(--accent); color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 0.8rem; }
.pager { display: flex; align-items: center; justify-content: center; gap: 16px; margin-top: 16px; }
.pager button { padding: 6px 16px; font-size: 0.85rem; }
.pager span { font-size: 0.85rem; color: var(--text-muted); }
</style>
