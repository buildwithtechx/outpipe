import type { AuthSession, OAuthProvider } from '#/interfaces/auth';
import type { Organization } from '#/interfaces/organization';
import { apiClient, getApiBaseURL } from '#/lib/api-client';

export function getAuthSession() {
  return apiClient.get<AuthSession>('/api/v1/auth/session');
}

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

export function getLastOrganizationSlug() {
  if (typeof window === 'undefined') return null;
  return window.localStorage.getItem('outpipe_last_organization');
}

export function getAuthReturnTo(defaultPath = '/select') {
  if (typeof window === 'undefined') return defaultPath;
  const requested = new URLSearchParams(window.location.search).get(
    'return_to',
  );
  return requested?.startsWith('/') && !requested.startsWith('//')
    ? requested
    : defaultPath;
}

export function startOAuthSignIn(
  provider: OAuthProvider,
  returnTo = '/select',
) {
  const url = new URL(`/api/v1/auth/oauth/${provider}`, `${getApiBaseURL()}/`);
  url.searchParams.set('return_to', returnTo);
  window.location.assign(url);
}
