import { Activity, Cable, LineChart, Lock } from 'lucide-react';
import type { NavGroup } from './constants';

export function getWorkspaceProducts(orgSlug: string): NavGroup[] {
  const base = `/${orgSlug}`;
  return [
    {
      name: 'Tunnels',
      items: [
        { label: 'Overview', to: base, icon: Cable, exact: true },
        { label: 'Tunnels', to: `${base}/tunnels`, icon: Cable },
        { label: 'Requests', to: `${base}/requests`, icon: Activity },
        { label: 'Agents', to: `${base}/agents`, icon: Activity },
        { label: 'Domains', to: `${base}/domains`, icon: Cable },
      ],
    },
    {
      name: 'Observability',
      items: [
        {
          label: 'Traces, logs & captures',
          to: `${base}/observability`,
          icon: Activity,
        },
      ],
    },
    {
      name: 'Secrets',
      items: [
        { label: 'Vaults & environments', to: `${base}/secrets`, icon: Lock },
      ],
    },
    {
      name: 'Uptime',
      items: [
        {
          label: 'Monitors & incidents',
          to: `${base}/uptime`,
          icon: LineChart,
          exact: true,
        },
        {
          label: 'Status page',
          to: `${base}/uptime/status-page`,
          icon: LineChart,
        },
      ],
    },
  ];
}

export function navItemIsActive(
  path: string,
  item: { to: string; exact?: boolean },
) {
  const normalized = path.replace(/\/+$/, '');
  return (
    normalized === item.to ||
    (!item.exact && normalized.startsWith(`${item.to}/`))
  );
}
