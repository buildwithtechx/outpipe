export type TunnelAPIClientOptions = {
  apiUrl: string;
  apiKey?: string;
  fetch?: typeof globalThis.fetch;
  apiPrefix?: string;
};

export type Tunnel = {
  id: string;
  publicHostname?: string;
  /** @deprecated Relay protocol messages use public_url; HTTP API responses use publicHostname. */
  public_url?: string;
  /** @deprecated Use publicHostname for HTTP API responses. */
  publicUrl?: string;
  status: string;
  [key: string]: unknown;
};
