import type { UsageEvent } from '#/interfaces/usage';
import { apiClient } from '#/lib/api-client';

export function getRequestEvents(organizationId: string) {
  return apiClient.get<{ events: UsageEvent[] }>(
    `/api/v1/organizations/${organizationId}/usage/requests?limit=100`,
  );
}
