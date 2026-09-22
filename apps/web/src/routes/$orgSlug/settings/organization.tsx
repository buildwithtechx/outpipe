import { createFileRoute } from '@tanstack/react-router';
import { OrganizationSettingsPage } from '#/features/organizations';

export const Route = createFileRoute('/$orgSlug/settings/organization')({
  component: OrganizationSettingsRoute,
});

function OrganizationSettingsRoute() {
  const { orgSlug } = Route.useParams();
  return <OrganizationSettingsPage orgSlug={orgSlug} />;
}
