<script setup lang="ts">
import { ref } from 'vue';
import { subscribeToStatus } from '../api';

const props = defineProps<{
  slug: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const subType = ref<'email' | 'webhook'>('email');
const target = ref('');
const isSubmitting = ref(false);
const errorMsg = ref('');
const isSuccess = ref(false);

async function handleSubmit() {
  if (!target.value.trim()) {
    errorMsg.value = subType.value === 'email' ? 'Please enter a valid email address' : 'Please enter a valid webhook URL';
    return;
  }
  isSubmitting.value = true;
  errorMsg.value = '';
  try {
    const ok = await subscribeToStatus(props.slug, subType.value, target.value.trim());
    if (ok) {
      isSuccess.value = true;
    } else {
      errorMsg.value = 'Failed to subscribe. Please try again.';
    }
  } catch {
    errorMsg.value = 'Network error while subscribing.';
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div class="modal-backdrop" @click.self="emit('close')">
    <div class="modal-content" role="dialog" aria-modal="true" aria-labelledby="subscribe-title">
      <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 0.5rem;">
        <h2 id="subscribe-title" class="modal-title">Subscribe to updates</h2>
        <button
          type="button"
          style="background: none; border: none; color: var(--text-muted); cursor: pointer; font-size: 1.25rem;"
          aria-label="Close modal"
          @click="emit('close')"
        >
          &times;
        </button>
      </div>
      <p class="modal-desc">
        Get real-time notifications whenever an incident is reported, updated, or resolved.
      </p>

      <div v-if="isSuccess" style="text-align: center; padding: 1.5rem 0;">
        <div style="width: 44px; height: 44px; background-color: rgba(16, 185, 129, 0.15); border-radius: 50%; display: flex; align-items: center; justify-content: center; margin: 0 auto 1rem; color: #10b981;">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <polyline points="20 6 9 17 4 12" />
          </svg>
        </div>
        <h3 style="font-size: 1.125rem; font-weight: 600; margin-bottom: 0.5rem;">Subscribed!</h3>
        <p style="font-size: 0.875rem; color: var(--text-secondary); margin-bottom: 1.5rem;">
          You will now receive automated updates for Outpipe services.
        </p>
        <button type="button" class="btn-subscribe" style="width: 100%; justify-content: center;" @click="emit('close')">
          Done
        </button>
      </div>

      <form v-else @submit.prevent="handleSubmit">
        <div style="display: flex; gap: 0.5rem; margin-bottom: 1.25rem; background-color: var(--bg-input); padding: 4px; border-radius: var(--radius-md);">
          <button
            type="button"
            :style="{
              flex: 1,
              padding: '0.4rem',
              borderRadius: '6px',
              border: 'none',
              fontSize: '0.8125rem',
              fontWeight: 600,
              cursor: 'pointer',
              backgroundColor: subType === 'email' ? 'var(--border-primary)' : 'transparent',
              color: subType === 'email' ? '#fff' : 'var(--text-muted)'
            }"
            @click="subType = 'email'"
          >
            Email
          </button>
          <button
            type="button"
            :style="{
              flex: 1,
              padding: '0.4rem',
              borderRadius: '6px',
              border: 'none',
              fontSize: '0.8125rem',
              fontWeight: 600,
              cursor: 'pointer',
              backgroundColor: subType === 'webhook' ? 'var(--border-primary)' : 'transparent',
              color: subType === 'webhook' ? '#fff' : 'var(--text-muted)'
            }"
            disabled
            title="Webhook subscriptions are not available"
          >
            Webhook
          </button>
        </div>

        <div style="margin-bottom: 1.25rem;">
          <label for="subscriber-target" style="display: block; font-size: 0.8125rem; font-weight: 500; color: var(--text-secondary); margin-bottom: 0.5rem;">
            {{ subType === 'email' ? 'Email Address' : 'Webhook Endpoint URL' }}
          </label>
          <input
            id="subscriber-target"
            v-model="target"
            :type="subType === 'email' ? 'email' : 'url'"
            :placeholder="subType === 'email' ? 'devops@example.com' : 'https://api.example.com/webhooks/status'"
            style="width: 100%; height: 42px; background-color: var(--bg-input); border: 1px solid var(--border-primary); border-radius: var(--radius-md); padding: 0 0.875rem; color: #fff; font-size: 0.875rem;"
            required
          />
        </div>

        <div v-if="errorMsg" style="color: #fda4af; font-size: 0.8125rem; margin-bottom: 1rem;">
          {{ errorMsg }}
        </div>

        <div style="display: flex; gap: 0.75rem;">
          <button
            type="button"
            style="flex: 1; height: 42px; background: none; border: 1px solid var(--border-primary); color: var(--text-secondary); border-radius: var(--radius-md); cursor: pointer; font-size: 0.875rem;"
            @click="emit('close')"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="btn-subscribe"
            :disabled="isSubmitting"
            style="flex: 2; height: 42px; justify-content: center;"
          >
            {{ isSubmitting ? 'Subscribing...' : 'Subscribe' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
