import {
  Activity,
  BookOpen,
  Cable,
  CreditCard,
  Globe,
  Key,
  Layers,
  LineChart,
  Radio,
  ScrollText,
  Settings,
  Shield,
  Users,
  Webhook,
  Zap,
} from 'lucide-react';
import type { ComponentType } from 'react';

export interface NavItem {
  label: string;
  to: string;
  icon: ComponentType<{ className?: string }>;
  exact?: boolean;
  desc?: string;
}

export function getWorkspaceNavItems(orgSlug: string): NavItem[] {
  return [
    {
      label: 'Overview',
      to: `/${orgSlug}`,
      icon: Layers,
      exact: true,
      desc: 'Workspace summary & connection health',
    },
    {
      label: 'Tunnels',
      to: `/${orgSlug}/tunnels`,
      icon: Cable,
      desc: 'Manage active tunnels & ports',
    },
    {
      label: 'Agents',
      to: `/${orgSlug}/agents`,
      icon: Radio,
      desc: 'Connected Outpipe CLI agents',
    },
    {
      label: 'Custom Domains',
      to: `/${orgSlug}/domains`,
      icon: Globe,
      desc: 'Custom hostnames & certificates',
    },
    {
      label: 'Live Requests',
      to: `/${orgSlug}/requests`,
      icon: Activity,
      desc: 'Realtime tunnel traffic stream',
    },
    {
      label: 'Usage',
      to: `/${orgSlug}/usage`,
      icon: ScrollText,
      desc: 'Bandwidth & concurrent connection limits',
    },
    {
      label: 'Billing',
      to: `/${orgSlug}/billing`,
      icon: CreditCard,
      desc: 'Subscription tier & invoice history',
    },
    {
      label: 'Members',
      to: `/${orgSlug}/members`,
      icon: Users,
      desc: 'Team members, roles & invitations',
    },
    {
      label: 'API Keys',
      to: `/${orgSlug}/api-keys`,
      icon: Key,
      desc: 'Service tokens for CLI & CI/CD workflows',
    },
    {
      label: 'Webhooks',
      to: `/${orgSlug}/webhooks`,
      icon: Webhook,
      desc: 'Workspace lifecycle event dispatches',
    },
    {
      label: 'Audit Logs',
      to: `/${orgSlug}/audit-logs`,
      icon: ScrollText,
      desc: 'Security event logs & activity trail',
    },
    {
      label: 'Settings',
      to: `/${orgSlug}/settings`,
      icon: Settings,
      desc: 'Workspace settings & danger zone',
    },
  ];
}

export const ADMIN_NAV_ITEMS: NavItem[] = [
  {
    label: 'Platform Overview',
    to: '/admin',
    icon: Shield,
    exact: true,
    desc: 'Platform control plane metrics',
  },
  {
    label: 'User Accounts',
    to: '/admin/users',
    icon: Users,
    desc: 'Global user accounts & sessions',
  },
  {
    label: 'Organizations',
    to: '/admin/organizations',
    icon: Layers,
    desc: 'All platform workspaces & owners',
  },
  {
    label: 'Global Tunnels',
    to: '/admin/tunnels',
    icon: Cable,
    desc: 'All active tunnels & public hostnames',
  },
  {
    label: 'Subscriptions',
    to: '/admin/subscriptions',
    icon: CreditCard,
    desc: 'Customer subscription tiers & status',
  },
  {
    label: 'System Usage',
    to: '/admin/usage',
    icon: Activity,
    desc: 'Global relay traffic & edge capacity',
  },
  {
    label: 'Telemetry & Charts',
    to: '/admin/charts',
    icon: LineChart,
    desc: 'Platform analytics & visual charts',
  },
  {
    label: 'Audit Logs',
    to: '/admin/audit-logs',
    icon: ScrollText,
    desc: 'Global security audit trail',
  },
  {
    label: 'Control Actions',
    to: '/admin/actions',
    icon: Zap,
    desc: 'Operations, cache invalidations & health',
  },
];

export const EXTERNAL_NAV_ITEMS: NavItem[] = [
  {
    label: 'Documentation',
    to: '/docs/$',
    icon: BookOpen,
    desc: 'Guides, CLI manual & architecture',
  },
];
