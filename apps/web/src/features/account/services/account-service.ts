import type { AuthUser } from '#/interfaces/auth';
import { apiClient } from '#/lib/api-client';

export function getAccount() {
  return apiClient.get<AuthUser>('/api/v1/account');
}

export function deleteAccount() {
  return apiClient.delete<void>('/api/v1/account');
}

export function transferOrganizationOwnership(
  organizationId: string,
  newOwnerId: string,
) {
  return apiClient.post<void>(
    `/api/v1/organizations/${organizationId}/transfer`,
    { newOwnerId },
  );
}
