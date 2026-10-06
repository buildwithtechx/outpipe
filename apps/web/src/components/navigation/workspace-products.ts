import { getWorkspaceNavItems, type NavGroup } from './constants';

export function getWorkspaceProducts(orgSlug: string): NavGroup[] {
  const items = getWorkspaceNavItems(orgSlug);
  return ['Tunnels', 'Observability', 'Secrets', 'Uptime'].map((name) => ({
    name,
    items: items.filter((item) => item.product === name),
  }));
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
