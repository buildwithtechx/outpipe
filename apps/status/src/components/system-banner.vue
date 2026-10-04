<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  systemStatus: 'operational' | 'degraded' | 'outage';
  systemStatusMessage: string;
  lastUpdated: string;
}>();

const formattedTime = computed(() => {
  try {
    return new Date(props.lastUpdated).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  } catch {
    return 'Just now';
  }
});
</script>

<template>
  <div :class="['status-banner', systemStatus]">
    <div class="status-banner-left">
      <div :class="['status-dot', systemStatus]" />
      <span class="status-banner-text">{{ systemStatusMessage }}</span>
    </div>
    <span class="status-banner-time">Updated {{ formattedTime }}</span>
  </div>
</template>
