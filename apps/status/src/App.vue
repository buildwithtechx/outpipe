<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { fetchStatusResult } from './api';
import { publicSiteUrl } from '@outpipe/shared';
import outpipeMark from '@outpipe/shared/outpipe.svg?url';
import type { StatusPageData } from './types';
import StatusHeader from './components/status-header.vue';
import SystemBanner from './components/system-banner.vue';
import MonitorCard from './components/monitor-card.vue';
import IncidentTimeline from './components/incident-timeline.vue';
import SubscribeModal from './components/subscribe-modal.vue';

const statusData = ref<StatusPageData | null>(null);
const isLoading = ref(true);
const isRefreshing = ref(false);
const isSubscribeOpen = ref(false);
const slug = ref('default');
const pageState = ref<'ready' | 'not-found' | 'unavailable'>('unavailable');
const siteUrl = publicSiteUrl(import.meta.env.DEV, import.meta.env.VITE_PUBLIC_SITE_URL);
let icon: HTMLLinkElement | null = null;
let timer: ReturnType<typeof setTimeout> | null = null;
let disposed = false;

async function loadData() {
  if (isRefreshing.value) return;
  isRefreshing.value = true;
  const pathParts = window.location.pathname.split('/').filter(Boolean);
  slug.value = (pathParts[0] === 'status' ? pathParts[1] : pathParts[0]) || 'default';
  if (timer) clearTimeout(timer);
  const result = await fetchStatusResult(slug.value);
  if (disposed) return;
  isRefreshing.value = false;
  pageState.value = result.state;
  statusData.value = result.state === 'ready' ? result.data : null;
  isLoading.value = false;
  if (result.state !== 'not-found') timer = setTimeout(loadData, 30000);
}

onMounted(() => {
  loadData();
  icon = document.createElement('link');
  icon.rel = 'icon';
  icon.type = 'image/svg+xml';
  icon.href = outpipeMark;
  document.head.appendChild(icon);
});

onUnmounted(() => {
  disposed = true;
  icon?.remove();
  if (timer) clearTimeout(timer);
});
</script>

<template>
  <div class="status-app">
    <StatusHeader :can-subscribe="Boolean(statusData)" @open-subscribe="isSubscribeOpen = true" />
    <main class="status-container status-main">
      <section class="product-intro" aria-labelledby="status-title">
        <p class="product-eyebrow">Outpipe status</p>
        <h1 id="status-title">{{ statusData?.title || 'Service status' }}</h1>
        <p>{{ statusData?.description || 'Current availability and incident updates, in one place.' }}</p>
      </section>
    <div v-if="isLoading" style="display: flex; justify-content: center; align-items: center; min-height: 50vh;">
      <span style="color: var(--text-muted); font-size: 0.875rem;">Loading status...</span>
    </div>

    <div v-else-if="statusData">
      <div>
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
      </div>

      <SubscribeModal
        v-if="isSubscribeOpen"
        :slug="slug"
        @close="isSubscribeOpen = false"
      />
    </div>
    <section v-else class="status-empty" aria-live="polite">
      <h2>{{ pageState === 'not-found' ? 'This status page is not published' : 'Status is temporarily unavailable' }}</h2>
      <p v-if="pageState === 'not-found'">Check the link or contact the page owner. If you manage this page, publish it from Uptime in your dashboard.</p>
      <p v-else>We could not load the latest service status. Try again in a moment.</p>
      <button type="button" class="btn-subscribe" :disabled="isRefreshing" @click="loadData">{{ isRefreshing ? 'Checking...' : 'Check again' }}</button>
    </section>
    </main>
    <footer class="product-footer">
      <p>Service availability, powered by Outpipe.</p>
      <nav class="product-footer-links" aria-label="Footer navigation">
        <a :href="siteUrl">Outpipe</a>
        <a :href="`${siteUrl}/docs`">Docs</a>
        <a :href="`${siteUrl}/contact`">Support</a>
      </nav>
    </footer>
  </div>
</template>
