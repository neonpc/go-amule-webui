<script setup lang="ts">
defineProps<{
  label: string
  speed: number
  color: string
}>()

function formatSpeed(bytes: number): string {
  if (bytes < 1024) return `${bytes} B/s`
  if (bytes < 1048576) return `${(bytes / 1024).toFixed(1)} KB/s`
  return `${(bytes / 1048576).toFixed(2)} MB/s`
}
</script>

<template>
  <div class="speed-card" :style="{ borderTop: `3px solid ${color}` }">
    <div class="speed-label">{{ label }}</div>
    <div class="speed-value" :style="{ color }">{{ formatSpeed(speed) }}</div>
  </div>
</template>

<style scoped>
.speed-card {
  background: var(--bg-card);
  border-radius: 12px;
  padding: 20px;
  min-width: 180px;
}

@media (max-width: 768px) {
  .speed-card {
    padding: 14px;
    min-width: 130px;
    flex: 1;
  }
  .speed-value {
    font-size: 1.2rem;
  }
}

.speed-label {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.speed-value {
  font-size: 1.5rem;
  font-weight: 700;
}
</style>
