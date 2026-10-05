import { describe, expect, it, vi } from 'vitest';
import type { TanStackStartTunnel } from '../src/interfaces/options';
import {
  normalizeTanStackRoute,
  tunnelLifecycle,
  tunnelStatus,
} from '../src/services/middleware';

describe('TanStack Start tunnel integration', () => {
  it('normalizes parameterized route segments', () => {
    expect(normalizeTanStackRoute('/users/42/details')).toBe(
      '/users/:id/details',
    );
    expect(normalizeTanStackRoute('/tunnels/01ARZ3NDEKTSV4RRFFQ69G5FAV')).toBe(
      '/tunnels/:id',
    );
    expect(normalizeTanStackRoute('/healthz')).toBe('/healthz');
  });

  it('returns current tunnel status response', async () => {
    const tunnel = createTunnel();
    const response = tunnelStatus(tunnel);
    expect(response.status).toBe(200);
    const body = await response.json();
    expect(body).toEqual({ status: 'idle' });
  });

  it('starts tunnel during lifecycle and calls next handler', async () => {
    const tunnel = createTunnel();
    const next = vi.fn().mockResolvedValue(new Response('ok'));
    const response = await tunnelLifecycle(tunnel, next);
    expect(next).toHaveBeenCalledOnce();
    expect(await response.text()).toBe('ok');
  });

  it('skips automatic startup when autoStart is disabled', async () => {
    const tunnel = createTunnel();
    tunnel.autoStart = false;
    tunnel.start = vi.fn();
    await tunnelLifecycle(tunnel, () => new Response('ok'));
    expect(tunnel.start).not.toHaveBeenCalled();
  });
});

function createTunnel(): TanStackStartTunnel {
  return {
    start: async () => ({ status: 'active' }),
    stop: async () => undefined,
    state: () => ({ status: 'idle' }),
  };
}
