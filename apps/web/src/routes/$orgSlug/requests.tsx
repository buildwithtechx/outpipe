import { createFileRoute } from '@tanstack/react-router';
import { RequestsPage } from '#/features/requests';

export const Route = createFileRoute('/$orgSlug/requests')({
  component: RequestsRoute,
});

function RequestsRoute() {
  const { orgSlug } = Route.useParams();
  return <RequestsPage orgSlug={orgSlug} />;
}
