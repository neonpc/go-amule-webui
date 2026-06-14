import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, Status, Download, SearchResult } from '../lib/api'

export const useAmuleStore = defineStore('amule', () => {
  const status = ref<Status | null>(null)
  const downloads = ref<Download[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchStatus() {
    try {
      status.value = await api.status()
    } catch (e: any) {
      error.value = e.message
    }
  }

  async function fetchDownloads() {
    try {
      downloads.value = await api.downloads()
    } catch (e: any) {
      error.value = e.message
    }
  }

  async function refresh() {
    loading.value = true
    error.value = null
    await Promise.allSettled([fetchStatus(), fetchDownloads()])
    loading.value = false
  }

  const searchResults = ref<SearchResult[]>([])
  const searchQuery = ref('')
  const searchType = ref('global')
  const searchActive = ref(false)
  const searchExtFilter = ref('')

  function setSearchResults(r: SearchResult[]) { searchResults.value = r }
  function clearSearch() {
    searchResults.value = []
    searchQuery.value = ''
    searchActive.value = false
  }

  return { status, downloads, loading, error, refresh, fetchStatus, fetchDownloads,
    searchResults, searchQuery, searchType, searchActive, searchExtFilter,
    setSearchResults, clearSearch }
})
