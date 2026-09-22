import type { Domain } from '#/interfaces/domain';
import { apiClient } from '#/lib/api-client';

export function getDomains(organizationId: string) {
  return apiClient.get<Domain[]>(
    `/api/v1/organizations/${organizationId}/domains`,
  );
}

export function createDomain(
  organizationId: string,
  input: { hostname: string; verificationMethod: string; tunnelId?: string },
) {
  return apiClient.post<{ domain: Domain; verificationToken: string }>(
    `/api/v1/organizations/${organizationId}/domains`,
    input,
  );
}

export function verifyDomain(domainId: string, token: string) {
  return apiClient.post<void>(`/api/v1/domains/${domainId}/verify`, { token });
}
