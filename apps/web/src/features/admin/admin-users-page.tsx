import { Button } from '#/components/ui/button';
import { useAuthSession } from '#/features/auth/hooks/use-auth-session';
import { AdminShell } from './admin-overview-page';
import { useAdminUserStatus, useAdminUsers } from './hooks/use-admin-resources';

export function AdminUsersPage() {
  const query = useAdminUsers();
  const status = useAdminUserStatus();
  const { user: currentUser } = useAuthSession();

  if (query.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading platform users…</p>;
  }

  if (query.isError) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load platform users.
      </p>
    );
  }

  return (
    <AdminShell
      title="Users"
      subtitle={`${query.data?.total ?? 0} accounts across the platform.`}
    >
      {status.isError && (
        <div className="mb-4 rounded-xl border border-rose-500/20 bg-rose-500/10 p-4 text-xs text-rose-300">
          Failed to update user status:{' '}
          {status.error instanceof Error
            ? status.error.message
            : 'An error occurred'}
        </div>
      )}
      <div className="overflow-hidden rounded-2xl border border-white/10 bg-white/2.5">
        {query.data?.items.map((user) => {
          const isCurrentAdmin = user.id === currentUser?.id;
          return (
            <div
              key={user.id}
              className="flex flex-wrap items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
            >
              <div>
                <p className="text-sm text-white/85">
                  {user.name || 'Unnamed user'}{' '}
                  {isCurrentAdmin && (
                    <span className="ml-2 rounded-full border border-purple-500/30 bg-purple-500/20 px-2 py-0.5 text-[10px] text-purple-300">
                      You
                    </span>
                  )}
                </p>
                <p className="mt-1 text-xs text-white/45">{user.email}</p>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-xs text-white/45">{user.status}</span>
                {isCurrentAdmin ? (
                  <div className="flex flex-col items-end gap-1">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled
                      aria-describedby={`user-self-disable-hint-${user.id}`}
                    >
                      Active
                    </Button>
                    <span
                      id={`user-self-disable-hint-${user.id}`}
                      className="text-[10px] text-white/40"
                    >
                      You cannot disable your own administrator account.
                    </span>
                  </div>
                ) : (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    disabled={status.isPending}
                    onClick={() =>
                      status.mutate({
                        userId: user.id,
                        status:
                          user.status === 'active' ? 'disabled' : 'active',
                      })
                    }
                  >
                    {user.status === 'active' ? 'Disable' : 'Enable'}
                  </Button>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </AdminShell>
  );
}
