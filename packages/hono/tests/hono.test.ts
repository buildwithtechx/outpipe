import { describe, expect, it, vi } from 'vitest';
import type { HonoTunnel } from '../src/interfaces/options';
import {
  normalizeHonoRoute,
  tunnelLifecycle,
  tunnelStatus,
} from '../src/services/middleware';

describe('Hono tunnel middleware', () => {
  it('normalizes parameterized route segments', () => {
    expect(normalizeHonoRoute('/users/123/profile')).toBe('/users/:id/profile');
    expect(
      normalizeHonoRoute('/tunnels/550e8400-e29b-41d4-a716-446655440000'),
    ).toBe('/tunnels/:id');
    expect(normalizeHonoRoute('/status')).toBe('/status');
  });

  it('returns current tunnel state via JSON', () => {
    const tunnel = createTunnel();
    const json = vi.fn();
    const c = { json } as never;
    tunnelStatus(tunnel)(c, (() => undefined) as never);
    expect(json).toHaveBeenCalledWith({ status: 'idle' });
  });

  it('starts the tunnel before continuing', async () => {
    const tunnel = createTunnel();
    const next = vi.fn();
    const c = {} as never;
    await tunnelLifecycle(tunnel)(c, next);
    expect(next).toHaveBeenCalledOnce();
  });
});

function createTunnel(): HonoTunnel {
  return {
    start: async () => ({ status: 'active' }),
    stop: async () => undefined,
    state: () => ({ status: 'idle' }),
  };
}
