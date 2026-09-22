import type { APIKey } from '#/interfaces/identity';
import { apiClient } from '#/lib/api-client';

export function getApiKeys(organizationId: string) {
  return apiClient.get<APIKey[]>(
    `/api/v1/organizations/${organizationId}/api-keys`,
  );
}

export function createApiKey(
  organizationId: string,
  input: { name: string; scopes: string[]; expiresAt?: string },
) {
  return apiClient.post<{ key: APIKey; token: string }>(
    `/api/v1/organizations/${organizationId}/api-keys`,
    input,
  );
}

export function revokeApiKey(organizationId: string, apiKeyId: string) {
  return apiClient.delete<void>(
    `/api/v1/organizations/${organizationId}/api-keys/${apiKeyId}`,
  );
}
