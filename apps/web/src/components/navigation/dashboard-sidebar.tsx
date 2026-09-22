import { Link, useLocation } from '@tanstack/react-router';
import { Check, Copy, ShieldCheck, Terminal } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Badge } from '#/components/ui/badge';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { getWorkspaceNavItems } from './constants';

interface DashboardSidebarProps {
  orgSlug: string;
  mobileOpen: boolean;
  setMobileOpen: (open: boolean) => void;
}

export function DashboardSidebar({
  orgSlug,
  mobileOpen,
  setMobileOpen,
}: DashboardSidebarProps) {
  const location = useLocation();
  const { data: session } = useAuthSession();
  const { organization } = useOrganization(orgSlug);
  const [copied, setCopied] = useState(false);

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

  const isPlatformAdmin = Boolean(session?.isPlatformAdmin);
  const navItems = getWorkspaceNavItems(orgSlug);

  const copyCliCommand = async () => {
    await navigator.clipboard.writeText('outpipe http 3000');
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const content = (
    <div className="flex h-full flex-col justify-between p-4 text-white">
      <div className="space-y-5">
        {/* Workspace Quick Card */}
        {organization && (
          <div className="flex items-center justify-between gap-2 rounded-2xl border border-white/10 bg-white/2.5 p-3 shadow-xs">
            <div className="flex items-center gap-2.5 min-w-0">
              <div className="flex size-8 shrink-0 items-center justify-center rounded-xl bg-indigo-600/20 border border-indigo-400/30 text-indigo-300 font-bold text-xs">
                {organization.name.slice(0, 2).toUpperCase()}
              </div>
              <div className="min-w-0">
                <h4 className="truncate text-xs font-semibold text-white/90">
                  {organization.name}
                </h4>
                <p className="truncate text-[10px] text-white/40 font-mono">
                  /{organization.slug}
                </p>
              </div>
            </div>
            <Badge
              variant="outline"
              className="shrink-0 border-emerald-500/30 bg-emerald-500/10 text-[10px] text-emerald-400 py-0.5 px-1.5"
            >
              Active
            </Badge>
          </div>
        )}

        {/* Navigation Items */}
        <nav className="space-y-0.5" aria-label="Workspace navigation">
          {navItems.map((item) => {
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
                    ? 'bg-indigo-600 text-white font-semibold shadow-xs'
                    : 'text-white/60 hover:bg-white/6 hover:text-white'
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

          {/* Superadmin Quick Access in Sidebar */}
          {isPlatformAdmin && (
            <div className="pt-3 mt-3 border-t border-white/10 space-y-0.5">
              <span className="px-3 text-[10px] uppercase font-bold tracking-wider text-purple-400/80 flex items-center gap-1.5">
                <ShieldCheck className="size-3" /> Superadmin
              </span>
              <Link
                to="/admin"
                onClick={() => setMobileOpen(false)}
                className="flex items-center gap-3 rounded-xl px-3 py-2 text-xs font-medium text-purple-300 transition-all hover:bg-purple-500/15 hover:text-purple-200"
              >
                <ShieldCheck className="size-4 shrink-0 text-purple-400" />
                <span className="truncate">Control Plane</span>
              </Link>
            </div>
          )}
        </nav>
      </div>

      {/* Footer Quickstart Card */}
      <div className="pt-4 border-t border-white/10 space-y-2">
        <div className="rounded-xl border border-white/10 bg-white/2 p-2.5 text-xs">
          <div className="flex items-center justify-between text-white/50 text-[10px] font-mono mb-1.5">
            <span className="flex items-center gap-1">
              <Terminal className="size-3 text-indigo-300" /> Quick Tunnel
            </span>
            <button
              type="button"
              onClick={copyCliCommand}
              className="text-white/40 hover:text-white transition flex items-center gap-0.5"
              title="Copy CLI snippet"
            >
              {copied ? (
                <Check className="size-3 text-emerald-400" />
              ) : (
                <Copy className="size-3" />
              )}
            </button>
          </div>
          <code className="block rounded-lg bg-black/60 border border-white/5 p-1.5 font-mono text-[11px] text-indigo-200 truncate">
            outpipe http 3000
          </code>
        </div>
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
