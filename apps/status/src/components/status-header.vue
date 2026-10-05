<script setup lang="ts">
defineProps<{
  title: string;
}>();

const emit = defineEmits<{
  (e: 'open-subscribe'): void;
}>();
</script>

<template>
  <header class="status-header">
    <div class="status-header-inner">
      <div class="brand-wrapper">
        <div class="brand-logo" aria-hidden="true">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#ffffff" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
          </svg>
        </div>
        <div>
          <h1 class="brand-title">{{ title }}</h1>
          <p class="brand-subtitle">Powered by Outpipe Network</p>
        </div>
      </div>
      <button
        type="button"
        class="btn-subscribe"
        @click="emit('open-subscribe')"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
          <path d="M13.73 21a2 2 0 0 1-3.46 0" />
        </svg>
        Subscribe to Updates
      </button>
    </div>
  </header>
</template>
