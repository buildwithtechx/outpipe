import type { Plan, Subscription } from '#/interfaces/billing';
import { apiClient } from '#/lib/api-client';

export function getBilling(organizationId: string) {
  return apiClient.get<{ plan: Plan; subscription: Subscription }>(
    `/api/v1/organizations/${organizationId}/billing`,
  );
}
