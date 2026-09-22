import { createFileRoute } from '@tanstack/react-router';
import { AdminOrganizationPage } from '#/features/admin/admin-organization-page';

export const Route = createFileRoute('/admin/organizations/$organizationID')({
  component: AdminOrganizationRoute,
});

function AdminOrganizationRoute() {
  const { organizationID } = Route.useParams();
  return <AdminOrganizationPage organizationId={organizationID} />;
}
