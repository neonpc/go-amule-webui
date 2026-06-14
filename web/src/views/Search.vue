<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, SearchResult as SR } from '../lib/api'

const query = ref('')
const searchType = ref('global')
const results = ref<SR[]>([])
const searching = ref(false)

type SearchType = { value: string; label: string }
const searchTypes: SearchType[] = [
  { value: 'local', label: 'Local' },
  { value: 'global', label: 'Global' },
  { value: 'kad', label: 'Kad' },
]

async function doSearch() {
  if (!query.value.trim()) return
  searching.value = true
  try {
    await api.search(query.value, searchType.value)
  } catch (e) {
    console.error('search failed', e)
    searching.value = false
    return
  }
  setTimeout(async () => {
    try {
      const r = await api.searchResults()
      results.value = r
    } catch {}
    searching.value = false
  }, 15000)
}

async function download(hash: string) {
  try {
    await api.searchDownload(hash)
  } catch (e: any) {
    alert(e.message)
  }
}

function fmtSize(bytes: number): string {
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1048576).toFixed(1)} MB`
}
</script>

<template>
  <div>
    <h1>Search</h1>
    <div class="search-box">
      <input v-model="query" @keyup.enter="doSearch" placeholder="Search files..." />
      <select v-model="searchType" class="search-type">
        <option v-for="t in searchTypes" :key="t.value" :value="t.value">{{ t.label }}</option>
      </select>
      <button @click="doSearch" :disabled="searching">Search</button>
    </div>
    <p v-if="searching">Searching...</p>
    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr><th>Name</th><th>Size</th><th>Sources</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="r in results" :key="r.hash">
            <td>{{ r.name }}</td><td>{{ fmtSize(r.size) }}</td><td>{{ r.sources }}</td>
            <td><button class="btn-dl" @click="download(r.hash)">Download</button></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.search-box { display: flex; gap: 8px; margin: 12px 0; }
input { flex: 1; padding: 10px 14px; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-card); color: var(--text); }
select { padding: 10px 14px; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-card); color: var(--text); cursor: pointer; }
button { padding: 10px 20px; background: var(--accent); color: white; border: none; border-radius: 8px; cursor: pointer; }
button:disabled { opacity: 0.5; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; margin-top: 12px; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--border); font-size: 0.9rem; }
th { color: var(--text-muted); font-weight: 600; }
.btn-dl { padding: 4px 10px; background: var(--accent); color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 0.8rem; }
</style>
