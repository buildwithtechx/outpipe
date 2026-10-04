export interface CreateSharePayload {
  ciphertext: string;
  iv: string;
  keyVerifier: string;
  passwordSalt?: string;
  passwordVerifier?: string;
  contentFormat: string;
  ttlSeconds: number;
  maxViews: number;
}

export interface CreateShareResponse {
  id: string;
  expiresAt: string;
  maxViews: number;
}

export interface ShareMetaResponse {
  id: string;
  contentFormat: string;
  expiresAt: string;
  maxViews: number;
  viewsRemaining: number;
  needsPassword: boolean;
}

export interface RevealShareResponse {
  ciphertext: string;
  iv: string;
  contentFormat: string;
}

export function getApiBaseUrl(): string {
  const metaEnv = (import.meta as { env?: Record<string, string> }).env;
  if (metaEnv?.PUBLIC_API_URL) {
    return metaEnv.PUBLIC_API_URL;
  }
  if (typeof window !== 'undefined') {
    if (
      window.location.hostname === 'localhost' &&
      window.location.port === '4321'
    ) {
      return 'http://localhost:8080/api/v1';
    }
    return `${window.location.origin}/api/v1`;
  }
  return 'http://localhost:8080/api/v1';
}

export async function createShareApi(
  payload: CreateSharePayload,
): Promise<CreateShareResponse> {
  const url = `${getApiBaseUrl()}/shares`;
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(
      `Failed to create secret share: ${errorText || response.statusText}`,
    );
  }

  return response.json() as Promise<CreateShareResponse>;
}

export async function getShareMetaApi(id: string): Promise<ShareMetaResponse> {
  const url = `${getApiBaseUrl()}/shares/${encodeURIComponent(id)}`;
  const response = await fetch(url, {
    method: 'GET',
    headers: {
      Accept: 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error(
      'This secret link is invalid, expired, or was already burned.',
    );
  }

  return response.json() as Promise<ShareMetaResponse>;
}

export async function revealShareApi(
  id: string,
  verifier: string,
  passwordVerifier?: string,
): Promise<RevealShareResponse> {
  const url = `${getApiBaseUrl()}/shares/${encodeURIComponent(id)}/reveal`;
  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify({
      verifier,
      passwordVerifier,
    }),
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(
      errorText ||
        'Failed to reveal secret. Verification failed or secret expired.',
    );
  }

  return response.json() as Promise<RevealShareResponse>;
}
