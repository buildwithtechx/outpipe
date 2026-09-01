import { createFileRoute } from '@tanstack/react-router';
import { TunnelsPage } from '#/features/tunnels';

export const Route = createFileRoute('/$orgSlug/tunnels/')({
  component: TunnelsRoute,
});

function TunnelsRoute() {
  const { orgSlug } = Route.useParams();
  return <TunnelsPage orgSlug={orgSlug} />;
}
