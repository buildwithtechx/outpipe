import { createFileRoute } from '@tanstack/react-router';
import { ObservabilityPage } from '#/features/observability';

export const Route = createFileRoute('/$orgSlug/observability')({
  component: ObservabilityRoute,
});

function ObservabilityRoute() {
  const { orgSlug } = Route.useParams();
  return <ObservabilityPage orgSlug={orgSlug} />;
}
