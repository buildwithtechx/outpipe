import type { UseQueryResult } from '@tanstack/react-query';
import type { Tunnel } from '#/interfaces/tunnel';
import { TunnelEmptyState } from './tunnel-empty-state';
import { TunnelListItem } from './tunnel-list-item';
import { TunnelPageState } from './tunnel-page-state';

type TunnelListProps = {
  query: UseQueryResult<Tunnel[], Error>;
  orgSlug: string;
};

export function TunnelList({ query, orgSlug }: TunnelListProps) {
  if (query.isLoading) {
    return <TunnelPageState label="Loading tunnels…" compact />;
  }

  if (query.isError) {
    return (
      <TunnelPageState
        error="We could not retrieve tunnels for this workspace."
        compact
      />
    );
  }

  if (!query.data?.length) {
    return <TunnelEmptyState />;
  }

  return (
    <div className="grid gap-3">
      {query.data.map((tunnel) => (
        <TunnelListItem key={tunnel.id} tunnel={tunnel} orgSlug={orgSlug} />
      ))}
    </div>
  );
}
