import { Link } from '@tanstack/react-router';
import {
  BookOpen,
  Command,
  LogOut,
  Menu,
  Search,
  ShieldCheck,
  X,
} from 'lucide-react';
import { useState } from 'react';
import { BrandLockup } from '#/components/layout';
import { Button } from '#/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '#/components/ui/dialog';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { useLogout } from '#/features/auth/hooks/use-logout';
import { WorkspaceSwitcher } from './workspace-switcher';

interface DashboardHeaderProps {
  mobileOpen: boolean;
  setMobileOpen: (open: boolean) => void;
  orgSlug: string;
  onOpenCommandPalette: () => void;
}

export function DashboardHeader({
  mobileOpen,
  setMobileOpen,
  orgSlug,
  onOpenCommandPalette,
}: DashboardHeaderProps) {
  const { data: session } = useAuthSession();
  const [logoutDialogOpen, setLogoutDialogOpen] = useState(false);
  const logoutMutation = useLogout();

  const isPlatformAdmin = Boolean(session?.isPlatformAdmin);
  const user = session?.user;

  return (
    <>
      <header className="sticky top-0 z-40 flex h-16 w-full items-center justify-between border-b border-white/10 bg-black/80 px-3 sm:px-6 backdrop-blur-xl">
        {/* Left Section: Mobile toggle, Brand, and Workspace Switcher */}
        <div className="flex items-center gap-2 sm:gap-4 min-w-0">
          <button
            type="button"
            onClick={() => setMobileOpen(!mobileOpen)}
            className="flex size-9 items-center justify-center rounded-xl border border-white/10 bg-white/5 text-white/70 hover:bg-white/10 hover:text-white transition md:hidden shrink-0"
            aria-label="Toggle navigation menu"
          >
            {mobileOpen ? (
              <X className="size-5" />
            ) : (
              <Menu className="size-5" />
            )}
          </button>

          <BrandLockup
            className="shrink-0 hidden xs:inline-flex"
            iconClassName="size-8 rounded-xl"
            nameClassName="text-base font-bold tracking-tight text-white hidden sm:inline"
          />

          <div className="h-5 w-px bg-white/10 hidden sm:block shrink-0" />

          <WorkspaceSwitcher currentOrgSlug={orgSlug} />
        </div>

        {/* Right Section: Search, Docs, Admin Indicator, User & Logout */}
        <div className="flex items-center gap-2 sm:gap-3 shrink-0">
          {/* Quick Search / Command Palette Button */}
          <button
            type="button"
            onClick={onOpenCommandPalette}
            className="hidden sm:flex items-center gap-2 rounded-xl border border-white/10 bg-white/3 px-3 py-1.5 text-xs text-white/50 transition hover:border-white/20 hover:bg-white/6 hover:text-white"
            title="Search sections (⌘K)"
          >
            <Search className="size-3.5" />
            <span className="hidden md:inline text-[11px]">Jump to...</span>
            <kbd className="inline-flex items-center gap-0.5 rounded border border-white/10 bg-white/5 px-1 py-0.5 text-[9px] font-mono text-white/40">
              <Command className="size-2.5" />K
            </kbd>
          </button>

          {/* Docs Shortcut */}
          <Link
            to="/docs/$"
            params={{ _splat: '' }}
            className="hidden lg:flex items-center gap-1.5 rounded-xl border border-white/10 bg-white/2 px-2.5 py-1.5 text-xs text-white/60 transition hover:border-white/20 hover:bg-white/5 hover:text-white"
            title="Read Documentation"
          >
            <BookOpen className="size-3.5 text-indigo-300" />
            <span>Docs</span>
          </Link>

          {/* Superadmin Quick Access */}
          {isPlatformAdmin && (
            <Link
              to="/admin"
              className="flex items-center gap-1.5 rounded-xl border border-purple-400/30 bg-purple-500/10 px-2.5 py-1.5 text-xs font-semibold text-purple-300 transition hover:bg-purple-500/20"
              title="Platform Admin Control Plane"
            >
              <ShieldCheck className="size-3.5" />
              <span className="hidden sm:inline">Admin</span>
            </Link>
          )}

          {/* User Status & Sign Out Button */}
          <div className="flex items-center gap-2 pl-2 border-l border-white/10">
            <div className="hidden xl:flex flex-col text-right">
              <span className="text-xs font-semibold text-white/90 truncate max-w-30">
                {user?.name || user?.email?.split('@')[0] || 'User'}
              </span>
              <span className="text-[10px] text-white/40 truncate max-w-30">
                {user?.email || ''}
              </span>
            </div>

            <button
              type="button"
              onClick={() => setLogoutDialogOpen(true)}
              className="flex size-9 items-center justify-center rounded-xl border border-white/10 bg-white/3 text-white/60 hover:border-rose-400/30 hover:bg-rose-500/10 hover:text-rose-300 transition shrink-0"
              title="Sign Out"
              aria-label="Sign Out"
            >
              <LogOut className="size-4" />
            </button>
          </div>
        </div>
      </header>

      {/* Logout Confirmation Dialog */}
      <Dialog open={logoutDialogOpen} onOpenChange={setLogoutDialogOpen}>
        <DialogContent className="max-w-md bg-neutral-900 border-white/15 text-white">
          <DialogHeader>
            <DialogTitle className="text-lg font-bold text-white flex items-center gap-2">
              <LogOut className="size-5 text-rose-400" />
              Confirm Sign Out
            </DialogTitle>
            <DialogDescription className="text-white/55 text-xs mt-1.5">
              Are you sure you want to sign out of your Outpipe account? Any
              active local CLI sessions with stored credentials will remain
              valid until revoked.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="mt-4 gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setLogoutDialogOpen(false)}
              disabled={logoutMutation.isPending}
              className="border-white/10 bg-white/5 text-white hover:bg-white/10 text-xs"
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="destructive"
              size="sm"
              onClick={() => logoutMutation.mutate()}
              disabled={logoutMutation.isPending}
              className="bg-rose-600 hover:bg-rose-700 text-white text-xs font-medium"
            >
              {logoutMutation.isPending ? 'Signing Out...' : 'Sign Out'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
