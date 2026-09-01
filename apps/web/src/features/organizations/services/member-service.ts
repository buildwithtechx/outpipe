import type { OrganizationMember } from '#/interfaces/organization';
import { apiClient } from '#/lib/api-client';

export function getMembers(organizationId: string) {
  return apiClient.get<OrganizationMember[]>(
    `/api/v1/organizations/${organizationId}/members`,
  );
}

export function inviteMember(
  organizationId: string,
  input: { email: string; role: 'admin' | 'member' | 'viewer' },
) {
  return apiClient.post(
    `/api/v1/organizations/${organizationId}/invitations`,
    input,
  );
}

export function removeMember(organizationId: string, memberId: string) {
  return apiClient.delete<void>(
    `/api/v1/organizations/${organizationId}/members/${memberId}`,
  );
}
