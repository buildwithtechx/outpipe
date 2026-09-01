import type { Entity } from '#/interfaces/api';

export type WebhookSubscription = Entity & {
  organizationId: string;
  name: string;
  url: string;
  events: string;
  lastDeliveredAt?: string;
};

export type WebhookDelivery = Entity & {
  subscriptionId: string;
  eventId: string;
  eventType: string;
  status: string;
  attempts: number;
  deliveredAt?: string;
};
