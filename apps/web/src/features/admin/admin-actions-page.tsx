import { Button } from '#/components/ui/button';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { AdminShell } from './admin-overview-page';
import { useAdminUserStatus, useAdminUsers } from './hooks/use-admin-resources';

export function AdminActionsPage() {
  const users = useAdminUsers();
  const status = useAdminUserStatus();
  const { user: currentUser } = useAuthSession();

  if (users.isLoading) {
    return (
      <p className="p-8 text-sm text-white/55">Loading available actions…</p>
    );
  }

  if (users.isError) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load available actions.
      </p>
    );
  }

  return (
    <AdminShell
      title="Admin actions"
      subtitle="Controlled account operations are recorded by the platform audit trail."
    >
      <div className="overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {users.data?.items.slice(0, 10).map((user) => {
          const isCurrentAdmin = user.id === currentUser?.id;
          return (
            <div
              key={user.id}
              className="flex flex-wrap items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
            >
              <div>
                <p className="text-sm text-white/80">
                  {user.email}{' '}
                  {isCurrentAdmin && (
                    <span className="ml-2 rounded-full border border-purple-500/30 bg-purple-500/20 px-2 py-0.5 text-[10px] text-purple-300">
                      You
                    </span>
                  )}
                </p>
                <p className="mt-1 text-xs text-white/40">
                  Current status: {user.status}
                </p>
              </div>
              {isCurrentAdmin && user.status === 'active' ? (
                <div className="flex flex-col items-start sm:items-end gap-1">
                  <Button
                    type="button"
                    variant="outline"
                    disabled
                    aria-describedby={`self-disable-hint-${user.id}`}
                  >
                    Active
                  </Button>
                  <span
                    id={`self-disable-hint-${user.id}`}
                    className="text-[11px] text-white/40"
                  >
                    You cannot disable your own administrator account.
                  </span>
                </div>
              ) : (
                <Button
                  type="button"
                  variant="outline"
                  disabled={status.isPending}
                  onClick={() =>
                    status.mutate({
                      userId: user.id,
                      status: user.status === 'active' ? 'disabled' : 'active',
                    })
                  }
                >
                  {user.status === 'active'
                    ? 'Disable account'
                    : 'Enable account'}
                </Button>
              )}
            </div>
          );
        })}
      </div>
    </AdminShell>
  );
}
