import type {
  CreateSecretInput,
  SecretEnvironment,
  SecretItem,
  SecretProject,
  SecretShare,
} from '#/interfaces';
import { apiClient } from '#/lib/api-client';

export type {
  CreateSecretInput,
  SecretEnvironment,
  SecretItem,
  SecretProject,
  SecretShare,
};

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

export function rollbackSecretVersion(
  organizationId: string,
  secretId: string,
  version: number,
) {
  return apiClient.post<SecretItem>(
    `/api/v1/organizations/${organizationId}/secrets/${secretId}/rollback`,
    { version },
  );
}

export function batchDeleteSecrets(
  organizationId: string,
  secretIds: string[],
) {
  return apiClient.post<void>(
    `/api/v1/organizations/${organizationId}/secrets/batch-delete`,
    { secretIds },
  );
}

export function batchMoveSecrets(
  organizationId: string,
  projectId: string,
  sourceEnvironmentId: string,
  targetEnvironmentId: string,
  secretIds: string[],
  conflictMode: 'skip' | 'overwrite',
) {
  return apiClient.post<void>(
    `/api/v1/organizations/${organizationId}/secrets/projects/${projectId}/environments/${sourceEnvironmentId}/move`,
    {
      targetEnvironmentId,
      secretIds,
      conflictMode,
    },
  );
}

export function createSecretShare(
  organizationId: string,
  input: {
    encryptedPayload: string;
    expiresAt: string;
    maxViews: number;
  },
) {
  return apiClient.post<SecretShare>(
    `/api/v1/organizations/${organizationId}/secrets/shares`,
    input,
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
