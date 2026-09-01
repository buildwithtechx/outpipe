import type { Organization } from '#/interfaces/organization';
import { apiClient } from '#/lib/api-client';

export function getOrganizations() {
  return apiClient.get<Organization[]>('/api/v1/organizations');
}

export function createOrganization(name: string, slug: string) {
  return apiClient.post<Organization>('/api/v1/organizations', {
    name,
    slug,
  });
}

export function checkOrganizationSlug(slug: string) {
  return apiClient.get<{ available: boolean }>(
    `/api/v1/organizations/slug-availability?slug=${encodeURIComponent(slug)}`,
  );
}
