import type { Tunnel, TunnelAPIClientOptions } from '../interfaces/api';
import { TunnelAPIError } from '../utils/errors';

export class TunnelAPIClient {
  private readonly apiUrl: string;
  private readonly apiKey?: string;
  private readonly apiPrefix: string;
  private readonly request: typeof globalThis.fetch;

  constructor(options: TunnelAPIClientOptions) {
    this.apiUrl = options.apiUrl.replace(/\/$/, '');
    this.apiKey = options.apiKey;
    this.apiPrefix = `/${(options.apiPrefix ?? 'api/v1').replace(/^\/+|\/+$/g, '')}`;
    this.request = options.fetch ?? globalThis.fetch;
    if (!this.request) {
      throw new TunnelAPIError(0, 'fetch is not available in this runtime');
    }
  }

  async createTunnel(
    organizationId: string,
    request: Record<string, unknown>,
  ): Promise<Tunnel> {
    return this.call<Tunnel>(
      `/organizations/${encodeURIComponent(organizationId)}/tunnels`,
      { method: 'POST', body: JSON.stringify(request) },
    );
  }

  async listTunnels(organizationId: string): Promise<Tunnel[]> {
    return this.call<Tunnel[]>(
      `/organizations/${encodeURIComponent(organizationId)}/tunnels`,
      { method: 'GET' },
    );
  }

  async inspectTunnel(tunnelId: string): Promise<Tunnel> {
    return this.call<Tunnel>(`/tunnels/${encodeURIComponent(tunnelId)}`, {
      method: 'GET',
    });
  }

  async closeTunnel(tunnelId: string): Promise<void> {
    await this.call<void>(`/tunnels/${encodeURIComponent(tunnelId)}`, {
      method: 'DELETE',
    });
  }

  private async call<T>(path: string, init: RequestInit): Promise<T> {
    const headers = new Headers(init.headers);
    headers.set('Accept', 'application/json');
    if (init.body) {
      headers.set('Content-Type', 'application/json');
    }
    if (this.apiKey) {
      headers.set('Authorization', `Bearer ${this.apiKey}`);
    }
    const response = await this.request(
      `${this.apiUrl}${this.apiPrefix}${path}`,
      { ...init, headers },
    );
    if (!response.ok) {
      const body = await response.text();
      let message = body;
      try {
        const payload = JSON.parse(body) as {
          message?: unknown;
          error?: unknown;
        };
        if (typeof payload.message === 'string' && payload.message !== '') {
          message = payload.message;
        } else if (typeof payload.error === 'string' && payload.error !== '') {
          message = payload.error;
        }
      } catch {
        // Preserve non-JSON error bodies as the exception message.
      }
      throw new TunnelAPIError(response.status, message);
    }
    if (response.status === 204) {
      return undefined as T;
    }
    return (await response.json()) as T;
  }
}
