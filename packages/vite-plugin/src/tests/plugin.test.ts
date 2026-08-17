import { describe, expect, it } from 'vitest';
import { outpipeTunnel } from '../services/plugin';

describe('outpipeTunnel', () => {
  it('is enabled only for Vite development by default', () => {
    const plugin = outpipeTunnel({
      relayUrl: 'wss://relay.test',
      agentToken: 'token',
    });
    expect(plugin.name).toBe('outpipe');
    expect(plugin.apply).toBe('serve');
  });

  it('does not configure a server when disabled', () => {
    const plugin = outpipeTunnel({
      relayUrl: 'wss://relay.test',
      agentToken: 'token',
      enabled: false,
    });
    const configureServer = plugin.configureServer;
    if (typeof configureServer !== 'function') {
      throw new Error('configureServer hook is required');
    }
    expect(configureServer.call({} as never, {} as never)).toBeUndefined();
  });
});
