<script setup lang="ts">
import outpipeMark from '@outpipe/shared/outpipe.svg?url';
import { publicSiteUrl } from '@outpipe/shared';

const siteUrl = publicSiteUrl(import.meta.env.DEV);

defineProps<{
  canSubscribe: boolean;
}>();

const emit = defineEmits<{
  (e: 'open-subscribe'): void;
}>();
</script>

<template>
  <header class="product-header">
    <div class="product-header-inner">
      <a :href="siteUrl" class="product-brand" aria-label="Outpipe home">
        <img :src="outpipeMark" alt="" width="40" height="40" />
        <span>Outpipe</span>
        <span class="product-label">Status</span>
      </a>
      <nav class="product-nav" aria-label="Product navigation">
        <a :href="`${siteUrl}/docs`">Docs</a>
        <a :href="`${siteUrl}/contact`">Support</a>
      <button
        v-if="canSubscribe"
        type="button"
        class="btn-subscribe"
        @click="emit('open-subscribe')"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
          <path d="M13.73 21a2 2 0 0 1-3.46 0" />
        </svg>
        Subscribe
      </button>
      </nav>
    </div>
  </header>
</template>
