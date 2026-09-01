import type { OrganizationMember } from '#/interfaces/organization';
import { apiClient } from '#/lib/api-client';

export function getMembers(organizationId: string) {
  return apiClient.get<OrganizationMember[]>(
    `/api/v1/organizations/${organizationId}/members`,
  );
}
