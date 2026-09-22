import { createFileRoute } from '@tanstack/react-router';
import { OrganizationOverviewPage } from '#/features/organizations';

export const Route = createFileRoute('/$orgSlug/')({
  component: OrganizationOverviewRoute,
});

function OrganizationOverviewRoute() {
  const { orgSlug } = Route.useParams();
  return <OrganizationOverviewPage orgSlug={orgSlug} />;
}
