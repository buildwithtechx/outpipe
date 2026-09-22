import type { UsageSnapshot } from '#/interfaces/usage';
import { apiClient } from '#/lib/api-client';

export function getUsageSnapshot(organizationId: string) {
  return apiClient.get<UsageSnapshot>(
    `/api/v1/organizations/${organizationId}/usage/snapshot`,
  );
}
