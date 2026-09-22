import { Outlet, useParams } from '@tanstack/react-router';
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

        <main className="flex-1 min-w-0 overflow-y-auto bg-black p-4 sm:p-6 lg:p-8">
          <div className="mx-auto max-w-6xl">{children || <Outlet />}</div>
        </main>
      </div>

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
