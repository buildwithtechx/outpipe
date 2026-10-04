<script setup lang="ts">
import { ref } from 'vue';
import type { DayUptime, StatusMonitor } from '../types';

defineProps<{
  monitors: StatusMonitor[];
}>();

const hoveredDay = ref<{ monitorId: string; day: DayUptime } | null>(null);

function formatStatus(status: string): string {
  switch (status) {
    case 'operational': return 'Operational';
    case 'degraded': return 'Degraded';
    case 'outage': return 'Major Outage';
    case 'maintenance': return 'Maintenance';
    default: return 'Active';
  }
}

function getStatusColor(status: string): string {
  switch (status) {
    case 'operational': return 'var(--status-operational)';
    case 'degraded': return 'var(--status-degraded)';
    case 'outage': return 'var(--status-outage)';
    default: return 'var(--status-unknown)';
  }
}
</script>

<template>
  <div class="monitors-card">
    <div class="monitors-header">
      <h2 class="monitors-title">Services & Relays</h2>
      <span class="monitors-subtitle">Uptime over the past 90 days</span>
    </div>

    <div v-for="monitor in monitors" :key="monitor.id" class="monitor-item">
      <div class="monitor-top">
        <div class="monitor-meta">
          <span class="monitor-name">{{ monitor.name }}</span>
          <span class="monitor-protocol">{{ monitor.type }}</span>
        </div>
        <div class="monitor-status-badge" :style="{ color: getStatusColor(monitor.status) }">
          <span class="status-dot" :class="monitor.status" style="width: 8px; height: 8px;" />
          {{ formatStatus(monitor.status) }}
        </div>
      </div>

      <div class="uptime-bars-wrapper">
        <div
          v-for="(day, idx) in monitor.history"
          :key="idx"
          :class="['uptime-bar', day.status]"
          :title="`${day.date}: ${day.uptimePercentage}% (${day.avgLatencyMs}ms)`"
          @mouseenter="hoveredDay = { monitorId: monitor.id, day }"
          @mouseleave="hoveredDay = null"
        />
      </div>

      <div class="uptime-footer">
        <span>90 days ago</span>
        <span v-if="hoveredDay && hoveredDay.monitorId === monitor.id" style="color: var(--text-primary); font-weight: 500;">
          {{ hoveredDay.day.date }}: {{ hoveredDay.day.uptimePercentage }}% ({{ hoveredDay.day.avgLatencyMs }}ms)
        </span>
        <span v-else>{{ monitor.uptime90Days }}% uptime</span>
        <span>Today</span>
      </div>
    </div>
  </div>
</template>
