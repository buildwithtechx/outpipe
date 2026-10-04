import { apiClient } from '#/lib/api-client';

export interface SecretProject {
  id: string;
  organizationId: string;
  slug: string;
  name: string;
  description: string;
  createdById: string;
  createdAt: string;
  updatedAt: string;
}

export interface SecretEnvironment {
  id: string;
  organizationId: string;
  projectId: string;
  slug: string;
  name: string;
  createdById: string;
  createdAt: string;
  updatedAt: string;
}

export interface SecretItem {
  id: string;
  key: string;
  value?: string;
  comment: string;
  version: number;
  updatedAt: string;
}

export interface SecretShare {
  id: string;
  organizationId?: string;
  contentFormat: string;
  expiresAt: string;
  maxViews: number;
  views: number;
  revokedAt?: string;
  createdAt: string;
}

export function getProjects(organizationId: string) {
  return apiClient.get<SecretProject[]>(
    `/api/v1/organizations/${organizationId}/secrets/projects`,
  );
}

export function createProject(
  organizationId: string,
  input: { slug: string; name: string; description?: string },
) {
  return apiClient.post<SecretProject>(
    `/api/v1/organizations/${organizationId}/secrets/projects`,
    input,
  );
}

export function deleteProject(organizationId: string, projectId: string) {
  return apiClient.delete<void>(
    `/api/v1/organizations/${organizationId}/secrets/projects/${projectId}`,
  );
}

export function getEnvironments(organizationId: string, projectId: string) {
  return apiClient.get<SecretEnvironment[]>(
    `/api/v1/organizations/${organizationId}/secrets/projects/${projectId}/environments`,
  );
}

export function createEnvironment(
  organizationId: string,
  projectId: string,
  input: { slug: string; name: string },
) {
  return apiClient.post<SecretEnvironment>(
    `/api/v1/organizations/${organizationId}/secrets/projects/${projectId}/environments`,
    input,
  );
}

export function getSecrets(
  organizationId: string,
  projectId: string,
  environmentId: string,
  reveal = false,
) {
  return apiClient.get<SecretItem[]>(
    `/api/v1/organizations/${organizationId}/secrets/projects/${projectId}/environments/${environmentId}?reveal=${reveal}`,
  );
}

export function setSecret(
  organizationId: string,
  projectId: string,
  environmentId: string,
  input: { key: string; value: string; comment?: string },
) {
  return apiClient.post<SecretItem>(
    `/api/v1/organizations/${organizationId}/secrets/projects/${projectId}/environments/${environmentId}`,
    input,
  );
}

export function deleteSecret(organizationId: string, secretId: string) {
  return apiClient.delete<void>(
    `/api/v1/organizations/${organizationId}/secrets/${secretId}`,
  );
}

export function restoreSecret(organizationId: string, secretId: string) {
  return apiClient.post<void>(
    `/api/v1/organizations/${organizationId}/secrets/${secretId}/restore`,
    {},
  );
}

export function getShares(organizationId: string) {
  return apiClient.get<SecretShare[]>(
    `/api/v1/organizations/${organizationId}/secrets/shares`,
  );
}

export function revokeShare(organizationId: string, shareId: string) {
  return apiClient.delete<void>(
    `/api/v1/organizations/${organizationId}/secrets/shares/${shareId}`,
  );
}
