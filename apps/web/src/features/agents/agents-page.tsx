import { useQuery } from '@tanstack/react-query';
import { Cable, Radio } from 'lucide-react';
import { Badge } from '#/components/ui/badge';
import { getOrganizations } from '#/features/organizations/services/organization-service';
import type { Agent, AgentStatus } from '#/interfaces/agent';
import { apiClient } from '#/lib/api-client';

const statusStyles: Record<AgentStatus, string> = {
  online: 'bg-emerald-400/10 text-emerald-200',
  pending: 'bg-amber-400/10 text-amber-100',
  offline: 'bg-white/5 text-white/55',
  revoked: 'bg-rose-400/10 text-rose-200',
};

export function AgentsPage({ orgSlug }: { orgSlug: string }) {
  const organizations = useQuery({
    queryKey: ['organizations'],
    queryFn: getOrganizations,
  });

  const organization = organizations.data?.find(
    (item) => item.slug === orgSlug,
  );

  const agents = useQuery({
    queryKey: ['agents', organization?.id],
    queryFn: () =>
      apiClient.get<Agent[]>(
        `/api/v1/organizations/${organization?.id}/agents`,
      ),
    enabled: Boolean(organization?.id),
  });

  if (organizations.isLoading || agents.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading agents…</p>;
  }

  if (organizations.isError || agents.isError || !organization) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load agents for this workspace.
      </p>
    );
  }

  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="flex flex-col gap-5 border-b border-white/10 pb-8 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="mb-3 text-sm font-medium text-indigo-200">
            {organization.name}
          </p>
          <h1 className="text-3xl font-semibold tracking-[-0.04em] sm:text-4xl">
            Agents
          </h1>
          <p className="mt-3 max-w-xl text-sm leading-6 text-white/55">
            Clients that keep your tunnels connected and report their health.
          </p>
        </div>
        <p className="text-sm text-white/45">
          Agents are registered from the Outpipe CLI.
        </p>
      </header>
      {!agents.data?.length ? (
        <div className="mt-8 rounded-2xl border border-dashed border-white/15 px-5 py-12 text-center">
          <Cable className="mx-auto size-6 text-indigo-200" />
          <h2 className="mt-4 font-semibold">No agents connected</h2>
          <p className="mt-2 text-sm text-white/50">
            Register an agent to manage tunnel connections from your
            infrastructure.
          </p>
        </div>
      ) : (
        <div className="mt-8 grid gap-3">
          {agents.data.map((agent) => (
            <article
              key={agent.id}
              className="grid gap-4 rounded-2xl border border-white/10 bg-white/[0.025] p-5 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center"
            >
              <div className="flex min-w-0 items-start gap-3">
                <Radio className="mt-0.5 size-5 shrink-0 text-indigo-200" />
                <div className="min-w-0">
                  <div className="flex flex-wrap items-center gap-2">
                    <h2 className="font-medium">{agent.name}</h2>
                    <Badge className={`border-0 ${statusStyles[agent.status]}`}>
                      {agent.status}
                    </Badge>
                  </div>
                  <p className="mt-2 truncate text-sm text-white/50">
                    {agent.hostname || 'Hostname not reported'} ·{' '}
                    {agent.platform || 'Platform unknown'}
                  </p>
                </div>
              </div>
              <p className="text-sm text-white/45 sm:text-right">
                {agent.version ? `v${agent.version}` : 'Version pending'}
                <br />
                {formatLastSeen(agent.lastSeenAt)}
              </p>
            </article>
          ))}
        </div>
      )}
    </main>
  );
}

function formatLastSeen(value?: string) {
  return value
    ? `Last seen ${new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value))}`
    : 'Not seen yet';
}
