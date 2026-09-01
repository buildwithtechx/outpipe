import type { PaginatedResponse } from '#/interfaces/api';
import type { AuthUser } from '#/interfaces/auth';
import type { Subscription } from '#/interfaces/billing';
import type { Organization } from '#/interfaces/organization';
import type { Tunnel } from '#/interfaces/tunnel';
import { apiClient } from '#/lib/api-client';

export type AdminOverview = {
  users: number;
  organizations: number;
  tunnels: number;
  subscriptions: number;
};
export type AdminUsage = {
  bandwidthBytes: number;
  requestCount: number;
  errorCount: number;
};
export function getAdminOverview() {
  return apiClient.get<AdminOverview>('/api/v1/admin/overview');
}
export function getAdminUsage() {
  return apiClient.get<AdminUsage>('/api/v1/admin/usage');
}
export function getAdminUser(userId: string) {
  return apiClient.get<AuthUser>(`/api/v1/admin/users/${userId}`);
}
export function getAdminOrganization(organizationId: string) {
  return apiClient.get<Organization>(
    `/api/v1/admin/organizations/${organizationId}`,
  );
}
export function getAdminUsers() {
  return apiClient.get<PaginatedResponse<AuthUser>>(
    '/api/v1/admin/users?limit=50',
  );
}
export function getAdminOrganizations() {
  return apiClient.get<PaginatedResponse<Organization>>(
    '/api/v1/admin/organizations?limit=50',
  );
}
export function getAdminTunnels() {
  return apiClient.get<PaginatedResponse<Tunnel>>(
    '/api/v1/admin/tunnels?limit=50',
  );
}
export function getAdminSubscriptions() {
  return apiClient.get<PaginatedResponse<Subscription>>(
    '/api/v1/admin/subscriptions?limit=50',
  );
}
export function setAdminUserStatus(
  userId: string,
  status: 'active' | 'disabled',
) {
  return apiClient.patch<void>(`/api/v1/admin/users/${userId}/status`, {
    status,
  });
}
