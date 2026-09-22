import type {
  WebhookDelivery,
  WebhookSubscription,
} from '#/interfaces/webhook';
import { apiClient } from '#/lib/api-client';

export function getWebhooks(organizationId: string) {
  return apiClient.get<WebhookSubscription[]>(
    `/api/v1/organizations/${organizationId}/webhooks`,
  );
}

export function createWebhook(
  organizationId: string,
  input: { name: string; url: string; events: string[] },
) {
  return apiClient.post<{ subscription: WebhookSubscription; secret: string }>(
    `/api/v1/organizations/${organizationId}/webhooks`,
    input,
  );
}

export function deleteWebhook(organizationId: string, webhookId: string) {
  return apiClient.delete<void>(
    `/api/v1/organizations/${organizationId}/webhooks/${webhookId}`,
  );
}

export function getWebhookDeliveries(
  organizationId: string,
  webhookId: string,
) {
  return apiClient.get<WebhookDelivery[]>(
    `/api/v1/organizations/${organizationId}/webhooks/${webhookId}/deliveries`,
  );
}
