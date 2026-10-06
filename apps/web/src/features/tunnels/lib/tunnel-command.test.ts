import { expect, it } from 'vitest';
import { publicTunnelEndpoint, tunnelConnectCommand } from './tunnel-command';

it('connects an existing tunnel with flags accepted by the CLI', () => {
  expect(
    tunnelConnectCommand({
      id: 'tunnel_123',
      protocol: 'https',
      targetPort: 3000,
    }),
  ).toBe('outpipe open --port 3000 --protocol https --tunnel-id tunnel_123');
  expect(
    tunnelConnectCommand({ id: 'raw-id', protocol: 'udp', targetPort: 65535 }),
  ).toContain('--protocol udp');
});

it.each([0, -1, 65536, 1.5, Number.NaN])(
  'rejects invalid local port %s',
  (targetPort) => {
    expect(() =>
      tunnelConnectCommand({ id: 'valid', protocol: 'http', targetPort }),
    ).toThrow();
  },
);

it.each(['id; curl evil', 'id --agent-token=secret', '$(command)', ''])(
  'rejects unsafe tunnel id %s',
  (id) => {
    expect(() =>
      tunnelConnectCommand({ id, protocol: 'http', targetPort: 3000 }),
    ).toThrow();
  },
);

it('preserves public ports and waits for raw endpoints to be allocated', () => {
  expect(
    publicTunnelEndpoint({
      protocol: 'http',
      publicHostname: 'demo.outpipe.dev',
      publicPort: 8443,
    }),
  ).toBe('https://demo.outpipe.dev:8443');
  expect(
    publicTunnelEndpoint({
      protocol: 'https',
      publicHostname: 'demo.outpipe.dev',
      publicPort: 443,
    }),
  ).toBe('https://demo.outpipe.dev');
  expect(
    publicTunnelEndpoint({
      protocol: 'tcp',
      publicHostname: 'demo.outpipe.dev',
    }),
  ).toBeNull();
  expect(
    publicTunnelEndpoint({
      protocol: 'udp',
      publicHostname: 'demo.outpipe.dev',
      publicPort: 9000,
    }),
  ).toBe('demo.outpipe.dev:9000');
});
