import { Outlet, useParams } from '@tanstack/react-router';
import { useState } from 'react';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { CommandPalette } from './command-palette';
import { DashboardHeader } from './dashboard-header';
import { DashboardSidebar } from './dashboard-sidebar';
import { MobileProductNav } from './mobile-product-nav';

export function DashboardLayout({ children }: { children?: React.ReactNode }) {
  const params = useParams({ strict: false }) as { orgSlug?: string };
  const orgSlug = params.orgSlug || '';
  const [mobileOpen, setMobileOpen] = useState(false);
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const { data: session } = useAuthSession();
  return (
    <div className="dark product-surface flex h-dvh flex-col overflow-hidden bg-background text-foreground antialiased">
      <a
        href="#workspace-content"
        className="sr-only focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus:z-50 focus:rounded-lg focus:bg-primary focus:p-3 focus:text-primary-foreground"
      >
        Skip to content
      </a>
      <DashboardHeader
        mobileOpen={mobileOpen}
        setMobileOpen={setMobileOpen}
        orgSlug={orgSlug}
        onOpenCommandPalette={() => setCommandPaletteOpen(true)}
      />
      <div className="relative flex min-h-0 flex-1 overflow-hidden">
        <DashboardSidebar
          orgSlug={orgSlug}
          mobileOpen={mobileOpen}
          setMobileOpen={setMobileOpen}
        />
        <main
          id="workspace-content"
          tabIndex={-1}
          className="min-w-0 flex-1 overflow-y-auto p-4 pb-28 sm:px-6 sm:pt-6 md:pb-8 lg:p-8"
        >
          <div className="mx-auto max-w-6xl">{children || <Outlet />}</div>
        </main>
      </div>
      <MobileProductNav
        orgSlug={orgSlug}
        open={mobileOpen}
        onOpen={() => setMobileOpen(true)}
      />
      <CommandPalette
        open={commandPaletteOpen}
        onOpenChange={setCommandPaletteOpen}
        orgSlug={orgSlug}
        isPlatformAdmin={Boolean(session?.isPlatformAdmin)}
      />
    </div>
  );
}
