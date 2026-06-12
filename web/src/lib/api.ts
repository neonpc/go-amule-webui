export interface Status {
  ed2k_connected: boolean
  ed2k_server: string
  kad_connected: boolean
  kad_firewalled: boolean
  dl_speed: number
  ul_speed: number
  queue_count: number
  source_count: number
}

export interface Download {
  hash: string
  name: string
  size: number
  done: number
  speed: number
  progress: number
  status: string
  sources: number
  priority: number
  category: string
  paused: boolean
}

export interface Upload {
  name: string
  client: string
  speed: number
  uploaded: number
}

export interface SharedFile {
  hash: string
  name: string
  size: number
  requests: number
  transfers: number
  priority: number
  last_xfer: number
  all_xfer: number
}

export interface SearchResult {
  hash: string
  name: string
  size: number
  sources: number
}

export interface ServerEntry {
  name: string
  desc: string
  address: string
  ip: string
  port: number
  users: number
  files: number
}

export interface Prefs {
  [key: string]: any
}

const BASE = ''

async function fetchJSON<T>(url: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${url}`, {
    headers: { 'Content-Type': 'application/json' },
    ...opts,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  return res.json()
}

export const api = {
  status: () => fetchJSON<Status>('/api/status'),
  downloads: () => fetchJSON<Download[]>('/api/downloads'),
  downloadAction: (hash: string, action: string) =>
    fetchJSON<{ status: string }>(`/api/downloads/${hash}?hash=${hash}&action=${action}`, { method: 'POST' }),
  uploads: () => fetchJSON<Upload[]>('/api/uploads'),
  shared: () => fetchJSON<SharedFile[]>('/api/shared'),
  search: (query: string, type: string = 'global') =>
    fetchJSON<{ status: string }>('/api/search', {
      method: 'POST',
      body: JSON.stringify({ query, searchType: type }),
    }),
  searchResults: () => fetchJSON<SearchResult[]>('/api/search/results'),
  searchStop: () => fetchJSON<{ status: string }>('/api/search/stop'),
  servers: () => fetchJSON<ServerEntry[]>('/api/servers'),
  kad: () => fetchJSON<{ connected: boolean; firewalled: boolean }>('/api/kad'),
  kadAction: (action: string) =>
    fetchJSON<{ status: string }>(`/api/kad?action=${action}`, { method: 'POST' }),
  stats: () => fetchJSON<Record<string, number>>('/api/stats'),
  log: () => fetchJSON<string[]>('/api/log'),
}
