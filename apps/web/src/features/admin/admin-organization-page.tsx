import { AdminShell } from './admin-overview-page';
import { useAdminOrganization } from './hooks/use-admin-resources';
export function AdminOrganizationPage({
  organizationId,
}: {
  organizationId: string;
}) {
  const query = useAdminOrganization(organizationId);
  if (query.isLoading)
    return (
      <p className="p-8 text-sm text-white/55">Loading organization details…</p>
    );
  if (query.isError || !query.data)
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load this organization.
      </p>
    );
  const organization = query.data;
  return (
    <AdminShell
      title={organization.name}
      subtitle="Workspace identity and ownership details."
    >
      <div className="grid gap-4 rounded-2xl border border-white/10 bg-white/[0.025] p-6 sm:grid-cols-2">
        <Detail label="Slug" value={organization.slug} />
        <Detail label="Owner" value={organization.ownerId} />
        <Detail label="Created" value={formatDate(organization.createdAt)} />
        <Detail label="Organization ID" value={organization.id} />
      </div>
    </AdminShell>
  );
}
function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs uppercase tracking-wider text-white/40">{label}</p>
      <p className="mt-2 break-all text-sm text-white/80">{value}</p>
    </div>
  );
}
function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
