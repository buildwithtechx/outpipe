import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CirclePlus, RefreshCw } from 'lucide-react';
import { useState } from 'react';
import { Button } from '#/components/ui/button';
import type { CreateTunnelRequest } from '#/interfaces/tunnel';
import { TunnelCreateForm } from './components/tunnel-create-form';
import { TunnelList } from './components/tunnel-list';
import { TunnelPageState } from './components/tunnel-page-state';
import { useOrganizationTunnels } from './hooks/use-organization-tunnels';
import { createTunnel } from './services/tunnel-service';

const initialRequest: CreateTunnelRequest = {
  name: '',
  protocol: 'http',
  targetHost: 'localhost',
  targetPort: 3000,
};

export function TunnelsPage({ orgSlug }: { orgSlug: string }) {
  const queryClient = useQueryClient();
  const [isCreating, setIsCreating] = useState(false);
  const [request, setRequest] = useState<CreateTunnelRequest>(initialRequest);
  const { organization, organizationsQuery, tunnelsQuery } =
    useOrganizationTunnels(orgSlug);
  const createMutation = useMutation({
    mutationFn: () => createTunnel(organization?.id ?? '', request),
    onSuccess: async () => {
      setRequest(initialRequest);
      setIsCreating(false);
      await queryClient.invalidateQueries({
        queryKey: ['tunnels', organization?.id],
      });
    },
  });

  if (organizationsQuery.isLoading) {
    return <TunnelPageState label="Loading workspace…" />;
  }

  if (organizationsQuery.isError) {
    return (
      <TunnelPageState error="Your session has expired or we could not load this workspace." />
    );
  }

  if (!organization) {
    return (
      <TunnelPageState error="This workspace is not available to your account." />
    );
  }

  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="flex flex-col gap-6 border-b border-white/10 pb-8 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="mb-3 text-sm font-medium text-indigo-200">
            {organization.name}
          </p>
          <h1 className="text-3xl font-semibold tracking-[-0.04em] sm:text-4xl">
            Tunnels
          </h1>
          <p className="mt-3 max-w-xl text-sm leading-6 text-white/55 sm:text-base">
            Create public endpoints and keep an eye on every local service your
            workspace shares.
          </p>
        </div>
        <Button
          type="button"
          size="lg"
          onClick={() => setIsCreating((value) => !value)}
          className="rounded-full bg-white px-5 text-black hover:bg-indigo-100"
        >
          <CirclePlus />
          {isCreating ? 'Close form' : 'New tunnel'}
        </Button>
      </header>

      {isCreating && (
        <TunnelCreateForm
          request={request}
          isPending={createMutation.isPending}
          error={
            createMutation.isError
              ? 'We could not create this tunnel. Check the target and try again.'
              : null
          }
          onChange={setRequest}
          onSubmit={() => createMutation.mutate()}
        />
      )}

      <section className="pt-8" aria-labelledby="tunnel-list-heading">
        <div className="mb-5 flex items-center justify-between gap-4">
          <div>
            <h2 id="tunnel-list-heading" className="text-base font-semibold">
              Workspace tunnels
            </h2>
            <p className="mt-1 text-sm text-white/45">
              {tunnelsQuery.data?.length ?? 0} configured endpoint
              {(tunnelsQuery.data?.length ?? 0) === 1 ? '' : 's'}
            </p>
          </div>
          <Button
            type="button"
            variant="outline"
            size="icon"
            aria-label="Refresh tunnels"
            onClick={() => void tunnelsQuery.refetch()}
            disabled={tunnelsQuery.isFetching}
            className="rounded-full border-white/10 bg-transparent text-white/60 hover:border-indigo-300/50 hover:bg-indigo-300/10 hover:text-white"
          >
            <RefreshCw
              className={tunnelsQuery.isFetching ? 'animate-spin' : undefined}
            />
          </Button>
        </div>
        <TunnelList query={tunnelsQuery} orgSlug={orgSlug} />
      </section>
    </main>
  );
}
