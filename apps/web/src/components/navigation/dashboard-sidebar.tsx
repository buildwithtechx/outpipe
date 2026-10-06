import { Link, useLocation } from '@tanstack/react-router';
import { PanelLeftClose, PanelLeftOpen, Search } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { Button } from '#/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from '#/components/ui/dialog';
import { Input } from '#/components/ui/input';
import { getWorkspaceNavGroups } from './constants';
import { ProductNavigation } from './product-navigation';
import { getWorkspaceProducts, navItemIsActive } from './workspace-products';

export function DashboardSidebar({
  orgSlug,
  mobileOpen,
  setMobileOpen,
}: {
  orgSlug: string;
  mobileOpen: boolean;
  setMobileOpen: (open: boolean) => void;
}) {
  const { pathname } = useLocation();
  const [collapsed, setCollapsed] = useState(false);
  const [search, setSearch] = useState('');
  const previousPath = useRef(pathname);
  useEffect(() => {
    try {
      setCollapsed(
        localStorage.getItem('outpipe.sidebar.collapsed') === 'true',
      );
    } catch {}
  }, []);
  useEffect(() => {
    if (previousPath.current === pathname) return;
    previousPath.current = pathname;
    setMobileOpen(false);
    setSearch('');
  }, [pathname, setMobileOpen]);
  const products = getWorkspaceProducts(orgSlug);
  const workspace = getWorkspaceNavGroups(orgSlug).find(
    (group) => group.name === 'Settings & Administration',
  );
  const query = search.trim().toLowerCase();
  const compact = collapsed && !mobileOpen;
  const content = (
    <div className="flex h-full flex-col gap-5 p-3">
      <div className="flex items-center justify-between gap-2">
        {!compact && (
          <span className="text-xs font-medium text-muted-foreground">
            Workspace navigation
          </span>
        )}
        <Button
          className="hidden md:inline-flex"
          variant="ghost"
          size="icon"
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          onClick={() => {
            const next = !collapsed;
            setCollapsed(next);
            try {
              localStorage.setItem('outpipe.sidebar.collapsed', String(next));
            } catch {}
          }}
        >
          {collapsed ? <PanelLeftOpen /> : <PanelLeftClose />}
        </Button>
      </div>
      {!compact && (
        <div className="relative">
          <Search className="absolute left-3 top-3 size-4 text-muted-foreground" />
          <Input
            aria-label="Search navigation"
            placeholder="Find a page…"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            className="pl-9"
          />
        </div>
      )}
      <nav
        aria-label="Workspace navigation"
        className="min-h-0 flex-1 space-y-4 overflow-y-auto"
      >
        {products.map((product) => (
          <ProductNavigation
            key={orgSlug + product.name}
            product={product}
            orgSlug={orgSlug}
            collapsed={compact}
            query={query}
            pathname={pathname}
          />
        ))}
        <div className="space-y-1 border-t border-border pt-4">
          {!compact && (
            <p className="px-3 py-2 text-xs text-muted-foreground">Workspace</p>
          )}
          {workspace?.items
            .filter(
              (item) => !query || item.label.toLowerCase().includes(query),
            )
            .map((item) => (
              <Link
                key={item.to}
                to={item.to}
                title={item.label}
                aria-label={item.label}
                aria-current={
                  navItemIsActive(pathname, item) ? 'page' : undefined
                }
                className={`flex min-h-10 items-center gap-3 rounded-lg px-3 text-sm ${navItemIsActive(pathname, item) ? 'bg-accent text-foreground' : 'text-muted-foreground hover:bg-accent hover:text-foreground'}`}
              >
                <item.icon className="size-4 shrink-0" />
                {!compact && item.label}
              </Link>
            ))}
        </div>
        {query &&
          !products.some(
            (product) =>
              product.name.toLowerCase().includes(query) ||
              product.items.some((item) =>
                item.label.toLowerCase().includes(query),
              ),
          ) &&
          !workspace?.items.some((item) =>
            item.label.toLowerCase().includes(query),
          ) && (
            <p role="status" className="p-3 text-sm text-muted-foreground">
              No matching pages.
            </p>
          )}
      </nav>
      {!compact && (
        <Link
          to="/docs/$"
          params={{ _splat: 'installation' }}
          className="rounded-xl border border-border p-3 text-sm text-muted-foreground hover:text-foreground"
        >
          Install the CLI →
        </Link>
      )}
    </div>
  );
  return (
    <>
      <aside
        className={`hidden h-full shrink-0 border-r border-border bg-card md:block ${collapsed ? 'w-20' : 'w-64'}`}
      >
        {content}
      </aside>
      <Dialog open={mobileOpen} onOpenChange={setMobileOpen}>
        <DialogContent
          id="mobile-workspace-navigation"
          className="inset-y-0 left-0 top-0 h-dvh w-[min(90vw,22rem)] max-w-none translate-x-0 translate-y-0 rounded-none p-0"
        >
          <DialogTitle className="sr-only">Workspace navigation</DialogTitle>
          <DialogDescription className="sr-only">
            Browse products and workspace settings.
          </DialogDescription>
          {content}
        </DialogContent>
      </Dialog>
    </>
  );
}
