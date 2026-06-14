<script setup lang="ts">
import { ref } from 'vue'
import { api, FSEntry } from '../lib/api'

const rootPath = '/media'
const rootEntries = ref<FSEntry[] | null>(null)
const loading = ref(true)
const err = ref('')

const dirs = ref<Map<string, { entries: FSEntry[]; loading: boolean }>>(new Map())
const openDirs = ref<Set<string>>(new Set())

async function loadDir(path: string) {
  if (dirs.value.has(path)) return
  dirs.value.set(path, { entries: [], loading: true })
  try {
    const entries = await api.fsBrowse(path)
    dirs.value.set(path, { entries, loading: false })
  } catch (e: any) {
    dirs.value.set(path, { entries: [], loading: false })
  }
}

async function toggleDir(path: string) {
  if (openDirs.value.has(path)) {
    openDirs.value.delete(path)
    openDirs.value = new Set(openDirs.value)
    return
  }
  openDirs.value.add(path)
  openDirs.value = new Set(openDirs.value)
  await loadDir(path)
}

async function loadRoot() {
  loading.value = true
  try {
    rootEntries.value = await api.fsBrowse(rootPath)
  } catch (e: any) {
    err.value = e.message
  }
  loading.value = false
}

loadRoot()

function fmtSize(bytes: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1073741824) return `${(bytes / 1048576).toFixed(1)} MB`
  return `${(bytes / 1073741824).toFixed(2)} GB`
}

function extClass(name: string): string {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  if (['mp4', 'mkv', 'avi', 'mov', 'wmv', 'flv', 'webm'].includes(ext)) return 'ext-video'
  if (['mp3', 'flac', 'wav', 'aac', 'ogg', 'wma'].includes(ext)) return 'ext-audio'
  if (['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp'].includes(ext)) return 'ext-image'
  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) return 'ext-archive'
  if (['pdf', 'epub', 'mobi', 'cbr', 'cbz'].includes(ext)) return 'ext-doc'
  return 'ext-other'
}
</script>

<template>
  <div>
    <h1>Shared Files</h1>
    <p class="root-hint">Browsing: {{ rootPath }}</p>

    <div v-if="loading" class="loading">Loading...</div>
    <p v-else-if="err" class="err">{{ err }}</p>
    <div v-else class="tree">
      <div v-for="e in rootEntries" :key="e.path" class="tree-item">
        <template v-if="e.is_dir">
          <div class="dir-row" @click="toggleDir(e.path)">
            <span class="arrow">{{ openDirs.has(e.path) ? '▼' : '▶' }}</span>
            <span class="icon-dir">📁</span>
            <span class="name">{{ e.name }}</span>
          </div>
          <div v-if="openDirs.has(e.path)" class="dir-children">
            <div v-if="dirs.get(e.path)?.loading" class="loading-sm">Loading...</div>
            <div v-else-if="!dirs.get(e.path)?.entries?.length" class="empty-sm">Empty</div>
            <template v-else>
              <div v-for="child in dirs.get(e.path)?.entries" :key="child.path" class="tree-item">
                <template v-if="child.is_dir">
                  <div class="dir-row" @click="toggleDir(child.path)">
                    <span class="arrow">{{ openDirs.has(child.path) ? '▼' : '▶' }}</span>
                    <span class="icon-dir">📁</span>
                    <span class="name">{{ child.name }}</span>
                  </div>
                  <div v-if="openDirs.has(child.path)" class="dir-children">
                    <div v-if="dirs.get(child.path)?.loading" class="loading-sm">Loading...</div>
                    <div v-else-if="!dirs.get(child.path)?.entries?.length" class="empty-sm">Empty</div>
                    <div v-else v-for="c2 in dirs.get(child.path)?.entries" :key="c2.path" class="tree-item">
                      <div v-if="c2.is_dir" class="dir-row" @click="toggleDir(c2.path)">
                        <span class="arrow">{{ openDirs.has(c2.path) ? '▼' : '▶' }}</span>
                        <span class="icon-dir">📁</span>
                        <span class="name">{{ c2.name }}</span>
                        <span class="meta">subdirectory</span>
                      </div>
                      <div v-else class="file-row">
                        <span class="icon-file" :class="extClass(c2.name)">📄</span>
                        <span class="name">{{ c2.name }}</span>
                        <span class="meta">{{ fmtSize(c2.size) }}</span>
                      </div>
                    </div>
                  </div>
                </template>
                <div v-else class="file-row">
                  <span class="icon-file" :class="extClass(child.name)">📄</span>
                  <span class="name">{{ child.name }}</span>
                  <span class="meta">{{ fmtSize(child.size) }}</span>
                </div>
              </div>
            </template>
          </div>
        </template>
        <div v-else class="file-row">
          <span class="icon-file" :class="extClass(e.name)">📄</span>
          <span class="name">{{ e.name }}</span>
          <span class="meta">{{ fmtSize(e.size) }}</span>
        </div>
      </div>
      <div v-if="!rootEntries?.length" class="empty">Empty directory</div>
    </div>
  </div>
</template>

<style scoped>
.root-hint { color: var(--text-muted); font-size: 0.8rem; margin: 4px 0 8px; }
.loading { color: var(--text-muted); margin-top: 12px; }
.loading-sm { padding: 4px 0 4px 24px; font-size: 0.8rem; color: var(--text-muted); }
.empty { color: var(--text-muted); margin-top: 12px; }
.empty-sm { padding: 4px 0 4px 24px; font-size: 0.8rem; color: var(--text-muted); }
.err { color: var(--err, #ef4444); font-size: 0.85rem; }
.tree { margin-top: 8px; }
.tree-item { user-select: none; }
.dir-row { display: flex; align-items: center; gap: 4px; padding: 3px 4px; border-radius: 4px; cursor: pointer; }
.dir-row:hover { background: var(--bg-hover, #ffffff08); }
.dir-children { padding-left: 20px; }
.file-row { display: flex; align-items: center; gap: 4px; padding: 2px 4px 2px 20px; }
.arrow { font-size: 0.6rem; width: 12px; text-align: center; color: var(--text-muted); }
.icon-dir { font-size: 0.9rem; }
.icon-file { font-size: 0.85rem; }
.name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 0.85rem; }
.meta { color: var(--text-muted); font-size: 0.78rem; white-space: nowrap; }
.ext-video { opacity: 0.9; }
.ext-audio { opacity: 0.8; }
.ext-image { opacity: 0.8; }
</style>
