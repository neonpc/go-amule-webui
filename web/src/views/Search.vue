<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, SearchResult as SR } from '../lib/api'

const query = ref('')
const results = ref<SR[]>([])
const searching = ref(false)

async function doSearch() {
  if (!query.value.trim()) return
  searching.value = true
  await api.search(query.value)
  setTimeout(async () => {
    try {
      const r = await api.searchResults()
      results.value = r
      await api.searchStop()
    } catch {}
    searching.value = false
  }, 3000)
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
      <button @click="doSearch" :disabled="searching">Search</button>
    </div>
    <p v-if="searching">Searching...</p>
    <div v-else class="table-wrap">
      <table>
        <thead>
          <tr><th>Name</th><th>Size</th><th>Sources</th></tr>
        </thead>
        <tbody>
          <tr v-for="r in results" :key="r.hash">
            <td>{{ r.name }}</td><td>{{ fmtSize(r.size) }}</td><td>{{ r.sources }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.search-box { display: flex; gap: 8px; margin: 12px 0; }
input { flex: 1; padding: 10px 14px; border: 1px solid var(--border); border-radius: 8px; background: var(--bg-card); color: var(--text); }
button { padding: 10px 20px; background: var(--accent); color: white; border: none; border-radius: 8px; cursor: pointer; }
button:disabled { opacity: 0.5; }
.table-wrap { overflow-x: auto; }
table { width: 100%; border-collapse: collapse; margin-top: 12px; }
th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid var(--border); font-size: 0.9rem; }
th { color: var(--text-muted); font-weight: 600; }
</style>
