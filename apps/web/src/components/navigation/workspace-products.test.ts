import { expect, it } from 'vitest';
import { getWorkspaceProducts, navItemIsActive } from './workspace-products';

it('matches nested pages without matching another product or organization', () => {
  const item = { to: '/acme/tunnels' };
  expect(navItemIsActive('/acme/tunnels/id', item)).toBe(true);
  expect(navItemIsActive('/acme/tunnels/', item)).toBe(true);
  expect(navItemIsActive('/acme/tunnels-other', item)).toBe(false);
  expect(navItemIsActive('/other/tunnels', item)).toBe(false);
  expect(navItemIsActive('/acme/tunnels', { to: '/acme', exact: true })).toBe(
    false,
  );
});

it('makes status page editing distinct from the monitor overview', () => {
  const uptime = getWorkspaceProducts('acme').find(
    (product) => product.name === 'Uptime',
  );
  expect(
    uptime?.items
      .filter((item) => navItemIsActive('/acme/uptime/status-page', item))
      .map((item) => item.label),
  ).toEqual(['Status page']);
  expect(getWorkspaceProducts('acme').map((product) => product.name)).toEqual([
    'Tunnels',
    'Observability',
    'Secrets',
    'Uptime',
  ]);
});
