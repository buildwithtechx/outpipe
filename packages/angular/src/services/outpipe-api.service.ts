import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Inject, Injectable } from '@angular/core';
import type { Tunnel } from '@outpipe/sdk';
import type { Observable } from 'rxjs';
import type { OutpipeAngularConfig } from '../interfaces';
import { OUTPIPE_API_CONFIG } from '../tokens';

@Injectable({ providedIn: 'root' })
export class OutpipeApiService {
  private readonly apiUrl: string;
  private readonly apiPrefix: string;

  constructor(
    @Inject(HttpClient) private readonly http: HttpClient,
    @Inject(OUTPIPE_API_CONFIG) private readonly config: OutpipeAngularConfig,
  ) {
    this.apiUrl = config.apiUrl.replace(/\/$/, '');
    this.apiPrefix = `/${(config.apiPrefix ?? 'api/v1').replace(/^\/+|\/+$/g, '')}`;
  }

  createTunnel(
    organizationId: string,
    request: Record<string, unknown>,
  ): Observable<Tunnel> {
    return this.request<Tunnel>(
      'POST',
      `/organizations/${encodeURIComponent(organizationId)}/tunnels`,
      request,
    );
  }

  listTunnels(organizationId: string): Observable<Tunnel[]> {
    return this.request<Tunnel[]>(
      'GET',
      `/organizations/${encodeURIComponent(organizationId)}/tunnels`,
    );
  }

  inspectTunnel(tunnelId: string): Observable<Tunnel> {
    return this.request<Tunnel>(
      'GET',
      `/tunnels/${encodeURIComponent(tunnelId)}`,
    );
  }

  closeTunnel(tunnelId: string): Observable<void> {
    return this.request<void>(
      'DELETE',
      `/tunnels/${encodeURIComponent(tunnelId)}`,
    );
  }

  private request<T>(
    method: string,
    path: string,
    body?: Record<string, unknown>,
  ): Observable<T> {
    let headers = new HttpHeaders({ Accept: 'application/json' });
    if (body !== undefined) {
      headers = headers.set('Content-Type', 'application/json');
    }
    if (this.config.apiKey) {
      headers = headers.set('Authorization', `Bearer ${this.config.apiKey}`);
    }
    return this.http.request<T>(
      method,
      `${this.apiUrl}${this.apiPrefix}${path}`,
      {
        body,
        headers,
        observe: 'body',
      },
    );
  }
}
