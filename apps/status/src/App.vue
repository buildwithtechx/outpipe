<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { fetchStatusData } from './api';
import type { StatusPageData } from './types';
import StatusHeader from './components/status-header.vue';
import SystemBanner from './components/system-banner.vue';
import MonitorCard from './components/monitor-card.vue';
import IncidentTimeline from './components/incident-timeline.vue';
import SubscribeModal from './components/subscribe-modal.vue';

const statusData = ref<StatusPageData | null>(null);
const isLoading = ref(true);
const isSubscribeOpen = ref(false);
const slug = ref('default');
let timer: ReturnType<typeof setInterval> | null = null;

async function loadData() {
  const pathParts = window.location.pathname.split('/').filter(Boolean);
  if (pathParts.length > 0 && pathParts[0] !== 'status') {
    slug.value = pathParts[0];
  }
  statusData.value = await fetchStatusData(slug.value);
  isLoading.value = false;
}

onMounted(() => {
  loadData();
  timer = setInterval(loadData, 30000);
});

onUnmounted(() => {
  if (timer) clearInterval(timer);
});
</script>

<template>
  <div class="status-container">
    <div v-if="isLoading" style="display: flex; justify-content: center; align-items: center; min-height: 50vh;">
      <span style="color: var(--text-muted); font-size: 0.875rem;">Loading status...</span>
    </div>

    <div v-else-if="statusData">
      <StatusHeader
        :title="statusData.title"
        @open-subscribe="isSubscribeOpen = true"
      />

      <main>
        <SystemBanner
          :system-status="statusData.systemStatus"
          :system-status-message="statusData.systemStatusMessage"
          :last-updated="statusData.lastUpdated"
        />

        <MonitorCard :monitors="statusData.monitors" />

        <IncidentTimeline
          :active-incidents="statusData.activeIncidents"
          :past-incidents="statusData.pastIncidents"
        />
      </main>

      <footer class="status-footer">
        <p>Outpipe Status &bull; Real-time infrastructure monitoring</p>
      </footer>

      <SubscribeModal
        v-if="isSubscribeOpen"
        :slug="slug"
        @close="isSubscribeOpen = false"
      />
    </div>
  </div>
</template>
