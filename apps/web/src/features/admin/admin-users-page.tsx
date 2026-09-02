import { Button } from '#/components/ui/button';
import { AdminShell } from './admin-overview-page';
import { useAdminUserStatus, useAdminUsers } from './hooks/use-admin-resources';
export function AdminUsersPage() {
  const query = useAdminUsers();
  const status = useAdminUserStatus();

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
      <div className="overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {query.data?.items.map((user) => (
          <div
            key={user.id}
            className="flex flex-wrap items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
          >
            <div>
              <p className="text-sm text-white/85">
                {user.name || 'Unnamed user'}
              </p>
              <p className="mt-1 text-xs text-white/45">{user.email}</p>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-xs text-white/45">{user.status}</span>
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={status.isPending}
                onClick={() =>
                  status.mutate({
                    userId: user.id,
                    status: user.status === 'active' ? 'disabled' : 'active',
                  })
                }
              >
                {user.status === 'active' ? 'Disable' : 'Enable'}
              </Button>
            </div>
          </div>
        ))}
      </div>
    </AdminShell>
  );
}
