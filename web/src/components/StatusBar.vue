<script setup lang="ts">
defineProps<{
  status: {
    ed2k_connected: boolean
    kad_connected: boolean
    kad_firewalled: boolean
    dl_speed: number
    ul_speed: number
  } | null
}>()
</script>

<template>
  <div class="status-bar">
    <div class="status-item">
      <span :class="['dot', status?.ed2k_connected ? 'green' : 'red']" />
      eD2K {{ status?.ed2k_connected ? 'Connected' : 'Disconnected' }}
    </div>
    <div class="status-item">
      <span :class="['dot', status?.kad_connected ? 'green' : 'red']" />
      Kad {{ status?.kad_connected ? (status?.kad_firewalled ? 'Firewalled' : 'OK') : 'Disconnected' }}
    </div>
  </div>
</template>

<style scoped>
.status-bar {
  display: flex;
  gap: 16px;
  padding: 12px 0;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.85rem;
  color: var(--text-muted);
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot.green { background: var(--success); }
.dot.red { background: var(--danger); }
</style>
