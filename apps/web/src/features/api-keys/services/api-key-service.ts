import type { APIKey } from '#/interfaces/identity';
import { apiClient } from '#/lib/api-client';

export function getApiKeys(organizationId: string) {
  return apiClient.get<APIKey[]>(
    `/api/v1/organizations/${organizationId}/api-keys`,
  );
}
