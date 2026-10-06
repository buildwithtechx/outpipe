import { Link } from '@tanstack/react-router';
import { ChevronDown } from 'lucide-react';
import { useEffect, useId, useState } from 'react';
import type { NavGroup } from './constants';
import { navItemIsActive } from './workspace-products';

export function ProductNavigation({
  product,
  orgSlug,
  collapsed,
  query,
  pathname,
  onNavigate,
}: {
  product: NavGroup;
  orgSlug: string;
  collapsed: boolean;
  query: string;
  pathname: string;
  onNavigate?: () => void;
}) {
  const active = product.items.some((item) => navItemIsActive(pathname, item));
  const [open, setOpen] = useState(active);
  const [ready, setReady] = useState(false);
  const id = useId();
  const storageKey = `outpipe.navigation.${orgSlug}.${product.name}`;
  useEffect(() => {
    try {
      const saved = localStorage.getItem(storageKey);
      setOpen(saved === null ? active : saved === 'true');
    } catch {
      setOpen(active);
    }
    setReady(true);
  }, [storageKey, active]);
  useEffect(() => {
    if (ready) {
      try {
        localStorage.setItem(storageKey, String(open));
      } catch {}
    }
  }, [open, ready, storageKey]);
  const items = product.items.filter(
    (item) =>
      !query ||
      product.name.toLowerCase().includes(query) ||
      item.label.toLowerCase().includes(query),
  );
  if (!items.length) return null;
  const Icon = product.items[0]?.icon;
  if (collapsed)
    return (
      <Link
        onClick={onNavigate}
        to={product.items[0]?.to ?? '/select'}
        title={product.name}
        aria-label={product.name}
        aria-current={
          product.items[0] && navItemIsActive(pathname, product.items[0])
            ? 'page'
            : undefined
        }
        className={`flex h-11 items-center justify-center rounded-lg ${active ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent'}`}
      >
        {Icon && <Icon className="size-5" />}
      </Link>
    );
  const expanded = open || Boolean(query);
  return (
    <section className="space-y-1">
      <button
        type="button"
        aria-expanded={expanded}
        aria-controls={id}
        onClick={() => {
          if (!query) setOpen(!open);
        }}
        className="flex min-h-11 w-full items-center gap-3 rounded-lg px-3 text-left text-sm font-medium hover:bg-accent"
      >
        {Icon && <Icon className="size-4" />}
        <span className="flex-1">{product.name}</span>
        <ChevronDown
          className={`size-4 text-muted-foreground motion-safe:transition-transform ${expanded ? '' : '-rotate-90'}`}
        />
      </button>
      <div id={id} hidden={!expanded} className="space-y-1">
        {items.map((item) => (
          <Link
            onClick={onNavigate}
            key={item.to}
            to={item.to}
            aria-current={navItemIsActive(pathname, item) ? 'page' : undefined}
            className={`block min-h-10 rounded-lg py-2.5 pl-10 pr-3 text-sm ${navItemIsActive(pathname, item) ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent hover:text-foreground'}`}
          >
            {item.label}
          </Link>
        ))}
      </div>
    </section>
  );
}
