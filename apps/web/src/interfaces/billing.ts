export type SubscriptionStatus =
  | 'active'
  | 'trialing'
  | 'past_due'
  | 'paused'
  | 'canceled'
  | 'expired';

export type BillingProvider = 'polar' | 'paystack' | 'internal';

import type { Entity } from '#/interfaces/api';

export type Plan = Entity & {
  key: string;
  name: string;
  description?: string;
  priceMinor: number;
  currency: string;
  billingInterval: string;
  maxTunnels: number;
  maxDomains: number;
  maxMembers: number;
  maxConnections: number;
  bandwidthBytes: number;
  retentionDays: number;
  features: string;
  active: boolean;
};

export type Subscription = Entity & {
  organizationId: string;
  planId: string;
  provider: BillingProvider;
  status: SubscriptionStatus;
  customerId?: string;
  billingInterval: string;
  currentPeriodEnd?: string;
  cancelAtPeriodEnd: boolean;
  canceledAt?: string;
  trialEndsAt?: string;
};
