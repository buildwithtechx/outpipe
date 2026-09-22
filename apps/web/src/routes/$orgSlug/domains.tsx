import { createFileRoute } from '@tanstack/react-router';
import { DomainsPage } from '#/features/domains';

export const Route = createFileRoute('/$orgSlug/domains')({
  component: DomainsRoute,
});

function DomainsRoute() {
  const { orgSlug } = Route.useParams();
  return <DomainsPage orgSlug={orgSlug} />;
}
