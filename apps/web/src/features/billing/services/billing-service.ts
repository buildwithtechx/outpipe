import type { Plan, Subscription } from '#/interfaces/billing';
import type { Invoice } from '#/interfaces/billing-records';
import { apiClient } from '#/lib/api-client';

export function getBilling(organizationId: string) {
  return apiClient.get<{ plan: Plan; subscription: Subscription }>(
    `/api/v1/organizations/${organizationId}/billing`,
  );
}

export function getInvoices(organizationId: string) {
  return apiClient.get<{ invoices: Invoice[] }>(
    `/api/v1/organizations/${organizationId}/billing/invoices`,
  );
}

export function getBillingPortal(organizationId: string) {
  return apiClient.get<{ url: string }>(
    `/api/v1/organizations/${organizationId}/billing/portal`,
  );
}

export function cancelBilling(organizationId: string) {
  return apiClient.post<void>(
    `/api/v1/organizations/${organizationId}/billing/cancel`,
  );
}

export function resumeBilling(organizationId: string) {
  return apiClient.post<void>(
    `/api/v1/organizations/${organizationId}/billing/resume`,
  );
}
