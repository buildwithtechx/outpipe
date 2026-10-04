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

export interface CreateSecretInput {
  key: string;
  value: string;
  comment?: string;
}
