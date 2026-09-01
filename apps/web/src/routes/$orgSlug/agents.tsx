import { createFileRoute } from '@tanstack/react-router';
import { AgentsPage } from '#/features/agents';

export const Route = createFileRoute('/$orgSlug/agents')({
  component: AgentsRoute,
});

function AgentsRoute() {
  const { orgSlug } = Route.useParams();
  return <AgentsPage orgSlug={orgSlug} />;
}
