import { useOrganization } from "#/features/organizations/hooks/use-organization";
import { useUsageSnapshot } from "./hooks/use-usage-snapshot";

export function UsagePage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const organizationId = organizationQuery.organization?.id;

  const query = useUsageSnapshot(organizationId);

  if (organizationQuery.isLoading || query.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading usage…</p>;
  }

  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  ) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load usage for this workspace.
      </p>
    );
  }

  const organization = organizationQuery.organization;
  const snapshot = query.data;

  const cards = [
    ["Requests", snapshot?.requestCount ?? 0],
    ["Errors", snapshot?.errorCount ?? 0],
    ["Connections", snapshot?.activeConnections ?? 0],
    ["Bandwidth", formatBytes(snapshot?.bandwidthBytes ?? 0)],
  ] as const;

  return (
    <div className="w-full max-w-6xl space-y-6 pb-12 text-white">
      <header className="border-b border-white/10 pb-6">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-indigo-400">
          {organization.name}
        </p>
        <h1 className="text-3xl font-semibold tracking-[-0.04em]">Usage</h1>
        <p className="mt-1 text-sm text-white/55">
          A snapshot of traffic and capacity for the current period.
        </p>
      </header>
      <section className="grid gap-3 pt-6 sm:grid-cols-2 lg:grid-cols-4">
        {cards.map(([label, value]) => (
          <div
            key={label}
            className="rounded-2xl border border-white/10 bg-white/2.5 p-5"
          >
            <p className="text-sm text-white/50">{label}</p>
            <p className="mt-5 text-3xl font-semibold">{value}</p>
          </div>
        ))}
      </section>
    </div>
  );
}

function formatBytes(value: number) {
  if (value < 1024) return `${value} B`;
  if (value < 1024 ** 2) return `${(value / 1024).toFixed(1)} KB`;
  if (value < 1024 ** 3) return `${(value / 1024 ** 2).toFixed(1)} MB`;
  return `${(value / 1024 ** 3).toFixed(1)} GB`;
}
