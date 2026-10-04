<script setup lang="ts">
import type { StatusIncident } from '../types';

defineProps<{
  activeIncidents: StatusIncident[];
  pastIncidents: StatusIncident[];
}>();

function formatDateTime(dateStr: string): string {
  try {
    return new Date(dateStr).toLocaleString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateStr;
  }
}

function getSeverityBadgeStyle(severity: string) {
  switch (severity) {
    case 'critical':
      return { backgroundColor: 'rgba(239, 68, 68, 0.15)', color: '#fca5a5', border: '1px solid rgba(239, 68, 68, 0.3)' };
    case 'major':
      return { backgroundColor: 'rgba(245, 158, 11, 0.15)', color: '#fcd34d', border: '1px solid rgba(245, 158, 11, 0.3)' };
    default:
      return { backgroundColor: 'rgba(59, 130, 246, 0.15)', color: '#93c5fd', border: '1px solid rgba(59, 130, 246, 0.3)' };
  }
}
</script>

<template>
  <div class="incidents-card">
    <div class="monitors-header">
      <h2 class="monitors-title">Incidents & Maintenance</h2>
      <span class="monitors-subtitle">Incident timeline and updates</span>
    </div>

    <!-- Active Incidents -->
    <div v-if="activeIncidents.length > 0">
      <div v-for="incident in activeIncidents" :key="incident.id" class="incident-entry">
        <div class="incident-title-row">
          <span class="incident-title">{{ incident.title }}</span>
          <span class="incident-badge" :style="getSeverityBadgeStyle(incident.severity)">
            {{ incident.severity }}
          </span>
        </div>
        <div class="incident-timeline">
          <div v-for="update in incident.updates" :key="update.id" class="timeline-step">
            <div class="timeline-status">{{ update.status }}</div>
            <div class="timeline-time">{{ formatDateTime(update.createdAt) }}</div>
            <p class="timeline-message">{{ update.message }}</p>
          </div>
        </div>
      </div>
    </div>
    <div v-else style="padding: 1rem 0; color: var(--text-muted); font-size: 0.875rem;">
      No active incidents or disruptions reported.
    </div>

    <!-- Past Incidents -->
    <div v-if="pastIncidents.length > 0" style="margin-top: 2rem; border-top: 1px solid var(--border-subtle); padding-top: 1.5rem;">
      <h3 style="font-size: 1rem; font-weight: 600; margin-bottom: 1rem;">Past Incidents</h3>
      <div v-for="incident in pastIncidents" :key="incident.id" class="incident-entry">
        <div class="incident-title-row">
          <span class="incident-title">{{ incident.title }}</span>
          <span class="incident-badge" style="background-color: rgba(16, 185, 129, 0.15); color: #6ee7b7; border: 1px solid rgba(16, 185, 129, 0.3)">
            Resolved
          </span>
        </div>
        <div class="incident-timeline">
          <div v-for="update in incident.updates" :key="update.id" class="timeline-step">
            <div class="timeline-status">{{ update.status }}</div>
            <div class="timeline-time">{{ formatDateTime(update.createdAt) }}</div>
            <p class="timeline-message">{{ update.message }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
