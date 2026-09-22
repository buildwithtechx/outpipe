import { AdminShell } from './admin-overview-page';
import { useAdminUser } from './hooks/use-admin-resources';

export function AdminUserPage({ userId }: { userId: string }) {
  const query = useAdminUser(userId);

  if (query.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading user details…</p>;
  }

  if (query.isError || !query.data) {
    return (
      <p className="p-8 text-sm text-rose-200">We could not load this user.</p>
    );
  }

  const user = query.data;

  return (
    <AdminShell
      title={user.name || 'User details'}
      subtitle="Identity and account status."
    >
      <div className="grid gap-4 rounded-2xl border border-white/10 bg-white/[0.025] p-6 sm:grid-cols-2">
        <Detail label="Email" value={user.email} />
        <Detail label="Status" value={user.status} />
        <Detail label="Created" value={formatDate(user.createdAt)} />
        <Detail
          label="Last login"
          value={
            user.lastLoginAt ? formatDate(user.lastLoginAt) : 'Not recorded'
          }
        />
      </div>
    </AdminShell>
  );
}
function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs uppercase tracking-wider text-white/40">{label}</p>
      <p className="mt-2 text-sm text-white/80">{value}</p>
    </div>
  );
}
function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
