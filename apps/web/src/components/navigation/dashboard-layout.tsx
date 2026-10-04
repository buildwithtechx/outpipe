import { Link, Outlet, useParams } from '@tanstack/react-router';
import { Activity, Cable, Layers, Menu, Radio } from 'lucide-react';
import { useState } from 'react';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { CommandPalette } from './command-palette';
import { DashboardHeader } from './dashboard-header';
import { DashboardSidebar } from './dashboard-sidebar';

export function DashboardLayout({ children }: { children?: React.ReactNode }) {
  const params = useParams({ strict: false }) as { orgSlug?: string };
  const orgSlug = params.orgSlug || '';
  const [mobileOpen, setMobileOpen] = useState(false);
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const { data: session } = useAuthSession();

  return (
    <div className="flex h-screen flex-col bg-black text-white antialiased overflow-hidden">
      {/* Top Header */}
      <DashboardHeader
        mobileOpen={mobileOpen}
        setMobileOpen={setMobileOpen}
        orgSlug={orgSlug}
        onOpenCommandPalette={() => setCommandPaletteOpen(true)}
      />

      {/* Main Container with Sidebar + Dynamic Page Content */}
      <div className="flex flex-1 min-h-0 overflow-hidden relative">
        <DashboardSidebar
          orgSlug={orgSlug}
          mobileOpen={mobileOpen}
          setMobileOpen={setMobileOpen}
        />

        <main className="flex-1 min-w-0 overflow-y-auto bg-black p-4 sm:p-6 lg:p-8 pb-20 md:pb-8">
          <div className="mx-auto max-w-6xl">{children || <Outlet />}</div>
        </main>
      </div>

      {/* Mobile Bottom Navigation Bar */}
      <nav
        aria-label="Mobile Navigation"
        className="md:hidden fixed bottom-0 left-0 right-0 z-30 flex items-center justify-around border-t border-white/10 bg-black/95 backdrop-blur-md px-2 py-2"
      >
        <Link
          to="/$orgSlug"
          params={{ orgSlug }}
          activeProps={{ className: 'text-indigo-400 font-semibold' }}
          className="flex flex-col items-center gap-1 text-[11px] text-white/60 hover:text-white transition-colors"
        >
          <Layers className="size-4" />
          <span>Overview</span>
        </Link>
        <Link
          to="/$orgSlug/tunnels"
          params={{ orgSlug }}
          activeProps={{ className: 'text-indigo-400 font-semibold' }}
          className="flex flex-col items-center gap-1 text-[11px] text-white/60 hover:text-white transition-colors"
        >
          <Cable className="size-4" />
          <span>Tunnels</span>
        </Link>
        <Link
          to="/$orgSlug/requests"
          params={{ orgSlug }}
          activeProps={{ className: 'text-indigo-400 font-semibold' }}
          className="flex flex-col items-center gap-1 text-[11px] text-white/60 hover:text-white transition-colors"
        >
          <Radio className="size-4" />
          <span>Requests</span>
        </Link>
        <Link
          to="/$orgSlug/observability"
          params={{ orgSlug }}
          activeProps={{ className: 'text-indigo-400 font-semibold' }}
          className="flex flex-col items-center gap-1 text-[11px] text-white/60 hover:text-white transition-colors"
        >
          <Activity className="size-4" />
          <span>Traces</span>
        </Link>
        <button
          type="button"
          onClick={() => setMobileOpen(true)}
          className="flex flex-col items-center gap-1 text-[11px] text-white/60 hover:text-white transition-colors"
        >
          <Menu className="size-4" />
          <span>More</span>
        </button>
      </nav>

      {/* Command Palette Modal */}
      <CommandPalette
        open={commandPaletteOpen}
        onOpenChange={setCommandPaletteOpen}
        orgSlug={orgSlug}
        isPlatformAdmin={Boolean(session?.isPlatformAdmin)}
      />
    </div>
  );
}
