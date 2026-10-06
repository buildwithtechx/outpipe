import { Link, useLocation } from '@tanstack/react-router';
import { Menu } from 'lucide-react';
import { getWorkspaceProducts, navItemIsActive } from './workspace-products';

export function MobileProductNav({
  orgSlug,
  open,
  onOpen,
}: {
  orgSlug: string;
  open: boolean;
  onOpen: () => void;
}) {
  const { pathname } = useLocation();
  const products = getWorkspaceProducts(orgSlug);
  const current = products.find((product) =>
    product.items.some((item) => navItemIsActive(pathname, item)),
  );
  const items =
    current?.items.slice(0, 3) ??
    products.flatMap((product) => product.items.slice(0, 1));
  return (
    <nav
      aria-label={current ? `${current.name} navigation` : 'Product navigation'}
      className="fixed inset-x-0 bottom-0 z-30 flex justify-around border-t border-border bg-card/95 px-2 pt-2 pb-[max(.5rem,env(safe-area-inset-bottom))] backdrop-blur md:hidden"
    >
      {items.map((item) => (
        <Link
          key={item.to}
          to={item.to}
          aria-current={navItemIsActive(pathname, item) ? 'page' : undefined}
          className={`flex min-h-11 flex-1 flex-col items-center justify-center gap-1 text-[11px] ${navItemIsActive(pathname, item) ? 'text-foreground' : 'text-muted-foreground'}`}
        >
          <item.icon className="size-4" />
          <span className="max-w-full truncate px-1">
            {products.find((product) => product.items[0]?.to === item.to)
              ?.name ?? item.label}
          </span>
        </Link>
      ))}
      <button
        type="button"
        aria-expanded={open}
        aria-controls="mobile-workspace-navigation"
        onClick={onOpen}
        className="flex min-h-11 flex-1 flex-col items-center justify-center gap-1 text-[11px] text-muted-foreground"
      >
        <Menu className="size-4" />
        <span>All pages</span>
      </button>
    </nav>
  );
}
