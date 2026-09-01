import { Link } from '@tanstack/react-router';
import { Activity, ArrowRight, Cable, CircleAlert, Radio } from 'lucide-react';
import { Button } from '#/components/ui/button';
import { TunnelPageState } from '#/features/tunnels/components/tunnel-page-state';
import { TunnelStatusBadge } from '#/features/tunnels/components/tunnel-status-badge';
import { useOrganizationTunnels } from '#/features/tunnels/hooks/use-organization-tunnels';

export function OrganizationOverviewPage({ orgSlug }: { orgSlug: string }) {
  const { organization, organizationsQuery, tunnelsQuery } =
    useOrganizationTunnels(orgSlug);

  if (organizationsQuery.isLoading) {
    return <TunnelPageState label="Loading workspace…" />;
  }

  if (organizationsQuery.isError || !organization) {
    return (
      <main className="mx-auto flex min-h-[55vh] max-w-xl flex-col items-center justify-center px-6 text-center text-white">
        <CircleAlert className="size-5 text-rose-200" />
        <h1 className="mt-4 text-xl font-semibold">Workspace unavailable</h1>
        <p className="mt-2 text-sm leading-6 text-white/55">
          This workspace is not available to your account, or your session has
          expired.
        </p>
        <Button
          asChild
          variant="outline"
          className="mt-6 border-white/10 bg-transparent text-white hover:bg-white/10 hover:text-white"
        >
          <Link to="/login">Sign in again</Link>
        </Button>
      </main>
    );
  }

  if (tunnelsQuery.isLoading) {
    return <TunnelPageState label="Loading workspace activity…" />;
  }

  if (tunnelsQuery.isError) {
    return (
      <TunnelPageState error="We could not retrieve the workspace activity." />
    );
  }

  const tunnels = tunnelsQuery.data ?? [];
  const activeCount = tunnels.filter(
    (tunnel) => tunnel.status === 'active',
  ).length;
  const connectingCount = tunnels.filter(
    (tunnel) => tunnel.status === 'connecting',
  ).length;

  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="flex flex-col gap-5 border-b border-white/10 pb-8 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="mb-3 text-sm font-medium text-indigo-200">
            {organization.name}
          </p>
          <h1 className="text-3xl font-semibold tracking-[-0.04em] sm:text-4xl">
            Workspace overview
          </h1>
          <p className="mt-3 max-w-2xl text-sm leading-6 text-white/55 sm:text-base">
            A quiet view of the services your team has made reachable.
          </p>
        </div>
        <Button
          asChild
          className="rounded-full bg-white px-5 text-black hover:bg-indigo-100"
        >
          <Link to="/$orgSlug/tunnels" params={{ orgSlug }}>
            Manage tunnels
            <ArrowRight />
          </Link>
        </Button>
      </header>

      <section
        className="grid gap-3 pt-8 sm:grid-cols-3"
        aria-label="Tunnel summary"
      >
        <div className="rounded-2xl border border-white/10 bg-white/[0.025] p-5">
          <Cable className="size-5 text-indigo-200" />
          <p className="mt-6 text-3xl font-semibold tracking-tight">
            {tunnels.length}
          </p>
          <p className="mt-1 text-sm text-white/50">Configured tunnels</p>
        </div>
        <div className="rounded-2xl border border-emerald-300/15 bg-emerald-400/[0.04] p-5">
          <Activity className="size-5 text-emerald-200" />
          <p className="mt-6 text-3xl font-semibold tracking-tight">
            {activeCount}
          </p>
          <p className="mt-1 text-sm text-white/50">Active connections</p>
        </div>
        <div className="rounded-2xl border border-amber-300/15 bg-amber-400/[0.04] p-5">
          <Radio className="size-5 text-amber-100" />
          <p className="mt-6 text-3xl font-semibold tracking-tight">
            {connectingCount}
          </p>
          <p className="mt-1 text-sm text-white/50">Waiting to connect</p>
        </div>
      </section>

      <section className="pt-10" aria-labelledby="recent-tunnels-heading">
        <div className="flex items-end justify-between gap-4">
          <div>
            <h2 id="recent-tunnels-heading" className="text-lg font-semibold">
              Recent tunnels
            </h2>
            <p className="mt-1 text-sm text-white/45">
              The latest endpoints configured in this workspace.
            </p>
          </div>
          <Link
            to="/$orgSlug/tunnels"
            params={{ orgSlug }}
            className="text-sm font-medium text-indigo-200 hover:text-indigo-100"
          >
            View all
          </Link>
        </div>
        {tunnels.length ? (
          <div className="mt-5 grid gap-3">
            {tunnels.slice(0, 4).map((tunnel) => (
              <Link
                key={tunnel.id}
                to="/$orgSlug/tunnels/$tunnelId"
                params={{ orgSlug, tunnelId: tunnel.id }}
                className="grid gap-3 rounded-2xl border border-white/10 bg-white/[0.025] p-4 transition-colors hover:border-indigo-300/40 hover:bg-indigo-300/[0.05] sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center"
              >
                <div className="min-w-0">
                  <p className="truncate font-medium">{tunnel.name}</p>
                  <p className="mt-1 truncate font-mono text-sm text-indigo-200">
                    {tunnel.publicHostname}
                  </p>
                </div>
                <TunnelStatusBadge status={tunnel.status} />
              </Link>
            ))}
          </div>
        ) : (
          <div className="mt-5 rounded-2xl border border-dashed border-white/15 px-5 py-8 text-sm text-white/50">
            No tunnel has been configured yet. Create one to expose a local
            service.
          </div>
        )}
      </section>
    </main>
  );
}
