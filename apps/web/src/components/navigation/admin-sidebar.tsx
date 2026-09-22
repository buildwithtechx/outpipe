import { Link, useLocation } from '@tanstack/react-router';
import { ArrowLeft, Server } from 'lucide-react';
import { useEffect } from 'react';
import { Badge } from '#/components/ui/badge';
import { useAdminOverview } from '#/features/admin/hooks/use-admin-resources';
import { ADMIN_NAV_ITEMS } from './constants';

interface AdminSidebarProps {
  mobileOpen: boolean;
  setMobileOpen: (open: boolean) => void;
}

export function AdminSidebar({ mobileOpen, setMobileOpen }: AdminSidebarProps) {
  const location = useLocation();
  const overview = useAdminOverview();

  useEffect(() => {
    if (!mobileOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setMobileOpen(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [mobileOpen, setMobileOpen]);

  const statusLabel = overview.isLoading
    ? 'Checking…'
    : overview.isError
      ? 'Degraded'
      : 'Operational';

  const statusBadgeClass = overview.isLoading
    ? 'border-purple-400/30 bg-purple-500/20 text-purple-300'
    : overview.isError
      ? 'border-rose-500/30 bg-rose-500/20 text-rose-300'
      : 'border-emerald-400/30 bg-emerald-500/20 text-emerald-300';

  const content = (
    <div className="flex h-full flex-col justify-between p-4 text-white">
      <div className="space-y-5">
        {/* Control Plane Status Card */}
        <div className="flex items-center justify-between gap-2 rounded-2xl border border-purple-500/20 bg-purple-950/20 p-3 shadow-xs">
          <div className="flex items-center gap-2.5 min-w-0">
            <div className="flex size-8 shrink-0 items-center justify-center rounded-xl bg-purple-600/30 border border-purple-400/40 text-purple-300">
              <Server className="size-4" />
            </div>
            <div className="min-w-0">
              <h4 className="truncate text-xs font-semibold text-purple-200">
                Control Plane
              </h4>
              <p className="truncate text-[10px] text-purple-300/60 font-mono">
                Global Edge Cluster
              </p>
            </div>
          </div>
          <Badge
            variant="outline"
            className={`shrink-0 text-[10px] py-0.5 px-1.5 ${statusBadgeClass}`}
          >
            {statusLabel}
          </Badge>
        </div>

        {/* Navigation Items */}
        <nav className="space-y-0.5" aria-label="Admin navigation">
          {ADMIN_NAV_ITEMS.map((item) => {
            const isActive = item.exact
              ? location.pathname === item.to ||
                location.pathname === `${item.to}/`
              : location.pathname.startsWith(item.to);
            const Icon = item.icon;

            return (
              <Link
                key={item.to}
                to={item.to}
                onClick={() => setMobileOpen(false)}
                className={`flex items-center gap-3 rounded-xl px-3 py-2 text-xs font-medium transition-all ${
                  isActive
                    ? 'bg-purple-600 text-white font-semibold shadow-xs'
                    : 'text-white/60 hover:bg-white/[0.06] hover:text-white'
                }`}
              >
                <Icon
                  className={`size-4 shrink-0 transition-colors ${
                    isActive ? 'text-white' : 'text-white/40'
                  }`}
                />
                <span className="truncate">{item.label}</span>
              </Link>
            );
          })}
        </nav>
      </div>

      {/* Footer Return Button */}
      <div className="pt-4 border-t border-white/10 space-y-2">
        <Link
          to="/select"
          onClick={() => setMobileOpen(false)}
          className="flex w-full items-center justify-between p-3 rounded-xl border border-white/10 bg-white/[0.02] text-white/70 hover:bg-white/[0.06] hover:text-white transition-all text-xs"
        >
          <div className="flex items-center gap-2">
            <ArrowLeft className="size-3.5" />
            <span className="font-medium">Exit to Workspace</span>
          </div>
        </Link>
      </div>
    </div>
  );

  return (
    <>
      {/* Desktop Fixed Sidebar */}
      <aside className="hidden md:flex w-64 border-r border-white/10 bg-neutral-950/70 flex-col justify-between shrink-0 h-full overflow-y-auto">
        {content}
      </aside>

      {/* Mobile Drawer Overlay */}
      {mobileOpen && (
        <div className="fixed inset-0 z-50 md:hidden flex">
          <div
            className="fixed inset-0 bg-black/75 backdrop-blur-sm animate-in fade-in duration-200"
            onClick={() => setMobileOpen(false)}
            aria-hidden="true"
          />
          <aside className="relative w-72 max-w-[80vw] bg-neutral-950 border-r border-white/15 flex flex-col justify-between h-full z-10 shadow-2xl overflow-y-auto animate-in slide-in-from-left duration-200">
            {content}
          </aside>
        </div>
      )}
    </>
  );
}
