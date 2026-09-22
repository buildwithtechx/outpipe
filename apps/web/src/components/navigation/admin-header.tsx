import { Link } from "@tanstack/react-router";
import { ArrowLeft, LogOut, Menu, ShieldCheck, X } from "lucide-react";
import { useState } from "react";
import { BrandLockup } from "#/components/layout";
import { Button } from "#/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "#/components/ui/dialog";
import { useAuthSession } from "#/features/auth/hooks/use-auth-session";
import { useLogout } from "#/features/auth/hooks/use-logout";

interface AdminHeaderProps {
  mobileOpen: boolean;
  setMobileOpen: (open: boolean) => void;
}

export function AdminHeader({ mobileOpen, setMobileOpen }: AdminHeaderProps) {
  const { data: session } = useAuthSession();
  const [logoutDialogOpen, setLogoutDialogOpen] = useState(false);
  const logoutMutation = useLogout();
  const user = session?.user;

  return (
    <>
      <header className="sticky top-0 z-40 flex h-16 w-full items-center justify-between border-b border-white/10 bg-black/80 px-3 sm:px-6 backdrop-blur-xl">
        {/* Left Section: Mobile toggle, Brand, and Admin Badge */}
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
            className="shrink-0"
            iconClassName="size-8 rounded-xl"
            nameClassName="text-base font-bold tracking-tight text-white hidden sm:inline"
          />

          <div className="flex items-center gap-1.5 rounded-full border border-purple-400/30 bg-purple-500/10 px-2.5 py-1 text-xs font-semibold text-purple-300">
            <ShieldCheck className="size-3.5" />
            <span>Control Plane</span>
          </div>
        </div>

        {/* Right Section: Back to Workspace link, Admin User Info, Logout */}
        <div className="flex items-center gap-2 sm:gap-3 shrink-0">
          <Link
            to="/select"
            className="flex items-center gap-1.5 rounded-xl border border-white/10 bg-white/3 px-3 py-1.5 text-xs text-white/70 transition hover:border-white/20 hover:bg-white/6 hover:text-white"
          >
            <ArrowLeft className="size-3.5" />
            <span className="hidden sm:inline">User Workspace</span>
          </Link>

          <div className="flex items-center gap-2 pl-2 border-l border-white/10">
            <div className="hidden md:flex flex-col text-right">
              <span className="text-xs font-semibold text-white/90 truncate max-w-30">
                {user?.name || user?.email?.split("@")[0] || "Admin"}
              </span>
              <span className="text-[10px] text-purple-400 font-mono">
                Superadmin
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
              Are you sure you want to sign out of the Admin Control Plane?
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
              {logoutMutation.isPending ? "Signing Out..." : "Sign Out"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
