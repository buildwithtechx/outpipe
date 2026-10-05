import { beforeEach, expect, it, vi } from 'vitest';
import { createHonoTunnel } from '../src/services/tunnel';

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

it('opens a fresh tunnel after stopping an active tunnel', async () => {
  const tunnel = createHonoTunnel({
    relayUrl: 'ws://example.com',
    localPort: 3000,
  });
  await tunnel.start();
  await tunnel.stop();
  expect(tunnel.state()).toEqual({ status: 'closed' });
  await tunnel.start();
  expect(relay.openTunnel).toHaveBeenCalledTimes(2);
});

it('closes state even when closeTunnel fails', async () => {
  const tunnel = createHonoTunnel({
    relayUrl: 'ws://example.com',
    localPort: 3000,
  });
  await tunnel.start();
  relay.closeTunnel.mockRejectedValueOnce(new Error('close failed'));
  await expect(tunnel.stop()).rejects.toThrow('close failed');
  expect(tunnel.state()).toEqual({ status: 'closed' });
  expect(relay.close).toHaveBeenCalledOnce();
});
