import { createFileRoute } from '@tanstack/react-router';
import { TunnelDetailsPage } from '#/features/tunnels';

export const Route = createFileRoute('/$orgSlug/tunnels/$tunnelId')({
  component: TunnelDetailsRoute,
});

function TunnelDetailsRoute() {
  const { orgSlug, tunnelId } = Route.useParams();
  return <TunnelDetailsPage orgSlug={orgSlug} tunnelID={tunnelId} />;
}
