import { beforeEach, expect, it, vi } from 'vitest';
import { createTanStackStartTunnel } from '../src/services/tunnel';

const relay = vi.hoisted(() => ({
  openTunnel: vi.fn(),
  closeTunnel: vi.fn(),
  close: vi.fn(),
  on: vi.fn(),
}));
vi.mock('@outpipe/sdk', () => ({
  RelayConnection: class {
    openTunnel = relay.openTunnel;
    closeTunnel = relay.closeTunnel;
    close = relay.close;
    on = relay.on;
  },
}));

beforeEach(() => {
  vi.clearAllMocks();
  relay.openTunnel.mockResolvedValue({
    tunnel_id: 'tunnel',
    public_url: 'https://example.com',
  });
  relay.closeTunnel.mockResolvedValue(undefined);
});

it('allows explicit startup with autoStart disabled and restart after stop', async () => {
  const tunnel = createTanStackStartTunnel({
    relayUrl: 'ws://example.com',
    localPort: 3000,
    autoStart: false,
  });
  expect((await tunnel.start()).status).toBe('active');
  await tunnel.stop();
  expect(tunnel.state().status).toBe('closed');
  expect((await tunnel.start()).status).toBe('active');
  expect(relay.openTunnel).toHaveBeenCalledTimes(2);
});
