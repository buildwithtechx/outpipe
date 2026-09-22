import { Link } from '@tanstack/react-router';
import { AdminShell } from './admin-overview-page';
import { useAdminOrganizations } from './hooks/use-admin-resources';

export function AdminOrganizationsPage() {
  const query = useAdminOrganizations();

  if (query.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading organizations…</p>;
  }

  if (query.isError) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load organizations.
      </p>
    );
  }

  return (
    <AdminShell
      title="Organizations"
      subtitle={`${query.data?.total ?? 0} workspaces registered on Outpipe.`}
    >
      <div className="grid gap-3">
        {query.data?.items.map((organization) => (
          <Link
            key={organization.id}
            to="/admin/organizations/$organizationID"
            params={{ organizationID: organization.id }}
            className="block rounded-2xl border border-white/10 bg-white/2.5 p-5 transition-colors hover:border-purple-500/30 hover:bg-white/5"
          >
            <div className="flex items-center justify-between gap-4">
              <p className="font-medium text-white">{organization.name}</p>
              <span className="font-mono text-xs text-indigo-200">
                {organization.slug}
              </span>
            </div>
            <p className="mt-2 text-xs text-white/40">
              Owner {organization.ownerId}
            </p>
          </Link>
        ))}
      </div>
    </AdminShell>
  );
}
